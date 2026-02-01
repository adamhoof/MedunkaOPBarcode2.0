package main

import (
	"encoding/json"
	"net/http"
	"sync"
)

type JobStatus struct {
	State   string `json:"state"`
	Message string `json:"message"`
}

// HandleJobStatus retrieves the current state of an import job using the "id" query parameter.
// It returns 404 if the job ID is not found in the memory store.
func HandleJobStatus(jobStore *sync.Map) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobID := r.URL.Query().Get("id")
		if jobID == "" {
			http.Error(w, "Missing id parameter", http.StatusBadRequest)
			return
		}

		status, found := jobStore.Load(jobID)
		if !found {
			http.Error(w, "Job not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	}
}
