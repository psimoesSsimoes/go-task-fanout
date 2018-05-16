package api

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"gitlab.com/mandalore/go-app/app"
	log "gitlab.com/vredens/go-logger"
)

// LogHandler ...
type LogHandler func(message string, fields map[string]interface{})

// WebAPIOption is the functional parameter type used to configure WebAPI.
type WebAPIOption func(api *WebAPI)

// WithWebAPILogger returns an option to set the WebAPI logger.
func WithWebAPILogger(l log.Logger) WebAPIOption {
	return func(api *WebAPI) {
		api.log = l
	}
}

// WithWebAPILogHandler allows seting a log handler for doing request logging instead of the default WebAPI logger.
func WithWebAPILogHandler(lh LogHandler) WebAPIOption {
	return func(api *WebAPI) {
		api.lh = lh
	}
}

// WebAPI ...
type WebAPI struct {
	EchoServer      *echo.Echo
	log             log.Logger
	control         chan int
	address         string
	running         bool
	lh              LogHandler
	fieldNamesToLog []string
}

// NewWebAPI ...
func NewWebAPI(address string, opts ...WebAPIOption) *WebAPI {
	api := &WebAPI{
		address: address,
		running: false,
		log:     log.Spawn(log.WithFields(map[string]interface{}{"component": "web"})),
	}

	api.lh = api.log.InfoData

	for _, opt := range opts {
		opt(api)
	}

	// Setup Echo server
	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = api.webErrorHandler
	e.Use(api.webLogger)
	e.Use(middleware.Recover())

	api.EchoServer = e

	api.log.Infof("configuring webserver [address:%s]", api.address)

	return api
}

// RegisterAdminRoutes registers preset handlers for <prefix>/admin/shutdown
func (api *WebAPI) RegisterAdminRoutes(prefix string) {
	api.RegisterRoute("POST", prefix+"/admin/shutdown", api.handleShutdown)
}

// RegisterHealthRoutes registers preset handlers for <prefix>/health/stats.
func (api *WebAPI) RegisterHealthRoutes(prefix string) {
	api.RegisterRoute("GET", prefix+"/health/stats", api.handleGetApplicationStats)
}

// RegisterDebugRoutes registers preset handlers for <prefix>/debug/profile/cpu and <prefix>/debug/profile/mem
func (api *WebAPI) RegisterDebugRoutes(prefix string) {
	api.RegisterRoute("GET", prefix+"/debug/profile/cpu", api.handleCPUProfiler)
	api.RegisterRoute("GET", prefix+"/debug/profile/mem", api.handleMemProfile)
}

// Static serves static files under the provided root folder.
// Deprecated since the underlying Echo server is exposed as .EchoServer.
func (api *WebAPI) Static(prefix, root string) {
	api.EchoServer.Static(prefix, root)
}

// RegisterRoute registers a new handler for a route.
func (api *WebAPI) RegisterRoute(method, route string, handler func(context echo.Context) error) {
	switch method {
	case "POST":
		api.EchoServer.POST(route, handler)
	case "PUT":
		api.EchoServer.PUT(route, handler)
	case "GET":
		api.EchoServer.GET(route, handler)
	case "DELETE":
		api.EchoServer.DELETE(route, handler)
	case "HEAD":
		api.EchoServer.HEAD(route, handler)
	case "PATCH":
		api.EchoServer.PATCH(route, handler)
	case "OPTIONS":
		api.EchoServer.OPTIONS(route, handler)
	default:
		panic("unexpected route method")
	}
}

// Start launches the HTTP Server and writes the exit
func (api *WebAPI) Start() error {
	api.running = true

	api.log.Infof("starting webserver [address:%s]", api.address)
	err := api.EchoServer.Start(api.address)
	api.log.Infof("shutting down webserver [address:%s]", api.address)
	api.running = false

	if err.Error() == "http: Server closed" {
		return nil
	}

	return err
}

// Stop performs a clean shutdown of the server.
func (api *WebAPI) Stop() error {
	if !api.running {
		return nil
	}

	return api.EchoServer.Server.Shutdown(context.Background())
}

func (api *WebAPI) handleShutdown(c echo.Context) error {
	// TODO: include some form of authorization

	go api.EchoServer.Server.Shutdown(context.Background())

	return c.NoContent(http.StatusNoContent)
}

func (api *WebAPI) handleGetApplicationStats(c echo.Context) error {
	var stats []byte
	var err error

	reset := c.QueryParam("reset")

	if reset == "true" {
		stats, err = app.StatsDumpAndReset()
	} else {
		stats, err = app.StatsDump()
	}

	if err != nil {
		return c.JSON(http.StatusInternalServerError, app.NewError(app.ErrorUnexpected, "failed to dump statistics", err))
	}

	return c.JSONBlob(http.StatusOK, stats)
}

func (api *WebAPI) handleCPUProfiler(c echo.Context) error {
	// this code was shamelessly stolen from the net/http/pprof core library and adapted to work with the Echo package.
	sec, _ := strconv.ParseInt(c.QueryParam("seconds"), 10, 64)
	if sec == 0 {
		sec = 30
	}

	w := c.Response().Writer

	// Set Content Type assuming StartCPUProfile will work,
	// because if it does it starts writing.
	w.Header().Set("Content-Type", "application/octet-stream")
	if err := pprof.StartCPUProfile(w); err != nil {
		// StartCPUProfile failed, so no writes yet.
		// Can change header back to text content
		// and send error code.
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Could not enable CPU profiling: %s\n", err)
		return nil
	}
	sleep(w, time.Duration(sec)*time.Second)
	pprof.StopCPUProfile()

	return nil
}

func (api *WebAPI) handleMemProfile(c echo.Context) error {
	w := c.Response().Writer

	debug, _ := strconv.Atoi(c.QueryParam("debug"))
	gc, _ := strconv.ParseInt(c.QueryParam("gc"), 10, 64)
	if gc == 0 {
		gc = 10
	}
	name := c.QueryParam("name")
	if name == "" {
		name = "heap"
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	p := pprof.Lookup(string(name))
	if p == nil {
		w.WriteHeader(404)
		fmt.Fprintf(w, "Unknown profile: %s\n", name)
		return nil
	}
	if name == "heap" && gc > 0 {
		runtime.GC()
	}
	p.WriteTo(w, debug)
	return nil
}

func (api *WebAPI) webErrorHandler(err error, c echo.Context) {
	var (
		code = http.StatusInternalServerError
		msg  interface{}
	)

	if e, ok := err.(*echo.HTTPError); ok {
		code = e.Code
		msg = map[string]interface{}{"message": e.Message}
	} else if e, ok := err.(app.Error); ok {
		msg = e
	} else {
		msg = map[string]interface{}{"message": http.StatusText(http.StatusInternalServerError)}
	}

	req := c.Request()

	if code == http.StatusInternalServerError {
		api.log.ErrorData("unhandled WebAPI error", log.Fields{
			"method": req.Method,
			"uri":    req.RequestURI,
			"cause":  app.StringifyError(err),
		})
	}

	if !c.Response().Committed {
		if c.Request().Method == echo.HEAD {
			if err := c.NoContent(code); err != nil {
				api.log.ErrorData("error sending response to client", log.Fields{"cause": app.StringifyError(err)})
			}
		} else {
			if err := c.JSON(code, msg); err != nil {
				api.log.ErrorData("error sending response to client", log.Fields{"cause": app.StringifyError(err)})
			}
		}
	}
}

func sleep(w http.ResponseWriter, d time.Duration) {
	var clientGone <-chan bool
	if cn, ok := w.(http.CloseNotifier); ok {
		clientGone = cn.CloseNotify()
	}
	select {
	case <-time.After(d):
	case <-clientGone:
	}
}

// func (api *WebAPI) webLogger() echo.MiddlewareFunc {
func (api *WebAPI) webLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		if api.lh == nil {
			return nil
		}

		req := c.Request()
		res := c.Response()
		start := time.Now()
		if err = next(c); err != nil {
			c.Error(err)
		}
		stop := time.Now()

		fields := make(map[string]interface{}, 13)

		id := req.Header.Get(echo.HeaderXRequestID)
		if id == "" {
			id = res.Header().Get(echo.HeaderXRequestID)
		}
		if id != "" {
			fields["id"] = id
		}

		req.URL.RequestURI()
		p := req.URL.Path
		if p == "" {
			p = "/"
		}
		fields["path"] = p
		fields["method"] = req.Method
		fields["uri"] = req.RequestURI

		cl := req.Header.Get(echo.HeaderContentLength)
		if cl == "" {
			cl = "0"
		}
		fields["bytes_in"] = cl
		fields["bytes_out"] = strconv.FormatInt(res.Size, 10)

		fields["remote_ip"] = c.RealIP()
		fields["status"] = res.Status

		fields["host"] = req.Host
		fields["referer"] = req.Referer()
		fields["user_agent"] = req.UserAgent()
		fields["route"] = c.Path()

		latency := stop.Sub(start)
		fields["latency_ns"] = int64(latency)

		api.lh(fmt.Sprintf("%s %s", req.Method, req.RequestURI), fields)

		return
	}
}
