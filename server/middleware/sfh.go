package middleware

import "net/http"

type StaticFileHandler struct {
	Filesystem http.FileSystem
}

func (h StaticFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.FileServer(h.Filesystem).ServeHTTP(w, r)
}
