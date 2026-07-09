package main

import (
	"course/api_server/internal/store"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserID(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user, err := app.store.Users.Get(r.Context(), userID)
	if err != nil {
		switch {
		case err == store.ErrNotFound:
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func parseUserID(r *http.Request) (int64, error) {
	idParam := chi.URLParam(r, "userID")
	return strconv.ParseInt(idParam, 10, 64)
}
