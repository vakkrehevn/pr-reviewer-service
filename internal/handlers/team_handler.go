package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"pr-reviewer-service/internal/models"
	"pr-reviewer-service/internal/storage"
)

func AddTeamHandler(w http.ResponseWriter, r *http.Request, store *storage.Storage) {
	if r.Method != "POST" {
		SendError(w, ErrorNotFound, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var team models.Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		SendError(w, ErrorNotFound, "Invalid JSON", http.StatusBadRequest)
		return
	}

	existingTeam, _ := store.GetTeam(team.TeamName)
	if existingTeam != nil {
		SendError(w, ErrorTeamExists, "team_name already exists", http.StatusBadRequest)
		return
	}

	err := store.CreateTeam(team)
	if err != nil {
		log.Printf("Error creating team: %v", err)
		SendError(w, ErrorNotFound, "Failed to create team", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"team": team,
	})
}

func GetTeamHandler(w http.ResponseWriter, r *http.Request, store *storage.Storage) {
	if r.Method != "GET" {
		SendError(w, ErrorNotFound, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		SendError(w, ErrorNotFound, "team_name is required", http.StatusBadRequest)
		return
	}

	team, err := store.GetTeam(teamName)
	if err != nil || team == nil {
		SendError(w, ErrorNotFound, "Team not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(team)
}