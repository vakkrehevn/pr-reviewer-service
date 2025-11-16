package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"pr-reviewer-service/internal/handlers"
	"pr-reviewer-service/internal/storage"

	"github.com/stretchr/testify/assert"
)

func TestIntegration_TeamLifecycle(t *testing.T) {
	db, err := storage.InitDB()
	assert.NoError(t, err)
	defer db.Close()

	store := storage.NewStorage(db)
	server := setupTestServer(store)
	defer server.Close()

	teamData := map[string]interface{}{
		"team_name": "integration-team-1",
		"members": []map[string]interface{}{
			{"user_id": "int-user-1", "username": "Int User 1", "is_active": true},
			{"user_id": "int-user-2", "username": "Int User 2", "is_active": true},
			{"user_id": "int-user-3", "username": "Int User 3", "is_active": true},
		},
	}

	jsonData, _ := json.Marshal(teamData)
	resp, err := http.Post(server.URL+"/team/add", "application/json", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	resp, err = http.Get(server.URL + "/team/get?team_name=integration-team-1")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var teamResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&teamResponse)
	assert.NoError(t, err)
	assert.Equal(t, "integration-team-1", teamResponse["team_name"])

	members := teamResponse["members"].([]interface{})
	assert.Equal(t, 3, len(members))
}

func TestIntegration_PRLifecycle(t *testing.T) {
	db, err := storage.InitDB()
	assert.NoError(t, err)
	defer db.Close()

	store := storage.NewStorage(db)
	server := setupTestServer(store)
	defer server.Close()

	teamData := map[string]interface{}{
		"team_name": "integration-team-2",
		"members": []map[string]interface{}{
			{"user_id": "pr-user-1", "username": "PR User 1", "is_active": true},
			{"user_id": "pr-user-2", "username": "PR User 2", "is_active": true},
			{"user_id": "pr-user-3", "username": "PR User 3", "is_active": true},
		},
	}

	jsonData, _ := json.Marshal(teamData)
	_, err = http.Post(server.URL+"/team/add", "application/json", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)

	prData := map[string]interface{}{
		"pull_request_id":   "integration-pr-1",
		"pull_request_name": "Integration Test PR",
		"author_id":         "pr-user-1",
	}

	jsonData, _ = json.Marshal(prData)
	resp, err := http.Post(server.URL+"/pullRequest/create", "application/json", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var prResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&prResponse)
	assert.NoError(t, err)

	pr := prResponse["pr"].(map[string]interface{})
	reviewers := pr["assigned_reviewers"].([]interface{})

	assert.Equal(t, "OPEN", pr["status"])
	assert.True(t, len(reviewers) > 0, "Reviewers should be assigned")

	mergeData := map[string]interface{}{
		"pull_request_id": "integration-pr-1",
	}
	jsonData, _ = json.Marshal(mergeData)

	resp, err = http.Post(server.URL+"/pullRequest/merge", "application/json", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestIntegration_UserReviews(t *testing.T) {
	db, err := storage.InitDB()
	assert.NoError(t, err)
	defer db.Close()

	store := storage.NewStorage(db)
	server := setupTestServer(store)
	defer server.Close()

	teamData := map[string]interface{}{
		"team_name": "integration-team-3",
		"members": []map[string]interface{}{
			{"user_id": "review-user-1", "username": "Review User 1", "is_active": true},
			{"user_id": "review-user-2", "username": "Review User 2", "is_active": true},
		},
	}

	jsonData, _ := json.Marshal(teamData)
	_, err = http.Post(server.URL+"/team/add", "application/json", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)

	prData := map[string]interface{}{
		"pull_request_id":   "integration-pr-review",
		"pull_request_name": "Review Test PR",
		"author_id":         "review-user-1",
	}

	jsonData, _ = json.Marshal(prData)
	_, err = http.Post(server.URL+"/pullRequest/create", "application/json", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)

	resp, err := http.Get(server.URL + "/users/getReview?user_id=review-user-2")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var reviewsResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&reviewsResponse)
	assert.NoError(t, err)
	assert.Equal(t, "review-user-2", reviewsResponse["user_id"])
}

func TestIntegration_Stats(t *testing.T) {
	db, err := storage.InitDB()
	assert.NoError(t, err)
	defer db.Close()

	store := storage.NewStorage(db)
	server := setupTestServer(store)
	defer server.Close()

	resp, err := http.Get(server.URL + "/stats")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var statsResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&statsResponse)
	assert.NoError(t, err)
	assert.NotNil(t, statsResponse["stats"])
}

func setupTestServer(store *storage.Storage) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			fmt.Fprintf(w, "PR Reviewer Service is working!")
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/team/add", func(w http.ResponseWriter, r *http.Request) {
		handlers.AddTeamHandler(w, r, store)
	})

	mux.HandleFunc("/team/get", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetTeamHandler(w, r, store)
	})

	mux.HandleFunc("/users/setIsActive", func(w http.ResponseWriter, r *http.Request) {
		handlers.SetUserActiveHandler(w, r, store)
	})

	mux.HandleFunc("/pullRequest/create", func(w http.ResponseWriter, r *http.Request) {
		handlers.CreatePRHandler(w, r, store)
	})

	mux.HandleFunc("/pullRequest/merge", func(w http.ResponseWriter, r *http.Request) {
		handlers.MergePRHandler(w, r, store)
	})

	mux.HandleFunc("/users/getReview", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetUserReviewsHandler(w, r, store)
	})

	mux.HandleFunc("/pullRequest/reassign", func(w http.ResponseWriter, r *http.Request) {
		handlers.ReassignReviewerHandler(w, r, store)
	})

	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		handlers.StatsHandler(w, r, store)
	})

	mux.HandleFunc("/users/bulkDeactivate", func(w http.ResponseWriter, r *http.Request) {
		handlers.BulkDeactivateHandler(w, r, store)
	})

	mux.HandleFunc("/pullRequest/get", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetPRHandler(w, r, store)
	})

	return httptest.NewServer(mux)
}
