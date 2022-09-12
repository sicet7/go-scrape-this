package app

import (
	"context"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/sicet7/go-scrape-this/app/database"
	"github.com/sicet7/go-scrape-this/app/logging"
	"github.com/sicet7/go-scrape-this/app/queue"
	"github.com/sicet7/go-scrape-this/app/utilities"
	goLog "log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"
)

var allowedContentTypes = []string{
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
	logger       *logging.LoggingHandler
	mux          *mux.Router
	server       *http.Server
	db           *database.Database
	queue        *queue.Queue
	version      string
	shutdownWait time.Duration
}

func initHandler(a *application) {
	r := mux.NewRouter()
	r.HandleFunc("/api/health", a.healthAction).Methods("GET")
	r.HandleFunc("/api/version", a.versionAction).Methods("GET")
	r.HandleFunc("/api/status", a.statusAction).Methods("GET")
	r.HandleFunc("/api/queue/work", a.queueWorkAction).Methods("GET")
	r.HandleFunc("/api/queue/workers", a.workerListAction).Methods("GET")
	a.Server().Handler = r
}

func initMiddleware(a *application) {
	h := a.Server().Handler

	h = handlers.ContentTypeHandler(h, allowedContentTypes...)

	if utilities.ReadBoolEnv("BEHIND_REVERSE_PROXY", false) {
		h = handlers.ProxyHeaders(h)
	}

	h = handlers.CompressHandler(h)

	h = logging.LoggingMiddleware(h, a.Logger().LoggerFromContext("http-access"))

	h = handlers.RecoveryHandler(handlers.RecoveryLogger(recoveryHandlerLogger{
		writer: a.Logger().LoggerFromContext("recovery-handler"),
	}))(h)

	a.Server().Handler = h
}

func NewApplication(version string) *application {
	loggingHandler := logging.NewHandler(os.Stdout, "application")
	appLogCtx := loggingHandler.Context("application").Str("version", version)
	loggingHandler.SetContext("application", &appLogCtx)

	httpAddressEnv := utilities.ReadStringEnv("HTTP_ADDR", "0.0.0.0:8080")
	workerAmountEnv := utilities.ReadIntEnv("MAX_QUEUE_WORKERS", runtime.NumCPU())
	shutdownWaitEnv := utilities.ReadIntEnv("SHUTDOWN_WAIT", 60)

	dbType, err := database.ParseDatabaseType(utilities.ReadStringEnv("DATABASE_TYPE", database.SQLITE.String()))
	if err != nil {
		loggingHandler.Default().Fatal().Msgf("db type: \"%v\"", err)
	}

	var dsn string
	if dbType != database.SQLITE {
		dbDsn, err := utilities.RequireStringEnv("DATABASE_DSN")
		if err != nil {
			loggingHandler.Default().Fatal().Msgf("environment: \"%v\"", err)
		}
		dsn = dbDsn
	} else {
		dsn = utilities.ReadStringEnv("DATABASE_DSN", "scraper.db")
	}

	db, err := database.NewDatabase(dbType, dsn, loggingHandler.LoggerFromContext("database"))
	if err != nil {
		loggingHandler.Default().Fatal().Msgf("failed to connect to database: \"%v\"", err)
	}

	shutdownWait := time.Second * time.Duration(shutdownWaitEnv)

	a := &application{
		version:      version,
		logger:       loggingHandler,
		shutdownWait: shutdownWait,
		db:           db,
		queue:        queue.NewQueue(workerAmountEnv, shutdownWait, loggingHandler.LoggerFromContext("queue")),
		server: &http.Server{
			Addr:         httpAddressEnv,
			WriteTimeout: time.Second * 15,
			ReadTimeout:  time.Second * 15,
			IdleTimeout:  time.Second * 60,
			ErrorLog: goLog.New(
				loggingHandler.LoggerFromContext("http-error"),
				"",
				goLog.Lmsgprefix|goLog.Llongfile,
			),
		},
	}
	initHandler(a)
	initMiddleware(a)
	return a
}

func (a *application) Server() *http.Server {
	return a.server
}

func (a *application) Logger() *logging.LoggingHandler {
	return a.logger
}

func (a *application) DefaultLogger() *zerolog.Logger {
	return a.Logger().Default()
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
			a.DefaultLogger().Fatal().Msgf("failed to start application server: %v\n", err)
		}
	}()
	a.queue.Start()
	a.DefaultLogger().Info().Msg("http server started")
}

func (a *application) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), a.shutdownWait)
	defer cancel()
	err := a.Server().Shutdown(ctx)
	if err != nil {
		a.DefaultLogger().Fatal().Msgf("http server shutdown threw errors: %v\n", err)
	}
	a.queue.Stop()
	a.DefaultLogger().Info().Msg("http server stopped")
}
