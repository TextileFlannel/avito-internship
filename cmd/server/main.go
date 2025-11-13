package main

import (
	"fmt"

	"avito-internship/api"
	"avito-internship/internal/config"
	"avito-internship/internal/handler"
	"avito-internship/internal/service"
	"avito-internship/internal/storage"
	"avito-internship/migrations"

	"github.com/labstack/echo/v4"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	e := echo.New()

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	storage, err := storage.NewStorage(connStr)
	if err != nil {
		panic(err)
	}
	if err := migrations.RunMigrations(storage.DB); err != nil {
		panic(err)
	}
	service := service.NewService(storage)
	handlers := handler.NewHandlers(service)

	api.RegisterHandlers(e, handlers)

	e.Logger.Fatal(e.Start(":" + cfg.HTTPPort))
}
