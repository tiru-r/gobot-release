package particle

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gobot.io/x/gobot/v2"
)

// Adaptor is the Gobot Adaptor for Particle
type Adaptor struct {
	name        string
	DeviceID    string
	AccessToken string
	APIServer   string
	servoPins   map[string]bool
	client      *http.Client
	gobot.Eventer
}

// Event is an event emitted by the Particle cloud
type Event struct {
	Name  string
	Data  string
	Error error
}

// SSEEvent represents a Server-Sent Event
type SSEEvent struct {
	event string
	data  string
}

// Event returns the event type
func (e *SSEEvent) Event() string {
	return e.event
}

// Data returns the event data
func (e *SSEEvent) Data() string {
	return e.data
}

var eventSource = func(url string) (chan *SSEEvent, chan error, error) {
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	eventChan := make(chan *SSEEvent)
	errorChan := make(chan error)

	go func() {
		defer resp.Body.Close()
		defer close(eventChan)
		defer close(errorChan)

		scanner := bufio.NewScanner(resp.Body)
		var currentEvent *SSEEvent

		for scanner.Scan() {
			line := scanner.Text()

			if line == "" {
				// Empty line signals end of event
				if currentEvent != nil && currentEvent.event != "" {
					eventChan <- currentEvent
				}
				currentEvent = nil
				continue
			}

			if currentEvent == nil {
				currentEvent = &SSEEvent{}
			}

			if strings.HasPrefix(line, "event:") {
				currentEvent.event = strings.TrimSpace(line[6:])
			} else if strings.HasPrefix(line, "data:") {
				data := strings.TrimSpace(line[5:])
				if currentEvent.data != "" {
					currentEvent.data += "\n" + data
				} else {
					currentEvent.data = data
				}
			}
			// Ignore other fields like id:, retry:, etc.
		}

		if err := scanner.Err(); err != nil {
			errorChan <- err
		}
	}()

	return eventChan, errorChan, nil
}

// NewAdaptor creates new Photon adaptor with deviceId and accessToken
// using api.particle.io server as default
func NewAdaptor(deviceID string, accessToken string) *Adaptor {
	return &Adaptor{
		name:        gobot.DefaultName("Particle"),
		DeviceID:    deviceID,
		AccessToken: accessToken,
		servoPins:   make(map[string]bool),
		APIServer:   "https://api.particle.io",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		Eventer:     gobot.NewEventer(),
	}
}

// Name returns the Adaptor name
func (s *Adaptor) Name() string { return s.name }

// SetName sets the Adaptor name
func (s *Adaptor) SetName(n string) { s.name = n }

// Connect returns true if connection to Particle Photon or Electron is successful
func (s *Adaptor) Connect() error { return nil }

// Finalize returns true if connection to Particle Photon or Electron is finalized successfully
func (s *Adaptor) Finalize() error { return nil }

// AnalogRead reads analog ping value using Particle cloud api
func (s *Adaptor) AnalogRead(pin string) (int, error) {
	params := url.Values{
		"params":       {pin},
		"access_token": {s.AccessToken},
	}

	url := fmt.Sprintf("%v/analogread", s.deviceURL())

	resp, err := s.request("POST", url, params)
	if err == nil {
		//nolint:forcetypeassert // ok here
		return int(resp["return_value"].(float64)), nil
	}

	return 0, err
}

// PwmWrite writes in pin using analog write api
func (s *Adaptor) PwmWrite(pin string, level byte) error {
	return s.AnalogWrite(pin, level)
}

// AnalogWrite writes analog pin with specified level using Particle cloud api
func (s *Adaptor) AnalogWrite(pin string, level byte) error {
	params := url.Values{
		"params":       {fmt.Sprintf("%v,%v", pin, level)},
		"access_token": {s.AccessToken},
	}
	url := fmt.Sprintf("%v/analogwrite", s.deviceURL())
	_, err := s.request("POST", url, params)
	return err
}

// DigitalWrite writes to a digital pin using Particle cloud api
func (s *Adaptor) DigitalWrite(pin string, level byte) error {
	params := url.Values{
		"params":       {fmt.Sprintf("%v,%v", pin, s.pinLevel(level))},
		"access_token": {s.AccessToken},
	}
	url := fmt.Sprintf("%v/digitalwrite", s.deviceURL())
	_, err := s.request("POST", url, params)
	return err
}

// DigitalRead reads from digital pin using Particle cloud api
func (s *Adaptor) DigitalRead(pin string) (int, error) {
	params := url.Values{
		"params":       {pin},
		"access_token": {s.AccessToken},
	}
	url := fmt.Sprintf("%v/digitalread", s.deviceURL())
	resp, err := s.request("POST", url, params)
	if err != nil {
		return -1, err
	}

	//nolint:forcetypeassert // ok here
	return int(resp["return_value"].(float64)), nil
}

// ServoWrite writes the 0-180 degree angle to the specified pin.
// To use it requires installing the "tinker-servo" sketch on your
// Particle device. not just the default "tinker".
func (s *Adaptor) ServoWrite(pin string, angle byte) error {
	if _, present := s.servoPins[pin]; !present {
		if err := s.servoPinOpen(pin); err != nil {
			return err
		}
	}

	params := url.Values{
		"params":       {fmt.Sprintf("%v,%v", pin, angle)},
		"access_token": {s.AccessToken},
	}
	url := fmt.Sprintf("%v/servoSet", s.deviceURL())
	_, err := s.request("POST", url, params)
	return err
}

// EventStream returns a gobot.Event based on the following params:
//
// * source - "all"/"devices"/"device" (More info at: http://docs.particle.io/api/#reading-data-from-a-core-events)
// * name  - Event name to subscribe for, leave blank to subscribe to all events.
//
// A new event is emitted as a particle.Event struct
func (s *Adaptor) EventStream(source string, name string) (*gobot.Event, error) {
	var url string

	switch source {
	case "all":
		url = fmt.Sprintf("%s/v1/events/%s?access_token=%s", s.APIServer, name, s.AccessToken)
	case "devices":
		url = fmt.Sprintf("%s/v1/devices/events/%s?access_token=%s", s.APIServer, name, s.AccessToken)
	case "device":
		url = fmt.Sprintf("%s/events/%s?access_token=%s", s.deviceURL(), name, s.AccessToken)
	default:
		return nil, errors.New("source param should be: all, devices or device")
	}

	events, errs, err := eventSource(url)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case ev := <-events:
				if ev != nil && ev.Event() != "" && ev.Data() != "" {
					s.Publish(ev.Event(), ev.Data())
				}
			case err := <-errs:
				if err != nil {
					// Handle error - could publish as event or log
					continue
				}
			}
		}
	}()

	return nil, nil //nolint:nilnil // seems ok here
}

// Variable returns a core variable value as a string
func (s *Adaptor) Variable(name string) (string, error) {
	url := fmt.Sprintf("%v/%s?access_token=%s", s.deviceURL(), name, s.AccessToken)
	resp, err := s.request("GET", url, nil)
	if err != nil {
		return "", err
	}

	var result string
	switch val := resp["result"].(type) {
	case bool:
		result = strconv.FormatBool(val)
	case float64:
		result = strconv.FormatFloat(val, 'f', -1, 64)
	case string:
		result = val
	}

	return result, nil
}

// Function executes a core function and
// returns value from request.
// Takes a String as the only argument and returns an Int.
// If function is not defined in core, it will time out
func (s *Adaptor) Function(name string, args string) (int, error) {
	params := url.Values{
		"args":         {args},
		"access_token": {s.AccessToken},
	}

	url := fmt.Sprintf("%s/%s", s.deviceURL(), name)
	resp, err := s.request("POST", url, params)
	if err != nil {
		return -1, err
	}

	//nolint:forcetypeassert // ok here
	return int(resp["return_value"].(float64)), nil
}

// setAPIServer sets Particle cloud api server, this can be used to change from default api.spark.io
func (s *Adaptor) setAPIServer(server string) {
	s.APIServer = server
}

// deviceURL constructs device url to make requests from Particle cloud api
func (s *Adaptor) deviceURL() string {
	if len(s.APIServer) == 0 {
		s.setAPIServer("https://api.particle.io")
	}
	return fmt.Sprintf("%v/v1/devices/%v", s.APIServer, s.DeviceID)
}

// pinLevel converts byte level to string expected in api
func (s *Adaptor) pinLevel(level byte) string {
	if level == 1 {
		return "HIGH"
	}
	return "LOW"
}

// request makes request to Particle cloud server, return err != nil if there is
// any issue with the request.
func (s *Adaptor) request(method string, url string, params url.Values) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	var req *http.Request
	var err error
	
	if method == "POST" {
		req, err = http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(params.Encode()))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else if method == "GET" {
		req, err = http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}
	
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	if err := json.Unmarshal(buf, &m); err != nil {
		return m, err
	}

	if resp.Status != "200 OK" {
		return m, fmt.Errorf("%v: error communicating to the Particle cloud", resp.Status)
	}

	if _, ok := m["error"]; ok {
		//nolint:forcetypeassert // ok here
		return m, errors.New(m["error"].(string))
	}

	return m, nil
}

func (s *Adaptor) servoPinOpen(pin string) error {
	params := url.Values{
		"params":       {pin},
		"access_token": {s.AccessToken},
	}
	url := fmt.Sprintf("%v/servoOpen", s.deviceURL())
	_, err := s.request("POST", url, params)
	if err != nil {
		return err
	}
	s.servoPins[pin] = true
	return nil
}
