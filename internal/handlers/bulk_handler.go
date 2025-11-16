package handlers

import (
	"encoding/json"
	"net/http"
	"pr-reviewer-service/internal/storage"
)

func BulkDeactivateHandler(w http.ResponseWriter, r *http.Request, store *storage.Storage) {
	if r.Method != "POST" {
		SendError(w, ErrorNotFound, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		TeamName string `json:"team_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		SendError(w, ErrorNotFound, "Invalid JSON", http.StatusBadRequest)
		return
	}

	result, err := store.BulkDeactivateTeamUsers(request.TeamName)
	if err != nil {
		SendError(w, ErrorNotFound, "Failed to deactivate users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"team_name": request.TeamName,
		"result":    result,
	})
}
