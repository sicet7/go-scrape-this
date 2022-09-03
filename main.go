package main

import (
	"flag"
	"github.com/sicet7/go-scrape-this/app"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var version = "no version"

func main() {
	var (
		//maxWorkersFlag   = flag.Int("max_workers", 5, "The number of workers to start.")
		//maxQueueSizeFlag = flag.Int("max_queue_size", 100, "The size of job queue.")
		addressFlag      = flag.String("address", "0.0.0.0:8080", "The server IP and Port.")
		shutdownWaitFlag = flag.Int64("shutdown_wait", 60, "The wait time before shutting down.")
	)
	flag.Parse()

	var (
		shutdownWait = time.Duration(*shutdownWaitFlag) * time.Second
		addr         = *addressFlag
	)

	application := app.NewApplication(version)
	application.SetHttpServerAddress(addr)
	application.SetShutdownWait(shutdownWait)

	application.StartHttpServer()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	<-c

	application.StopHttpServer()
	os.Exit(0)
}
