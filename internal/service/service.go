package service

import (
	"avito-internship/internal/models"
	"avito-internship/internal/storage"
	"errors"
	"math/rand"
	"time"
)

type Service struct {
	storage *storage.Storage
}

func NewService(storage *storage.Storage) *Service {
	return &Service{storage: storage}
}

func (s *Service) AddTeam(teamName string, members []models.TeamMember) error {
	return s.storage.AddTeam(teamName, members)
}

func (s *Service) GetTeam(teamName string) ([]models.TeamMember, error) {
	return s.storage.GetTeam(teamName)
}

func (s *Service) SetUserActive(userId string, isActive bool) error {
	user, err := s.storage.GetUser(userId)
	if err != nil {
		return err
	}
	if user.IsActive && !isActive {
		// Переназначить все открытые PR, где он ревьювер
		prs, err := s.storage.GetPRsByReviewer(userId)
		if err != nil {
			return err
		}
		for _, prShort := range prs {
			if prShort.Status == "OPEN" {
				_, _, err := s.ReassignPR(prShort.PullRequestId, userId)
				if err != nil {
					// Если переназначение невозможно, пропустить, но продолжить деактивацию
					continue
				}
			}
		}
	}
	return s.storage.SetUserActive(userId, isActive)
}

func (s *Service) CreatePR(prId, prName, authorId string) (models.PullRequest, error) {
	author, err := s.storage.GetUser(authorId)
	if err != nil {
		return models.PullRequest{}, err
	}
	teamUsers, err := s.storage.GetUsersByTeam(author.TeamName)
	if err != nil {
		return models.PullRequest{}, err
	}
	var activeReviewers []models.User
	for _, u := range teamUsers {
		if u.IsActive && u.UserId != authorId {
			activeReviewers = append(activeReviewers, u)
		}
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(activeReviewers), func(i, j int) {
		activeReviewers[i], activeReviewers[j] = activeReviewers[j], activeReviewers[i]
	})
	var reviewers []string
	for i := 0; i < len(activeReviewers) && i < 2; i++ {
		reviewers = append(reviewers, activeReviewers[i].UserId)
	}
	now := time.Now()
	pr := models.PullRequest{
		PullRequestId:     prId,
		PullRequestName:   prName,
		AuthorId:          authorId,
		Status:            "OPEN",
		AssignedReviewers: reviewers,
		CreatedAt:         &now,
	}
	err = s.storage.CreatePR(pr)
	if err != nil {
		return models.PullRequest{}, err
	}
	return pr, nil
}

func (s *Service) MergePR(prId string) (models.PullRequest, error) {
	pr, err := s.storage.GetPR(prId)
	if err != nil {
		return models.PullRequest{}, err
	}
	if pr.Status == "MERGED" {
		return pr, nil
	}
	pr.Status = "MERGED"
	now := time.Now()
	pr.MergedAt = &now
	return pr, s.storage.UpdatePR(pr)
}

func (s *Service) ReassignPR(prId, oldUserId string) (models.PullRequest, string, error) {
	pr, err := s.storage.GetPR(prId)
	if err != nil {
		return models.PullRequest{}, "", err
	}
	if pr.Status == "MERGED" {
		return models.PullRequest{}, "", errors.New("cannot reassign on merged PR")
	}
	found := false
	for i, r := range pr.AssignedReviewers {
		if r == oldUserId {
			pr.AssignedReviewers = append(pr.AssignedReviewers[:i], pr.AssignedReviewers[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return models.PullRequest{}, "", errors.New("reviewer is not assigned to this PR")
	}
	oldUser, err := s.storage.GetUser(oldUserId)
	if err != nil {
		return models.PullRequest{}, "", err
	}
	teamUsers, err := s.storage.GetUsersByTeam(oldUser.TeamName)
	if err != nil {
		return models.PullRequest{}, "", err
	}
	var activeReviewers []models.User
	for _, u := range teamUsers {
		if u.IsActive && u.UserId != pr.AuthorId {
			assigned := false
			for _, ar := range pr.AssignedReviewers {
				if ar == u.UserId {
					assigned = true
					break
				}
			}
			if !assigned {
				activeReviewers = append(activeReviewers, u)
			}
		}
	}
	if len(activeReviewers) == 0 {
		return models.PullRequest{}, "", errors.New("no active replacement candidate in team")
	}
	rand.Seed(time.Now().UnixNano())
	newReviewer := activeReviewers[rand.Intn(len(activeReviewers))]
	pr.AssignedReviewers = append(pr.AssignedReviewers, newReviewer.UserId)
	err = s.storage.UpdatePR(pr)
	return pr, newReviewer.UserId, err
}

func (s *Service) GetPRsByReviewer(userId string) ([]models.PullRequestShort, error) {
	_, err := s.storage.GetUser(userId)
	if err != nil {
		return nil, err
	}
	return s.storage.GetPRsByReviewer(userId)
}

func (s *Service) GetAssignmentStats() ([]models.AssignmentStat, error) {
	return s.storage.GetAssignmentStats()
}

func (s *Service) DeactivateTeam(teamName string) error {
	users, err := s.storage.GetUsersByTeam(teamName)
	if err != nil {
		return err
	}
	// Для каждого активного пользователя переназначить его PR
	for _, user := range users {
		if user.IsActive {
			prs, err := s.storage.GetPRsByReviewer(user.UserId)
			if err != nil {
				continue
			}
			for _, prShort := range prs {
				if prShort.Status == "OPEN" {
					_, _, err := s.ReassignPR(prShort.PullRequestId, user.UserId)
					if err != nil {
						continue // Пропустить если переназначение невозможно
					}
				}
			}
		}
	}
	// Деактивировать всех пользователей команды
	return s.storage.DeactivateTeam(teamName)
}
