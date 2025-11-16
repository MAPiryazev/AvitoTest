package handler

import (
	"AvitoTest/internal/adapter"
	"AvitoTest/internal/api"
	"AvitoTest/internal/models"
	"encoding/json"
	"net/http"
)

func (h *handler) PostTeamAdd(w http.ResponseWriter, r *http.Request) {
	var body api.Team
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, err)
		return
	}

	team := models.Team{
		Name: body.TeamName,
	}

	// преобразуем участников
	var users []models.User
	for _, m := range body.Members {
		users = append(users, models.User{
			UserID:   m.UserId,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}
	team.Members = users

	if err := h.services.Team.CreateTeam(r.Context(), team); err != nil {
		writeAPIError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *handler) GetTeamGet(w http.ResponseWriter, r *http.Request, params api.GetTeamGetParams) {
	members, err := h.services.Team.GetTeamWithMembers(r.Context(), params.TeamName)
	if err != nil {
		writeAPIError(w, err)
		return
	}

	resp := adapter.TeamWithMembersToAPI(params.TeamName, members)
	writeJSON(w, http.StatusOK, resp)
}
