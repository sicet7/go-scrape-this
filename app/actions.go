package app

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/sicet7/go-scrape-this/app/utilities"
	"net/http"
	"runtime"
)

var memoryUsage = utilities.NewMemoryUsage()

func (a *application) healthAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]string{
		"status": "OK",
	})
	if err != nil {
		panic(err)
	}
}

func (a *application) versionAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]string{
		"version": a.Version(),
	})
	if err != nil {
		panic(err)
	}
}

func (a *application) statusAction(w http.ResponseWriter, r *http.Request) {
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

func (a *application) workerListAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(a.queue.GetStates())
	if err != nil {
		panic(err)
	}
}

func (a *application) queueWorkAction(w http.ResponseWriter, r *http.Request) {
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
