package handlers

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"pr-reviewer-service/internal/models"
	"pr-reviewer-service/internal/storage"
	"time"
)

func CreatePRHandler(w http.ResponseWriter, r *http.Request, store *storage.Storage) {
	if r.Method != "POST" {
		SendError(w, ErrorNotFound, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		SendError(w, ErrorNotFound, "Invalid JSON", http.StatusBadRequest)
		return
	}

	existingPR, _ := store.GetPRByID(request.PullRequestID)
	if existingPR != nil {
		SendError(w, ErrorPRExists, "PR id already exists", http.StatusConflict)
		return
	}

	author, err := store.GetUserByID(request.AuthorID)
	if err != nil || author == nil {
		SendError(w, ErrorNotFound, "Author not found", http.StatusNotFound)
		return
	}

	teamMembers, err := store.GetActiveTeamMembers(author.TeamName, request.AuthorID)
	if err != nil {
		SendError(w, ErrorNotFound, "Failed to get team members", http.StatusInternalServerError)
		return
	}

	reviewers := assignReviewers(teamMembers, 2)

	pr := models.PullRequest{
		PullRequestID:    request.PullRequestID,
		PullRequestName:  request.PullRequestName,
		AuthorID:         request.AuthorID,
		Status:           "OPEN",
		AssignedReviewers: reviewers,
	}

	err = store.CreatePR(pr)
	if err != nil {
		SendError(w, ErrorNotFound, "Failed to create PR", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pr": pr,
	})
}

func assignReviewers(teamMembers []models.User, maxReviewers int) []string {
	if len(teamMembers) == 0 {
		return []string{}
	}

	rand.Seed(time.Now().UnixNano())

	shuffled := make([]models.User, len(teamMembers))
	copy(shuffled, teamMembers)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	count := min(len(shuffled), maxReviewers)
	reviewers := make([]string, count)
	for i := 0; i < count; i++ {
		reviewers[i] = shuffled[i].UserID
	}

	return reviewers
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func MergePRHandler(w http.ResponseWriter, r *http.Request, store *storage.Storage) {
    if r.Method != "POST" {
        SendError(w, ErrorNotFound, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var request struct {
        PullRequestID string `json:"pull_request_id"`
    }

    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        SendError(w, ErrorNotFound, "Invalid JSON", http.StatusBadRequest)
        return
    }

    pr, err := store.GetPRByID(request.PullRequestID)
    if err != nil || pr == nil {
        SendError(w, ErrorNotFound, "PR not found", http.StatusNotFound)
        return
    }

    if pr.Status == "MERGED" {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "pr": pr,
        })
        return
    }

    err = store.UpdatePRStatus(request.PullRequestID, "MERGED")
    if err != nil {
        SendError(w, ErrorNotFound, "Failed to merge PR", http.StatusInternalServerError)
        return
    }

    updatedPR, _ := store.GetPRByID(request.PullRequestID)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "pr": updatedPR,
    })
}

func stringPtr(s string) *string {
	return &s
}

func ReassignReviewerHandler(w http.ResponseWriter, r *http.Request, store *storage.Storage) {
    if r.Method != "POST" {
        SendError(w, ErrorNotFound, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var request struct {
        PullRequestID string `json:"pull_request_id"`
        OldUserID     string `json:"old_user_id"`
    }

    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        SendError(w, ErrorNotFound, "Invalid JSON", http.StatusBadRequest)
        return
    }

    pr, err := store.GetPRByID(request.PullRequestID)
    if err != nil || pr == nil {
        SendError(w, ErrorNotFound, "PR not found", http.StatusNotFound)
        return
    }

    if pr.Status == "MERGED" {
        SendError(w, ErrorPRMerged, "cannot reassign on merged PR", http.StatusConflict)
        return
    }

    isAssigned := false
    for _, reviewer := range pr.AssignedReviewers {
        if reviewer == request.OldUserID {
            isAssigned = true
            break
        }
    }
    
    if !isAssigned {
        SendError(w, ErrorNotAssigned, "reviewer is not assigned to this PR", http.StatusConflict)
        return
    }

    oldReviewerTeam, err := store.GetUserTeam(request.OldUserID)
    if err != nil {
        SendError(w, ErrorNotFound, "Old reviewer team not found", http.StatusNotFound)
        return
    }

    teamMembers, err := store.GetActiveTeamMembers(oldReviewerTeam, pr.AuthorID)
    if err != nil {
        SendError(w, ErrorNotFound, "Failed to get team members", http.StatusInternalServerError)
        return
    }

    availableMembers := []models.User{}
    for _, member := range teamMembers {
        isCurrentReviewer := false
        for _, reviewer := range pr.AssignedReviewers {
            if member.UserID == reviewer {
                isCurrentReviewer = true
                break
            }
        }
        if !isCurrentReviewer {
            availableMembers = append(availableMembers, member)
        }
    }

    if len(availableMembers) == 0 {
        SendError(w, ErrorNoCandidate, "no active replacement candidate in team", http.StatusConflict)
        return
    }

    rand.Seed(time.Now().UnixNano())
    newReviewer := availableMembers[rand.Intn(len(availableMembers))]

    newReviewers := make([]string, len(pr.AssignedReviewers))
    for i, reviewer := range pr.AssignedReviewers {
        if reviewer == request.OldUserID {
            newReviewers[i] = newReviewer.UserID
        } else {
            newReviewers[i] = reviewer
        }
    }

    err = store.UpdatePRReviewers(request.PullRequestID, newReviewers)
    if err != nil {
        SendError(w, ErrorNotFound, "Failed to update PR reviewers", http.StatusInternalServerError)
        return
    }

    updatedPR, _ := store.GetPRByID(request.PullRequestID)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "pr":          updatedPR,
        "replaced_by": newReviewer.UserID,
    })
}

func GetPRHandler(w http.ResponseWriter, r *http.Request, store *storage.Storage) {
    if r.Method != "GET" {
        SendError(w, ErrorNotFound, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    prID := r.URL.Query().Get("pull_request_id")
    if prID == "" {
        SendError(w, ErrorNotFound, "pull_request_id is required", http.StatusBadRequest)
        return
    }

    pr, err := store.GetPRByID(prID)
    if err != nil {
        SendError(w, ErrorNotFound, "Failed to get PR", http.StatusInternalServerError)
        return
    }
    if pr == nil {
        SendError(w, ErrorNotFound, "PR not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "pr": pr,
    })
}