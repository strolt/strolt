package ui

import (
	"embed"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"

	"github.com/go-chi/chi/v5"
)

//go:embed build/*
var Assets embed.FS

var mode = "embed"

type fsFunc func(name string) (fs.File, error)

func (f fsFunc) Open(name string) (fs.File, error) {
	return f(name)
}

func isDeny(name string) bool {
	return name == ".gitignore"
}

func assetHandler() http.Handler {
	handler := fsFunc(func(name string) (fs.File, error) {
		assetPath := path.Join("build/dist", name)

		f, err := Assets.Open(assetPath)
		if os.IsNotExist(err) || isDeny(name) {
			return Assets.Open("build/dist/index.html")
		}

		return f, err
	})

	return http.StripPrefix("/", http.FileServer(http.FS(handler)))
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	url, _ := url.Parse("http://127.0.0.1:5080")

	proxy := httputil.ReverseProxy{Director: func(r *http.Request) {
		r.URL.Scheme = url.Scheme
		r.URL.Host = url.Host
		r.URL.Path = url.Path + r.URL.Path
		r.Host = url.Host
	}}
	proxy.ServeHTTP(w, r)
}

func Router(r chi.Router) {
	if mode == "proxy" {
		r.Get("/*", proxyHandler)
	} else {
		uiHandler := assetHandler()
		r.Handle("/", uiHandler)
		r.Handle("/*", uiHandler)
	}
}
