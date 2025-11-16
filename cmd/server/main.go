package main

import (
	"fmt"
	"log"
	"net/http"
	"pr-reviewer-service/internal/config"
	"pr-reviewer-service/internal/handlers"
	"pr-reviewer-service/internal/storage"
)

func main() {
	fmt.Printf("Connecting to database: %s\n", config.DBName)
	db, err := storage.InitDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()
	fmt.Println("Connected to PostgreSQL!")

	store := storage.NewStorage(db)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			fmt.Fprintf(w, "PR Reviewer Service is working!")
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	
	http.HandleFunc("/team/add", func(w http.ResponseWriter, r *http.Request) {
		handlers.AddTeamHandler(w, r, store)
	})
	
	http.HandleFunc("/team/get", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetTeamHandler(w, r, store)
	})
	
	http.HandleFunc("/users/setIsActive", func(w http.ResponseWriter, r *http.Request) {
		handlers.SetUserActiveHandler(w, r, store)
	})
	
	http.HandleFunc("/pullRequest/create", func(w http.ResponseWriter, r *http.Request) {
		handlers.CreatePRHandler(w, r, store)
	})
	
	http.HandleFunc("/pullRequest/merge", func(w http.ResponseWriter, r *http.Request) {
		handlers.MergePRHandler(w, r, store)
	})

	http.HandleFunc("/users/getReview", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetUserReviewsHandler(w, r, store)
	})

	http.HandleFunc("/pullRequest/reassign", func(w http.ResponseWriter, r *http.Request) {
		handlers.ReassignReviewerHandler(w, r, store)
	})
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		handlers.StatsHandler(w, r, store)
	})

	http.HandleFunc("/users/bulkDeactivate", func(w http.ResponseWriter, r *http.Request) {
		handlers.BulkDeactivateHandler(w, r, store)
	})
	http.HandleFunc("/pullRequest/get", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetPRHandler(w, r, store)
	})

	fmt.Println("Server starting on port 8080...")
	http.ListenAndServe(":8080", nil)
}

