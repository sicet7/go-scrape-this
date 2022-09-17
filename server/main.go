package main

import (
	"embed"
	"go-scrape-this/server/app"
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
	application := app.NewApplication(version, http.FS(subFs))
	application.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	<-c

	application.Stop()
	os.Exit(0)
}
