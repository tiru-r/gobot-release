package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"gobot.io/x/gobot/v2"
)

// API represents an API server
type API struct {
	manager  *gobot.Manager
	router   *http.ServeMux
	Host     string
	Port     string
	Cert     string
	Key      string
	handlers []func(http.ResponseWriter, *http.Request)
	start    func(*API)
}

// NewAPI returns a new api instance
func NewAPI(m *gobot.Manager) *API {
	return &API{
		manager: m,
		router:  http.NewServeMux(),
		Port:    "3000",
		start: func(a *API) {
			log.Println("Initializing API on " + a.Host + ":" + a.Port + "...")
			http.Handle("/", a)
			server := &http.Server{
				Addr:              a.Host + ":" + a.Port,
				ReadHeaderTimeout: 30 * time.Second,
			}

			go func() {
				if a.Cert != "" && a.Key != "" {
					if err := server.ListenAndServeTLS(a.Cert, a.Key); err != nil {
						log.Printf("Server error: %v", err)
					}
				} else {
					log.Println("WARNING: API using insecure connection. " +
						"We recommend using an SSL certificate with Gobot.")
					if err := server.ListenAndServe(); err != nil {
						log.Printf("Server error: %v", err)
					}
				}
			}()
		},
	}
}

// ServeHTTP calls api handlers and then serves request using api router
func (a *API) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	for _, handler := range a.handlers {
		rec := httptest.NewRecorder()
		handler(rec, req)
		for k, v := range rec.Header() {
			res.Header()[k] = v
		}
		if rec.Code == http.StatusUnauthorized {
			http.Error(res, "Not Authorized", http.StatusUnauthorized)
			return
		}
	}
	a.router.ServeHTTP(res, req)
}

// Post wraps api router Post call
func (a *API) Post(path string, f func(http.ResponseWriter, *http.Request)) {
	a.router.HandleFunc("POST "+path, f)
}

// Put wraps api router Put call
func (a *API) Put(path string, f func(http.ResponseWriter, *http.Request)) {
	a.router.HandleFunc("PUT "+path, f)
}

// Delete wraps api router Delete call
func (a *API) Delete(path string, f func(http.ResponseWriter, *http.Request)) {
	a.router.HandleFunc("DELETE "+path, f)
}

// Options wraps api router Options call
func (a *API) Options(path string, f func(http.ResponseWriter, *http.Request)) {
	a.router.HandleFunc("OPTIONS "+path, f)
}

// Get wraps api router Get call
func (a *API) Get(path string, f func(http.ResponseWriter, *http.Request)) {
	a.router.HandleFunc("GET "+path, f)
}

// Head wraps api router Head call
func (a *API) Head(path string, f func(http.ResponseWriter, *http.Request)) {
	a.router.HandleFunc("HEAD "+path, f)
}

// AddHandler appends handler to api handlers
func (a *API) AddHandler(f func(http.ResponseWriter, *http.Request)) {
	a.handlers = append(a.handlers, f)
}

// Start initializes the api by setting up C3PIO API routes and basic web interface.
func (a *API) Start() {
	a.AddWebRoutes()

	a.start(a)
}

// StartWithoutDefaults initializes the api without setting up the default routes.
// Good for custom web interfaces.
func (a *API) StartWithoutDefaults() {
	a.start(a)
}

// AddC3PIORoutes adds all of the standard C3PIO routes to the API.
// For more information, please see:
// http://cppp.io/
func (a *API) AddC3PIORoutes() {
	mcpCommandRoute := "/api/commands/{command}"
	robotDeviceCommandRoute := "/api/robots/{robot}/devices/{device}/commands/{command}"
	robotCommandRoute := "/api/robots/{robot}/commands/{command}"

	a.Get("/api/commands", a.mcpCommands)
	a.Get(mcpCommandRoute, a.executeMcpCommand)
	a.Post(mcpCommandRoute, a.executeMcpCommand)
	a.Get("/api/robots", a.robots)
	a.Get("/api/robots/{robot}", a.robot)
	a.Get("/api/robots/{robot}/commands", a.robotCommands)
	a.Get(robotCommandRoute, a.executeRobotCommand)
	a.Post(robotCommandRoute, a.executeRobotCommand)
	a.Get("/api/robots/{robot}/devices", a.robotDevices)
	a.Get("/api/robots/{robot}/devices/{device}", a.robotDevice)
	a.Get("/api/robots/{robot}/devices/{device}/events/{event}", a.robotDeviceEvent)
	a.Get("/api/robots/{robot}/devices/{device}/commands", a.robotDeviceCommands)
	a.Get(robotDeviceCommandRoute, a.executeRobotDeviceCommand)
	a.Post(robotDeviceCommandRoute, a.executeRobotDeviceCommand)
	a.Get("/api/robots/{robot}/connections", a.robotConnections)
	a.Get("/api/robots/{robot}/connections/{connection}", a.robotConnection)
	a.Get("/api/", a.mcp)
}

// AddWebRoutes adds basic web routes for API documentation and status.
// This provides a simple web interface for API discovery.
func (a *API) AddWebRoutes() {
	a.AddC3PIORoutes()

	a.Get("/", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/html; charset=utf-8")
		html := `<!DOCTYPE html>
<html>
<head>
    <title>Gobot API</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .api-info { background: #f5f5f5; padding: 20px; border-radius: 5px; }
        a { color: #0066cc; }
    </style>
</head>
<body>
    <h1>Gobot API Server</h1>
    <div class="api-info">
        <h2>Available Endpoints:</h2>
        <ul>
            <li><a href="/api/">/api/</a> - API Root</li>
            <li><a href="/api/robots">/api/robots</a> - List all robots</li>
            <li><a href="/api/commands">/api/commands</a> - List all commands</li>
        </ul>
        <p>For a modern web interface, we recommend using external tools like:</p>
        <ul>
            <li><strong>ThingsBoard</strong> - Comprehensive IoT dashboard</li>
            <li><strong>Node-RED</strong> - Visual flow programming</li>
            <li><strong>Custom React/Vue.js apps</strong> - Connect to this API</li>
        </ul>
    </div>
</body>
</html>`
		if _, err := res.Write([]byte(html)); err != nil {
			log.Printf("Error: %v", err)
		}
	})
}

// mcp returns MCP route handler.
// Writes JSON with gobot representation
func (a *API) mcp(res http.ResponseWriter, req *http.Request) {
	a.writeJSON(map[string]interface{}{"MCP": gobot.NewJSONManager(a.manager)}, res)
}

// mcpCommands returns commands route handler.
// Writes JSON with global commands representation
func (a *API) mcpCommands(res http.ResponseWriter, req *http.Request) {
	a.writeJSON(map[string]interface{}{"commands": gobot.NewJSONManager(a.manager).Commands}, res)
}

// robots returns route handler.
// Writes JSON with robots representation
func (a *API) robots(res http.ResponseWriter, req *http.Request) {
	jsonRobots := []*gobot.JSONRobot{}
	a.manager.Robots().Each(func(r *gobot.Robot) {
		jsonRobots = append(jsonRobots, gobot.NewJSONRobot(r))
	})
	a.writeJSON(map[string]interface{}{"robots": jsonRobots}, res)
}

// robot returns route handler.
// Writes JSON with robot representation
func (a *API) robot(res http.ResponseWriter, req *http.Request) {
	if robot, err := a.jsonRobotFor(req.PathValue("robot")); err != nil {
		a.writeJSON(map[string]interface{}{"error": err.Error()}, res)
	} else {
		a.writeJSON(map[string]interface{}{"robot": robot}, res)
	}
}

// robotCommands returns commands route handler
// Writes JSON with robot commands representation
func (a *API) robotCommands(res http.ResponseWriter, req *http.Request) {
	if robot, err := a.jsonRobotFor(req.PathValue("robot")); err != nil {
		a.writeJSON(map[string]interface{}{"error": err.Error()}, res)
	} else {
		a.writeJSON(map[string]interface{}{"commands": robot.Commands}, res)
	}
}

// robotDevices returns devices route handler.
// Writes JSON with robot devices representation
func (a *API) robotDevices(res http.ResponseWriter, req *http.Request) {
	if robot := a.manager.Robot(req.PathValue("robot")); robot != nil {
		jsonDevices := []*gobot.JSONDevice{}
		robot.Devices().Each(func(d gobot.Device) {
			jsonDevices = append(jsonDevices, gobot.NewJSONDevice(d))
		})
		a.writeJSON(map[string]interface{}{"devices": jsonDevices}, res)
	} else {
		a.writeJSON(map[string]interface{}{"error": "No Robot found with the name " + req.PathValue("robot")}, res)
	}
}

// robotDevice returns device route handler.
// Writes JSON with robot device representation
func (a *API) robotDevice(res http.ResponseWriter, req *http.Request) {
	if device, err := a.jsonDeviceFor(req.PathValue("robot"), req.PathValue("device")); err != nil {
		a.writeJSON(map[string]interface{}{"error": err.Error()}, res)
	} else {
		a.writeJSON(map[string]interface{}{"device": device}, res)
	}
}

func (a *API) robotDeviceEvent(res http.ResponseWriter, req *http.Request) {
	f, _ := res.(http.Flusher)

	dataChan := make(chan string)

	res.Header().Set("Content-Type", "text/event-stream")
	res.Header().Set("Cache-Control", "no-cache")
	res.Header().Set("Connection", "keep-alive")

	device := a.manager.Robot(req.PathValue("robot")).
		Device(req.PathValue("device"))

	//nolint:forcetypeassert // no error return value, so there is no better way
	if event := a.manager.Robot(req.PathValue("robot")).
		Device(req.PathValue("device")).(gobot.Eventer).
		Event(req.PathValue("event")); len(event) > 0 {
		//nolint:forcetypeassert // no error return value, so there is no better way
		if err := device.(gobot.Eventer).On(event, func(data interface{}) {
			d, _ := json.Marshal(data)
			dataChan <- string(d)
		}); err != nil {
			log.Printf("Error: %v", err)
		}

		for {
			select {
			case data := <-dataChan:
				fmt.Fprintf(res, "data: %v\n\n", data)
				f.Flush()
			case <-req.Context().Done():
				log.Println("Closing connection")
				return
			}
		}
	} else {
		a.writeJSON(map[string]interface{}{
			"error": "No Event found with the name " + req.PathValue("event"),
		}, res)
	}
}

// robotDeviceCommands returns device commands route handler
// writes JSON with robot device commands representation
func (a *API) robotDeviceCommands(res http.ResponseWriter, req *http.Request) {
	if device, err := a.jsonDeviceFor(req.PathValue("robot"), req.PathValue("device")); err != nil {
		a.writeJSON(map[string]interface{}{"error": err.Error()}, res)
	} else {
		a.writeJSON(map[string]interface{}{"commands": device.Commands}, res)
	}
}

// robotConnections returns connections route handler
// writes JSON with robot connections representation
func (a *API) robotConnections(res http.ResponseWriter, req *http.Request) {
	jsonConnections := []*gobot.JSONConnection{}
	if robot := a.manager.Robot(req.PathValue("robot")); robot != nil {
		robot.Connections().Each(func(c gobot.Connection) {
			jsonConnections = append(jsonConnections, gobot.NewJSONConnection(c))
		})
		a.writeJSON(map[string]interface{}{"connections": jsonConnections}, res)
	} else {
		a.writeJSON(map[string]interface{}{"error": "No Robot found with the name " + req.PathValue("robot")}, res)
	}
}

// robotConnection returns connection route handler
// writes JSON with robot connection representation
func (a *API) robotConnection(res http.ResponseWriter, req *http.Request) {
	if conn, err := a.jsonConnectionFor(req.PathValue("robot"), req.PathValue("connection")); err != nil {
		a.writeJSON(map[string]interface{}{"error": err.Error()}, res)
	} else {
		a.writeJSON(map[string]interface{}{"connection": conn}, res)
	}
}

// executeMcpCommand calls a global command associated to requested route
func (a *API) executeMcpCommand(res http.ResponseWriter, req *http.Request) {
	a.executeCommand(a.manager.Command(req.PathValue("command")),
		res,
		req,
	)
}

// executeRobotDeviceCommand calls a device command associated to requested route
func (a *API) executeRobotDeviceCommand(res http.ResponseWriter, req *http.Request) {
	if _, err := a.jsonDeviceFor(req.PathValue("robot"),
		req.PathValue("device")); err != nil {
		a.writeJSON(map[string]interface{}{"error": err.Error()}, res)
	} else {
		a.executeCommand(
			//nolint:forcetypeassert // no error return value, so there is no better way
			a.manager.Robot(req.PathValue("robot")).
				Device(req.PathValue("device")).(gobot.Commander).
				Command(req.PathValue("command")),
			res,
			req,
		)
	}
}

// executeRobotCommand calls a robot command associated to requested route
func (a *API) executeRobotCommand(res http.ResponseWriter, req *http.Request) {
	if _, err := a.jsonRobotFor(req.PathValue("robot")); err != nil {
		a.writeJSON(map[string]interface{}{"error": err.Error()}, res)
	} else {
		a.executeCommand(
			a.manager.Robot(req.PathValue("robot")).
				Command(req.PathValue("command")),
			res,
			req,
		)
	}
}

// executeCommand writes JSON response with `f` returned value.
func (a *API) executeCommand(f func(map[string]interface{}) interface{},
	res http.ResponseWriter,
	req *http.Request,
) {
	body := make(map[string]interface{})
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		log.Printf("Error: %v", err)
	}

	if f != nil {
		a.writeJSON(map[string]interface{}{"result": f(body)}, res)
	} else {
		a.writeJSON(map[string]interface{}{"error": "Unknown Command"}, res)
	}
}

// writeJSON writes `j` as JSON in response
func (a *API) writeJSON(j interface{}, res http.ResponseWriter) {
	data, err := json.Marshal(j)
	if err != nil {
		log.Printf("Error: %v", err)
	}
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := res.Write(data); err != nil {
		log.Printf("Error: %v", err)
	}
}

// Debug add handler to api that prints each request
func (a *API) Debug() {
	a.AddHandler(func(res http.ResponseWriter, req *http.Request) {
		log.Println(req)
	})
}

func (a *API) jsonRobotFor(name string) (*gobot.JSONRobot, error) {
	if robot := a.manager.Robot(name); robot != nil {
		return gobot.NewJSONRobot(robot), nil
	}
	return nil, fmt.Errorf("No Robot found with the name %s", name)
}

func (a *API) jsonDeviceFor(robot string, name string) (*gobot.JSONDevice, error) {
	if device := a.manager.Robot(robot).Device(name); device != nil {
		return gobot.NewJSONDevice(device), nil
	}

	return nil, fmt.Errorf("No Device found with the name %s", name)
}

func (a *API) jsonConnectionFor(robot string, name string) (*gobot.JSONConnection, error) {
	if connection := a.manager.Robot(robot).Connection(name); connection != nil {
		return gobot.NewJSONConnection(connection), nil
	}

	return nil, fmt.Errorf("No Connection found with the name %s", name)
}
