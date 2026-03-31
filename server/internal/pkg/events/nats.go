package events

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
)

type Publisher interface {
	PublishJSON(ctx context.Context, subject string, payload any) error
}

type NATSPublisher struct {
	conn *nats.Conn
}

func NewNATSPublisher(conn *nats.Conn) *NATSPublisher {
	if conn == nil {
		return nil
	}
	return &NATSPublisher{conn: conn}
}

func (p *NATSPublisher) PublishJSON(ctx context.Context, subject string, payload any) error {
	if p == nil || p.conn == nil {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return p.conn.Publish(subject, body)
}
