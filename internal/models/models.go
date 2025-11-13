package models

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

type PullRequestShort struct {
	PullRequestId   string
	PullRequestName string
	AuthorId        string
	Status          string
}
