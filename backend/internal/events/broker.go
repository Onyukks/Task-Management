// Package events implements a tiny in-process pub/sub broker used to push
// real-time task updates to connected clients over Server-Sent Events.
//
// Events are scoped per user so a client only receives changes to its own
// tasks. This is intentionally in-memory: it is simple, dependency-free, and
// sufficient for a single API instance. Scaling to multiple instances would
// mean swapping this for Redis pub/sub or Postgres LISTEN/NOTIFY (noted as a
// trade-off in the README).
package events

import (
	"sync"

	"github.com/google/uuid"
)

// Event is a single change broadcast to subscribers.
type Event struct {
	Type string `json:"type"` // task.created | task.updated | task.deleted
	Data any    `json:"data"`
}

type subscriber struct {
	userID uuid.UUID
	ch     chan Event
}

type Broker struct {
	mu   sync.RWMutex
	subs map[*subscriber]struct{}
}

func NewBroker() *Broker {
	return &Broker{subs: make(map[*subscriber]struct{})}
}

// Subscribe registers a listener for a user's events and returns the channel
// plus an unsubscribe func the caller must defer.
func (b *Broker) Subscribe(userID uuid.UUID) (<-chan Event, func()) {
	s := &subscriber{userID: userID, ch: make(chan Event, 16)}
	b.mu.Lock()
	b.subs[s] = struct{}{}
	b.mu.Unlock()

	return s.ch, func() {
		b.mu.Lock()
		delete(b.subs, s)
		b.mu.Unlock()
		close(s.ch)
	}
}

// Publish delivers an event to every subscriber for the given user. Delivery
// is non-blocking: a slow client that has filled its buffer simply drops the
// event rather than stalling the publisher.
func (b *Broker) Publish(userID uuid.UUID, e Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for s := range b.subs {
		if s.userID != userID {
			continue
		}
		select {
		case s.ch <- e:
		default:
		}
	}
}
