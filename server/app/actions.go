package app

import (
	"encoding/json"
	"go-scrape-this/server/app/database/models"
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

func (a *Application) userListAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	limit := utils.GetIntOption(r, "limit", 10)
	offset := utils.GetIntOption(r, "offset", 0)
	if limit > 100 {
		limit = 100
	}
	db := a.Database().Connection()
	var users []models.User
	var count int64
	db.Model(models.User{}).Count(&count)
	db.Limit(limit).Offset(offset).Find(&users)
	err := json.NewEncoder(w).Encode(map[string]interface{}{
		"data":   users,
		"total":  count,
		"count":  len(users),
		"offset": offset,
		"limit":  limit,
	})
	if err != nil {
		panic(err)
	}
}
