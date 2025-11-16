package handlers

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func SendError(w http.ResponseWriter, code, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{}
	response.Error.Code = code
	response.Error.Message = message

	json.NewEncoder(w).Encode(response)
}

const (
	ErrorTeamExists  = "TEAM_EXISTS"
	ErrorPRExists    = "PR_EXISTS"
	ErrorPRMerged    = "PR_MERGED"
	ErrorNotAssigned = "NOT_ASSIGNED"
	ErrorNoCandidate = "NO_CANDIDATE"
	ErrorNotFound    = "NOT_FOUND"
)
