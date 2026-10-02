package web

import (
	"embed"
	"net/http"
)

//go:embed index.html
var files embed.FS

func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, files, "index.html")
	})
}
