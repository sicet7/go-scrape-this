package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func (a *application) healthAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]string{
		"status": "OK",
	})
	if err != nil {
		log.Printf("Error building the response, %v", err)
		w.WriteHeader(http.StatusInternalServerError)

		_, err := fmt.Fprintf(w, "{ \"error\": \"Error building the response\" }")
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (a *application) versionAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]string{
		"version": a.Version(),
	})
	if err != nil {
		log.Printf("Error building the response, %v", err)
		w.WriteHeader(http.StatusInternalServerError)

		_, err := fmt.Fprintf(w, "{ \"error\": \"Error building the response\" }")
		if err != nil {
			log.Fatal(err)
		}
	}
}
