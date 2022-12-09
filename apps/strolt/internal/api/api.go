package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/api/public"
	"github.com/strolt/strolt/apps/strolt/internal/api/services"
	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/apps/strolt/internal/env"
	"github.com/strolt/strolt/apps/strolt/internal/ldflags"
	"github.com/strolt/strolt/shared/apiu"
	"github.com/strolt/strolt/shared/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/docgen"
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
// @title           Strolt API.
func (api *API) handler() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	// r.Use(middleware.Logger)
	r.Use(Logger())
	r.Use(middleware.Compress(5)) //nolint:gomnd

	public.New().Router(r)

	r.Group(func(r chi.Router) {
		r.Use(middleware.BasicAuth("api", config.Get().API.Users))

		r.Get("/api/v1/config", api.getConfig)
		r.Get("/api/v1/metrics", api.getStroltMetrics)
		r.Get("/api/v1/info", api.getInfo)
		services.New().Router(r)
	})

	if env.IsDebug() {
		docgen.PrintRoutes(r)
	}

	return r
}

type getInfoResponse struct {
	Version string `json:"version"`
}

// getInfo godoc
// @Id					 getInfo
// @Summary      Get info
// @Tags         info
// @Security BasicAuth
// @success 200 {object} getInfoResponse
// @Router       /api/v1/info [get].
func (api *API) getInfo(w http.ResponseWriter, r *http.Request) {
	apiu.RenderJSON200(w, r, getInfoResponse{
		Version: ldflags.GetVersion(),
	})
}
