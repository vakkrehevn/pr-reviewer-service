package handlers

import (
	"encoding/json"
	"net/http"
	"pr-reviewer-service/internal/storage"
)

func StatsHandler(w http.ResponseWriter, r *http.Request, store *storage.Storage) {
	if r.Method != "GET" {
		SendError(w, ErrorNotFound, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := store.GetReviewStats()
	if err != nil {
		SendError(w, ErrorNotFound, "Failed to get stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"stats": stats,
	})
}