package xsuportal

import (
	"crypto/elliptic"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/golang/protobuf/proto"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/isucon/isucon10-final/webapp/golang/proto/xsuportal/resources"
)

const (
	WebpushVAPIDPrivateKeyPath = "../vapid_private.pem"
	WebpushSubject             = "xsuportal@example.com"
)

type Notifier struct {
	mu      sync.Mutex
	options *webpush.Options
}

func (n *Notifier) VAPIDKey() *webpush.Options {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.options == nil {
		pemBytes, err := ioutil.ReadFile(WebpushVAPIDPrivateKeyPath)
		if err != nil {
			return nil
		}
		block, _ := pem.Decode(pemBytes)
		if block == nil {
			return nil
		}
		priKey, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil
		}
		priBytes := priKey.D.Bytes()
		pubBytes := elliptic.Marshal(priKey.Curve, priKey.X, priKey.Y)
		pri := base64.RawURLEncoding.EncodeToString(priBytes)
		pub := base64.RawURLEncoding.EncodeToString(pubBytes)
		n.options = &webpush.Options{
			HTTPClient: &http.Client{
				Transport: &http.Transport{ MaxIdleConnsPerHost: 100 },
			},
			Subscriber:      WebpushSubject,
			VAPIDPrivateKey: pri,
			VAPIDPublicKey:  pub,
		}
	}
	return n.options
}

func (n *Notifier) NotifyClarificationAnswered(db sqlx.Ext, c *Clarification, updated bool) error {
	var contestants []struct {
		ID     string `db:"id"`
		TeamID int64  `db:"team_id"`
	}
	if c.Disclosed.Valid && c.Disclosed.Bool {
		err := sqlx.Select(
			db,
			&contestants,
			"SELECT `id`, `team_id` FROM `contestants` WHERE `team_id` IS NOT NULL",
		)
		if err != nil {
			return fmt.Errorf("select all contestants: %w", err)
		}
	} else {
		err := sqlx.Select(
			db,
			&contestants,
			"SELECT `id`, `team_id` FROM `contestants` WHERE `team_id` = ?",
			c.TeamID,
		)
		if err != nil {
			return fmt.Errorf("select contestants(team_id=%v): %w", c.TeamID, err)
		}
	}
	wg := sync.WaitGroup{}
	vapidKey := n.VAPIDKey()
	for _, contestant := range contestants {
		notificationPB := &resources.Notification{
			Content: &resources.Notification_ContentClarification{
				ContentClarification: &resources.Notification_ClarificationMessage{
					ClarificationId: c.ID,
					Owned:           c.TeamID == contestant.TeamID,
					Updated:         updated,
				},
			},
		}
		notification, err := n.notify(db, notificationPB, contestant.ID)
		if err != nil {
			return fmt.Errorf("notify: %w", err)
		}
		if vapidKey != nil {
			notificationPB.Id = notification.ID
			notificationPB.CreatedAt = timestamppb.New(notification.CreatedAt)
			b, err := proto.Marshal(notificationPB)
			if err != nil {
				return fmt.Errorf("marshal notification: %w", err)
			}
			message := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
			base64.StdEncoding.Encode(message, b)

			subscriptions, err := GetPushSubscriptions(db, contestant.ID)
			if err != nil {
				return fmt.Errorf("get push subscrptions: %w", err)
			}
			if len(subscriptions) == 0 {
				return fmt.Errorf("no push subscriptions found: contestant_id=%v", contestant.ID)
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				for _, subscription := range subscriptions {
					SendWebPush(vapidKey, message, &subscription)
				}
			}()
		}
	}
	wg.Wait()
	return nil
}

func GetPushSubscriptions(db sqlx.Queryer, contestantID string) ([]PushSubscription, error) {
	var subscriptions []PushSubscription
	err := sqlx.Select(
		db,
		&subscriptions,
		"SELECT * FROM `push_subscriptions` WHERE `contestant_id` = ?",
		contestantID,
	)
	if err != sql.ErrNoRows && err != nil {
		return nil, fmt.Errorf("select push subscriptions: %w", err)
	}
	return subscriptions, nil
}

func SendWebPush(options *webpush.Options, message []byte, pushSubscription *PushSubscription) error {
	resp, err := webpush.SendNotification(
		message,
		&webpush.Subscription{
			Endpoint: pushSubscription.Endpoint,
			Keys: webpush.Keys{
				Auth:   pushSubscription.Auth,
				P256dh: pushSubscription.P256DH,
			},
		},
		options,
	)
	if err != nil {
		return fmt.Errorf("send notification: %w", err)
	}
	defer resp.Body.Close()
	expired := resp.StatusCode == 410
	if expired {
		return fmt.Errorf("expired notification")
	}
	invalid := resp.StatusCode == 404
	if invalid {
		return fmt.Errorf("invalid notification")
	}
	return nil
}

func (n *Notifier) NotifyBenchmarkJobFinished(db sqlx.Ext, job *BenchmarkJob) error {
	var contestants []struct {
		ID     string `db:"id"`
		TeamID int64  `db:"team_id"`
	}
	err := sqlx.Select(
		db,
		&contestants,
		"SELECT `id`, `team_id` FROM `contestants` WHERE `team_id` = ?",
		job.TeamID,
	)
	if err != nil {
		return fmt.Errorf("select contestants(team_id=%v): %w", job.TeamID, err)
	}
	wg := sync.WaitGroup{}
	vapidKey := n.VAPIDKey()
	for _, contestant := range contestants {
		notificationPB := &resources.Notification{
			Content: &resources.Notification_ContentBenchmarkJob{
				ContentBenchmarkJob: &resources.Notification_BenchmarkJobMessage{
					BenchmarkJobId: job.ID,
				},
			},
		}
		notification, err := n.notify(db, notificationPB, contestant.ID)
		if err != nil {
			return fmt.Errorf("notify: %w", err)
		}
		if vapidKey != nil {
			notificationPB.Id = notification.ID
			notificationPB.CreatedAt = timestamppb.New(notification.CreatedAt)
			b, err := proto.Marshal(notificationPB)
			if err != nil {
				return fmt.Errorf("marshal notification: %w", err)
			}
			message := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
			base64.StdEncoding.Encode(message, b)

			subscriptions, err := GetPushSubscriptions(db, contestant.ID)
			if err != nil {
				return fmt.Errorf("get push subscrptions: %w", err)
			}
			if len(subscriptions) == 0 {
				return fmt.Errorf("no push subscriptions found: contestant_id=%v", contestant.ID)
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				for _, subscription := range subscriptions {
					SendWebPush(vapidKey, message, &subscription)
				}
			}()
		}
	}
	wg.Wait()
	return nil
}

var notificationId int64
func (n *Notifier) notify(db sqlx.Ext, notificationPB *resources.Notification, contestantID string) (*Notification, error) {
	/*
	m, err := proto.Marshal(notificationPB)
	if err != nil {
		return nil, fmt.Errorf("marshal notification: %w", err)
	}
	encodedMessage := base64.StdEncoding.EncodeToString(m)
	res, err := db.Exec(
		"INSERT INTO `notifications` (`contestant_id`, `encoded_message`, `read`, `created_at`, `updated_at`) VALUES (?, ?, FALSE, NOW(6), NOW(6))",
		contestantID,
		encodedMessage,
	)
	if err != nil {
		return nil, fmt.Errorf("insert notification: %w", err)
	}
	lastInsertID, _ := res.LastInsertId()
	var notification Notification
	err = sqlx.Get(
		db,
		&notification,
		"SELECT id, created_at FROM `notifications` WHERE `id` = ? LIMIT 1",
		lastInsertID,
	)
	if err != nil {
		return nil, fmt.Errorf("get inserted notification: %w", err)
	}
	*/
	notification := Notification{
		ID: atomic.AddInt64(&notificationId, 1),
		CreatedAt: time.Now(),
	}
	return &notification, nil
}


