package queue

import (
	"github.com/rs/zerolog"
	"golang.org/x/exp/slices"
	"sync"
)

const (
	initialized = "initialized"
	starting    = "starting"
	pending     = "pending"
	processing  = "processing"
	processed   = "processed"
	stopping    = "stopping"
)

type WorkerState struct {
	Id    int    `json:"worker-id"`
	State string `json:"state,omitempty"`
	JobId string `json:"job-id,omitempty"`
}

// Worker - the worker threads that process the jobs
type Worker struct {
	done             *sync.WaitGroup
	state            WorkerState
	logger           *zerolog.Logger
	readyPool        chan chan Job
	assignedJobQueue chan Job
	quit             chan bool
}

// NewWorker - creates a new worker
func NewWorker(id int, readyPool chan chan Job, done *sync.WaitGroup, logger *zerolog.Logger) *Worker {
	return &Worker{
		done:      done,
		logger:    logger,
		readyPool: readyPool,
		state: WorkerState{
			Id:    id,
			State: initialized,
		},
		assignedJobQueue: make(chan Job),
		quit:             make(chan bool),
	}
}

func (ws WorkerState) IsActive() bool {
	return slices.Contains([]string{processed, processing}, ws.State)
}

func (ws WorkerState) IsReady() bool {
	return ws.State == pending
}

// Finish - finish processing the given job
func (w *Worker) Finish(job Job) {
	w.state = WorkerState{
		Id:    w.state.Id,
		State: processed,
		JobId: job.ID().String(),
	}
	if r := recover(); r != nil {
		job.Error(w.LogWithState(), r)
		w.LogWithState().Error().Msgf("panicked while processing job. \"%v\"", r)
		return
	}
	w.LogWithState().Info().Msg("worker processed job")
}

// Process - Make the worker process a given job
func (w *Worker) Process(job Job) {
	w.state = WorkerState{
		Id:    w.state.Id,
		State: processing,
		JobId: job.ID().String(),
	}
	defer w.Finish(job)
	w.LogWithState().Info().Msg("worker processing job")
	job.Process(w.LogWithState())
}

// LogWithState - returns a logger with the current state already set in the context
func (w *Worker) LogWithState() *zerolog.Logger {
	l := w.logger.With().Interface("state", w.state).Logger()
	return &l
}

// Start - begins the job processing loop for the worker
func (w *Worker) Start() {
	go func() {
		w.done.Add(1)
		w.state = WorkerState{
			Id:    w.state.Id,
			State: starting,
		}
		w.LogWithState().Info().Msg("worker starting")
		for {
			w.readyPool <- w.assignedJobQueue // check the job queue in
			w.state = WorkerState{
				Id:    w.state.Id,
				State: pending,
			}
			w.LogWithState().Info().Msg("worker waiting for jobs")
			select {
			case job := <-w.assignedJobQueue: // see if anything has been assigned to the queue
				w.Process(job)
			case <-w.quit:
				w.state = WorkerState{
					Id:    w.state.Id,
					State: stopping,
				}
				w.LogWithState().Info().Msg("worker stopping")
				w.done.Done()
				return
			}
		}
	}()
}

// State - returns current state of the worker
func (w *Worker) State() WorkerState {
	return w.state
}

// Stop - stops the worker
func (w *Worker) Stop() {
	w.quit <- true
}
