package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"pr-reviewer-service/internal/config"
	"pr-reviewer-service/internal/models"

	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func (s *Storage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) CreateTeam(team models.Team) error {
	_, err := s.db.Exec("INSERT INTO teams (team_name) VALUES ($1) ON CONFLICT (team_name) DO NOTHING", team.TeamName)
	if err != nil {
		return err
	}

	for _, member := range team.Members {
		_, err := s.db.Exec(`
			INSERT INTO users (user_id, username, team_name, is_active) 
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id) DO UPDATE SET 
				username = $2, team_name = $3, is_active = $4
		`, member.UserID, member.Username, team.TeamName, member.IsActive) // ← Убедись что team.TeamName
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Storage) GetUserByID(userID string) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(`
		SELECT user_id, username, team_name, is_active 
		FROM users WHERE user_id = $1
	`, userID).Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Storage) GetActiveTeamMembers(teamName, excludeUserID string) ([]models.User, error) {
	rows, err := s.db.Query(`
		SELECT user_id, username, team_name, is_active 
		FROM users 
		WHERE team_name = $1 AND is_active = true AND user_id != $2
	`, teamName, excludeUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (s *Storage) CreatePR(pr models.PullRequest) error {
	reviewersJSON, _ := json.Marshal(pr.AssignedReviewers)

	_, err := s.db.Exec(`
		INSERT INTO pull_requests 
		(pull_request_id, pull_request_name, author_id, status, assigned_reviewers) 
		VALUES ($1, $2, $3, $4, $5)
	`, pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.Status, reviewersJSON)

	return err
}

func (s *Storage) GetPRByID(prID string) (*models.PullRequest, error) {
	var pr models.PullRequest
	var reviewersJSON string

	err := s.db.QueryRow(`
		SELECT pull_request_id, pull_request_name, author_id, status, assigned_reviewers, created_at, merged_at
		FROM pull_requests WHERE pull_request_id = $1
	`, prID).Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &reviewersJSON, &pr.CreatedAt, &pr.MergedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(reviewersJSON), &pr.AssignedReviewers)

	return &pr, nil
}

func (s *Storage) UpdateUserActive(userID string, isActive bool) error {
	_, err := s.db.Exec(`
		UPDATE users SET is_active = $1 WHERE user_id = $2
	`, isActive, userID)
	return err
}

func InitDB() (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (s *Storage) GetTeam(teamName string) (*models.Team, error) {
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)", teamName).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	rows, err := s.db.Query(`
		SELECT user_id, username, team_name, is_active 
		FROM users WHERE team_name = $1
	`, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)
		if err != nil {
			return nil, err
		}
		members = append(members, user)
	}

	return &models.Team{
		TeamName: teamName,
		Members:  members,
	}, nil
}

func (s *Storage) UpdatePRStatus(prID string, status string) error {
	if status == "MERGED" {
		_, err := s.db.Exec(`
            UPDATE pull_requests 
            SET status = $1, merged_at = CURRENT_TIMESTAMP 
            WHERE pull_request_id = $2 AND status != 'MERGED'
        `, status, prID)
		return err
	}

	_, err := s.db.Exec(`
        UPDATE pull_requests SET status = $1 WHERE pull_request_id = $2
    `, status, prID)
	return err
}

func (s *Storage) GetPRsByReviewer(userID string) ([]models.PullRequest, error) {
	rows, err := s.db.Query(`
        SELECT pull_request_id, pull_request_name, author_id, status, assigned_reviewers, created_at, merged_at
        FROM pull_requests 
        WHERE assigned_reviewers::jsonb ? $1
        ORDER BY created_at DESC
    `, userID)

	if err != nil {
		fmt.Printf("SQL error in GetPRsByReviewer: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var prs []models.PullRequest
	for rows.Next() {
		var pr models.PullRequest
		var reviewersJSON string

		err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &reviewersJSON, &pr.CreatedAt, &pr.MergedAt)
		if err != nil {
			fmt.Printf("Scan error in GetPRsByReviewer: %v\n", err)
			return nil, err
		}
		if reviewersJSON != "" {
			err = json.Unmarshal([]byte(reviewersJSON), &pr.AssignedReviewers)
			if err != nil {
				fmt.Printf("JSON unmarshal error in GetPRsByReviewer: %v\n", err)
			}
		}

		prs = append(prs, pr)
	}

	return prs, nil
}

func (s *Storage) GetUserTeam(userID string) (string, error) {
	var teamName string
	err := s.db.QueryRow("SELECT team_name FROM users WHERE user_id = $1", userID).Scan(&teamName)
	return teamName, err
}

func (s *Storage) UpdatePRReviewers(prID string, reviewers []string) error {
	reviewersJSON, _ := json.Marshal(reviewers)
	_, err := s.db.Exec(`
        UPDATE pull_requests SET assigned_reviewers = $1 WHERE pull_request_id = $2
    `, reviewersJSON, prID)
	return err
}

func (s *Storage) GetReviewStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	rows, err := s.db.Query(`
		SELECT u.user_id, u.username, 
			COUNT(CASE WHEN pr.assigned_reviewers IS NOT NULL AND pr.assigned_reviewers::jsonb ? u.user_id THEN 1 END) as assignment_count
		FROM users u
		LEFT JOIN pull_requests pr ON pr.assigned_reviewers IS NOT NULL AND pr.assigned_reviewers::jsonb ? u.user_id
		GROUP BY u.user_id, u.username
		ORDER BY assignment_count DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type UserStat struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		Count    int    `json:"assignment_count"`
	}

	var userStats []UserStat
	for rows.Next() {
		var stat UserStat
		err := rows.Scan(&stat.UserID, &stat.Username, &stat.Count)
		if err != nil {
			return nil, err
		}
		userStats = append(userStats, stat)
	}

	var totalPRs, openPRs, mergedPRs int
	s.db.QueryRow("SELECT COUNT(*) FROM pull_requests").Scan(&totalPRs)
	s.db.QueryRow("SELECT COUNT(*) FROM pull_requests WHERE status = 'OPEN'").Scan(&openPRs)
	s.db.QueryRow("SELECT COUNT(*) FROM pull_requests WHERE status = 'MERGED'").Scan(&mergedPRs)

	stats["user_assignments"] = userStats
	stats["total_prs"] = totalPRs
	stats["open_prs"] = openPRs
	stats["merged_prs"] = mergedPRs
	stats["total_users"] = len(userStats)

	return stats, nil
}

func (s *Storage) BulkDeactivateTeamUsers(teamName string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`
		UPDATE users SET is_active = false 
		WHERE team_name = $1
	`, teamName)
	if err != nil {
		return nil, err
	}

	deactivatedCount, _ := res.RowsAffected()
	result["deactivated_users"] = deactivatedCount

	var affectedPRs []string
	rows, err := tx.Query(`
		SELECT DISTINCT pr.pull_request_id 
		FROM pull_requests pr
		JOIN users u ON u.user_id = ANY(
		SELECT value->>0 
		FROM jsonb_array_elements(pr.assigned_reviewers)
		)
		WHERE pr.status = 'OPEN' 
		AND u.team_name = $1 
		AND u.is_active = false
	`, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var prID string
		err := rows.Scan(&prID)
		if err != nil {
			return nil, err
		}
		affectedPRs = append(affectedPRs, prID)
	}

	result["affected_prs"] = affectedPRs
	result["reassigned_prs_count"] = 0

	for _, prID := range affectedPRs {
		reassigned, err := s.safeReassignPRReviewers(tx, prID, teamName)
		if err != nil {
			return nil, err
		}
		if reassigned {
			result["reassigned_prs_count"] = result["reassigned_prs_count"].(int) + 1
		}
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *Storage) safeReassignPRReviewers(tx *sql.Tx, prID string, teamName string) (bool, error) {
	var pr models.PullRequest
	var reviewersJSON string

	err := tx.QueryRow(`
		SELECT pull_request_id, author_id, assigned_reviewers, status
		FROM pull_requests WHERE pull_request_id = $1
	`, prID).Scan(&pr.PullRequestID, &pr.AuthorID, &reviewersJSON, &pr.Status)

	if err != nil {
		return false, err
	}
	json.Unmarshal([]byte(reviewersJSON), &pr.AssignedReviewers)

	var deactivatedReviewers []string
	var activeReviewers []string

	for _, reviewer := range pr.AssignedReviewers {
		var isActive bool
		err := tx.QueryRow(`
			SELECT is_active FROM users WHERE user_id = $1
		`, reviewer).Scan(&isActive)

		if err == nil && isActive {
			activeReviewers = append(activeReviewers, reviewer)
		} else {
			deactivatedReviewers = append(deactivatedReviewers, reviewer)
		}
	}

	if len(deactivatedReviewers) > 0 {
		rows, err := tx.Query(`
			SELECT user_id FROM users 
			WHERE team_name = $1 
			AND is_active = true 
			AND user_id != $2
			AND user_id NOT IN (SELECT jsonb_array_elements_text($3::jsonb))
		`, teamName, pr.AuthorID, reviewersJSON)

		if err != nil {
			return false, err
		}
		defer rows.Close()

		var availableUsers []string
		for rows.Next() {
			var userID string
			err := rows.Scan(&userID)
			if err != nil {
				return false, err
			}
			availableUsers = append(availableUsers, userID)
		}
		newReviewers := activeReviewers
		needed := 2 - len(activeReviewers)
		if needed > 0 && len(availableUsers) > 0 {
			count := min(needed, len(availableUsers))
			newReviewers = append(newReviewers, availableUsers[:count]...)
		}

		newReviewersJSON, _ := json.Marshal(newReviewers)
		_, err = tx.Exec(`
			UPDATE pull_requests SET assigned_reviewers = $1 
			WHERE pull_request_id = $2
		`, newReviewersJSON, prID)

		if err != nil {
			return false, err
		}

		return true, nil
	}

	return false, nil
}
