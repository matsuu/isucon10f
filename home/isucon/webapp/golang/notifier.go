package xsuportal

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"sync"

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
		if n.VAPIDKey() != nil {
			notificationPB.Id = notification.ID
			notificationPB.CreatedAt = timestamppb.New(notification.CreatedAt)

			// TODO: Web Push IIKANJI NI SHITE
			subscriptions, err := GetPushSubscriptions(db, contestant.ID)
			if err != nil {
				return fmt.Errorf("get push subscrptions: %w", err)
			}
			if len(subscriptions) == 0 {
				return fmt.Errorf("no push subscriptions found: contestant_id=%v", contestant.ID)
			}

	                vapidKey, err := GetVAPIDKey(WebpushVAPIDPrivateKeyPath)
			if err != nil {
				panic(err)
			}
			wg := sync.WaitGroup{}
			for _, subscription := range subscriptions {
				wg.Add(1)
				go func(vapidKey *ecdsa.PrivateKey, notificationPB *resources.Notification, pushSubscription *PushSubscription) {
					defer wg.Done()
					SendWebPush(vapidKey, notificationPB, &subscription)
				}(vapidKey, notificationPB, &subscription)
			}
			wg.Wait()
		}
	}
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

func SendWebPush(vapidKey *ecdsa.PrivateKey, notificationPB *resources.Notification, pushSubscription *PushSubscription) error {
	b, err := proto.Marshal(notificationPB)
	if err != nil {
		return fmt.Errorf("marshal notification: %w", err)
	}
	message := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
	base64.StdEncoding.Encode(message, b)

	vapidPrivateKey := base64.RawURLEncoding.EncodeToString(vapidKey.D.Bytes())
	vapidPublicKey := base64.RawURLEncoding.EncodeToString(elliptic.Marshal(vapidKey.Curve, vapidKey.X, vapidKey.Y))

	resp, err := webpush.SendNotification(
		message,
		&webpush.Subscription{
			Endpoint: pushSubscription.Endpoint,
			Keys: webpush.Keys{
				Auth:   pushSubscription.Auth,
				P256dh: pushSubscription.P256DH,
			},
		},
		&webpush.Options{
			Subscriber:      WebpushSubject,
			VAPIDPublicKey:  vapidPublicKey,
			VAPIDPrivateKey: vapidPrivateKey,
		},
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
		return nil
		// return fmt.Errorf("invalid notification")
	}
	return nil
}

func GetVAPIDKey(path string) (*ecdsa.PrivateKey, error) {
	pemBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read pem: %w", err)
	}
	for {
		block, rest := pem.Decode(pemBytes)
		pemBytes = rest
		if block == nil {
			break
		}
		ecPrivateKey, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			continue
		}
		return ecPrivateKey, nil
	}
	return nil, fmt.Errorf("not found ec private key")
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
		if n.VAPIDKey() != nil {
			notificationPB.Id = notification.ID
			notificationPB.CreatedAt = timestamppb.New(notification.CreatedAt)
			// TODO: Web Push IIKANJI NI SHITE
			subscriptions, err := GetPushSubscriptions(db, contestant.ID)
			if err != nil {
				return fmt.Errorf("get push subscrptions: %w", err)
			}
			if len(subscriptions) == 0 {
				return fmt.Errorf("no push subscriptions found: contestant_id=%v", contestant.ID)
			}

		        vapidKey, err := GetVAPIDKey(WebpushVAPIDPrivateKeyPath)
			if err != nil {
				panic(err)
			}
			wg := sync.WaitGroup{}
			for _, subscription := range subscriptions {
				wg.Add(1)
				go func(vapidKey *ecdsa.PrivateKey, notificationPB *resources.Notification, pushSubscription *PushSubscription) {
					defer wg.Done()
					SendWebPush(vapidKey, notificationPB, &subscription)
				}(vapidKey, notificationPB, &subscription)
			}
			wg.Wait()
		}
	}
	return nil
}

func (n *Notifier) notify(db sqlx.Ext, notificationPB *resources.Notification, contestantID string) (*Notification, error) {
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
		"SELECT * FROM `notifications` WHERE `id` = ? LIMIT 1",
		lastInsertID,
	)
	if err != nil {
		return nil, fmt.Errorf("get inserted notification: %w", err)
	}
	return &notification, nil
}


