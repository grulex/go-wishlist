package eventmanager

import (
	"context"
	"encoding/json"
	"errors"
)

var ErrInvalidPayload = errors.New("invalid payload")

type EventName string

// Payload Must be json marshalable and unmarshalable
type Payload interface {
}

type Event interface {
	GetName() EventName
	GetPayload() Payload
}

type EventHandler func(ctx context.Context, payload json.RawMessage) error

type EventManager interface {
	Publish(ctx context.Context, event Event) error
	PublishMany(ctx context.Context, events ...Event) error
	Subscribe(eventName EventName, handler EventHandler)
	StartHandling(ctx context.Context) error
	Stop(ctx context.Context) error
}
