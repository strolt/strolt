package server

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Start() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("welcome"))
		if err != nil {
			log.Println("error")
		}
	})

	server := &http.Server{
		Addr:              ":3000",
		ReadHeaderTimeout: 30 * time.Second, //nolint:gomnd
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
