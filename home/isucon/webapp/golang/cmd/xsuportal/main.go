package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"google.golang.org/protobuf/types/known/timestamppb"

	xsuportal "github.com/isucon/isucon10-final/webapp/golang"
	xsuportalpb "github.com/isucon/isucon10-final/webapp/golang/proto/xsuportal"
	resourcespb "github.com/isucon/isucon10-final/webapp/golang/proto/xsuportal/resources"
	adminpb "github.com/isucon/isucon10-final/webapp/golang/proto/xsuportal/services/admin"
	audiencepb "github.com/isucon/isucon10-final/webapp/golang/proto/xsuportal/services/audience"
	commonpb "github.com/isucon/isucon10-final/webapp/golang/proto/xsuportal/services/common"
	contestantpb "github.com/isucon/isucon10-final/webapp/golang/proto/xsuportal/services/contestant"
	registrationpb "github.com/isucon/isucon10-final/webapp/golang/proto/xsuportal/services/registration"
	"github.com/isucon/isucon10-final/webapp/golang/util"
)

const (
	TeamCapacity               = 90
	AdminID                    = "admin"
	AdminPassword              = "admin"
	DebugContestStatusFilePath = "/tmp/XSUPORTAL_CONTEST_STATUS"
	MYSQL_ER_DUP_ENTRY         = 1062
	SessionName                = "xsucon_session"
)

var db *sqlx.DB
var rds *redis.Pool
var notifier xsuportal.Notifier
var defaultContestStatus xsuportal.ContestStatus
var teamPBMap sync.Map
var teamMap sync.Map
var contestantMap sync.Map

func main() {
	srv := echo.New()
	srv.Debug = util.GetEnv("DEBUG", "") != ""
	srv.Server.Addr = fmt.Sprintf(":%v", util.GetEnv("PORT", "9292"))
	srv.HideBanner = true

	srv.Binder = ProtoBinder{}
	srv.HTTPErrorHandler = func(err error, c echo.Context) {
		if !c.Response().Committed {
			c.Logger().Error(c.Request().Method, " ", c.Request().URL.Path, " ", err)
			_ = halt(c, http.StatusInternalServerError, "", err)
		}
	}

	var err error
	rds = xsuportal.GetRedis()
	if err != nil {
		panic(err)
	}

	db, _ = xsuportal.GetDB()
	db.SetMaxOpenConns(800)
	db.SetMaxIdleConns(800)

	err = sqlx.Get(db, &defaultContestStatus, "SELECT * FROM `contest_config`")
	if err != sql.ErrNoRows && err != nil {
		panic(err)
	}

	rows, _ := db.Queryx("SELECT * FROM teams")
	if err != sql.ErrNoRows && err != nil {
		panic(err)
	}
	for rows.Next() {
		var team xsuportal.Team
		rows.StructScan(&team)
		cacheTeam(&team)
	}
	rows.Close()

	rows, _ = db.Queryx("SELECT * FROM contestants")
	if err != sql.ErrNoRows && err != nil {
		panic(err)
	}
	for rows.Next() {
		var contestant xsuportal.Contestant
		rows.StructScan(&contestant)
		cacheContestant(&contestant)
	}
	rows.Close()

	ns, _ := makeNotificationsPB([]*xsuportal.Notification{})
	nsResponse = contestantpb.ListNotificationsResponse{
		Notifications:               ns,
		LastAnsweredClarificationId: 0,
	}

	// srv.Use(middleware.Logger())
	srv.Use(middleware.Recover())
	srv.Use(session.Middleware(sessions.NewCookieStore([]byte("tagomoris"))))

	srv.File("/", "public/audience.html")
	srv.File("/registration", "public/audience.html")
	srv.File("/signup", "public/audience.html")
	srv.File("/login", "public/audience.html")
	srv.File("/logout", "public/audience.html")
	srv.File("/teams", "public/audience.html")

	srv.File("/contestant", "public/contestant.html")
	srv.File("/contestant/benchmark_jobs", "public/contestant.html")
	srv.File("/contestant/benchmark_jobs/:id", "public/contestant.html")
	srv.File("/contestant/clarifications", "public/contestant.html")

	srv.File("/admin", "public/admin.html")
	srv.File("/admin/", "public/admin.html")
	srv.File("/admin/clarifications", "public/admin.html")
	srv.File("/admin/clarifications/:id", "public/admin.html")

	srv.Static("/", "public")

	admin := &AdminService{}
	audience := &AudienceService{}
	registration := &RegistrationService{}
	contestant := &ContestantService{}
	common := &CommonService{}

	srv.POST("/initialize", admin.Initialize)
	srv.GET("/api/admin/clarifications", admin.ListClarifications)
	srv.GET("/api/admin/clarifications/:id", admin.GetClarification)
	srv.PUT("/api/admin/clarifications/:id", admin.RespondClarification)
	srv.GET("/api/session", common.GetCurrentSession)
	srv.GET("/api/audience/teams", audience.ListTeams)
	srv.GET("/api/audience/dashboard", audience.Dashboard)
	srv.GET("/api/registration/session", registration.GetRegistrationSession)
	srv.POST("/api/registration/team", registration.CreateTeam)
	srv.POST("/api/registration/contestant", registration.JoinTeam)
	srv.PUT("/api/registration", registration.UpdateRegistration)
	srv.DELETE("/api/registration", registration.DeleteRegistration)
	srv.POST("/api/contestant/benchmark_jobs", contestant.EnqueueBenchmarkJob)
	srv.GET("/api/contestant/benchmark_jobs", contestant.ListBenchmarkJobs)
	srv.GET("/api/contestant/benchmark_jobs/:id", contestant.GetBenchmarkJob)
	srv.GET("/api/contestant/clarifications", contestant.ListClarifications)
	srv.POST("/api/contestant/clarifications", contestant.RequestClarification)
	srv.GET("/api/contestant/dashboard", contestant.Dashboard)
	srv.GET("/api/contestant/notifications", contestant.ListNotifications)
	srv.POST("/api/contestant/push_subscriptions", contestant.SubscribeNotification)
	srv.DELETE("/api/contestant/push_subscriptions", contestant.UnsubscribeNotification)
	srv.POST("/api/signup", contestant.Signup)
	srv.POST("/api/login", contestant.Login)
	srv.POST("/api/logout", contestant.Logout)

	srv.GET("/debug/pprof/*", echo.WrapHandler(http.DefaultServeMux))

	srv.Logger.Error(srv.StartServer(srv.Server))
}

type ProtoBinder struct{}

func (p ProtoBinder) Bind(i interface{}, e echo.Context) error {
	rc := e.Request().Body
	defer rc.Close()
	b, err := ioutil.ReadAll(rc)
	if err != nil {
		return halt(e, http.StatusBadRequest, "", fmt.Errorf("read request body: %w", err))
	}
	if err := proto.Unmarshal(b, i.(proto.Message)); err != nil {
		return halt(e, http.StatusBadRequest, "", fmt.Errorf("unmarshal request body: %w", err))
	}
	return nil
}

type AdminService struct{}

func (*AdminService) Initialize(e echo.Context) error {
	var req adminpb.InitializeRequest
	if err := e.Bind(&req); err != nil {
		return err
	}

	conn := rds.Get()
	defer conn.Close()
	if _, err := conn.Do("FLUSHALL"); err != nil {
		return err
	}

	queries := []string{
		"TRUNCATE `teams`",
		"TRUNCATE `contestants`",
		"TRUNCATE `benchmark_jobs`",
		"TRUNCATE `clarifications`",
		"TRUNCATE `notifications`",
		"TRUNCATE `push_subscriptions`",
		"TRUNCATE `contest_config`",
		"TRUNCATE `scores`",
	}
	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return fmt.Errorf("truncate table: %w", err)
		}
	}

	passwordHash := sha256.Sum256([]byte(AdminPassword))
	digest := hex.EncodeToString(passwordHash[:])
	_, err := db.Exec("INSERT `contestants` (`id`, `password`, `staff`, `created_at`) VALUES (?, ?, TRUE, NOW(6))", AdminID, digest)
	if err != nil {
		return fmt.Errorf("insert initial contestant: %w", err)
	}

	if req.Contest != nil {
		_, err := db.Exec(
			"INSERT `contest_config` (`registration_open_at`, `contest_starts_at`, `contest_freezes_at`, `contest_ends_at`) VALUES (?, ?, ?, ?)",
			req.Contest.RegistrationOpenAt.AsTime().Round(time.Microsecond),
			req.Contest.ContestStartsAt.AsTime().Round(time.Microsecond),
			req.Contest.ContestFreezesAt.AsTime().Round(time.Microsecond),
			req.Contest.ContestEndsAt.AsTime().Round(time.Microsecond),
		)
		if err != nil {
			return fmt.Errorf("insert contest: %w", err)
		}
	} else {
		_, err := db.Exec("INSERT `contest_config` (`registration_open_at`, `contest_starts_at`, `contest_freezes_at`, `contest_ends_at`) VALUES (TIMESTAMPADD(SECOND, 0, NOW(6)), TIMESTAMPADD(SECOND, 5, NOW(6)), TIMESTAMPADD(SECOND, 40, NOW(6)), TIMESTAMPADD(SECOND, 50, NOW(6)))")
		if err != nil {
			return fmt.Errorf("insert contest: %w", err)
		}
	}

	err = sqlx.Get(db, &defaultContestStatus, "SELECT * FROM `contest_config`")
	if err != nil {
		return fmt.Errorf("failed to get contest_config: %w", err)
	}

	host := util.GetEnv("BENCHMARK_SERVER_HOST", "localhost")
	port, _ := strconv.Atoi(util.GetEnv("BENCHMARK_SERVER_PORT", "50051"))
	res := &adminpb.InitializeResponse{
		Language: "go",
		BenchmarkServer: &adminpb.InitializeResponse_BenchmarkServer{
			Host: host,
			Port: int64(port),
		},
	}
	teamPBMap = sync.Map{}
	teamMap = sync.Map{}
	contestantMap = sync.Map{}
	rows, _ := db.Queryx("SELECT * FROM contestants")
	if err != sql.ErrNoRows && err != nil {
		panic(err)
	}
	for rows.Next() {
		var contestant xsuportal.Contestant
		rows.StructScan(&contestant)
		cacheContestant(&contestant)
	}
	rows.Close()

	finishedLeaderboard = nil

	return writeProto(e, http.StatusOK, res)
}

func (*AdminService) ListClarifications(e echo.Context) error {
	contestant, _, ok, err := loginRequired(e, db, &loginRequiredOption{})
	if !ok {
		return wrapError("check session", err)
	}
	if !contestant.Staff {
		return halt(e, http.StatusForbidden, "管理者権限が必要です", nil)
	}
	rows, err := db.Queryx("SELECT * FROM `clarifications` JOIN teams ON clarifications.team_id = teams.id ORDER BY clarifications.updated_at DESC")
	if err != sql.ErrNoRows && err != nil {
		return fmt.Errorf("query clarifications: %w", err)
	}
	res := &adminpb.ListClarificationsResponse{}
	defer rows.Close()
	for rows.Next() {
		var clarification xsuportal.Clarification
		var team xsuportal.Team
		err := rows.Scan(
			&clarification.ID,
			&clarification.TeamID,
			&clarification.Disclosed,
			&clarification.Question,
			&clarification.Answer,
			&clarification.AnsweredAt,
			&clarification.CreatedAt,
			&clarification.UpdatedAt,
			&team.ID,
			&team.Name,
			&team.LeaderID,
			&team.EmailAddress,
			&team.InviteToken,
			&team.Withdrawn,
			&team.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("query team(id=%v, clarification=%v): %w", clarification.TeamID, clarification.ID, err)
		}
		c, err := makeClarificationPB(db, &clarification, &team)
		if err != nil {
			return fmt.Errorf("make clarification: %w", err)
		}
		res.Clarifications = append(res.Clarifications, c)
	}
	return writeProto(e, http.StatusOK, res)
}

func (*AdminService) GetClarification(e echo.Context) error {
	contestant, _, ok, err := loginRequired(e, db, &loginRequiredOption{})
	if !ok {
		return wrapError("check session", err)
	}
	id, err := strconv.Atoi(e.Param("id"))
	if err != nil {
		return fmt.Errorf("parse id: %w", err)
	}
	// contestant, _ := getCurrentContestant(e, db, false)
	if !contestant.Staff {
		return halt(e, http.StatusForbidden, "管理者権限が必要です", nil)
	}
	var clarification xsuportal.Clarification
	var team xsuportal.Team
	err = db.QueryRowx(
		"SELECT * FROM `clarifications` JOIN teams ON clarifications.team_id = teams.id WHERE clarifications.id = ? LIMIT 1",
		id,
	).Scan(
		&clarification.ID,
		&clarification.TeamID,
		&clarification.Disclosed,
		&clarification.Question,
		&clarification.Answer,
		&clarification.AnsweredAt,
		&clarification.CreatedAt,
		&clarification.UpdatedAt,
		&team.ID,
		&team.Name,
		&team.LeaderID,
		&team.EmailAddress,
		&team.InviteToken,
		&team.Withdrawn,
		&team.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("get clarification: %w", err)
	}
	c, err := makeClarificationPB(db, &clarification, &team)
	if err != nil {
		return fmt.Errorf("make clarification: %w", err)
	}
	return writeProto(e, http.StatusOK, &adminpb.GetClarificationResponse{
		Clarification: c,
	})
}

func (*AdminService) RespondClarification(e echo.Context) error {
	contestant, _, ok, err := loginRequired(e, db, &loginRequiredOption{})
	if !ok {
		return wrapError("check session", err)
	}
	id, err := strconv.Atoi(e.Param("id"))
	if err != nil {
		return fmt.Errorf("parse id: %w", err)
	}
	// contestant, _ := getCurrentContestant(e, db, false)
	if !contestant.Staff {
		return halt(e, http.StatusForbidden, "管理者権限が必要です", nil)
	}
	var req adminpb.RespondClarificationRequest
	if err := e.Bind(&req); err != nil {
		return err
	}

	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var clarificationBefore xsuportal.Clarification
	err = tx.Get(
		&clarificationBefore,
		"SELECT * FROM `clarifications` WHERE `id` = ? LIMIT 1 FOR UPDATE",
		id,
	)
	if err == sql.ErrNoRows {
		return halt(e, http.StatusNotFound, "質問が見つかりません", nil)
	}
	if err != nil {
		return fmt.Errorf("get clarification with lock: %w", err)
	}
	wasAnswered := clarificationBefore.AnsweredAt.Valid
	wasDisclosed := clarificationBefore.Disclosed

	_, err = tx.Exec(
		"UPDATE `clarifications` SET `disclosed` = ?, `answer` = ?, `updated_at` = NOW(6), `answered_at` = NOW(6) WHERE `id` = ? LIMIT 1",
		req.Disclose,
		req.Answer,
		id,
	)
	if err != nil {
		return fmt.Errorf("update clarification: %w", err)
	}
	var clarification xsuportal.Clarification
	var team xsuportal.Team
	err = tx.QueryRowx(
		"SELECT * FROM `clarifications` JOIN teams ON clarifications.team_id = teams.id WHERE clarifications.id = ? LIMIT 1",
		id,
	).Scan(
		&clarification.ID,
		&clarification.TeamID,
		&clarification.Disclosed,
		&clarification.Question,
		&clarification.Answer,
		&clarification.AnsweredAt,
		&clarification.CreatedAt,
		&clarification.UpdatedAt,
		&team.ID,
		&team.Name,
		&team.LeaderID,
		&team.EmailAddress,
		&team.InviteToken,
		&team.Withdrawn,
		&team.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("get clarification: %w", err)
	}
	c, err := makeClarificationPB(tx, &clarification, &team)
	if err != nil {
		return fmt.Errorf("make clarification: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	updated := wasAnswered && wasDisclosed == clarification.Disclosed
	if err := notifier.NotifyClarificationAnswered(db, &clarification, updated); err != nil {
		return fmt.Errorf("notify clarification answered: %w", err)
	}
	return writeProto(e, http.StatusOK, &adminpb.RespondClarificationResponse{
		Clarification: c,
	})
}

type CommonService struct{}

func (*CommonService) GetCurrentSession(e echo.Context) error {
	res := &commonpb.GetCurrentSessionResponse{}
	currentContestant, err := getCurrentContestant(e, db, false)
	if err != nil {
		return fmt.Errorf("get current contestant: %w", err)
	}
	if currentContestant != nil {
		res.Contestant = makeContestantPB(currentContestant)
	}
	currentTeam, err := getCurrentTeam(e, db, false, currentContestant)
	if err != nil {
		return fmt.Errorf("get current team: %w", err)
	}
	if currentTeam != nil {
		res.Team, err = makeTeamPB(db, currentTeam, true, true)
		if err != nil {
			return fmt.Errorf("make team: %w", err)
		}
	}
	res.Contest, err = makeContestPB(e)
	if err != nil {
		return fmt.Errorf("make contest: %w", err)
	}
	vapidKey := notifier.VAPIDKey()
	if vapidKey != nil {
		res.PushVapidKey = vapidKey.VAPIDPublicKey
	}
	return writeProto(e, http.StatusOK, res)
}

type ContestantService struct{}

func (*ContestantService) EnqueueBenchmarkJob(e echo.Context) error {
	var req contestantpb.EnqueueBenchmarkJobRequest
	if err := e.Bind(&req); err != nil {
		return err
	}
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()
	_, team, ok, err := loginRequired(e, tx, &loginRequiredOption{Team: true});
	if !ok {
		return wrapError("check session", err)
	}
	if ok, err := contestStatusRestricted(e, tx, resourcespb.Contest_STARTED, "競技時間外はベンチマークを実行できません"); !ok {
		return wrapError("check contest status", err)
	}
	// team, _ := getCurrentTeam(e, tx, false, nil)

	var hasJob bool
	err = db.Get(
		&hasJob,
		"SELECT TRUE AS `cnt` FROM `benchmark_jobs` WHERE `team_id` = ? AND `finished_at` IS NULL LIMIT 1 LOCK IN SHARE MODE",
		team.ID,
	)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("count benchmark job: %w", err)
	}
	if hasJob {
		return halt(e, http.StatusForbidden, "既にベンチマークを実行中です", nil)
	}

	_, err = tx.Exec(
		"INSERT INTO `benchmark_jobs` (`team_id`, `target_hostname`, `status`, `updated_at`, `created_at`) VALUES (?, ?, ?, NOW(6), NOW(6))",
		team.ID,
		req.TargetHostname,
		int(resourcespb.BenchmarkJob_PENDING),
	)
	if err != nil {
		return fmt.Errorf("enqueue benchmark job: %w", err)
	}
	var job xsuportal.BenchmarkJob
	err = tx.Get(
		&job,
		"SELECT * FROM `benchmark_jobs` WHERE `id` = (SELECT LAST_INSERT_ID()) LIMIT 1",
	)
	if err != nil {
		return fmt.Errorf("get benchmark job: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	j := makeBenchmarkJobPB(&job)

	conn := rds.Get()
	defer conn.Close()
	jobStr := fmt.Sprintf("%d@%s@%d", job.ID, job.TargetHostName, job.CreatedAt.UnixNano())
	_, err = conn.Do("LPUSH", "jobs", jobStr)
	if err != nil {
		return fmt.Errorf("redis: %w", err)
	}

	return writeProto(e, http.StatusOK, &contestantpb.EnqueueBenchmarkJobResponse{
		Job: j,
	})
}

func (*ContestantService) ListBenchmarkJobs(e echo.Context) error {
	_, team, ok, err := loginRequired(e, db, &loginRequiredOption{Team: true})
	if !ok {
		return wrapError("check session", err)
	}
	jobs, err := makeBenchmarkJobsPB(e, db, team, 0)
	if err != nil {
		return fmt.Errorf("make benchmark jobs: %w", err)
	}
	return writeProto(e, http.StatusOK, &contestantpb.ListBenchmarkJobsResponse{
		Jobs: jobs,
	})
}

func (*ContestantService) GetBenchmarkJob(e echo.Context) error {
	_, team, ok, err := loginRequired(e, db, &loginRequiredOption{Team: true})
	if !ok {
		return wrapError("check session", err)
	}
	id, err := strconv.Atoi(e.Param("id"))
	if err != nil {
		return fmt.Errorf("parse id: %w", err)
	}
	// team, _ := getCurrentTeam(e, db, false, nil)
	var job xsuportal.BenchmarkJob
	err = db.Get(
		&job,
		"SELECT * FROM `benchmark_jobs` WHERE `team_id` = ? AND `id` = ? LIMIT 1",
		team.ID,
		id,
	)
	if err == sql.ErrNoRows {
		return halt(e, http.StatusNotFound, "ベンチマークジョブが見つかりません", nil)
	}
	if err != nil {
		return fmt.Errorf("get benchmark job: %w", err)
	}
	return writeProto(e, http.StatusOK, &contestantpb.GetBenchmarkJobResponse{
		Job: makeBenchmarkJobPB(&job),
	})
}

func (*ContestantService) ListClarifications(e echo.Context) error {
	_, team, ok, err := loginRequired(e, db, &loginRequiredOption{Team: true})
	if !ok {
		return wrapError("check session", err)
	}
	// team, _ := getCurrentTeam(e, db, false, nil)
	rows, err := db.Queryx(
		"SELECT * FROM `clarifications` JOIN teams ON clarifications.team_id = teams.id WHERE `team_id` = ? OR `disclosed` = TRUE ORDER BY clarifications.id DESC",
		team.ID,
	)
	if err != sql.ErrNoRows && err != nil {
		return fmt.Errorf("select clarifications: %w", err)
	}
	res := &contestantpb.ListClarificationsResponse{}
	defer rows.Close()
	for rows.Next() {
		var clarification xsuportal.Clarification
		var team xsuportal.Team
		err := rows.Scan(
			&clarification.ID,
			&clarification.TeamID,
			&clarification.Disclosed,
			&clarification.Question,
			&clarification.Answer,
			&clarification.AnsweredAt,
			&clarification.CreatedAt,
			&clarification.UpdatedAt,
			&team.ID,
			&team.Name,
			&team.LeaderID,
			&team.EmailAddress,
			&team.InviteToken,
			&team.Withdrawn,
			&team.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("get team(id=%v): %w", clarification.TeamID, err)
		}
		c, err := makeClarificationPB(db, &clarification, &team)
		if err != nil {
			return fmt.Errorf("make clarification: %w", err)
		}
		res.Clarifications = append(res.Clarifications, c)
	}
	return writeProto(e, http.StatusOK, res)
}

func (*ContestantService) RequestClarification(e echo.Context) error {
	_, team, ok, err := loginRequired(e, db, &loginRequiredOption{Team: true})
	if !ok {
		return wrapError("check session", err)
	}
	var req contestantpb.RequestClarificationRequest
	if err := e.Bind(&req); err != nil {
		return err
	}
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()
	// team, _ := getCurrentTeam(e, tx, false, nil)
	_, err = tx.Exec(
		"INSERT INTO `clarifications` (`team_id`, `question`, `created_at`, `updated_at`) VALUES (?, ?, NOW(6), NOW(6))",
		team.ID,
		req.Question,
	)
	if err != nil {
		return fmt.Errorf("insert clarification: %w", err)
	}
	var clarification xsuportal.Clarification
	err = tx.Get(&clarification, "SELECT * FROM `clarifications` WHERE `id` = LAST_INSERT_ID() LIMIT 1")
	if err != nil {
		return fmt.Errorf("get clarification: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	c, err := makeClarificationPB(db, &clarification, team)
	if err != nil {
		return fmt.Errorf("make clarification: %w", err)
	}
	return writeProto(e, http.StatusOK, &contestantpb.RequestClarificationResponse{
		Clarification: c,
	})
}

func (*ContestantService) Dashboard(e echo.Context) error {
	_, team, ok, err := loginRequired(e, db, &loginRequiredOption{Team: true})
	if !ok {
		return wrapError("check session", err)
	}
	// team, _ := getCurrentTeam(e, db, false, nil)
	leaderboard, err := makeLeaderboardPB(e, team.ID)
	if err != nil {
		return fmt.Errorf("make leaderboard: %w", err)
	}
	contestStatus, err := getCurrentContestStatus(e)
	if err != nil {
		return fmt.Errorf("get current contest status: %w", err)
	}
	if contestStatus.Status == resourcespb.Contest_FINISHED {
		e.Response().Header().Set("Cache-Control", "private, max-age=11")
	}
	return writeProto(e, http.StatusOK, &contestantpb.DashboardResponse{
		Leaderboard: leaderboard,
	})
}


var nsResponse contestantpb.ListNotificationsResponse
func (*ContestantService) ListNotifications(e echo.Context) error {
	_, _, ok, err := loginRequired(e, db, &loginRequiredOption{Team: true})
	if !ok {
		return wrapError("check session", err)
	}
	return writeProto(e, http.StatusOK, &nsResponse)
}

/*
func (*ContestantService) ListNotifications(e echo.Context) error {
	contestant, team, ok, err := loginRequired(e, db, &loginRequiredOption{Team: true})
	if !ok {
		return wrapError("check session", err)
	}

	afterStr := e.QueryParam("after")

	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()
	// contestant, _ := getCurrentContestant(e, tx, false)

	var notifications []*xsuportal.Notification
	if afterStr != "" {
		after, err := strconv.Atoi(afterStr)
		if err != nil {
			return fmt.Errorf("parse after: %w", err)
		}
		err = tx.Select(
			&notifications,
			"SELECT * FROM `notifications` WHERE `contestant_id` = ? AND `id` > ? ORDER BY `id`",
			contestant.ID,
			after,
		)
		if err != sql.ErrNoRows && err != nil {
			return fmt.Errorf("select notifications(after=%v): %w", after, err)
		}
	} else {
		err = tx.Select(
			&notifications,
			"SELECT * FROM `notifications` WHERE `contestant_id` = ? ORDER BY `id`",
			contestant.ID,
		)
		if err != sql.ErrNoRows && err != nil {
			return fmt.Errorf("select notifications: %w", err)
		}
	}
	_, err = tx.Exec(
		"UPDATE `notifications` SET `read` = TRUE WHERE `contestant_id` = ? AND `read` = FALSE",
		contestant.ID,
	)
	if err != nil {
		return fmt.Errorf("update notifications: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	// team, _ := getCurrentTeam(e, db, false, nil)

	var lastAnsweredClarificationID int64
	err = db.Get(
		&lastAnsweredClarificationID,
		"SELECT `id` FROM `clarifications` WHERE (`team_id` = ? OR `disclosed` = TRUE) AND `answered_at` IS NOT NULL ORDER BY `id` DESC LIMIT 1",
		team.ID,
	)
	if err != sql.ErrNoRows && err != nil {
		return fmt.Errorf("get last answered clarification: %w", err)
	}
	ns, err := makeNotificationsPB(notifications)
	if err != nil {
		return fmt.Errorf("make notifications: %w", err)
	}
	return writeProto(e, http.StatusOK, &contestantpb.ListNotificationsResponse{
		Notifications:               ns,
		LastAnsweredClarificationId: lastAnsweredClarificationID,
	})
}
*/

func (*ContestantService) SubscribeNotification(e echo.Context) error {
	contestant, _, ok, err := loginRequired(e, db, &loginRequiredOption{Team: true})
	if !ok {
		return wrapError("check session", err)
	}

	if notifier.VAPIDKey() == nil {
		return halt(e, http.StatusServiceUnavailable, "WebPush は未対応です", nil)
	}

	var req contestantpb.SubscribeNotificationRequest
	if err := e.Bind(&req); err != nil {
		return err
	}

	// contestant, _ := getCurrentContestant(e, db, false)
	_, err = db.Exec(
		"INSERT INTO `push_subscriptions` (`contestant_id`, `endpoint`, `p256dh`, `auth`, `created_at`, `updated_at`) VALUES (?, ?, ?, ?, NOW(6), NOW(6))",
		contestant.ID,
		req.Endpoint,
		req.P256Dh,
		req.Auth,
	)
	if err != nil {
		return fmt.Errorf("insert push_subscription: %w", err)
	}
	return writeProto(e, http.StatusOK, &contestantpb.SubscribeNotificationResponse{})
}

func (*ContestantService) UnsubscribeNotification(e echo.Context) error {
	contestant, _, ok, err := loginRequired(e, db, &loginRequiredOption{Team: true})
	if !ok {
		return wrapError("check session", err)
	}

	if notifier.VAPIDKey() == nil {
		return halt(e, http.StatusServiceUnavailable, "WebPush は未対応です", nil)
	}

	var req contestantpb.UnsubscribeNotificationRequest
	if err := e.Bind(&req); err != nil {
		return err
	}

	// contestant, _ := getCurrentContestant(e, db, false)
	_, err = db.Exec(
		"DELETE FROM `push_subscriptions` WHERE `contestant_id` = ? AND `endpoint` = ? LIMIT 1",
		contestant.ID,
		req.Endpoint,
	)
	if err != nil {
		return fmt.Errorf("delete push_subscription: %w", err)
	}
	return writeProto(e, http.StatusOK, &contestantpb.UnsubscribeNotificationResponse{})
}

func (*ContestantService) Signup(e echo.Context) error {
	var req contestantpb.SignupRequest
	if err := e.Bind(&req); err != nil {
		return err
	}

	hash := sha256.Sum256([]byte(req.Password))
	_, err := db.Exec(
		"INSERT INTO `contestants` (`id`, `password`, `staff`, `created_at`) VALUES (?, ?, FALSE, NOW(6))",
		req.ContestantId,
		hex.EncodeToString(hash[:]),
	)
	if mErr, ok := err.(*mysql.MySQLError); ok && mErr.Number == MYSQL_ER_DUP_ENTRY {
		return halt(e, http.StatusBadRequest, "IDが既に登録されています", nil)
	}
	if err != nil {
		return fmt.Errorf("insert contestant: %w", err)
	}

	var contestant xsuportal.Contestant
	db.Get(&contestant, "SELECT * FROM contestants WHERE id = ?", req.ContestantId)
	cacheContestant(&contestant)

	sess, err := session.Get(SessionName, e)
	if err != nil {
		return fmt.Errorf("get session: %w", err)
	}
	sess.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 3600,
	}
	sess.Values["contestant_id"] = req.ContestantId
	if err := sess.Save(e.Request(), e.Response()); err != nil {
		return fmt.Errorf("save session: %w", err)
	}
	return writeProto(e, http.StatusOK, &contestantpb.SignupResponse{})
}

func (*ContestantService) Login(e echo.Context) error {
	var req contestantpb.LoginRequest
	if err := e.Bind(&req); err != nil {
		return err
	}
	var password string
	err := db.Get(
		&password,
		"SELECT `password` FROM `contestants` WHERE `id` = ? LIMIT 1",
		req.ContestantId,
	)
	if err != sql.ErrNoRows && err != nil {
		return fmt.Errorf("get contestant: %w", err)
	}
	passwordHash := sha256.Sum256([]byte(req.Password))
	digest := hex.EncodeToString(passwordHash[:])
	if err != sql.ErrNoRows && subtle.ConstantTimeCompare([]byte(digest), []byte(password)) == 1 {
		sess, err := session.Get(SessionName, e)
		if err != nil {
			return fmt.Errorf("get session: %w", err)
		}
		sess.Options = &sessions.Options{
			Path:   "/",
			MaxAge: 3600,
		}
		sess.Values["contestant_id"] = req.ContestantId
		if err := sess.Save(e.Request(), e.Response()); err != nil {
			return fmt.Errorf("save session: %w", err)
		}
	} else {
		return halt(e, http.StatusBadRequest, "ログインIDまたはパスワードが正しくありません", nil)
	}
	return writeProto(e, http.StatusOK, &contestantpb.LoginResponse{})
}

func (*ContestantService) Logout(e echo.Context) error {
	sess, err := session.Get(SessionName, e)
	if err != nil {
		return fmt.Errorf("get session: %w", err)
	}
	if _, ok := sess.Values["contestant_id"]; ok {
		delete(sess.Values, "contestant_id")
		sess.Options = &sessions.Options{
			Path:   "/",
			MaxAge: -1,
		}
		if err := sess.Save(e.Request(), e.Response()); err != nil {
			return fmt.Errorf("delete session: %w", err)
		}
	} else {
		return halt(e, http.StatusUnauthorized, "ログインしていません", nil)
	}
	return writeProto(e, http.StatusOK, &contestantpb.LogoutResponse{})
}

type RegistrationService struct{}

func (*RegistrationService) GetRegistrationSession(e echo.Context) error {
	var team *xsuportal.Team
	currentTeam, err := getCurrentTeam(e, db, false, nil)
	if err != nil {
		return fmt.Errorf("get current team: %w", err)
	}
	team = currentTeam
	if team == nil {
		teamIDStr := e.QueryParam("team_id")
		inviteToken := e.QueryParam("invite_token")
		if teamIDStr != "" && inviteToken != "" {
			teamID, err := strconv.Atoi(teamIDStr)
			if err != nil {
				return fmt.Errorf("parse team id: %w", err)
			}
			var t xsuportal.Team
			err = db.Get(
				&t,
				"SELECT * FROM `teams` WHERE `id` = ? AND `invite_token` = ? AND `withdrawn` = FALSE LIMIT 1",
				teamID,
				inviteToken,
			)
			if err == sql.ErrNoRows {
				return halt(e, http.StatusNotFound, "招待URLが無効です", nil)
			}
			if err != nil {
				return fmt.Errorf("get team: %w", err)
			}
			team = &t
		}
	}

	var members []xsuportal.Contestant
	if team != nil {
		err := db.Select(
			&members,
			"SELECT * FROM `contestants` WHERE `team_id` = ?",
			team.ID,
		)
		if err != nil {
			return fmt.Errorf("select members: %w", err)
		}
	}

	res := &registrationpb.GetRegistrationSessionResponse{
		Status: 0,
	}
	contestant, err := getCurrentContestant(e, db, false)
	if err != nil {
		return fmt.Errorf("get current contestant: %w", err)
	}
	switch {
	case contestant != nil && contestant.TeamID.Valid:
		res.Status = registrationpb.GetRegistrationSessionResponse_JOINED
	case team != nil && len(members) >= 3:
		res.Status = registrationpb.GetRegistrationSessionResponse_NOT_JOINABLE
	case contestant == nil:
		res.Status = registrationpb.GetRegistrationSessionResponse_NOT_LOGGED_IN
	case team != nil:
		res.Status = registrationpb.GetRegistrationSessionResponse_JOINABLE
	case team == nil:
		res.Status = registrationpb.GetRegistrationSessionResponse_CREATABLE
	default:
		return fmt.Errorf("undeterminable status")
	}
	if team != nil {
		res.Team, err = makeTeamPB(db, team, contestant != nil && currentTeam != nil && contestant.ID == currentTeam.LeaderID.String, true)
		if err != nil {
			return fmt.Errorf("make team: %w", err)
		}
		res.MemberInviteUrl = fmt.Sprintf("/registration?team_id=%v&invite_token=%v", team.ID, team.InviteToken)
		res.InviteToken = team.InviteToken
	}
	return writeProto(e, http.StatusOK, res)
}

func (*RegistrationService) CreateTeam(e echo.Context) error {
	var req registrationpb.CreateTeamRequest
	if err := e.Bind(&req); err != nil {
		return err
	}
	contestant, _, ok, err := loginRequired(e, db, &loginRequiredOption{})
	if !ok {
		return wrapError("check session", err)
	}
	ok, err = contestStatusRestricted(e, db, resourcespb.Contest_REGISTRATION, "チーム登録期間ではありません")
	if !ok {
		return wrapError("check contest status", err)
	}

	ctx := context.Background()
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin: %w", err)
	}
	defer tx.Rollback()

	randomBytes := make([]byte, 64)
	_, err = rand.Read(randomBytes)
	if err != nil {
		return fmt.Errorf("read random: %w", err)
	}
	inviteToken := base64.URLEncoding.EncodeToString(randomBytes)

	res, err := tx.ExecContext(
		ctx,
		"INSERT INTO `teams` (`name`, `email_address`, `invite_token`, `created_at`) VALUES (?, ?, ?, NOW(6))",
		req.TeamName,
		req.EmailAddress,
		inviteToken,
	)
	if err != nil {
		return fmt.Errorf("insert team: %w", err)
	}
	teamID, err := res.LastInsertId()
	if err != nil || teamID == 0 {
		return halt(e, http.StatusInternalServerError, "チームを登録できませんでした", nil)
	}

	var withinCapacity bool
	err = tx.QueryRowContext(
		ctx,
		"SELECT COUNT(*) <= ? AS `within_capacity` FROM `teams`",
		TeamCapacity,
	).Scan(&withinCapacity)
	if err != nil {
		return fmt.Errorf("check capacity: %w", err)
	}
	if !withinCapacity {
		return halt(e, http.StatusForbidden, "チーム登録数上限です", nil)
	}
	// contestant, _ := getCurrentContestant(e, tx, false)

	_, err = tx.ExecContext(
		ctx,
		"UPDATE `contestants` SET `name` = ?, `student` = ?, `team_id` = ? WHERE id = ? LIMIT 1",
		req.Name,
		req.IsStudent,
		teamID,
		contestant.ID,
	)
	if err != nil {
		return fmt.Errorf("update contestant: %w", err)
	}

	_, err = tx.ExecContext(
		ctx,
		"UPDATE `teams` SET `leader_id` = ? WHERE `id` = ? LIMIT 1",
		contestant.ID,
		teamID,
	)
	if err != nil {
		return fmt.Errorf("update team: %w", err)
	}

	var team xsuportal.Team
	err = tx.QueryRowxContext(
		ctx,
		"SELECT * FROM teams WHERE id = ?",
		teamID,
	).StructScan(&team)
	if err != nil {
		return fmt.Errorf("select team: %w", err)
	}
	err = tx.Get(contestant, "SELECT * FROM contestants WHERE id = ?", contestant.ID)
	if err != nil {
		return fmt.Errorf("select contestant: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	cacheTeam(&team)
	cacheContestant(contestant)

	return writeProto(e, http.StatusOK, &registrationpb.CreateTeamResponse{
		TeamId: teamID,
	})
}

func (*RegistrationService) JoinTeam(e echo.Context) error {
	var req registrationpb.JoinTeamRequest
	if err := e.Bind(&req); err != nil {
		return err
	}
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	contestant, _, ok, err := loginRequired(e, tx, &loginRequiredOption{Lock: true})
	if !ok {
		return wrapError("check session", err)
	}
	if ok, err := contestStatusRestricted(e, tx, resourcespb.Contest_REGISTRATION, "チーム登録期間ではありません"); !ok {
		return wrapError("check contest status", err)
	}
	var team xsuportal.Team
	err = tx.Get(
		&team,
		"SELECT * FROM `teams` WHERE `id` = ? AND `invite_token` = ? AND `withdrawn` = FALSE LIMIT 1 FOR UPDATE",
		req.TeamId,
		req.InviteToken,
	)
	if err == sql.ErrNoRows {
		return halt(e, http.StatusBadRequest, "招待URLが不正です", nil)
	}
	if err != nil {
		return fmt.Errorf("get team with lock: %w", err)
	}
	var memberCount int
	err = tx.Get(
		&memberCount,
		"SELECT COUNT(*) AS `cnt` FROM `contestants` WHERE `team_id` = ?",
		req.TeamId,
	)
	if err != nil {
		return fmt.Errorf("count team member: %w", err)
	}
	if memberCount >= 3 {
		return halt(e, http.StatusBadRequest, "チーム人数の上限に達しています", nil)
	}

	// contestant, _ := getCurrentContestant(e, tx, false)
	_, err = tx.Exec(
		"UPDATE `contestants` SET `team_id` = ?, `name` = ?, `student` = ? WHERE `id` = ? LIMIT 1",
		req.TeamId,
		req.Name,
		req.IsStudent,
		contestant.ID,
	)
	if err != nil {
		return fmt.Errorf("update contestant: %w", err)
	}

	tx.Get(contestant, "SELECT * FROM contestants WHERE id = ?", contestant.ID)

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	cacheTeam(&team)
	cacheContestant(contestant)
	return writeProto(e, http.StatusOK, &registrationpb.JoinTeamResponse{})
}

func (*RegistrationService) UpdateRegistration(e echo.Context) error {
	var req registrationpb.UpdateRegistrationRequest
	if err := e.Bind(&req); err != nil {
		return err
	}
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()
	contestant, team, ok, err := loginRequired(e, tx, &loginRequiredOption{Team: true, Lock: true})
	if !ok {
		return wrapError("check session", err)
	}
	// team, _ := getCurrentTeam(e, tx, false, nil)
	// contestant, _ := getCurrentContestant(e, tx, false)
	if team.LeaderID.Valid && team.LeaderID.String == contestant.ID {
		_, err := tx.Exec(
			"UPDATE `teams` SET `name` = ?, `email_address` = ? WHERE `id` = ? LIMIT 1",
			req.TeamName,
			req.EmailAddress,
			team.ID,
		)
		if err != nil {
			return fmt.Errorf("update team: %w", err)
		}
	}
	_, err = tx.Exec(
		"UPDATE `contestants` SET `name` = ?, `student` = ? WHERE `id` = ? LIMIT 1",
		req.Name,
		req.IsStudent,
		contestant.ID,
	)
	if err != nil {
		return fmt.Errorf("update contestant: %w", err)
	}

	tx.Get(contestant, "SELECT * FROM contestants WHERE id = ?", contestant.ID)

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	cacheTeam(team)
	cacheContestant(contestant)
	return writeProto(e, http.StatusOK, &registrationpb.UpdateRegistrationResponse{})
}

func (*RegistrationService) DeleteRegistration(e echo.Context) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()
	contestant, team, ok, err := loginRequired(e, tx, &loginRequiredOption{Team: true, Lock: true})
	if !ok {
		return wrapError("check session", err)
	}
	if ok, err := contestStatusRestricted(e, tx, resourcespb.Contest_REGISTRATION, "チーム登録期間外は辞退できません"); !ok {
		return wrapError("check contest status", err)
	}
	// team, _ := getCurrentTeam(e, tx, false, nil)
	// contestant, _ := getCurrentContestant(e, tx, false)
	if team.LeaderID.Valid && team.LeaderID.String == contestant.ID {
		_, err := tx.Exec(
			"UPDATE `teams` SET `withdrawn` = TRUE, `leader_id` = NULL WHERE `id` = ? LIMIT 1",
			team.ID,
		)
		if err != nil {
			return fmt.Errorf("withdrawn team(id=%v): %w", team.ID, err)
		}
		_, err = tx.Exec(
			"UPDATE `contestants` SET `team_id` = NULL WHERE `team_id` = ?",
			team.ID,
		)
		if err != nil {
			return fmt.Errorf("withdrawn members(team_id=%v): %w", team.ID, err)
		}
	} else {
		_, err := tx.Exec(
			"UPDATE `contestants` SET `team_id` = NULL WHERE `id` = ? LIMIT 1",
			contestant.ID,
		)
		if err != nil {
			return fmt.Errorf("withdrawn contestant(id=%v): %w", contestant.ID, err)
		}
	}

	tx.Get(contestant, "SELECT * FROM contestants WHERE id = ?", contestant.ID)

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	cacheTeam(team)
	cacheContestant(contestant)
	return writeProto(e, http.StatusOK, &registrationpb.DeleteRegistrationResponse{})
}

type AudienceService struct{}

func (*AudienceService) ListTeams(e echo.Context) error {
	var teams []xsuportal.Team
	err := db.Select(&teams, "SELECT * FROM `teams` WHERE `withdrawn` = FALSE ORDER BY `created_at` DESC")
	if err != nil {
		return fmt.Errorf("select teams: %w", err)
	}
	res := &audiencepb.ListTeamsResponse{}
	for _, team := range teams {
		var members []xsuportal.Contestant
		err := db.Select(
			&members,
			"SELECT * FROM `contestants` WHERE `team_id` = ? ORDER BY `created_at`",
			team.ID,
		)
		if err != nil {
			return fmt.Errorf("select members(team_id=%v): %w", team.ID, err)
		}
		var memberNames []string
		isStudent := true
		for _, member := range members {
			memberNames = append(memberNames, member.Name.String)
			isStudent = isStudent && member.Student
		}
		res.Teams = append(res.Teams, &audiencepb.ListTeamsResponse_TeamListItem{
			TeamId:      team.ID,
			Name:        team.Name,
			MemberNames: memberNames,
			IsStudent:   isStudent,
		})
	}
	return writeProto(e, http.StatusOK, res)
}

func (*AudienceService) Dashboard(e echo.Context) error {
	leaderboard, err := makeLeaderboardPB(e, 0)
	if err != nil {
		return fmt.Errorf("make leaderboard: %w", err)
	}
	contestStatus, err := getCurrentContestStatus(e)
	if err != nil {
		return fmt.Errorf("get current contest status: %w", err)
	}
	maxAge := 1
	if contestStatus.Status == resourcespb.Contest_FINISHED {
		maxAge = 11
	}
	cacheControl := fmt.Sprintf("public, max-age=%d", maxAge)
	e.Response().Header().Set("Cache-Control", cacheControl)
	return writeProto(e, http.StatusOK, &audiencepb.DashboardResponse{
		Leaderboard: leaderboard,
	})
}

type XsuportalContext struct {
	Contestant *xsuportal.Contestant
	Team       *xsuportal.Team
}

func getXsuportalContext(e echo.Context) *XsuportalContext {
	xc := e.Get("xsucon_context")
	if xc == nil {
		xc = &XsuportalContext{}
		e.Set("xsucon_context", xc)
	}
	return xc.(*XsuportalContext)
}

func getCurrentContestant(e echo.Context, db sqlx.Queryer, lock bool) (*xsuportal.Contestant, error) {
	xc := getXsuportalContext(e)
	if xc.Contestant != nil {
		return xc.Contestant, nil
	}
	sess, err := session.Get(SessionName, e)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	contestantID, ok := sess.Values["contestant_id"]
	if !ok {
		return nil, nil
	}
	if lock {
		var contestant xsuportal.Contestant
		query := "SELECT * FROM `contestants` WHERE `id` = ? LIMIT 1 FOR UPDATE"
		err := sqlx.Get(db, &contestant, query, contestantID)
		if err == sql.ErrNoRows {
			return nil, nil
		}
		if err != nil {
			return nil, fmt.Errorf("query contestant: %w", err)
		}
		xc.Contestant = &contestant
	} else {
		if v, ok := contestantMap.Load(contestantID); ok {
			xc.Contestant = v.(*xsuportal.Contestant)
		} else {
			return nil, nil
		}
	}
	return xc.Contestant, nil
}

func getCurrentTeam(e echo.Context, db sqlx.Queryer, lock bool, contestantParam *xsuportal.Contestant) (*xsuportal.Team, error) {
	xc := getXsuportalContext(e)
	if xc.Team != nil {
		return xc.Team, nil
	}
	var contestant *xsuportal.Contestant
	var err error
	if contestantParam != nil {
		contestant = contestantParam
	} else {
		contestant, err = getCurrentContestant(e, db, false)
		if err != nil {
			return nil, fmt.Errorf("current contestant: %w", err)
		}
	}
	if contestant == nil {
		return nil, nil
	}
	if lock {
		var team xsuportal.Team
		query := "SELECT * FROM `teams` WHERE `id` = ? LIMIT 1 FOR UPDATE"
		err = sqlx.Get(db, &team, query, contestant.TeamID.Int64)
		if err == sql.ErrNoRows {
			return nil, nil
		}
		if err != nil {
			return nil, fmt.Errorf("query team: %w", err)
		}
		xc.Team = &team
	} else {
		if v, ok := teamMap.Load(contestant.TeamID.Int64); ok {
			xc.Team = v.(*xsuportal.Team)
		} else {
			return nil, nil
		}
	}
	return xc.Team, nil
}

func getCurrentContestStatus(e echo.Context) (*xsuportal.ContestStatus, error) {
	contestStatus := defaultContestStatus
	now := time.Now().Round(time.Microsecond)
	contestStatus.CurrentTime = now
	switch {
	case now.Before(contestStatus.RegistrationOpenAt):
		contestStatus.StatusStr = "standby"
	case now.Before(contestStatus.ContestStartsAt):
		contestStatus.StatusStr = "registration"
	case now.Before(contestStatus.ContestEndsAt):
		contestStatus.StatusStr = "started"
	default:
		contestStatus.StatusStr = "finished"
	}
	contestStatus.Frozen = !now.Before(contestStatus.ContestStartsAt) && now.Before(contestStatus.ContestFreezesAt)

	statusStr := contestStatus.StatusStr
	if e.Echo().Debug {
		b, err := ioutil.ReadFile(DebugContestStatusFilePath)
		if err == nil {
			statusStr = string(b)
		}
	}
	switch statusStr {
	case "standby":
		contestStatus.Status = resourcespb.Contest_STANDBY
	case "registration":
		contestStatus.Status = resourcespb.Contest_REGISTRATION
	case "started":
		contestStatus.Status = resourcespb.Contest_STARTED
	case "finished":
		contestStatus.Status = resourcespb.Contest_FINISHED
	default:
		return nil, fmt.Errorf("unexpected contest status: %q", contestStatus.StatusStr)
	}
	return &contestStatus, nil
}

type loginRequiredOption struct {
	Team bool
	Lock bool
}

func loginRequired(e echo.Context, db sqlx.Queryer, option *loginRequiredOption) (*xsuportal.Contestant, *xsuportal.Team, bool, error) {
	contestant, err := getCurrentContestant(e, db, option.Lock)
	if err != nil {
		return contestant, nil, false, fmt.Errorf("current contestant: %w", err)
	}
	if contestant == nil {
		return contestant, nil, false, halt(e, http.StatusUnauthorized, "ログインが必要です", nil)
	}
	if option.Team {
		t, err := getCurrentTeam(e, db, option.Lock, contestant)
		if err != nil {
			return contestant, t, false, fmt.Errorf("current team: %w", err)
		}
		if t == nil {
			return contestant, t, false, halt(e, http.StatusForbidden, "参加登録が必要です", nil)
		}
		return contestant, t, true, nil
	}
	return contestant, nil, true, nil
}

func contestStatusRestricted(e echo.Context, db sqlx.Queryer, status resourcespb.Contest_Status, message string) (bool, error) {
	contestStatus, err := getCurrentContestStatus(e)
	if err != nil {
		return false, fmt.Errorf("get current contest status: %w", err)
	}
	if contestStatus.Status != status {
		return false, halt(e, http.StatusForbidden, message, nil)
	}
	return true, nil
}

func writeProto(e echo.Context, code int, m proto.Message) error {
	res, _ := proto.Marshal(m)
	return e.Blob(code, "application/vnd.google.protobuf", res)
}

func halt(e echo.Context, code int, humanMessage string, err error) error {
	message := &xsuportalpb.Error{
		Code: int32(code),
	}
	if err != nil {
		message.Name = fmt.Sprintf("%T", err)
		message.HumanMessage = err.Error()
		message.HumanDescriptions = strings.Split(fmt.Sprintf("%+v", err), "\n")
	}
	if humanMessage != "" {
		message.HumanMessage = humanMessage
		message.HumanDescriptions = []string{humanMessage}
	}
	res, _ := proto.Marshal(message)
	return e.Blob(code, "application/vnd.google.protobuf; proto=xsuportal.proto.Error", res)
}

func makeClarificationPB(db sqlx.Queryer, c *xsuportal.Clarification, t *xsuportal.Team) (*resourcespb.Clarification, error) {
	team, err := makeTeamPB(db, t, false, true)
	if err != nil {
		return nil, fmt.Errorf("make team: %w", err)
	}
	pb := &resourcespb.Clarification{
		Id:        c.ID,
		TeamId:    c.TeamID,
		Answered:  c.AnsweredAt.Valid,
		Disclosed: c.Disclosed.Bool,
		Question:  c.Question.String,
		Answer:    c.Answer.String,
		CreatedAt: timestamppb.New(c.CreatedAt),
		Team:      team,
	}
	if c.AnsweredAt.Valid {
		pb.AnsweredAt = timestamppb.New(c.AnsweredAt.Time)
	}
	return pb, nil
}

func makeTeamPB(db sqlx.Queryer, t *xsuportal.Team, detail bool, enableMembers bool) (*resourcespb.Team, error) {
	var pbv resourcespb.Team
	if v, ok := teamPBMap.Load(t.ID); ok {
		pbv = *v.(*resourcespb.Team)
	} else {
		var team xsuportal.Team
		db.QueryRowx("SELECT * FROM teams WHERE id = ?", t.ID).Scan(&team)
		v, err := cacheTeam(&team)
		if err != nil {
			return nil, err
		}
		pbv = *v
	}

	if !detail {
		pbv.Detail = nil
	}
	if !enableMembers {
		if !t.LeaderID.Valid {
			pbv.Leader = nil
		}
		pbv.Members = []*resourcespb.Contestant{}
		pbv.MemberIds = []string{}
	}
	if !t.Student.Valid {
		pbv.Student = nil
	}

	return &pbv, nil
}

func makeContestantPB(c *xsuportal.Contestant) *resourcespb.Contestant {
	return &resourcespb.Contestant{
		Id:        c.ID,
		TeamId:    c.TeamID.Int64,
		Name:      c.Name.String,
		IsStudent: c.Student,
		IsStaff:   c.Staff,
	}
}

func makeContestPB(e echo.Context) (*resourcespb.Contest, error) {
	contestStatus, err := getCurrentContestStatus(e)
	if err != nil {
		return nil, fmt.Errorf("get current contest status: %w", err)
	}
	return &resourcespb.Contest{
		RegistrationOpenAt: timestamppb.New(contestStatus.RegistrationOpenAt),
		ContestStartsAt:    timestamppb.New(contestStatus.ContestStartsAt),
		ContestFreezesAt:   timestamppb.New(contestStatus.ContestFreezesAt),
		ContestEndsAt:      timestamppb.New(contestStatus.ContestEndsAt),
		Status:             contestStatus.Status,
		Frozen:             contestStatus.Frozen,
	}, nil
}

func cacheTeam(t *xsuportal.Team) (*resourcespb.Team, error) {
	// fmt.Printf("[DEBUG]cacheTeam %s\n", t.ID)
	pb := &resourcespb.Team{
		Id:        t.ID,
		Name:      t.Name,
		LeaderId:  t.LeaderID.String,
		Withdrawn: t.Withdrawn,
	}
	pb.Detail = &resourcespb.Team_TeamDetail{
		EmailAddress: t.EmailAddress,
		InviteToken:  t.InviteToken,
	}
	rows, err := db.Queryx("SELECT * FROM `contestants` WHERE `team_id` = ? ORDER BY `created_at`", t.ID)
	if err != nil {
		return nil, fmt.Errorf("select members: %w", err)
	}
	defer rows.Close()
	isStudent := true
	for rows.Next() {
		var member xsuportal.Contestant
		rows.StructScan(&member)
		pb.Members = append(pb.Members, makeContestantPB(&member))
		pb.MemberIds = append(pb.MemberIds, member.ID)
		if t.LeaderID.Valid && t.LeaderID.String == member.ID {
			pb.Leader = makeContestantPB(&member)
		}
		isStudent = isStudent && member.Student
	}
	pb.Student = &resourcespb.Team_StudentStatus{
		Status: isStudent,
	}
	teamMap.Store(t.ID, t)
	teamPBMap.Store(t.ID, pb)
	return pb, nil
}

func cacheContestant(c *xsuportal.Contestant) (error) {
	contestantMap.Store(c.ID, c)
	return nil
}


var finishedLeaderboard *resourcespb.Leaderboard
func makeLeaderboardPB(e echo.Context, teamID int64) (*resourcespb.Leaderboard, error) {
	if finishedLeaderboard != nil {
		return finishedLeaderboard, nil
	}

	contestStatus, err := getCurrentContestStatus(e)
	if err != nil {
		return nil, fmt.Errorf("get current contest status: %w", err)
	}
	contestFinished := contestStatus.Status == resourcespb.Contest_FINISHED
	contestFreezesAt := contestStatus.ContestFreezesAt

	tx, err := db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()
	var leaderboard []xsuportal.LeaderBoardTeam
	if contestFinished {
		query := "SELECT\n" +
			"  `teams`.`id` AS `id`,\n" +
			"  `teams`.`name` AS `name`,\n" +
			"  `teams`.`leader_id` AS `leader_id`,\n" +
			"  `teams`.`withdrawn` AS `withdrawn`,\n" +
			"  `student`,\n" +
			"  best_score AS `best_score`,\n" +
			"  best_score_started_at AS `best_score_started_at`,\n" +
			"  best_score_marked_at AS `best_score_marked_at`,\n" +
			"  latest_score AS `latest_score`,\n" +
			"  latest_score_started_at AS `latest_score_started_at`,\n" +
			"  latest_score_marked_at AS `latest_score_marked_at`,\n" +
			"  finish_count AS `finish_count`\n" +
			"FROM\n" +
			"  `teams`\n" +
			"  JOIN scores ON teams.id = scores.team_id\n" +
			"ORDER BY\n" +
			"  `latest_score` DESC,\n" +
			"  `latest_score_marked_at` ASC\n"
		err := tx.Select(&leaderboard, query)
		if err != sql.ErrNoRows && err != nil {
			return nil, fmt.Errorf("select leaderboard: %w", err)
		}
	} else {
		query := "SELECT\n" +
			"  `teams`.`id` AS `id`,\n" +
			"  `teams`.`name` AS `name`,\n" +
			"  `teams`.`leader_id` AS `leader_id`,\n" +
			"  `teams`.`withdrawn` AS `withdrawn`,\n" +
			"  `student`,\n" +
			"  CASE WHEN team_id = ? THEN best_score ELSE freeze_best_score END AS `best_score`,\n" +
			"  CASE WHEN team_id = ? THEN best_score_started_at ELSE freeze_best_score_started_at END AS `best_score_started_at`,\n" +
			"  CASE WHEN team_id = ? THEN best_score_marked_at ELSE freeze_best_score_marked_at END AS `best_score_marked_at`,\n" +
			"  CASE WHEN team_id = ? THEN latest_score ELSE freeze_latest_score END AS `latest_score`,\n" +
			"  CASE WHEN team_id = ? THEN latest_score_started_at ELSE freeze_latest_score_started_at END AS `latest_score_started_at`,\n" +
			"  CASE WHEN team_id = ? THEN latest_score_marked_at ELSE freeze_latest_score_marked_at END AS `latest_score_marked_at`,\n" +
			"  CASE WHEN team_id = ? THEN finish_count ELSE freeze_finish_count END AS `finish_count`\n" +
			"FROM\n" +
			"  `teams`\n" +
			"  JOIN scores ON teams.id = scores.team_id\n" +
			"ORDER BY\n" +
			"  `latest_score` DESC,\n" +
			"  `latest_score_marked_at` ASC\n"
		err := tx.Select(&leaderboard, query, teamID, teamID, teamID, teamID, teamID, teamID, teamID)
		if err != sql.ErrNoRows && err != nil {
			return nil, fmt.Errorf("select leaderboard: %w", err)
		}
	}
	var jobResults []xsuportal.JobResult
	if contestFinished {
		jobResultsQuery := "SELECT\n" +
			"  `team_id` AS `team_id`,\n" +
			"  (`score_raw` - `score_deduction`) AS `score`,\n" +
			"  `started_at` AS `started_at`,\n" +
			"  `finished_at` AS `finished_at`\n" +
			"FROM\n" +
			"  `benchmark_jobs`\n" +
			"WHERE\n" +
			"  `started_at` IS NOT NULL\n" +
			"  AND `finished_at` IS NOT NULL\n" +
			"ORDER BY\n" +
			"  `finished_at`"
		err := tx.Select(&jobResults, jobResultsQuery)
		if err != sql.ErrNoRows && err != nil {
			return nil, fmt.Errorf("select job results: %w", err)
		}
	} else {
		jobResultsQuery := "SELECT\n" +
			"  `team_id` AS `team_id`,\n" +
			"  (`score_raw` - `score_deduction`) AS `score`,\n" +
			"  `started_at` AS `started_at`,\n" +
			"  `finished_at` AS `finished_at`\n" +
			"FROM\n" +
			"  `benchmark_jobs`\n" +
			"WHERE\n" +
			"  `started_at` IS NOT NULL\n" +
			"  AND (\n" +
			"    `finished_at` IS NOT NULL\n" +
			"    -- score freeze\n" +
			"    AND (`team_id` = ? OR (`team_id` != ? AND `finished_at` < ?))\n" +
			"  )\n" +
			"ORDER BY\n" +
			"  `finished_at`"
		err := tx.Select(&jobResults, jobResultsQuery, teamID, teamID, contestFreezesAt)
		if err != sql.ErrNoRows && err != nil {
			return nil, fmt.Errorf("select job results: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	teamGraphScores := make(map[int64][]*resourcespb.Leaderboard_LeaderboardItem_LeaderboardScore)
	for _, jobResult := range jobResults {
		teamGraphScores[jobResult.TeamID] = append(teamGraphScores[jobResult.TeamID], &resourcespb.Leaderboard_LeaderboardItem_LeaderboardScore{
			Score:     jobResult.Score,
			StartedAt: timestamppb.New(jobResult.StartedAt),
			MarkedAt:  timestamppb.New(jobResult.FinishedAt),
		})
	}
	pb := &resourcespb.Leaderboard{}
	for _, team := range leaderboard {
		t, _ := makeTeamPB(db, team.Team(), false, false)
		item := &resourcespb.Leaderboard_LeaderboardItem{
			Scores: teamGraphScores[team.ID],
			BestScore: &resourcespb.Leaderboard_LeaderboardItem_LeaderboardScore{
				Score:     team.BestScore.Int64,
				StartedAt: toTimestamp(team.BestScoreStartedAt),
				MarkedAt:  toTimestamp(team.BestScoreMarkedAt),
			},
			LatestScore: &resourcespb.Leaderboard_LeaderboardItem_LeaderboardScore{
				Score:     team.LatestScore.Int64,
				StartedAt: toTimestamp(team.LatestScoreStartedAt),
				MarkedAt:  toTimestamp(team.LatestScoreMarkedAt),
			},
			Team:        t,
			FinishCount: team.FinishCount.Int64,
		}
		if team.Student.Valid && team.Student.Bool {
			pb.StudentTeams = append(pb.StudentTeams, item)
		} else {
			pb.GeneralTeams = append(pb.GeneralTeams, item)
		}
		pb.Teams = append(pb.Teams, item)
	}
	if contestFinished {
		finishedLeaderboard = pb
	}
	return pb, nil
}

func makeBenchmarkJobPB(job *xsuportal.BenchmarkJob) *resourcespb.BenchmarkJob {
	pb := &resourcespb.BenchmarkJob{
		Id:             job.ID,
		TeamId:         job.TeamID,
		Status:         resourcespb.BenchmarkJob_Status(job.Status),
		TargetHostname: job.TargetHostName,
		CreatedAt:      timestamppb.New(job.CreatedAt),
		UpdatedAt:      timestamppb.New(job.UpdatedAt),
	}
	if job.StartedAt.Valid {
		pb.StartedAt = timestamppb.New(job.StartedAt.Time)
	}
	if job.FinishedAt.Valid {
		pb.FinishedAt = timestamppb.New(job.FinishedAt.Time)
		pb.Result = makeBenchmarkResultPB(job)
	}
	return pb
}

func makeBenchmarkResultPB(job *xsuportal.BenchmarkJob) *resourcespb.BenchmarkResult {
	hasScore := job.ScoreRaw.Valid && job.ScoreDeduction.Valid
	pb := &resourcespb.BenchmarkResult{
		Finished: job.FinishedAt.Valid,
		Passed:   job.Passed.Bool,
		Reason:   job.Reason.String,
	}
	if hasScore {
		pb.Score = int64(job.ScoreRaw.Int32 - job.ScoreDeduction.Int32)
		pb.ScoreBreakdown = &resourcespb.BenchmarkResult_ScoreBreakdown{
			Raw:       int64(job.ScoreRaw.Int32),
			Deduction: int64(job.ScoreDeduction.Int32),
		}
	}
	return pb
}

func makeBenchmarkJobsPB(e echo.Context, db sqlx.Queryer, team *xsuportal.Team, limit int) ([]*resourcespb.BenchmarkJob, error) {
	// team, _ := getCurrentTeam(e, db, false, nil)
	query := "SELECT * FROM `benchmark_jobs` WHERE `team_id` = ? ORDER BY `created_at` DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	var jobs []xsuportal.BenchmarkJob
	if err := sqlx.Select(db, &jobs, query, team.ID); err != nil {
		return nil, fmt.Errorf("select benchmark jobs: %w", err)
	}
	var benchmarkJobs []*resourcespb.BenchmarkJob
	for _, job := range jobs {
		benchmarkJobs = append(benchmarkJobs, makeBenchmarkJobPB(&job))
	}
	return benchmarkJobs, nil
}

func makeNotificationsPB(notifications []*xsuportal.Notification) ([]*resourcespb.Notification, error) {
	var ns []*resourcespb.Notification
	for _, notification := range notifications {
		decoded, err := base64.StdEncoding.DecodeString(notification.EncodedMessage)
		if err != nil {
			return nil, fmt.Errorf("decode message: %w", err)
		}
		var message resourcespb.Notification
		if err := proto.Unmarshal(decoded, &message); err != nil {
			return nil, fmt.Errorf("unmarshal message: %w", err)
		}
		message.Id = notification.ID
		message.CreatedAt = timestamppb.New(notification.CreatedAt)
		ns = append(ns, &message)
	}
	return ns, nil
}

func wrapError(message string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

func toTimestamp(t sql.NullTime) *timestamppb.Timestamp {
	if t.Valid {
		return timestamppb.New(t.Time)
	}
	return nil
}
