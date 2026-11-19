// Package api provides the strolt HTTP API server and its routes.
package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/api/public"
	"github.com/strolt/strolt/apps/strolt/internal/api/services"
	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/apps/strolt/internal/env"
	"github.com/strolt/strolt/shared/apiu"
	"github.com/strolt/strolt/shared/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// API is the strolt HTTP API server.
type API struct {
	addr       string
	httpServer *http.Server
	log        *logger.Logger
	operations operationTracker
}

type operationTracker struct {
	mu       sync.Mutex
	wg       sync.WaitGroup
	stopping bool
}

func (tracker *operationTracker) begin() (func(), bool) {
	tracker.mu.Lock()
	defer tracker.mu.Unlock()
	if tracker.stopping {
		return nil, false
	}
	tracker.wg.Add(1)
	return tracker.wg.Done, true
}

func (tracker *operationTracker) stopAndWait() {
	tracker.mu.Lock()
	tracker.stopping = true
	tracker.mu.Unlock()
	tracker.wg.Wait()
}

// New creates an API server listening on the configured host and port.
func New(processContext context.Context) *API {
	addr := fmt.Sprintf("%s:%d", env.Host(), env.Port())
	api := API{addr: addr, httpServer: nil, log: logger.New(), operations: operationTracker{mu: sync.Mutex{}, wg: sync.WaitGroup{}, stopping: false}}
	api.httpServer = api.makeHTTPServer(processContext)
	return &api
}

// Shutdown api http server.
func (api *API) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if api.httpServer != nil {
		if err := api.httpServer.Shutdown(ctx); err != nil {
			api.log.Debugf("api shutdown error, %s", err)
		}

		api.log.Debug("shutdown api server completed")
	}
}

// Run serves requests until ctx is canceled or the server exits.
func (api *API) Run(ctx context.Context, cancel func()) {
	done := make(chan struct{})
	go func() {
		<-ctx.Done()
		api.log.Debug("stop api server...")
		api.Shutdown() //nolint:contextcheck
		close(done)
	}()

	go func() {
		api.log.Infof("api server started on: %s", api.addr)
		if err := api.httpServer.ListenAndServe(); err != nil {
			api.log.Warnf("http server terminated, %s", err)
			cancel()
		}
	}()

	<-done
	api.operations.stopAndWait()
	api.log.Debug("api server was stopped")
}

func (api *API) makeHTTPServer(processContext context.Context) *http.Server {
	return &http.Server{
		Addr:                         api.addr,
		Handler:                      api.handler(processContext),
		DisableGeneralOptionsHandler: false,
		TLSConfig:                    nil,
		ReadTimeout:                  0,
		ReadHeaderTimeout:            5 * time.Second,   //nolint:mnd
		WriteTimeout:                 120 * time.Second, //nolint:mnd
		IdleTimeout:                  30 * time.Second,  //nolint:mnd
		MaxHeaderBytes:               0,
		MaxHeaderValueCount:          0,
		TLSNextProto:                 nil,
		ConnState:                    nil,
		ErrorLog:                     nil,
		BaseContext:                  nil,
		ConnContext:                  nil,
		HTTP2:                        nil,
		Protocols:                    nil,
		DisableClientPriority:        false,
	}
}

// handler godoc
//
//	@version					1.0
//	@securityDefinitions.basic	BasicAuth
//	@title						Strolt API.

func (api *API) handler(processContext context.Context) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Compress(5)) //nolint:mnd

	if env.IsAPILogEnabled() {
		r.Use(apiu.Logger())
	}

	public.New().Router(r)

	r.Group(func(r chi.Router) {
		r.Use(middleware.BasicAuth("api", config.Get().API.Users))

		r.Get("/api/v1/config", api.getConfig)
		r.Get("/api/v1/info", api.getInfo)
		services.New(processContext, api.operations.begin).Router(r)
	})

	return r
}
