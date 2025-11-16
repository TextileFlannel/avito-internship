package service

import (
	"avito-internship/internal/models"
	"avito-internship/internal/storage"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_AddTeam(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	storage := &storage.Storage{DB: db}
	svc := NewService(storage)

	teamName := "test-team"
	members := []models.TeamMember{
		{UserId: "1", Username: "user1", IsActive: true},
		{UserId: "2", Username: "user2", IsActive: false},
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO teams").WithArgs(teamName).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO users").WithArgs("1", "user1", teamName, true).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO users").WithArgs("2", "user2", teamName, false).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = svc.AddTeam(teamName, members)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_GetTeam(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	storage := &storage.Storage{DB: db}
	svc := NewService(storage)

	teamName := "test-team"
	expectedMembers := []models.TeamMember{
		{UserId: "1", Username: "user1", IsActive: true},
		{UserId: "2", Username: "user2", IsActive: false},
	}

	rows := sqlmock.NewRows([]string{"user_id", "username", "is_active"}).
		AddRow("1", "user1", true).
		AddRow("2", "user2", false)
	mock.ExpectQuery("SELECT user_id, username, is_active FROM users WHERE team_name = \\$1").WithArgs(teamName).WillReturnRows(rows)

	members, err := svc.GetTeam(teamName)
	assert.NoError(t, err)
	assert.Equal(t, expectedMembers, members)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_CreatePR(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	storage := &storage.Storage{DB: db}
	svc := NewService(storage)

	prId := "pr1"
	prName := "Test PR"
	authorId := "author1"

	authorRows := sqlmock.NewRows([]string{"user_id", "username", "team_name", "is_active"}).
		AddRow("author1", "author", "team1", true)
	mock.ExpectQuery("SELECT user_id, username, team_name, is_active FROM users WHERE user_id = \\$1").WithArgs(authorId).WillReturnRows(authorRows)

	teamRows := sqlmock.NewRows([]string{"user_id", "username", "team_name", "is_active"}).
		AddRow("rev1", "rev1", "team1", true).
		AddRow("rev2", "rev2", "team1", true)
	mock.ExpectQuery("SELECT user_id, username, team_name, is_active FROM users WHERE team_name = \\$1").WithArgs("team1").WillReturnRows(teamRows)

	mock.ExpectExec("INSERT INTO pull_requests").WithArgs(prId, prName, authorId, "OPEN", sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))

	pr, err := svc.CreatePR(prId, prName, authorId)
	assert.NoError(t, err)
	assert.Equal(t, prId, pr.PullRequestId)
	assert.Equal(t, prName, pr.PullRequestName)
	assert.Equal(t, authorId, pr.AuthorId)
	assert.Equal(t, "OPEN", pr.Status)
	assert.Len(t, pr.AssignedReviewers, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_MergePR(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	storage := &storage.Storage{DB: db}
	svc := NewService(storage)

	prId := "pr1"
	createdAt := time.Now()

	prRows := sqlmock.NewRows([]string{"pull_request_id", "pull_request_name", "author_id", "status", "assigned_reviewers", "created_at", "merged_at"}).
		AddRow(prId, "Test PR", "author1", "OPEN", `["rev1", "rev2"]`, createdAt, nil)
	mock.ExpectQuery("SELECT pull_request_id, pull_request_name, author_id, status, assigned_reviewers, created_at, merged_at FROM pull_requests WHERE pull_request_id = \\$1").WithArgs(prId).WillReturnRows(prRows)

	mock.ExpectExec("UPDATE pull_requests SET pull_request_name = \\$1, author_id = \\$2, status = \\$3, assigned_reviewers = \\$4, merged_at = \\$5 WHERE pull_request_id = \\$6").
		WithArgs("Test PR", "author1", "MERGED", sqlmock.AnyArg(), sqlmock.AnyArg(), prId).WillReturnResult(sqlmock.NewResult(1, 1))

	pr, err := svc.MergePR(prId)
	assert.NoError(t, err)
	assert.Equal(t, "MERGED", pr.Status)
	assert.NotNil(t, pr.MergedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_ReassignPR(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	storage := &storage.Storage{DB: db}
	svc := NewService(storage)

	prId := "pr1"
	oldUserId := "rev1"
	createdAt := time.Now()

	prRows := sqlmock.NewRows([]string{"pull_request_id", "pull_request_name", "author_id", "status", "assigned_reviewers", "created_at", "merged_at"}).
		AddRow(prId, "Test PR", "author1", "OPEN", `["rev1", "rev2"]`, createdAt, nil)
	mock.ExpectQuery("SELECT pull_request_id, pull_request_name, author_id, status, assigned_reviewers, created_at, merged_at FROM pull_requests WHERE pull_request_id = \\$1").WithArgs(prId).WillReturnRows(prRows)

	userRows := sqlmock.NewRows([]string{"user_id", "username", "team_name", "is_active"}).
		AddRow("rev1", "rev1", "team1", true)
	mock.ExpectQuery("SELECT user_id, username, team_name, is_active FROM users WHERE user_id = \\$1").WithArgs(oldUserId).WillReturnRows(userRows)

	teamRows := sqlmock.NewRows([]string{"user_id", "username", "team_name", "is_active"}).
		AddRow("rev3", "rev3", "team1", true)
	mock.ExpectQuery("SELECT user_id, username, team_name, is_active FROM users WHERE team_name = \\$1").WithArgs("team1").WillReturnRows(teamRows)

	mock.ExpectExec("UPDATE pull_requests SET pull_request_name = \\$1, author_id = \\$2, status = \\$3, assigned_reviewers = \\$4, merged_at = \\$5 WHERE pull_request_id = \\$6").
		WithArgs("Test PR", "author1", "OPEN", sqlmock.AnyArg(), nil, prId).WillReturnResult(sqlmock.NewResult(1, 1))

	pr, newReviewer, err := svc.ReassignPR(prId, oldUserId)
	assert.NoError(t, err)
	assert.Equal(t, "OPEN", pr.Status)
	assert.Contains(t, pr.AssignedReviewers, "rev2")
	assert.Contains(t, pr.AssignedReviewers, newReviewer)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_DeactivateTeam(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	storage := &storage.Storage{DB: db}
	svc := NewService(storage)

	teamName := "test-team"

	usersRows := sqlmock.NewRows([]string{"user_id", "username", "team_name", "is_active"}).
		AddRow("user1", "user1", teamName, true).
		AddRow("user2", "user2", teamName, false)
	mock.ExpectQuery("SELECT user_id, username, team_name, is_active FROM users WHERE team_name = \\$1").WithArgs(teamName).WillReturnRows(usersRows)

	mock.ExpectExec("UPDATE users SET is_active = false WHERE team_name = \\$1").WithArgs(teamName).WillReturnResult(sqlmock.NewResult(0, 2))

	err = svc.DeactivateTeam(teamName)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_GetAssignmentStats(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	storage := &storage.Storage{DB: db}
	svc := NewService(storage)

	expectedStats := []models.AssignmentStat{
		{UserId: "user1", Count: 5},
		{UserId: "user2", Count: 3},
	}

	rows := sqlmock.NewRows([]string{"user_id", "count"}).
		AddRow("user1", 5).
		AddRow("user2", 3)
	mock.ExpectQuery("SELECT elem AS user_id, COUNT\\(\\*\\) AS count FROM pull_requests pr CROSS JOIN jsonb_array_elements_text\\(pr.assigned_reviewers\\) AS elem GROUP BY elem ORDER BY count DESC").WillReturnRows(rows)

	stats, err := svc.GetAssignmentStats()
	assert.NoError(t, err)
	assert.Equal(t, expectedStats, stats)
	assert.NoError(t, mock.ExpectationsWereMet())
}
