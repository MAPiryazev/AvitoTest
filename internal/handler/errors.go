package handler

import (
	"AvitoTest/internal/api"
	customerrors "AvitoTest/internal/custom-errors"
	"encoding/json"
	"errors"
	"net/http"
)

func writeAPIError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	var resp api.ErrorResponse

	// маппинг ошибок в коды
	switch {
	case errors.Is(err, customerrors.ErrNotFound):
		w.WriteHeader(http.StatusNotFound)
		resp.Error.Code = api.NOTFOUND
		resp.Error.Message = err.Error()

	case errors.Is(err, customerrors.ErrAlreadyExists):
		w.WriteHeader(http.StatusConflict)
		resp.Error.Code = api.PREXISTS
		resp.Error.Message = err.Error()

	case errors.Is(err, customerrors.ErrForbidden):
		w.WriteHeader(http.StatusForbidden)
		resp.Error.Code = api.NOTASSIGNED
		resp.Error.Message = err.Error()

	case errors.Is(err, customerrors.ErrValidation):
		w.WriteHeader(http.StatusBadRequest)
		resp.Error.Code = api.NOCANDIDATE
		resp.Error.Message = err.Error()

	default:
		w.WriteHeader(http.StatusInternalServerError)
		resp.Error.Code = api.NOTFOUND
		resp.Error.Message = err.Error()
	}

	_ = json.NewEncoder(w).Encode(resp)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
