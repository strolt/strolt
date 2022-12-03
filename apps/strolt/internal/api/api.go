package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/api/public"
	"github.com/strolt/strolt/apps/strolt/internal/api/services"
	"github.com/strolt/strolt/apps/strolt/internal/env"
	"github.com/strolt/strolt/apps/strolt/internal/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/docgen"
)

func Serve(ctx context.Context, cancel func()) {
	log := logger.New()

	addr := fmt.Sprintf("%s:%d", env.Host(), env.Port())

	log.Info("API server started on:", addr)

	server := &http.Server{Addr: addr, Handler: service()} //nolint

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	go func() { //nolint
		// fmt.Println("Serve go func...")

		<-ctx.Done()
		// fmt.Println("Serve go func <-ctx.Done()")

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second) //nolint

		go func() {
			// fmt.Println("Serve go func + go func")
			<-shutdownCtx.Done()
			// fmt.Println("Serve go func + go func", <-shutdownCtx.Done())
			if shutdownCtx.Err() == context.DeadlineExceeded { //nolint
				log.Error("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// fmt.Println("Trigger graceful shutdown")
		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Error(err)
		}
		// fmt.Println("serverStopCtx")
		serverStopCtx()
	}()

	// Run the server
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Error(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
	// fmt.Println("<-serverCtx.Done()")
} //nolint

// @title           Strolt API
// @version         1.0
// @BasePath  /
// @securityDefinitions.basic  BasicAuth

func service() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	// r.Use(middleware.Logger)
	r.Use(Logger())

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("sup")) //nolint
	})

	// r.Mount("/swagger", httpSwagger.WrapHandler)

	r.Mount("/service", servicesHandler())

	r.Route("/api", func(r chi.Router) {
		r.Get("/config", getConfig)
		services.Router(r)
	})

	public.Router(r)
	// r.Get("/api/services/status", services.GetStatus)
	// r.Post("/api/services/{serviceName}/tasks/{taskName}/backup", services.PostBackup)
	// r.Get("/api/services/{serviceName}/tasks/{taskName}/destinations/{destinationName}/snapshots", services.GetSnapshots)
	// r.Get("/api/services/{serviceName}/tasks/{taskName}/destinations/{destinationName}/prune", services.GetPrune)
	// r.Post("/api/services/{serviceName}/tasks/{taskName}/destinations/{destinationName}/prune", services.PostPrune)

	r.Get("/slow", func(w http.ResponseWriter, r *http.Request) {
		// Simulates some hard work.
		//
		// We want this handler to complete successfully during a shutdown signal,
		// so consider the work here as some background routine to fetch a long running
		// search query to find as many results as possible, but, instead we cut it short
		// and respond with what we have so far. How a shutdown is handled is entirely
		// up to the developer, as some code blocks are preemptable, and others are not.
		time.Sleep(5 * time.Second) //nolint:gomnd

		w.Write([]byte(fmt.Sprintf("all done.\n"))) //nolint
	})

	docgen.PrintRoutes(r)
	// os.WriteFile("docs.md", []byte(docgen.MarkdownRoutesDoc(r, docgen.MarkdownOpts{})), 0o755)
	// os.WriteFile("docs.json", []byte(docgen.JSONRoutesDoc(r)), 0o755)

	return r
}
