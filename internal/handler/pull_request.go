package handler

import (
	"AvitoTest/internal/adapter"
	"AvitoTest/internal/api"
	"AvitoTest/internal/models"
	"encoding/json"
	"net/http"
)

func (h *handler) PostPullRequestCreate(w http.ResponseWriter, r *http.Request) {
	var body api.PostPullRequestCreateJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, err)
		return
	}

	// Получаем пользователя по user_id (строке) для получения внутреннего ID
	author, err := h.services.User.GetUserByPublicID(r.Context(), body.AuthorId)
	if err != nil {
		writeAPIError(w, err)
		return
	}

	pr := models.PullRequest{
		PullRequestID: body.PullRequestId,
		Name:          body.PullRequestName,
		AuthorID:      author.ID,
	}

	created, err := h.services.PR.CreatePullRequest(r.Context(), pr)
	if err != nil {
		writeAPIError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, adapter.PullRequestModelToAPI(*created))
}

func (h *handler) PostPullRequestMerge(w http.ResponseWriter, r *http.Request) {
	var body api.PostPullRequestMergeJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, err)
		return
	}

	merged, err := h.services.PR.MergePullRequest(r.Context(), body.PullRequestId)
	if err != nil {
		writeAPIError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, adapter.PullRequestModelToAPI(*merged))
}

func (h *handler) PostPullRequestReassign(w http.ResponseWriter, r *http.Request) {
	var body api.PostPullRequestReassignJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, err)
		return
	}

	// Получаем пользователя по user_id (строке) для получения внутреннего ID
	oldUser, err := h.services.User.GetUserByPublicID(r.Context(), body.OldUserId)
	if err != nil {
		writeAPIError(w, err)
		return
	}

	updated, err := h.services.PR.ReassignReviewer(
		r.Context(),
		body.PullRequestId,
		oldUser.ID,
	)
	if err != nil {
		writeAPIError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, adapter.PullRequestModelToAPI(*updated))
}
