package queue

import (
	"github.com/rs/zerolog"
	"sync"
	"time"
)

type QueueStatus struct {
	TotalWorkers  int `json:"total-workers"`
	ActiveWorkers int `json:"active-workers"`
	ReadyWorkers  int `json:"ready-workers"`
}

// Queue - a queue for enqueueing jobs to be processed
type Queue struct {
	internalQueue     chan Job
	readyPool         chan chan Job
	workers           []*Worker
	dispatcherStopped *sync.WaitGroup
	workersStopped    *sync.WaitGroup
	quit              chan bool
	shutdownTimeout   time.Duration
}

// NewQueue - creates a new job queue
func NewQueue(maxWorkers int, shutdownTimeout time.Duration, logger *zerolog.Logger) *Queue {
	workersStopped := sync.WaitGroup{}
	readyPool := make(chan chan Job, maxWorkers)
	workers := make([]*Worker, maxWorkers, maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		logger.Debug().Int("worker-id", i).Msg("initializing worker.")
		workers[i] = NewWorker(i, readyPool, &workersStopped, logger)
		logger.Debug().Int("worker-id", i).Msg("initialized worker.")
	}
	return &Queue{
		internalQueue:     make(chan Job),
		readyPool:         readyPool,
		workers:           workers,
		dispatcherStopped: &sync.WaitGroup{},
		workersStopped:    &workersStopped,
		quit:              make(chan bool),
		shutdownTimeout:   shutdownTimeout,
	}
}

// Start - starts the worker routines and dispatcher routine
func (q *Queue) Start() {
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
				if waitTimeout(q.workersStopped, q.shutdownTimeout) {
					panic("Failed to stop all queue workers within the timeout")
				}
				q.dispatcherStopped.Done()
				return
			}
		}
	}()
}

// Stop - stops the workers and dispatcher routine
func (q *Queue) Stop() {
	q.quit <- true
	q.dispatcherStopped.Wait()
}

// Submit - adds a new job to be processed
func (q *Queue) Submit(job Job) {
	q.internalQueue <- job
}

// GetStates - returns the states of all the workers
func (q *Queue) GetStates() []WorkerState {
	output := []WorkerState{}
	for i := 0; i < len(q.workers); i++ {
		output = append(output, q.workers[i].State())
	}
	return output
}

func (q *Queue) QueueStatus() QueueStatus {
	totalWorkers := len(q.workers)
	activeWorkers := 0
	rdyWorkers := 0
	for i := 0; i < len(q.workers); i++ {
		state := q.workers[i].State()
		if state.IsActive() {
			activeWorkers++
		}
		if state.IsReady() {
			rdyWorkers++
		}
	}
	return QueueStatus{
		TotalWorkers:  totalWorkers,
		ActiveWorkers: activeWorkers,
		ReadyWorkers:  rdyWorkers,
	}
}
