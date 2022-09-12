package main

import (
	"github.com/sicet7/go-scrape-this/app"
	"os"
	"os/signal"
	"syscall"
)

var version = "no version"

func main() {
	application := app.NewApplication(version)

	application.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	<-c

	application.Stop()
	os.Exit(0)
}
