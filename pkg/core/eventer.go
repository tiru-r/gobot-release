package core

import (
	"context"
	"sync"
	"time"
)

type eventChannel chan *Event

type eventer struct {
	// map of valid Event names
	eventnames map[string]string

	// new events get put in to the event channel
	in eventChannel

	// map of out channels used by subscribers
	outs map[eventChannel]eventChannel

	// mutex to protect the eventChannel map
	eventsMutex sync.Mutex

	// context for graceful shutdown
	ctx context.Context
	cancel context.CancelFunc

	// done channel to signal shutdown completion
	done chan struct{}
}

const eventChanBufferSize = 10

// Eventer is the interface which describes how a Driver or Adaptor
// handles events.
type Eventer interface {
	// Events returns the map of valid Event names.
	Events() (eventnames map[string]string)

	// Event returns an Event string from map of valid Event names.
	// Mostly used to validate that an Event name is valid.
	Event(name string) string

	// AddEvent registers a new Event name.
	AddEvent(name string)

	// DeleteEvent removes a previously registered Event name.
	DeleteEvent(name string)

	// Publish new events to any subscriber
	Publish(name string, data interface{})

	// Subscribe to events
	Subscribe() (events eventChannel)

	// Unsubscribe from an event channel
	Unsubscribe(events eventChannel)

	// Event handler
	On(name string, f func(s interface{})) error

	// Event handler, only executes one time
	Once(name string, f func(s interface{})) error

	// Shutdown gracefully stops the eventer
	Shutdown(ctx context.Context) error
}

// NewEventer returns a new Eventer.
func NewEventer() Eventer {
	ctx, cancel := context.WithCancel(context.Background())
	evtr := &eventer{
		eventnames: make(map[string]string),
		in:         make(eventChannel, eventChanBufferSize),
		outs:       make(map[eventChannel]eventChannel),
		ctx:        ctx,
		cancel:     cancel,
		done:       make(chan struct{}),
	}

	// goroutine to cascade "in" events to all "out" event channels
	go func() {
		defer close(evtr.done)
		for {
			select {
			case evt := <-evtr.in:
				evtr.eventsMutex.Lock()
				for _, out := range evtr.outs {
					select {
					case out <- evt:
					case <-evtr.ctx.Done():
						evtr.eventsMutex.Unlock()
						return
					default:
						// Drop event if channel is full to prevent blocking
					}
				}
				evtr.eventsMutex.Unlock()
			case <-evtr.ctx.Done():
				return
			}
		}
	}()

	return evtr
}

// Events returns the map of valid Event names.
func (e *eventer) Events() map[string]string {
	return e.eventnames
}

// Event returns an Event string from map of valid Event names.
// Mostly used to validate that an Event name is valid.
func (e *eventer) Event(name string) string {
	return e.eventnames[name]
}

// AddEvent registers a new Event name.
func (e *eventer) AddEvent(name string) {
	e.eventnames[name] = name
}

// DeleteEvent removes a previously registered Event name.
func (e *eventer) DeleteEvent(name string) {
	delete(e.eventnames, name)
}

// Publish new events to anyone that is subscribed
func (e *eventer) Publish(name string, data interface{}) {
	evt := NewEvent(name, data)
	select {
	case e.in <- evt:
	case <-e.ctx.Done():
		// Eventer is shutting down, drop the event
	case <-time.After(100 * time.Millisecond):
		// Drop event if channel is full to prevent blocking
	}
}

// Subscribe to any events from this eventer
func (e *eventer) Subscribe() eventChannel {
	e.eventsMutex.Lock()
	defer e.eventsMutex.Unlock()
	out := make(eventChannel, eventChanBufferSize)
	e.outs[out] = out
	return out
}

// Unsubscribe from the event channel
func (e *eventer) Unsubscribe(events eventChannel) {
	e.eventsMutex.Lock()
	defer e.eventsMutex.Unlock()
	delete(e.outs, events)
}

// On executes the event handler f when e is Published to.
func (e *eventer) On(n string, f func(s interface{})) error {
	out := e.Subscribe()
	go func() {
		defer e.Unsubscribe(out)
		for {
			select {
			case evt := <-out:
				if evt.Name == n {
					f(evt.Data)
				}
			case <-e.ctx.Done():
				return
			}
		}
	}()

	return nil
}

// Once is similar to On except that it only executes f one time.
func (e *eventer) Once(n string, f func(s interface{})) error {
	out := e.Subscribe()
	go func() {
		defer e.Unsubscribe(out)
		for {
			select {
			case evt := <-out:
				if evt.Name == n {
					f(evt.Data)
					return
				}
			case <-e.ctx.Done():
				return
			}
		}
	}()

	return nil
}

// Shutdown gracefully stops the eventer
func (e *eventer) Shutdown(ctx context.Context) error {
	// Cancel the context to stop all goroutines
	e.cancel()
	
	// Wait for the main dispatcher to finish with timeout
	select {
	case <-e.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
