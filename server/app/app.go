package app

import (
	"context"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"go-scrape-this/server/app/database"
	"go-scrape-this/server/app/database/models"
	"go-scrape-this/server/app/middleware"
	"go-scrape-this/server/app/queue"
	"go-scrape-this/server/app/utils"
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

type Application struct {
	logger       *LoggingHandler
	server       http.Server
	db           database.Database
	queue        *queue.Queue
	version      string
	shutdownWait time.Duration
}

func NewApplication(version string, filesystem http.FileSystem) *Application {
	loggingHandler := NewLoggingHandler(os.Stdout, "application")
	appLogCtx := loggingHandler.Context("application").Str("version", version)
	loggingHandler.SetContext("application", &appLogCtx)

	httpAddressEnv := utils.ReadStringEnv("HTTP_ADDR", "0.0.0.0:8080")
	workerAmountEnv := utils.ReadIntEnv("MAX_QUEUE_WORKERS", runtime.NumCPU())
	shutdownWaitEnv := utils.ReadIntEnv("SHUTDOWN_WAIT", 60)

	dbType, err := database.ParseDatabaseType(utils.ReadStringEnv("DATABASE_TYPE", database.SQLITE.String()))
	if err != nil {
		loggingHandler.Default().Fatal().Msgf("db type: \"%v\"", err)
	}

	var dsn string
	if dbType != database.SQLITE {
		dbDsn, err := utils.RequireStringEnv("DATABASE_DSN")
		if err != nil {
			loggingHandler.Default().Fatal().Msgf("environment: \"%v\"", err)
		}
		dsn = dbDsn
	} else {
		dsn = utils.ReadStringEnv("DATABASE_DSN", "scraper.db")
	}

	db, err := database.NewDatabase(dbType, dsn, loggingHandler.LoggerFromContext("database"))
	if err != nil {
		loggingHandler.Default().Fatal().Msgf("failed to connect to database: \"%v\"", err)
	}

	shutdownWait := time.Second * time.Duration(shutdownWaitEnv)

	a := &Application{
		version:      version,
		logger:       loggingHandler,
		shutdownWait: shutdownWait,
		db:           db,
		queue: queue.NewQueue(
			workerAmountEnv,
			shutdownWait,
			loggingHandler.LoggerFromContext("queue"),
		),
		server: http.Server{
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
	a.initHandlers(filesystem)
	a.initMiddleware()
	return a
}

func (a *Application) Server() *http.Server {
	return &a.server
}

func (a *Application) Logger() *LoggingHandler {
	return a.logger
}

func (a *Application) DefaultLogger() *zerolog.Logger {
	return a.Logger().Default()
}

func (a *Application) Version() string {
	return a.version
}

func (a *Application) Database() *database.Database {
	return &a.db
}

func (a *Application) Start() {
	err := a.Database().RunMigrations()
	if err != nil {
		a.DefaultLogger().Fatal().Msgf("failed to run database migrations: %v\n", err)
	}
	go func() {
		if err := a.Server().ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.DefaultLogger().Fatal().Msgf("failed to start application server: %v\n", err)
		}
	}()
	a.queue.Start()
	a.DefaultLogger().Info().Msg("http server started")
	db := a.Database().Connection()
	rootUser, err := models.NewUser("root", "root")
	if err != nil {
		a.DefaultLogger().Error().Msgf("failed to create root user: %v\n", err)
	}
	var user models.User
	result := db.Where(models.User{Username: "root"}).Attrs(rootUser).FirstOrCreate(&user)
	if result.Error != nil {
		a.DefaultLogger().Error().Msgf("failed to create root user: %v\n", result.Error.Error())
	} else {
		if result.RowsAffected > 0 {
			a.DefaultLogger().Info().Interface("User", user).Msg("root user created")
		}
	}
}

func (a *Application) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), a.shutdownWait)
	defer cancel()
	err := a.Server().Shutdown(ctx)
	if err != nil {
		a.DefaultLogger().Fatal().Msgf("http server shutdown threw errors: %v\n", err)
	}
	a.queue.Stop()
	a.DefaultLogger().Info().Msg("http server stopped")
}

func (a *Application) initHandlers(filesystem http.FileSystem) {
	r := mux.NewRouter()

	r.HandleFunc("/api/health", a.healthAction).Methods("GET")
	r.HandleFunc("/api/version", a.versionAction).Methods("GET")
	r.HandleFunc("/api/status", a.statusAction).Methods("GET")

	r.HandleFunc("/api/workers", a.workerListAction).Methods("GET")

	r.HandleFunc("/api/users", a.userListAction).Methods("GET")

	r.PathPrefix("/").Handler(middleware.StaticFileHandler{
		Filesystem: filesystem,
	})
	a.Server().Handler = r
}

func (a *Application) initMiddleware() {
	h := a.Server().Handler

	h = handlers.ContentTypeHandler(h, allowedContentTypes...)

	if utils.ReadBoolEnv("BEHIND_REVERSE_PROXY", false) {
		h = handlers.ProxyHeaders(h)
	}

	h = handlers.CompressHandler(h)

	h = middleware.LoggingMiddleware(h, a.Logger().LoggerFromContext("http-access"))

	h = handlers.RecoveryHandler(handlers.RecoveryLogger(recoveryHandlerLogger{
		writer: a.Logger().LoggerFromContext("recovery-handler"),
	}))(h)

	a.Server().Handler = h
}
