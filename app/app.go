package app

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

type application struct {
	logger       *zerolog.Logger
	mux          *mux.Router
	server       *http.Server
	version      string
	shutdownWait time.Duration
}

func NewApplication(version string) *application {
	zerolog.TimeFieldFormat = time.RFC3339
	r := mux.NewRouter()
	a := &application{
		version:      version,
		logger:       &log.Logger,
		shutdownWait: time.Minute,
		server: &http.Server{
			Addr:         "0.0.0.0:8080",
			WriteTimeout: time.Second * 15,
			ReadTimeout:  time.Second * 15,
			IdleTimeout:  time.Second * 60,
			Handler:      r,
		},
	}
	r.HandleFunc("/api/health", a.healthAction)
	r.HandleFunc("/api/version", a.versionAction)
	return a
}

func (a *application) Server() *http.Server {
	return a.server
}

func (a *application) Logger() *zerolog.Logger {
	return a.logger
}

func (a *application) SetHttpServerAddress(addr string) {
	a.Server().Addr = addr
}

func (a *application) SetShutdownWait(duration time.Duration) {
	a.shutdownWait = duration
}

func (a *application) Version() string {
	return a.version
}

func (a *application) StartHttpServer() {
	go func() {
		if err := a.Server().ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.Logger().Fatal().Msgf("Failed to start application server: %v\n", err)
		}
	}()
	a.Logger().Info().Msg("HTTP server started.")
}

func (a *application) StopHttpServer() {
	ctx, cancel := context.WithTimeout(context.Background(), a.shutdownWait)
	defer cancel()
	err := a.Server().Shutdown(ctx)
	if err != nil {
		panic(err)
	}
	a.Logger().Info().Msg("HTTP server started.")
}
