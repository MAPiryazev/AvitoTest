package handler

import (
	"AvitoTest/internal/api"
	"AvitoTest/internal/service"
	"net/http"
)

// Handler интерфейс HTTP слоя.
type Handler interface {
	PostPullRequestCreate(w http.ResponseWriter, r *http.Request)
	PostPullRequestMerge(w http.ResponseWriter, r *http.Request)
	PostPullRequestReassign(w http.ResponseWriter, r *http.Request)

	PostTeamAdd(w http.ResponseWriter, r *http.Request)
	GetTeamGet(w http.ResponseWriter, r *http.Request, params api.GetTeamGetParams)

	GetUsersGetReview(w http.ResponseWriter, r *http.Request, params api.GetUsersGetReviewParams)
	PostUsersSetIsActive(w http.ResponseWriter, r *http.Request)
}

// handler реализация интерфейса Handler.
type handler struct {
	services *service.Service
}

// NewHandler возвращает интерфейс Handler.
func NewHandler(s *service.Service) Handler {
	return &handler{services: s}
}
