package api

import (
	"avito-internship/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handlers struct {
	service *service.Service
}

func NewHandlers(service *service.Service) *Handlers {
	return &Handlers{service: service}
}

func (h *Handlers) PostPullRequestCreate(ctx echo.Context) error {
	return ctx.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (h *Handlers) PostPullRequestMerge(ctx echo.Context) error {
	return ctx.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (h *Handlers) PostPullRequestReassign(ctx echo.Context) error {
	return ctx.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (h *Handlers) PostTeamAdd(ctx echo.Context) error {
	return ctx.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (h *Handlers) GetTeamGet(ctx echo.Context, params GetTeamGetParams) error {
	return ctx.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (h *Handlers) GetUsersGetReview(ctx echo.Context, params GetUsersGetReviewParams) error {
	return ctx.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (h *Handlers) PostUsersSetIsActive(ctx echo.Context) error {
	return ctx.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}
