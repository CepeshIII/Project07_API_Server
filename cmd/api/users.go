package main

import (
	"context"
	"course/api_server/internal/store"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type userKey string

const userCtx userKey = "user"

type FollowUserPayload struct {
	UserID int64 `json:"user_id"`
}

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	followingUser := getUserFromCtx(r)

	// Try read updated post data from request
	var payload FollowUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Check if the user that follow exists
	_, err := app.store.Users.Get(r.Context(), payload.UserID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	err = app.store.Followers.FollowUser(r.Context(), followingUser.ID, payload.UserID)
	if err != nil {
		if err == store.ErrorConflict {
			app.conflictResponse(w, r, err)
			return
		}

		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, "OK"); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) getFollowersHandler(w http.ResponseWriter, r *http.Request) {
	followingUser := getUserFromCtx(r)

	followers, err := app.store.Followers.GetFollowers(r.Context(), followingUser.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, followers); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)

	// Try read updated post data from request
	var payload FollowUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Check if the user that follow exists
	_, err := app.store.Users.Get(r.Context(), payload.UserID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	err = app.store.Followers.UnfollowUser(r.Context(), user.ID, payload.UserID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, "OK"); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUserID(r)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		ctx := r.Context()
		user, err := app.store.Users.Get(ctx, id)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundResponse(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserFromCtx(r *http.Request) *store.User {
	user, _ := r.Context().Value(userCtx).(*store.User)
	return user
}

func parseUserID(r *http.Request) (int64, error) {
	idParam := chi.URLParam(r, "userID")
	return strconv.ParseInt(idParam, 10, 64)
}
