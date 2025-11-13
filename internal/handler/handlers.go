package handler

import (
	"avito-internship/api"
	"avito-internship/internal/models"
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

func (h *Handlers) PostTeamAdd(ctx echo.Context) error {
	var req api.PostTeamAddJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: struct {
			Code    api.ErrorResponseErrorCode `json:"code"`
			Message string                     `json:"message"`
		}{Code: api.NOTFOUND, Message: "invalid request"}})
	}
	members := make([]models.TeamMember, len(req.Members))
	for i, m := range req.Members {
		members[i] = models.TeamMember{
			UserId:   m.UserId,
			Username: m.Username,
			IsActive: m.IsActive,
		}
	}
	err := h.service.AddTeam(req.TeamName, members)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: struct {
			Code    api.ErrorResponseErrorCode `json:"code"`
			Message string                     `json:"message"`
		}{Code: api.TEAMEXISTS, Message: err.Error()}})
	}
	return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handlers) GetTeamGet(ctx echo.Context, params api.GetTeamGetParams) error {
	members, err := h.service.GetTeam(params.TeamName)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, api.ErrorResponse{Error: struct {
			Code    api.ErrorResponseErrorCode `json:"code"`
			Message string                     `json:"message"`
		}{Code: api.NOTFOUND, Message: err.Error()}})
	}
	resp := api.Team{
		TeamName: params.TeamName,
		Members:  make([]api.TeamMember, len(members)),
	}
	for i, m := range members {
		resp.Members[i] = api.TeamMember{
			UserId:   m.UserId,
			Username: m.Username,
			IsActive: m.IsActive,
		}
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (h *Handlers) PostUsersSetIsActive(ctx echo.Context) error {
	var req api.PostUsersSetIsActiveJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: struct {
			Code    api.ErrorResponseErrorCode `json:"code"`
			Message string                     `json:"message"`
		}{Code: api.NOTFOUND, Message: "invalid request"}})
	}
	err := h.service.SetUserActive(req.UserId, req.IsActive)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, api.ErrorResponse{Error: struct {
			Code    api.ErrorResponseErrorCode `json:"code"`
			Message string                     `json:"message"`
		}{Code: api.NOTFOUND, Message: err.Error()}})
	}
	return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handlers) PostPullRequestCreate(ctx echo.Context) error {
	var req api.PostPullRequestCreateJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: struct {
			Code    api.ErrorResponseErrorCode `json:"code"`
			Message string                     `json:"message"`
		}{Code: api.NOTFOUND, Message: "invalid request"}})
	}
	pr, err := h.service.CreatePR(req.PullRequestId, req.PullRequestName, req.AuthorId)
	if err != nil {
		if err.Error() == "user not found" {
			return ctx.JSON(http.StatusNotFound, api.ErrorResponse{Error: struct {
				Code    api.ErrorResponseErrorCode `json:"code"`
				Message string                     `json:"message"`
			}{Code: api.NOTFOUND, Message: err.Error()}})
		}
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: struct {
			Code    api.ErrorResponseErrorCode `json:"code"`
			Message string                     `json:"message"`
		}{Code: api.PREXISTS, Message: err.Error()}})
	}
	resp := api.PullRequest{
		PullRequestId:     pr.PullRequestId,
		PullRequestName:   pr.PullRequestName,
		AuthorId:          pr.AuthorId,
		Status:            api.PullRequestStatus(pr.Status),
		AssignedReviewers: pr.AssignedReviewers,
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (h *Handlers) PostPullRequestMerge(ctx echo.Context) error {
	var req api.PostPullRequestMergeJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: struct {
			Code    api.ErrorResponseErrorCode `json:"code"`
			Message string                     `json:"message"`
		}{Code: api.NOTFOUND, Message: "invalid request"}})
	}
	pr, err := h.service.MergePR(req.PullRequestId)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, api.ErrorResponse{Error: struct {
			Code    api.ErrorResponseErrorCode `json:"code"`
			Message string                     `json:"message"`
		}{Code: api.NOTFOUND, Message: err.Error()}})
	}
	resp := api.PullRequest{
		PullRequestId:     pr.PullRequestId,
		PullRequestName:   pr.PullRequestName,
		AuthorId:          pr.AuthorId,
		Status:            api.PullRequestStatus(pr.Status),
		AssignedReviewers: pr.AssignedReviewers,
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (h *Handlers) PostPullRequestReassign(ctx echo.Context) error {
	var req api.PostPullRequestReassignJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: struct {
			Code    api.ErrorResponseErrorCode `json:"code"`
			Message string                     `json:"message"`
		}{Code: api.NOTFOUND, Message: "invalid request"}})
	}
	pr, newReviewer, err := h.service.ReassignPR(req.PullRequestId, req.OldUserId)
	if err != nil {
		code := api.NOTFOUND
		if err.Error() == "cannot reassign on merged PR" {
			code = api.PRMERGED
		} else if err.Error() == "reviewer is not assigned to this PR" {
			code = api.NOTASSIGNED
		} else if err.Error() == "no active replacement candidate in team" {
			code = api.NOCANDIDATE
		}
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{Error: struct {
			Code    api.ErrorResponseErrorCode `json:"code"`
			Message string                     `json:"message"`
		}{Code: code, Message: err.Error()}})
	}
	resp := struct {
		Pr         api.PullRequest `json:"pr"`
		ReplacedBy string          `json:"replaced_by"`
	}{
		Pr: api.PullRequest{
			PullRequestId:     pr.PullRequestId,
			PullRequestName:   pr.PullRequestName,
			AuthorId:          pr.AuthorId,
			Status:            api.PullRequestStatus(pr.Status),
			AssignedReviewers: pr.AssignedReviewers,
		},
		ReplacedBy: newReviewer,
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (h *Handlers) GetUsersGetReview(ctx echo.Context, params api.GetUsersGetReviewParams) error {
	prs, _ := h.service.GetPRsByReviewer(params.UserId)
	resp := struct {
		UserId       string                 `json:"user_id"`
		PullRequests []api.PullRequestShort `json:"pull_requests"`
	}{
		UserId:       params.UserId,
		PullRequests: make([]api.PullRequestShort, len(prs)),
	}
	for i, pr := range prs {
		resp.PullRequests[i] = api.PullRequestShort{
			PullRequestId:   pr.PullRequestId,
			PullRequestName: pr.PullRequestName,
			AuthorId:        pr.AuthorId,
			Status:          api.PullRequestShortStatus(pr.Status),
		}
	}
	return ctx.JSON(http.StatusOK, resp)
}
