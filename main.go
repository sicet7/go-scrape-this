package main

import (
	"embed"
	"github.com/sicet7/go-scrape-this/server"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var version = "0.0.0"

//go:embed all:dist
var content embed.FS

func main() {
	subFs, err := fs.Sub(content, "dist")
	if err != nil {
		panic(err)
	}
	application := server.NewApplication(version, http.FS(subFs))
	application.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	<-c

	application.Stop()
	os.Exit(0)
}
