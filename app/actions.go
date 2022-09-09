package app

import (
	"encoding/json"
	"net/http"
)

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
