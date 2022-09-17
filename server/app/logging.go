package app

import (
	"github.com/rs/zerolog"
	"io"
	"sync"
	"time"
)

type LoggingHandler struct {
	writer         io.Writer
	contextLock    sync.Mutex
	defaultContext string
	contextList    map[string]*zerolog.Context
}

func NewLoggingHandler(writer io.Writer, defaultContext string) *LoggingHandler {
	if zerolog.TimeFieldFormat != time.RFC3339 {
		zerolog.TimeFieldFormat = time.RFC3339
	}
	return &LoggingHandler{
		writer:         writer,
		defaultContext: defaultContext,
		contextList:    map[string]*zerolog.Context{},
	}
}

func (l *LoggingHandler) SetContext(contextName string, context *zerolog.Context) {
	l.contextLock.Lock()
	defer l.contextLock.Unlock()
	l.contextList[contextName] = context
}

func (l *LoggingHandler) Context(contextName string) *zerolog.Context {
	context, exist := l.contextList[contextName]
	if exist {
		return context
	}
	l.contextLock.Lock()
	defer l.contextLock.Unlock()
	context, exist = l.contextList[contextName]
	if exist {
		return context
	}
	newContext := zerolog.New(l.writer).With().Str("context", contextName).Timestamp()
	l.contextList[contextName] = &newContext
	return &newContext
}

func (l *LoggingHandler) LoggerFromContext(contextName string) *zerolog.Logger {
	logger := l.Context(contextName).Logger()
	return &logger
}

func (l *LoggingHandler) Default() *zerolog.Logger {
	return l.LoggerFromContext(l.defaultContext)
}
