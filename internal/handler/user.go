package handler

import (
	"AvitoTest/internal/adapter"
	"AvitoTest/internal/api"
	"encoding/json"
	"net/http"
)

func (h *handler) GetUsersGetReview(w http.ResponseWriter, r *http.Request, params api.GetUsersGetReviewParams) {
	// Получаем пользователя по user_id (строке) для получения внутреннего ID
	user, err := h.services.User.GetUserByPublicID(r.Context(), params.UserId)
	if err != nil {
		writeAPIError(w, err)
		return
	}

	prs, err := h.services.PR.GetPRsForReviewer(r.Context(), user.ID)
	if err != nil {
		writeAPIError(w, err)
		return
	}

	out := make([]api.PullRequestShort, len(prs))
	for i, p := range prs {
		out[i] = adapter.PullRequestShortModelToAPI(p)
	}

	writeJSON(w, http.StatusOK, out)
}

func (h *handler) PostUsersSetIsActive(w http.ResponseWriter, r *http.Request) {
	var body api.PostUsersSetIsActiveJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, err)
		return
	}

	if err := h.services.User.SetUserActive(r.Context(), body.UserId, body.IsActive); err != nil {
		writeAPIError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
