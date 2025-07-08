package nats

import (
	"context"
	"errors"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"gobot.io/x/gobot/v2"
)

var (
	// ErrDriverNotReady indicates the driver is not ready for operation
	ErrDriverNotReady = errors.New("driver not ready")
)

const (
	// Data event when data is available for Driver
	Data = "data"

	// Error event when error occurs in Driver
	Error = "error"
)

// Driver for NATS with modern JetStream support
type Driver struct {
	name       string
	topic      string
	connection gobot.Connection
	gobot.Eventer
	gobot.Commander
	stream     string
	consumer   string
}

// NewDriver returns a new Gobot NATS Driver with modern JetStream support
func NewDriver(a *Adaptor, topic string) *Driver {
	m := &Driver{
		name:       gobot.DefaultName("NATS"),
		topic:      topic,
		connection: a,
		Eventer:    gobot.NewEventer(),
		Commander:  gobot.NewCommander(),
		stream:     "DEFAULT",
		consumer:   gobot.DefaultName("CONSUMER"),
	}

	return m
}

// NewDriverWithJetStream creates a new NATS driver with JetStream configuration
func NewDriverWithJetStream(a *Adaptor, topic, stream, consumer string) *Driver {
	m := NewDriver(a, topic)
	m.stream = stream
	m.consumer = consumer
	return m
}

// Name returns name for the Driver
func (m *Driver) Name() string { return m.name }

// SetName sets name for the Driver
func (m *Driver) SetName(name string) { m.name = name }

// Connection returns Connections used by the Driver
func (m *Driver) Connection() gobot.Connection {
	return m.connection
}

func (m *Driver) adaptor() *Adaptor {
	//nolint:forcetypeassert // ok here
	return m.Connection().(*Adaptor)
}

// Start starts the Driver
func (m *Driver) Start() error {
	return nil
}

// Halt halts the Driver
func (m *Driver) Halt() error {
	return nil
}

// Topic returns the current topic for the Driver
func (m *Driver) Topic() string { return m.topic }

// SetTopic sets the current topic for the Driver
func (m *Driver) SetTopic(topic string) { m.topic = topic }

// Publish a message to the current device topic
func (m *Driver) Publish(data any) bool {
	//nolint:forcetypeassert // ok here
	message := data.([]byte)
	return m.adaptor().Publish(m.topic, message)
}

// On subscribes to data updates for the current device topic,
// and then calls the message handler function when data is received
// Uses modern context-aware subscription patterns
func (m *Driver) On(n string, f func(msg Message)) error {
	// TODO: also be able to subscribe to Error updates
	m.adaptor().On(m.topic, f)
	return nil
}

// SetStream sets the JetStream stream name for this driver
func (m *Driver) SetStream(stream string) {
	m.stream = stream
}

// GetStream returns the current JetStream stream name
func (m *Driver) GetStream() string {
	return m.stream
}

// SetConsumer sets the JetStream consumer name for this driver
func (m *Driver) SetConsumer(consumer string) {
	m.consumer = consumer
}

// GetConsumer returns the current JetStream consumer name
func (m *Driver) GetConsumer() string {
	return m.consumer
}

// PublishWithContext publishes a message using context for timeout control
func (m *Driver) PublishWithContext(ctx context.Context, data any) error {
	//nolint:forcetypeassert // ok here
	message := data.([]byte)
	adaptor := m.adaptor()
	
	if js := adaptor.JetStream(); js != nil {
		// Use JetStream publish with context
		_, err := js.Publish(ctx, m.topic, message)
		return err
	}
	
	// Fallback to core NATS
	if adaptor.Publish(m.topic, message) {
		return nil
	}
	return ErrDriverNotReady
}

// CreateStream creates a JetStream stream for this driver's topic
func (m *Driver) CreateStream(ctx context.Context) (jetstream.Stream, error) {
	adaptor := m.adaptor()
	js := adaptor.JetStream()
	if js == nil {
		return nil, ErrDriverNotReady
	}
	
	return js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     m.stream,
		Subjects: []string{m.topic},
		Storage:  jetstream.MemoryStorage,
	})
}

// CreateConsumer creates a JetStream consumer for this driver
func (m *Driver) CreateConsumer(ctx context.Context) (jetstream.Consumer, error) {
	adaptor := m.adaptor()
	js := adaptor.JetStream()
	if js == nil {
		return nil, ErrDriverNotReady
	}
	
	return js.CreateOrUpdateConsumer(ctx, m.stream, jetstream.ConsumerConfig{
		Durable:       m.consumer,
		FilterSubject: m.topic,
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
}

// ConsumeMessages starts consuming messages using modern JetStream pull-based approach
func (m *Driver) ConsumeMessages(ctx context.Context, batchSize int, f func(msg Message)) error {
	consumer, err := m.CreateConsumer(ctx)
	if err != nil {
		return err
	}
	
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				msgs, err := consumer.Fetch(batchSize, jetstream.FetchMaxWait(5*time.Second))
				if err != nil {
					time.Sleep(100 * time.Millisecond)
					continue
				}
				for msg := range msgs.Messages() {
					// Convert JetStream message to core NATS message format
					natsMsg := &nats.Msg{
						Subject: msg.Subject(),
						Data:    msg.Data(),
						Header:  msg.Headers(),
					}
					f(natsMsg)
					msg.Ack()
				}
			}
		}
	}()
	
	return nil
}