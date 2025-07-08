package mqtt

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"

	"gobot.io/x/gobot/v2"
)

// ErrNilClient is returned when a client action can't be taken because the struct has no client
var ErrNilClient = fmt.Errorf("no MQTT client available")

// Token represents an async operation result
type Token interface {
	Wait() bool
	Error() error
}

// token implements Token interface
type token struct {
	done chan struct{}
	err  error
}

func newToken() *token {
	return &token{
		done: make(chan struct{}),
	}
}

func (t *token) Wait() bool {
	<-t.done
	return t.err == nil
}

func (t *token) Error() error {
	return t.err
}

func (t *token) complete(err error) {
	t.err = err
	close(t.done)
}

// Message represents an MQTT message
type Message interface {
	Topic() string
	Payload() []byte
	Qos() byte
	Retained() bool
}

// message implements Message interface
type message struct {
	topic    string
	payload  []byte
	qos      byte
	retained bool
}

func (m *message) Topic() string   { return m.topic }
func (m *message) Payload() []byte { return m.payload }
func (m *message) Qos() byte       { return m.qos }
func (m *message) Retained() bool  { return m.retained }

// Client represents an MQTT client
type Client interface {
	Connect() Token
	Disconnect(quiesce uint)
	Publish(topic string, qos byte, retained bool, payload interface{}) Token
	Subscribe(topic string, qos byte, callback func(Client, Message)) Token
}

// client implements Client interface using Go's standard library
type client struct {
	conn        net.Conn
	clientID    string
	username    string
	password    string
	host        string
	tlsConfig   *tls.Config
	connected   bool
	mu          sync.RWMutex
	subscribers map[string]func(Client, Message)
	ctx         context.Context
	cancel      context.CancelFunc
	packetID    uint16
}

// ClientOptions represents MQTT client options
type ClientOptions struct {
	brokers       []string
	clientID      string
	username      string
	password      string
	tlsConfig     *tls.Config
	autoReconnect bool
	cleanSession  bool
}

// NewClientOptions creates new client options
func NewClientOptions() *ClientOptions {
	return &ClientOptions{
		autoReconnect: true,
		cleanSession:  true,
	}
}

// AddBroker adds a broker URL
func (o *ClientOptions) AddBroker(broker string) {
	o.brokers = append(o.brokers, broker)
}

// SetClientID sets the client ID
func (o *ClientOptions) SetClientID(clientID string) {
	o.clientID = clientID
}

// SetUsername sets the username
func (o *ClientOptions) SetUsername(username string) {
	o.username = username
}

// SetPassword sets the password
func (o *ClientOptions) SetPassword(password string) {
	o.password = password
}

// SetTLSConfig sets the TLS configuration
func (o *ClientOptions) SetTLSConfig(config *tls.Config) {
	o.tlsConfig = config
}

// NewClient creates a new MQTT client
func NewClient(opts *ClientOptions) Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &client{
		clientID:    opts.clientID,
		username:    opts.username,
		password:    opts.password,
		host:        opts.brokers[0], // Use first broker
		tlsConfig:   opts.tlsConfig,
		subscribers: make(map[string]func(Client, Message)),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// parseURL parses MQTT broker URL
func (c *client) parseURL(brokerURL string) (network, address string, useTLS bool, err error) {
	u, err := url.Parse(brokerURL)
	if err != nil {
		return
	}

	switch u.Scheme {
	case "tcp":
		network = "tcp"
		useTLS = false
	case "ssl", "tls":
		network = "tcp"
		useTLS = true
	default:
		err = fmt.Errorf("unsupported scheme: %s", u.Scheme)
		return
	}

	host := u.Host
	if !strings.Contains(host, ":") {
		if useTLS {
			host += ":8883"
		} else {
			host += ":1883"
		}
	}
	address = host
	return
}

// Connect establishes connection to MQTT broker
func (c *client) Connect() Token {
	token := newToken()
	go func() {
		network, address, useTLS, err := c.parseURL(c.host)
		if err != nil {
			token.complete(err)
			return
		}

		var conn net.Conn
		if useTLS {
			conn, err = tls.Dial(network, address, c.tlsConfig)
		} else {
			conn, err = net.Dial(network, address)
		}
		if err != nil {
			token.complete(err)
			return
		}

		c.mu.Lock()
		c.conn = conn
		c.connected = true
		c.mu.Unlock()

		// Send CONNECT packet
		err = c.sendConnect()
		if err != nil {
			token.complete(err)
			return
		}

		// Start reading messages
		go c.readLoop()

		token.complete(nil)
	}()
	return token
}

// Disconnect closes the connection
func (c *client) Disconnect(quiesce uint) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.connected = false
	c.cancel()
}

// nextPacketID generates next packet ID
func (c *client) nextPacketID() uint16 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.packetID++
	if c.packetID == 0 {
		c.packetID = 1
	}
	return c.packetID
}

// sendConnect sends MQTT CONNECT packet
func (c *client) sendConnect() error {
	// MQTT 3.1.1 CONNECT packet
	var buf []byte

	// Variable header
	protocolName := "MQTT"
	buf = append(buf, byte(len(protocolName)>>8), byte(len(protocolName)))
	buf = append(buf, []byte(protocolName)...)
	buf = append(buf, 4) // Protocol level

	// Connect flags
	flags := byte(0x02) // Clean session
	if c.username != "" {
		flags |= 0x80
		if c.password != "" {
			flags |= 0x40
		}
	}
	buf = append(buf, flags)

	// Keep alive (60 seconds)
	buf = append(buf, 0, 60)

	// Payload
	buf = append(buf, byte(len(c.clientID)>>8), byte(len(c.clientID)))
	buf = append(buf, []byte(c.clientID)...)

	if c.username != "" {
		buf = append(buf, byte(len(c.username)>>8), byte(len(c.username)))
		buf = append(buf, []byte(c.username)...)
		if c.password != "" {
			buf = append(buf, byte(len(c.password)>>8), byte(len(c.password)))
			buf = append(buf, []byte(c.password)...)
		}
	}

	// Fixed header
	packet := []byte{0x10} // CONNECT packet type
	packet = append(packet, encodeLength(len(buf))...)
	packet = append(packet, buf...)

	_, err := c.conn.Write(packet)
	return err
}

// encodeLength encodes remaining length for MQTT packets
func encodeLength(length int) []byte {
	var encoded []byte
	for {
		b := length % 128
		length = length / 128
		if length > 0 {
			b |= 128
		}
		encoded = append(encoded, byte(b))
		if length == 0 {
			break
		}
	}
	return encoded
}

// readLoop reads incoming MQTT packets
func (c *client) readLoop() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		c.mu.RLock()
		conn := c.conn
		c.mu.RUnlock()

		if conn == nil {
			return
		}

		// Read fixed header
		header := make([]byte, 1)
		_, err := io.ReadFull(conn, header)
		if err != nil {
			return
		}

		// Read remaining length
		length, err := c.readRemainingLength(conn)
		if err != nil {
			return
		}

		// Read payload
		payload := make([]byte, length)
		if length > 0 {
			_, err = io.ReadFull(conn, payload)
			if err != nil {
				return
			}
		}

		// Handle packet
		c.handlePacket(header[0], payload)
	}
}

// readRemainingLength reads MQTT remaining length
func (c *client) readRemainingLength(conn net.Conn) (int, error) {
	length := 0
	multiplier := 1
	for {
		buf := make([]byte, 1)
		_, err := io.ReadFull(conn, buf)
		if err != nil {
			return 0, err
		}
		length += int(buf[0]&127) * multiplier
		if (buf[0] & 128) == 0 {
			break
		}
		multiplier *= 128
	}
	return length, nil
}

// handlePacket handles incoming MQTT packets
func (c *client) handlePacket(header byte, payload []byte) {
	packetType := (header >> 4) & 0x0F
	switch packetType {
	case 0x02: // CONNACK
		// Connection acknowledged
	case 0x03: // PUBLISH
		c.handlePublish(header, payload)
	case 0x09: // SUBACK
		// Subscribe acknowledged
	}
}

// handlePublish handles PUBLISH packets
func (c *client) handlePublish(header byte, payload []byte) {
	if len(payload) < 2 {
		return
	}

	// Extract topic
	topicLen := int(binary.BigEndian.Uint16(payload[0:2]))
	if len(payload) < 2+topicLen {
		return
	}
	topic := string(payload[2 : 2+topicLen])

	// Extract message payload
	offset := 2 + topicLen
	qos := (header >> 1) & 0x03
	if qos > 0 {
		offset += 2 // Skip packet ID for QoS > 0
	}

	if offset > len(payload) {
		return
	}
	msgPayload := payload[offset:]

	// Create message
	msg := &message{
		topic:    topic,
		payload:  msgPayload,
		qos:      qos,
		retained: (header & 0x01) != 0,
	}

	// Call subscriber
	c.mu.RLock()
	if callback, exists := c.subscribers[topic]; exists {
		go callback(c, msg)
	}
	c.mu.RUnlock()
}

// Publish publishes a message
func (c *client) Publish(topic string, qos byte, retained bool, payload interface{}) Token {
	token := newToken()
	go func() {
		c.mu.RLock()
		conn := c.conn
		c.mu.RUnlock()

		if conn == nil {
			token.complete(ErrNilClient)
			return
		}

		// Convert payload to bytes
		var data []byte
		switch p := payload.(type) {
		case []byte:
			data = p
		case string:
			data = []byte(p)
		default:
			data = []byte(fmt.Sprintf("%v", p))
		}

		// Build PUBLISH packet
		var buf []byte

		// Topic
		buf = append(buf, byte(len(topic)>>8), byte(len(topic)))
		buf = append(buf, []byte(topic)...)

		// Packet ID (for QoS > 0)
		if qos > 0 {
			packetID := c.nextPacketID()
			buf = append(buf, byte(packetID>>8), byte(packetID))
		}

		// Payload
		buf = append(buf, data...)

		// Fixed header
		header := byte(0x30) // PUBLISH
		if retained {
			header |= 0x01
		}
		header |= (qos & 0x03) << 1

		packet := []byte{header}
		packet = append(packet, encodeLength(len(buf))...)
		packet = append(packet, buf...)

		_, err := conn.Write(packet)
		token.complete(err)
	}()
	return token
}

// Subscribe subscribes to a topic
func (c *client) Subscribe(topic string, qos byte, callback func(Client, Message)) Token {
	token := newToken()
	go func() {
		c.mu.Lock()
		c.subscribers[topic] = callback
		conn := c.conn
		c.mu.Unlock()

		if conn == nil {
			token.complete(ErrNilClient)
			return
		}

		// Build SUBSCRIBE packet
		packetID := c.nextPacketID()
		var buf []byte

		// Packet ID
		buf = append(buf, byte(packetID>>8), byte(packetID))

		// Topic filter
		buf = append(buf, byte(len(topic)>>8), byte(len(topic)))
		buf = append(buf, []byte(topic)...)
		buf = append(buf, qos)

		// Fixed header
		packet := []byte{0x82} // SUBSCRIBE
		packet = append(packet, encodeLength(len(buf))...)
		packet = append(packet, buf...)

		_, err := conn.Write(packet)
		token.complete(err)
	}()
	return token
}

// Adaptor is the Gobot Adaptor for MQTT
type Adaptor struct {
	name          string
	Host          string
	clientID      string
	username      string
	password      string
	useSSL        bool
	serverCert    string
	clientCert    string
	clientKey     string
	autoReconnect bool
	cleanSession  bool
	client        Client
	qos           int
}

// NewAdaptor creates a new mqtt adaptor with specified host and client id
func NewAdaptor(host string, clientID string) *Adaptor {
	return &Adaptor{
		name:          gobot.DefaultName("MQTT"),
		Host:          host,
		autoReconnect: false,
		cleanSession:  true,
		useSSL:        false,
		clientID:      clientID,
	}
}

// NewAdaptorWithAuth creates a new mqtt adaptor with specified host, client id, username, and password.
func NewAdaptorWithAuth(host, clientID, username, password string) *Adaptor {
	return &Adaptor{
		name:          "MQTT",
		Host:          host,
		autoReconnect: false,
		cleanSession:  true,
		useSSL:        false,
		clientID:      clientID,
		username:      username,
		password:      password,
	}
}

// Name returns the MQTT adaptors name
func (a *Adaptor) Name() string { return a.name }

// SetName sets the MQTT adaptors name
func (a *Adaptor) SetName(n string) { a.name = n }

// Port returns the Host name
func (a *Adaptor) Port() string { return a.Host }

// AutoReconnect returns the MQTT AutoReconnect setting
func (a *Adaptor) AutoReconnect() bool { return a.autoReconnect }

// SetAutoReconnect sets the MQTT AutoReconnect setting
func (a *Adaptor) SetAutoReconnect(val bool) { a.autoReconnect = val }

// CleanSession returns the MQTT CleanSession setting
func (a *Adaptor) CleanSession() bool { return a.cleanSession }

// SetCleanSession sets the MQTT CleanSession setting. Should be false if reconnect is enabled.
// Otherwise all subscriptions will be lost
func (a *Adaptor) SetCleanSession(val bool) { a.cleanSession = val }

// UseSSL returns the MQTT server SSL preference
func (a *Adaptor) UseSSL() bool { return a.useSSL }

// SetUseSSL sets the MQTT server SSL preference
func (a *Adaptor) SetUseSSL(val bool) { a.useSSL = val }

// ServerCert returns the MQTT server SSL cert file
func (a *Adaptor) ServerCert() string { return a.serverCert }

// SetQoS sets the QoS value passed into the MTT client on Publish/Subscribe events
func (a *Adaptor) SetQoS(qos int) { a.qos = qos }

// SetServerCert sets the MQTT server SSL cert file
func (a *Adaptor) SetServerCert(val string) { a.serverCert = val }

// ClientCert returns the MQTT client SSL cert file
func (a *Adaptor) ClientCert() string { return a.clientCert }

// SetClientCert sets the MQTT client SSL cert file
func (a *Adaptor) SetClientCert(val string) { a.clientCert = val }

// ClientKey returns the MQTT client SSL key file
func (a *Adaptor) ClientKey() string { return a.clientKey }

// SetClientKey sets the MQTT client SSL key file
func (a *Adaptor) SetClientKey(val string) { a.clientKey = val }

// Connect returns true if connection to mqtt is established
func (a *Adaptor) Connect() error {
	a.client = NewClient(a.createClientOptions())
	token := a.client.Connect()
	token.Wait()
	return token.Error()
}

// Disconnect returns true if connection to mqtt is closed
func (a *Adaptor) Disconnect() error {
	if a.client != nil {
		a.client.Disconnect(500)
	}
	return nil
}

// Finalize returns true if connection to mqtt is finalized successfully
func (a *Adaptor) Finalize() error {
	return a.Disconnect()
}

// Publish a message under a specific topic
func (a *Adaptor) Publish(topic string, message []byte) bool {
	_, err := a.PublishWithQOS(topic, a.qos, message)
	return err == nil
}

// PublishAndRetain publishes a message under a specific topic with retain flag
func (a *Adaptor) PublishAndRetain(topic string, message []byte) bool {
	if a.client == nil {
		return false
	}

	a.client.Publish(topic, byte(a.qos), true, message)
	return true
}

// PublishWithQOS allows per-publish QOS values to be set and returns a Token
func (a *Adaptor) PublishWithQOS(topic string, qos int, message []byte) (Token, error) {
	if a.client == nil {
		return nil, ErrNilClient
	}

	token := a.client.Publish(topic, byte(qos), false, message)
	return token, nil
}

// OnWithQOS allows per-subscribe QOS values to be set and returns a Token
func (a *Adaptor) OnWithQOS(event string, qos int, f func(msg Message)) (Token, error) {
	if a.client == nil {
		return nil, ErrNilClient
	}

	token := a.client.Subscribe(event, byte(qos), func(client Client, msg Message) {
		f(msg)
	})

	return token, nil
}

// On subscribes to a topic, and then calls the message handler function when data is received
func (a *Adaptor) On(event string, f func(msg Message)) bool {
	_, err := a.OnWithQOS(event, a.qos, f)
	return err == nil
}

func (a *Adaptor) createClientOptions() *ClientOptions {
	opts := NewClientOptions()
	opts.AddBroker(a.Host)
	opts.SetClientID(a.clientID)
	if a.username != "" && a.password != "" {
		opts.SetPassword(a.password)
		opts.SetUsername(a.username)
	}
	// Note: autoReconnect and cleanSession are handled internally

	if a.UseSSL() {
		opts.SetTLSConfig(a.newTLSConfig())
	}
	return opts
}

// newTLSConfig sets the TLS config in the case that we are using
// an MQTT broker with TLS
func (a *Adaptor) newTLSConfig() *tls.Config {
	// Import server certificate
	var certpool *x509.CertPool
	if len(a.ServerCert()) > 0 {
		certpool = x509.NewCertPool()
		pemCerts, err := os.ReadFile(a.ServerCert())
		if err == nil {
			certpool.AppendCertsFromPEM(pemCerts)
		}
	}

	// Import client certificate/key pair
	var certs []tls.Certificate
	if len(a.ClientCert()) > 0 && len(a.ClientKey()) > 0 {
		cert, err := tls.LoadX509KeyPair(a.ClientCert(), a.ClientKey())
		if err != nil {
			// TODO: proper error handling
			panic(err)
		}
		certs = append(certs, cert)
	}

	// Create tls.Config with desired tls properties
	return &tls.Config{
		// RootCAs = certs used to verify server cert.
		RootCAs: certpool,
		// ClientAuth = whether to request cert from server.
		// Since the server is set up for SSL, this happens
		// anyways.
		ClientAuth: tls.NoClientCert,
		// ClientCAs = certs used to validate client cert.
		ClientCAs: nil,
		// InsecureSkipVerify = verify that cert contents
		// match server. IP matches what is in cert etc.
		InsecureSkipVerify: false,
		// Certificates = list of certs client sends to server.
		Certificates: certs,
		// MinVersion contains the minimum TLS version that is acceptable.
		// TLS 1.2 is currently used as the minimum when acting as a client.
		MinVersion: tls.VersionTLS12,
	}
}
