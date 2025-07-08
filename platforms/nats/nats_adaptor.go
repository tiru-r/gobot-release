package nats

import (
	"context"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"gobot.io/x/gobot/v2"
)

// Adaptor is a configuration struct for interacting with a NATS server.
// Name is a logical name for the adaptor/nats server connection.
// Host is in the form "localhost:4222" which is the hostname/ip and port of the nats server.
// ClientID is a unique identifier integer that specifies the identity of the client.
type Adaptor struct {
	name      string
	Host      string
	clientID  int
	username  string
	password  string
	client    *nats.Conn
	js        jetstream.JetStream
	connect   func() (*nats.Conn, error)
	ctx       context.Context
	cancel    context.CancelFunc
	timeout   time.Duration
}

// Message is a message received from the server.
type Message *nats.Msg

// NewAdaptor populates a new NATS Adaptor with modern JetStream support.
func NewAdaptor(host string, clientID int, options ...nats.Option) *Adaptor {
	hosts, err := processHostString(host)
	ctx, cancel := context.WithCancel(context.Background())

	return &Adaptor{
		name:     gobot.DefaultName("NATS"),
		Host:     hosts,
		clientID: clientID,
		ctx:      ctx,
		cancel:   cancel,
		timeout:  30 * time.Second,
		connect: func() (*nats.Conn, error) {
			if err != nil {
				return nil, err
			}
			return nats.Connect(hosts, options...)
		},
	}
}

// NewAdaptorWithAuth populates a NATS Adaptor including username and password with modern JetStream support.
func NewAdaptorWithAuth(host string, clientID int, username string, password string, options ...nats.Option) *Adaptor {
	hosts, err := processHostString(host)
	ctx, cancel := context.WithCancel(context.Background())

	return &Adaptor{
		name:     gobot.DefaultName("NATS"),
		Host:     hosts,
		clientID: clientID,
		username: username,
		password: password,
		ctx:      ctx,
		cancel:   cancel,
		timeout:  30 * time.Second,
		connect: func() (*nats.Conn, error) {
			if err != nil {
				return nil, err
			}
			return nats.Connect(hosts, append(options, nats.UserInfo(username, password))...)
		},
	}
}

func processHostString(host string) (string, error) {
	urls := strings.Split(host, ",")
	for i, s := range urls {
		s = strings.TrimSpace(s)
		if !strings.HasPrefix(s, "tls://") && !strings.HasPrefix(s, "nats://") {
			s = "nats://" + s
		}

		u, err := url.Parse(s)
		if err != nil {
			return "", err
		}

		urls[i] = u.String()
	}

	return strings.Join(urls, ","), nil
}

// Name returns the logical client name.
func (a *Adaptor) Name() string { return a.name }

// SetName sets the logical client name.
func (a *Adaptor) SetName(n string) { a.name = n }

// Connect makes a connection to the Nats server with modern JetStream support.
func (a *Adaptor) Connect() error {
	var err error
	a.client, err = a.connect()
	if err != nil {
		return err
	}

	// Initialize JetStream with modern API
	a.js, err = jetstream.New(a.client)
	if err != nil {
		// JetStream not available, but core NATS still works
		log.Printf("JetStream not available: %v", err)
	}

	return nil
}

// Disconnect from the nats server with proper cleanup.
func (a *Adaptor) Disconnect() error {
	if a.cancel != nil {
		a.cancel()
	}
	if a.client != nil {
		a.client.Close()
	}
	return nil
}

// Finalize is simply a helper method for the disconnect.
func (a *Adaptor) Finalize() error {
	return a.Disconnect()
}

// Publish sends a message with the particular topic to the nats server with modern context support.
func (a *Adaptor) Publish(topic string, message []byte) bool {
	if a.client == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(a.ctx, a.timeout)
	defer cancel()

	// Try JetStream publish first if available
	if a.js != nil {
		if _, err := a.js.Publish(ctx, topic, message); err != nil {
			// Fallback to core NATS publish
			if err := a.client.Publish(topic, message); err != nil {
				log.Println(err)
				return false
			}
		}
	} else {
		// Core NATS publish
		if err := a.client.Publish(topic, message); err != nil {
			log.Println(err)
			return false
		}
	}

	return true
}

// On is an event-handler style subscriber to a particular topic (named event).
// Supply a handler function to use the bytes returned by the server.
// Uses modern context-aware subscription with fallback to core NATS.
func (a *Adaptor) On(event string, f func(msg Message)) bool {
	if a.client == nil {
		return false
	}

	// Try JetStream consumer first if available
	if a.js != nil {
		ctx, cancel := context.WithTimeout(a.ctx, a.timeout)
		defer cancel()

		// Create ephemeral consumer for the subject
		if cons, err := a.js.CreateOrUpdateConsumer(ctx, "DEFAULT", jetstream.ConsumerConfig{
			FilterSubject: event,
			AckPolicy:     jetstream.AckExplicitPolicy,
		}); err == nil {
			// Use modern pull-based consumption
			go func() {
				for {
					select {
					case <-a.ctx.Done():
						return
					default:
						msgs, err := cons.FetchNoWait(1)
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
			return true
		}
	}

	// Fallback to core NATS subscription
	if _, err := a.client.Subscribe(event, func(msg *nats.Msg) {
		f(msg)
	}); err != nil {
		log.Println(err)
		return false
	}

	return true
}

// SetTimeout sets the timeout for context operations
func (a *Adaptor) SetTimeout(timeout time.Duration) {
	a.timeout = timeout
}

// GetTimeout returns the current timeout for context operations
func (a *Adaptor) GetTimeout() time.Duration {
	return a.timeout
}

// JetStream returns the JetStream interface for advanced operations
func (a *Adaptor) JetStream() jetstream.JetStream {
	return a.js
}

// Context returns the base context for operations
func (a *Adaptor) Context() context.Context {
	return a.ctx
}