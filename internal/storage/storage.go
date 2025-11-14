package storage

import (
	"avito-internship/internal/models"
	"database/sql"
	"encoding/json"
	"errors"

	_ "github.com/lib/pq"
)

type Storage struct {
	DB *sql.DB
}

func NewStorage(connStr string) (*Storage, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &Storage{DB: db}, nil
}

func (s *Storage) AddTeam(teamName string, members []models.TeamMember) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("INSERT INTO teams (team_name) VALUES ($1)", teamName)
	if err != nil {
		return err
	}

	for _, m := range members {
		_, err = tx.Exec("INSERT INTO users (user_id, username, team_name, is_active) VALUES ($1, $2, $3, $4)",
			m.UserId, m.Username, teamName, m.IsActive)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) GetTeam(teamName string) ([]models.TeamMember, error) {
	rows, err := s.DB.Query("SELECT user_id, username, is_active FROM users WHERE team_name = $1", teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.TeamMember
	for rows.Next() {
		var m models.TeamMember
		if err := rows.Scan(&m.UserId, &m.Username, &m.IsActive); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	if len(members) == 0 {
		return nil, errors.New("team not found")
	}
	return members, nil
}

func (s *Storage) SetUserActive(userId string, isActive bool) error {
	result, err := s.DB.Exec("UPDATE users SET is_active = $1 WHERE user_id = $2", isActive, userId)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (s *Storage) CreatePR(pr models.PullRequest) error {
	reviewersJSON, err := json.Marshal(pr.AssignedReviewers)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec("INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, assigned_reviewers, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		pr.PullRequestId, pr.PullRequestName, pr.AuthorId, pr.Status, reviewersJSON, pr.CreatedAt)
	return err
}

func (s *Storage) GetPR(prId string) (models.PullRequest, error) {
	var pr models.PullRequest
	var reviewersJSON []byte
	err := s.DB.QueryRow("SELECT pull_request_id, pull_request_name, author_id, status, assigned_reviewers, created_at, merged_at FROM pull_requests WHERE pull_request_id = $1", prId).
		Scan(&pr.PullRequestId, &pr.PullRequestName, &pr.AuthorId, &pr.Status, &reviewersJSON, &pr.CreatedAt, &pr.MergedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.PullRequest{}, errors.New("PR not found")
		}
		return models.PullRequest{}, err
	}
	err = json.Unmarshal(reviewersJSON, &pr.AssignedReviewers)
	if err != nil {
		return models.PullRequest{}, err
	}
	return pr, nil
}

func (s *Storage) UpdatePR(pr models.PullRequest) error {
	reviewersJSON, err := json.Marshal(pr.AssignedReviewers)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec("UPDATE pull_requests SET pull_request_name = $1, author_id = $2, status = $3, assigned_reviewers = $4, merged_at = $5 WHERE pull_request_id = $6",
		pr.PullRequestName, pr.AuthorId, pr.Status, reviewersJSON, pr.MergedAt, pr.PullRequestId)
	return err
}

func (s *Storage) GetUser(userId string) (models.User, error) {
	var u models.User
	err := s.DB.QueryRow("SELECT user_id, username, team_name, is_active FROM users WHERE user_id = $1", userId).
		Scan(&u.UserId, &u.Username, &u.TeamName, &u.IsActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, errors.New("user not found")
		}
		return models.User{}, err
	}
	return u, nil
}

func (s *Storage) GetUsersByTeam(teamName string) ([]models.User, error) {
	rows, err := s.DB.Query("SELECT user_id, username, team_name, is_active FROM users WHERE team_name = $1", teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.UserId, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (s *Storage) GetPRsByReviewer(userId string) ([]models.PullRequestShort, error) {
	rows, err := s.DB.Query(`
		SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
		FROM pull_requests pr
		WHERE EXISTS (
			SELECT 1
			FROM jsonb_array_elements_text(pr.assigned_reviewers) AS elem
			WHERE elem = $1
		)
	`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []models.PullRequestShort
	for rows.Next() {
		var pr models.PullRequestShort
		if err := rows.Scan(&pr.PullRequestId, &pr.PullRequestName, &pr.AuthorId, &pr.Status); err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}
	return prs, nil
}

func (s *Storage) GetAssignmentStats() ([]models.AssignmentStat, error) {
	rows, err := s.DB.Query(`
		SELECT elem AS user_id, COUNT(*) AS count
		FROM pull_requests pr
		CROSS JOIN jsonb_array_elements_text(pr.assigned_reviewers) AS elem
		GROUP BY elem
		ORDER BY count DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []models.AssignmentStat
	for rows.Next() {
		var stat models.AssignmentStat
		if err := rows.Scan(&stat.UserId, &stat.Count); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	return stats, nil
}
