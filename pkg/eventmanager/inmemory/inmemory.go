package inmemory

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/grulex/go-wishlist/pkg/eventmanager"
	"sync"
)

type storageEvent struct {
	Name    eventmanager.EventName
	Payload json.RawMessage
}
type EventManager struct {
	handlers   map[eventmanager.EventName][]eventmanager.EventHandler
	eventsChan chan storageEvent
	mu         *sync.RWMutex
}

func NewEventManager(bufferSize int) *EventManager {
	return &EventManager{
		handlers:   make(map[eventmanager.EventName][]eventmanager.EventHandler),
		eventsChan: make(chan storageEvent, bufferSize),
		mu:         &sync.RWMutex{},
	}
}

func (m EventManager) Publish(_ context.Context, event eventmanager.Event) error {
	jsonPayload, err := json.Marshal(event.GetPayload())
	if err != nil {
		return fmt.Errorf("marshal event payload: %w", err)
	}
	m.eventsChan <- storageEvent{
		Name:    event.GetName(),
		Payload: jsonPayload,
	}
	return nil
}

func (m EventManager) PublishMany(ctx context.Context, events ...eventmanager.Event) error {
	for _, event := range events {
		err := m.Publish(ctx, event)
		if err != nil {
			return fmt.Errorf("publish event: %w", err)
		}
	}
	return nil
}

func (m EventManager) Subscribe(eventName eventmanager.EventName, handler eventmanager.EventHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[eventName] = append(m.handlers[eventName], handler)
}

func (m EventManager) StartHandling(ctx context.Context) error {
	for {
		select {
		case event := <-m.eventsChan:
			m.mu.RLock()
			handlers, ok := m.handlers[event.Name]
			m.mu.RUnlock()
			if !ok {
				continue
			}
			for _, handler := range handlers {
				err := handler(context.Background(), event.Payload)
				if err != nil {
					fmt.Println(err)
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (m EventManager) Stop(_ context.Context) error {
	close(m.eventsChan)
	return nil
}
