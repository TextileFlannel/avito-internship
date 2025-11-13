package storage

import (
	"sync"
)

type Team struct {
	TeamName string
	Members  []TeamMember
}

type TeamMember struct {
	UserId   string
	Username string
	IsActive bool
}

type User struct {
	UserId   string
	Username string
	TeamName string
	IsActive bool
}

type PullRequest struct {
	PullRequestId     string
	PullRequestName   string
	AuthorId          string
	Status            string
	AssignedReviewers []string
}

type Storage struct {
	teams map[string]Team
	users map[string]User
	prs   map[string]PullRequest
	mu    sync.RWMutex
}

func NewStorage() *Storage {
	return &Storage{
		teams: make(map[string]Team),
		users: make(map[string]User),
		prs:   make(map[string]PullRequest),
	}
}
