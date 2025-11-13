package main

import (
	"avito-internship/internal/api"
	"avito-internship/internal/service"
	"avito-internship/internal/storage"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	storage := storage.NewStorage()
	service := service.NewService(storage)
	handlers := api.NewHandlers(service)

	api.RegisterHandlers(e, handlers)

	e.Logger.Fatal(e.Start(":8080"))
}
