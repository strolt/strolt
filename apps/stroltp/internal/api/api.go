package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/strolt/strolt/apps/stroltp/internal/api/managerh"
	"github.com/strolt/strolt/apps/stroltp/internal/api/public"
	"github.com/strolt/strolt/apps/stroltp/internal/config"
	"github.com/strolt/strolt/apps/stroltp/internal/env"

	"github.com/strolt/strolt/shared/apiu"
	"github.com/strolt/strolt/shared/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type API struct {
	addr       string
	httpServer *http.Server
	log        *logger.Logger
}

func New() *API {
	addr := fmt.Sprintf("%s:%d", env.Host(), env.Port())

	api := API{
		addr:       addr,
		httpServer: &http.Server{}, //nolint
		log:        logger.New(),
	}

	api.httpServer = api.makeHTTPServer()

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

func (api *API) makeHTTPServer() *http.Server {
	return &http.Server{
		Addr:              api.addr,
		Handler:           api.handler(),
		ReadHeaderTimeout: 5 * time.Second,   //nolint:gomnd
		WriteTimeout:      120 * time.Second, //nolint:gomnd
		IdleTimeout:       30 * time.Second,  //nolint:gomnd
	}
}

// Run the lister and request's router, activate api server.
func (api *API) Run(ctx context.Context, cancel func()) {
	done := make(chan bool)

	go func() {
		<-ctx.Done()

		api.log.Debug("stop api server...")
		api.Shutdown() //nolint:contextcheck
		done <- true
	}()

	go func() {
		api.log.Infof("api server started on: %s", api.addr)

		err := api.httpServer.ListenAndServe()
		if err != nil {
			api.log.Warnf("http server terminated, %s", err)
			cancel()
		}
	}()

	<-done
	api.log.Debug("api server was stopped")
}

// @version         1.0
// @securityDefinitions.basic  BasicAuth
// @title           Strolt Proxy API.
func (api *API) handler() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Compress(5)) //nolint:gomnd

	if env.IsAPILogEnabled() {
		r.Use(apiu.Logger())
	}

	public.New().Router(r)

	r.Group(func(r chi.Router) {
		r.Use(middleware.BasicAuth("api", config.GetUsers()))

		// r.Get("/api/v1/config", api.getConfig)
		r.Get("/api/v1/info", api.getInfo)
		// services.New().Router(r)
		managerh.New().Router(r)
	})

	return r
}
