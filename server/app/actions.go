package app

import (
	"encoding/json"
	"github.com/google/uuid"
	"go-scrape-this/server/app/utils"
	"net/http"
	"runtime"
)

var memoryUsage = utils.NewMemoryUsage()

func (a *Application) healthAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]string{
		"status": "OK",
	})
	if err != nil {
		panic(err)
	}
}

func (a *Application) versionAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]string{
		"version": a.Version(),
	})
	if err != nil {
		panic(err)
	}
}

func (a *Application) statusAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]interface{}{
		"queue-status": a.queue.QueueStatus(),
		"goroutines":   runtime.NumGoroutine(),
		"memory-usage": memoryUsage.Get().Alloc,
	})
	if err != nil {
		panic(err)
	}
}

func (a *Application) workerListAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(a.queue.GetStates())
	if err != nil {
		panic(err)
	}
}

func (a *Application) queueWorkAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	job := TestJob{
		Id:      uuid.New(),
		Message: "test",
	}
	a.queue.Submit(job)
	err := json.NewEncoder(w).Encode(job)
	if err != nil {
		panic(err)
	}
}
