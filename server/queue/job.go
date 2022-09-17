package queue

import (
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Job - interface for job processing
type Job interface {
	ID() uuid.UUID
	Process(*zerolog.Logger)
	Error(*zerolog.Logger, interface{})
}
