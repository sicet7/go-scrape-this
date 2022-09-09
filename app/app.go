package app

import (
	"context"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sicet7/go-scrape-this/app/database"
	"github.com/sicet7/go-scrape-this/app/middlewares"
	goLog "log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var allowedContentTypes []string = []string{
	"application/json",
}

type recoveryHandlerLogger struct {
	writer *zerolog.Logger
}

func (h recoveryHandlerLogger) Println(v ...interface{}) {
	event := h.writer.Error()
	event.Int("code", http.StatusInternalServerError)
	for i, p := range v {
		event = event.Interface("panic "+strconv.Itoa(i), p)
	}
	event.Msg("handled panic")
}

type application struct {
	logger       *zerolog.Logger
	mux          *mux.Router
	server       *http.Server
	db           *database.Database
	version      string
	shutdownWait time.Duration
}

func initHandler(a *application) {
	r := mux.NewRouter()
	r.HandleFunc("/api/health", a.healthAction)
	r.HandleFunc("/api/version", a.versionAction)
	a.Server().Handler = r
}

func initMiddleware(a *application) {
	h := a.Server().Handler

	h = handlers.ContentTypeHandler(h, allowedContentTypes...)

	reverseProxy, ok := os.LookupEnv("BEHIND_REVERSE_PROXY")
	if !ok {
		reverseProxy = "false"
	}

	if strings.ToLower(reverseProxy) == "true" {
		h = handlers.ProxyHeaders(h)
	}

	h = handlers.CompressHandler(h)

	h = middlewares.LoggingMiddleware(h, a.Logger())

	h = handlers.RecoveryHandler(handlers.RecoveryLogger(recoveryHandlerLogger{
		writer: a.Logger(),
	}))(h)

	a.Server().Handler = h
}

func NewApplication(version string) *application {
	zerolog.TimeFieldFormat = time.RFC3339
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	db, err := database.NewDatabase(&logger)
	if err != nil {
		log.Fatal().Msgf("failed to connect to database: \"%v\"", err)
	}

	a := &application{
		version:      version,
		logger:       &logger,
		shutdownWait: time.Minute,
		db:           db,
		server: &http.Server{
			Addr:         "0.0.0.0:8080",
			WriteTimeout: time.Second * 15,
			ReadTimeout:  time.Second * 15,
			IdleTimeout:  time.Second * 60,
			ErrorLog:     goLog.New(logger, "http error:", goLog.Lmsgprefix|goLog.Llongfile),
		},
	}
	initHandler(a)
	initMiddleware(a)
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

func (a *application) Database() *database.Database {
	return a.db
}

func (a *application) Start() {
	go func() {
		if err := a.Server().ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.Logger().Fatal().Msgf("failed to start application server: %v\n", err)
		}
	}()
	a.Logger().Info().Msg("http server started")
}

func (a *application) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), a.shutdownWait)
	defer cancel()
	err := a.Server().Shutdown(ctx)
	if err != nil {
		a.Logger().Error().Msgf("http server shutdown threw errors: %v\n", err)
		return err
	}
	a.Logger().Info().Msg("http server stopped")
	return nil
}
