package service

import (
	"avito-internship/internal/storage"
)

type Service struct {
	storage *storage.Storage
}

func NewService(storage *storage.Storage) *Service {
	return &Service{storage: storage}
}
