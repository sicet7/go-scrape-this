package queue

import (
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"sync"
)

// Job - interface for job processing
type Job interface {
	ID() uuid.UUID
	Process(*zerolog.Logger)
	Error(...interface{})
}

type WorkerState struct {
	Id    int    `json:"worker-id,omitempty"`
	State string `json:"state,omitempty"`
	JobId string `json:"job-id,omitempty"`
}

// Worker - the worker threads that actually process the jobs
type Worker struct {
	done             *sync.WaitGroup
	state            WorkerState
	logger           *zerolog.Logger
	readyPool        chan chan Job
	assignedJobQueue chan Job
	quit             chan bool
}

// JobQueue - a queue for enqueueing jobs to be processed
type JobQueue struct {
	internalQueue     chan Job
	readyPool         chan chan Job
	workers           []*Worker
	dispatcherStopped *sync.WaitGroup
	workersStopped    *sync.WaitGroup
	quit              chan bool
}

// NewJobQueue - creates a new job queue
func NewJobQueue(maxWorkers int, logger *zerolog.Logger) *JobQueue {
	workersStopped := sync.WaitGroup{}
	readyPool := make(chan chan Job, maxWorkers)
	workers := make([]*Worker, maxWorkers, maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		logger.Debug().Int("worker-id", i).Msg("initializing worker.")
		workers[i] = NewWorker(i, readyPool, &workersStopped, logger)
		logger.Debug().Int("worker-id", i).Msg("initialized worker.")
	}
	return &JobQueue{
		internalQueue:     make(chan Job),
		readyPool:         readyPool,
		workers:           workers,
		dispatcherStopped: &sync.WaitGroup{},
		workersStopped:    &workersStopped,
		quit:              make(chan bool),
	}
}

// Start - starts the worker routines and dispatcher routine
func (q *JobQueue) Start() {
	for i := 0; i < len(q.workers); i++ {
		q.workers[i].Start()
	}
	go func() {
		q.dispatcherStopped.Add(1)
		for {
			select {
			case job := <-q.internalQueue: // We got something in on our queue
				workerChannel := <-q.readyPool // Check out an available worker
				workerChannel <- job           // Send the request to the channel
			case <-q.quit:
				for i := 0; i < len(q.workers); i++ {
					q.workers[i].Stop()
				}
				q.workersStopped.Wait()
				q.dispatcherStopped.Done()
				return
			}
		}
	}()
}

// Stop - stops the workers and dispatcher routine
func (q *JobQueue) Stop() {
	q.quit <- true
	q.dispatcherStopped.Wait()
}

// Submit - adds a new job to be processed
func (q *JobQueue) Submit(job Job) {
	q.internalQueue <- job
}

// NewWorker - creates a new worker
func NewWorker(id int, readyPool chan chan Job, done *sync.WaitGroup, logger *zerolog.Logger) *Worker {
	return &Worker{
		done:      done,
		logger:    logger,
		readyPool: readyPool,
		state: WorkerState{
			Id:    id,
			State: "initialized",
		},
		assignedJobQueue: make(chan Job),
		quit:             make(chan bool),
	}
}

func (w *Worker) HandleError(job Job) {
	w.state = WorkerState{
		Id:    w.state.Id,
		State: "processed",
		JobId: job.ID().String(),
	}
	if r := recover(); r != nil {
		job.Error(r)
		w.LogWithState().Error().Msgf("panicked while processing job. \"%v\"", r)
		return
	}
	w.LogWithState().Info().Msg("worker processed job")
}

func (w *Worker) Process(job Job) {
	w.state = WorkerState{
		Id:    w.state.Id,
		State: "processing",
		JobId: job.ID().String(),
	}
	defer w.HandleError(job)
	w.LogWithState().Info().Msg("worker processing job")
	job.Process(w.LogWithState())
}

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
			State: "starting",
		}
		w.LogWithState().Info().Msg("worker starting")
		for {
			w.readyPool <- w.assignedJobQueue // check the job queue in
			w.state = WorkerState{
				Id:    w.state.Id,
				State: "pending",
			}
			w.LogWithState().Info().Msg("worker waiting for jobs")
			select {
			case job := <-w.assignedJobQueue: // see if anything has been assigned to the queue
				w.Process(job)
			case <-w.quit:
				w.state = WorkerState{
					Id:    w.state.Id,
					State: "stopping",
				}
				w.LogWithState().Info().Msg("worker stopping")
				w.done.Done()
				return
			}
		}
	}()
}

// Stop - stops the worker
func (w *Worker) Stop() {
	w.quit <- true
}
