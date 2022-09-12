package app

import (
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type TestJob struct {
	Id      uuid.UUID `json:"id"`
	Message string    `json:"message"`
}

func (t TestJob) ID() uuid.UUID {
	return t.Id
}

func (t TestJob) Process(logger *zerolog.Logger) {
	panic(t.Message)
}

func (t TestJob) Error(logger *zerolog.Logger, v interface{}) {
	logger.Error().Interface("panic", v).Msg("failed to execute job.")
}
