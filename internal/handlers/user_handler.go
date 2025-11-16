package handlers

import (
	"encoding/json"
	"net/http"
	"fmt"
	"pr-reviewer-service/internal/models"
	"pr-reviewer-service/internal/storage"
)

func SetUserActiveHandler(w http.ResponseWriter, r *http.Request, store *storage.Storage) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err := store.UpdateUserActive(request.UserID, request.IsActive)
	if err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	user, err := store.GetUserByID(request.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user": user,
	})
}

func GetUserReviewsHandler(w http.ResponseWriter, r *http.Request, store *storage.Storage) {
    if r.Method != "GET" {
        SendError(w, ErrorNotFound, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    userID := r.URL.Query().Get("user_id")
    if userID == "" {
        SendError(w, ErrorNotFound, "user_id is required", http.StatusBadRequest)
        return
    }

    fmt.Printf("GetUserReviewsHandler called for user: %s\n", userID)

    user, err := store.GetUserByID(userID)
    if err != nil {
        fmt.Printf("GetUserByID error: %v\n", err)
        SendError(w, ErrorNotFound, "Failed to get user", http.StatusInternalServerError)
        return
    }
    if user == nil {
        fmt.Printf("User not found: %s\n", userID)
        SendError(w, ErrorNotFound, "User not found", http.StatusNotFound)
        return
    }

    fmt.Printf("User found: %s, active: %v\n", userID, user.IsActive)

    prs, err := store.GetPRsByReviewer(userID)
    if err != nil {
        fmt.Printf("GetPRsByReviewer error: %v\n", err)
        SendError(w, ErrorNotFound, "Failed to get user reviews", http.StatusInternalServerError)
        return
    }

    fmt.Printf("Found %d PRs for user %s\n", len(prs), userID)

    prsShort := make([]models.PullRequestShort, len(prs))
    for i, pr := range prs {
        prsShort[i] = models.PullRequestShort{
            PullRequestID:   pr.PullRequestID,
            PullRequestName: pr.PullRequestName,
            AuthorID:        pr.AuthorID,
            Status:          pr.Status,
        }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "user_id":       userID,
        "pull_requests": prsShort,
    })
}