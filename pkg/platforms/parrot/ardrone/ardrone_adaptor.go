package ardrone

import (
	"fmt"
	"net"
	"sync"
	"time"

	"gobot.io/x/gobot/v2"
)

// Config represents drone configuration
type Config struct {
	IP   string
	Port string
}

// DefaultConfig returns default drone configuration
func DefaultConfig() Config {
	return Config{
		IP:   "192.168.1.1",
		Port: "5556",
	}
}

// DroneClient represents a modern AR.Drone client using Go's standard library
type DroneClient struct {
	config Config
	conn   net.Conn
	mu     sync.Mutex
	seq    int
}

// Connect establishes connection to the drone
func Connect(config Config) (*DroneClient, error) {
	addr := fmt.Sprintf("%s:%s", config.IP, config.Port)
	conn, err := net.DialTimeout("udp", addr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to drone: %w", err)
	}

	return &DroneClient{
		config: config,
		conn:   conn,
	}, nil
}

// sendCommand sends AT command to the drone
func (c *DroneClient) sendCommand(cmd string, args ...interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.seq++
	command := fmt.Sprintf("AT*%s=%d", cmd, c.seq)
	if len(args) > 0 {
		for _, arg := range args {
			command += fmt.Sprintf(",%v", arg)
		}
	}
	command += "\r"

	_, err := c.conn.Write([]byte(command))
	return err
}

// drone defines expected drone behaviour
type drone interface {
	Takeoff() bool
	Land()
	Up(n float64)
	Down(n float64)
	Left(n float64)
	Right(n float64)
	Forward(n float64)
	Backward(n float64)
	Clockwise(n float64)
	Counterclockwise(n float64)
	Hover()
}

// Implement drone interface for DroneClient
func (c *DroneClient) Takeoff() bool {
	err := c.sendCommand("REF", "290718208")
	return err == nil
}

func (c *DroneClient) Land() {
	c.sendCommand("REF", "290717696")
}

func (c *DroneClient) Up(speed float64) {
	c.sendCommand("PCMD", "1", "0", "0", fmt.Sprintf("%d", int(speed*1000)), "0")
}

func (c *DroneClient) Down(speed float64) {
	c.sendCommand("PCMD", "1", "0", "0", fmt.Sprintf("%d", -int(speed*1000)), "0")
}

func (c *DroneClient) Left(speed float64) {
	c.sendCommand("PCMD", "1", fmt.Sprintf("%d", -int(speed*1000)), "0", "0", "0")
}

func (c *DroneClient) Right(speed float64) {
	c.sendCommand("PCMD", "1", fmt.Sprintf("%d", int(speed*1000)), "0", "0", "0")
}

func (c *DroneClient) Forward(speed float64) {
	c.sendCommand("PCMD", "1", "0", fmt.Sprintf("%d", -int(speed*1000)), "0", "0")
}

func (c *DroneClient) Backward(speed float64) {
	c.sendCommand("PCMD", "1", "0", fmt.Sprintf("%d", int(speed*1000)), "0", "0")
}

func (c *DroneClient) Clockwise(speed float64) {
	c.sendCommand("PCMD", "1", "0", "0", "0", fmt.Sprintf("%d", int(speed*1000)))
}

func (c *DroneClient) Counterclockwise(speed float64) {
	c.sendCommand("PCMD", "1", "0", "0", "0", fmt.Sprintf("%d", -int(speed*1000)))
}

func (c *DroneClient) Hover() {
	c.sendCommand("PCMD", "1", "0", "0", "0", "0")
}

// Adaptor is gobot.Adaptor representation for the Ardrone
type Adaptor struct {
	name    string
	drone   drone
	config  Config
	connect func(*Adaptor) (drone, error)
}

// NewAdaptor returns a new ardrone.Adaptor and optionally accepts:
//
//	string: The ardrones IP Address
func NewAdaptor(v ...string) *Adaptor {
	a := &Adaptor{
		name: gobot.DefaultName("ARDrone"),
		connect: func(a *Adaptor) (drone, error) {
			return Connect(a.config)
		},
	}

	a.config = DefaultConfig()
	if len(v) > 0 {
		a.config.IP = v[0]
	}

	return a
}

// Name returns the Adaptor Name
func (a *Adaptor) Name() string { return a.name }

// SetName sets the Adaptor Name
func (a *Adaptor) SetName(n string) { a.name = n }

// Connect establishes a connection to the ardrone
func (a *Adaptor) Connect() error {
	d, err := a.connect(a)
	if err != nil {
		return err
	}
	a.drone = d
	return nil
}

// Finalize terminates the connection to the ardrone
func (a *Adaptor) Finalize() error { return nil }
