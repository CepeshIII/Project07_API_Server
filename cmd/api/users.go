package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"

	"github.com/CepeshIII/Project07_API_Server/internal/store"

	"github.com/go-chi/chi/v5"
)

type userKey string

const userCtx userKey = "user"

type FollowUserPayload struct {
	UserID int64 `json:"user_id" example:"1"`
}

// GetUser godoc
//
//	@Summary		Fetches a user profile
//	@Description	Fetches a user profile by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	UserEnvelope
//	@Failure		400	{object}	ErrorEnvelope
//	@Failure		404	{object}	ErrorEnvelope
//	@Failure		500	{object}	ErrorEnvelope
//	@Security		ApiKeyAuth
//	@Router			/users/{id} [get]
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// FollowUser godoc
//
//	@Summary		Follows a user
//	@Description	Follows a user by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int					true	"User ID"
//	@Param			body	body		FollowUserPayload	true	"Follow request"
//	@Success		200		{object}	MessageEnvelope		"User followed"
//	@Failure		400		{object}	ErrorEnvelope		"User not found"
//	@Failure		404		{object}	ErrorEnvelope		"User payload missing"
//	@Failure		409		{object}	ErrorEnvelope		"Status Conflict"
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}/follow [put]
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

	if err := app.jsonResponse(w, http.StatusOK, "User followed"); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// GetFollowers godoc
//
//	@Summary		Get Followers of a user
//	@Description	Get Followers of a user by User ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int	true	"User ID"
//	@Success		200		{object}	FollowersEnvelope
//	@Failure		400		{object}	ErrorEnvelope
//	@Failure		404		{object}	ErrorEnvelope
//	@Failure		500		{object}	ErrorEnvelope
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}/followers [get]
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

// UnfollowUser godoc
//
//	@Summary		unfollow a user
//	@Description	unfollow a user by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int					true	"User ID"
//	@Param			body	body		FollowUserPayload	true	"Unfollow request"
//	@Success		200		{object}	MessageEnvelope		"User unfollowed"
//	@Failure		400		{object}	ErrorEnvelope		"User not found"
//	@Failure		404		{object}	ErrorEnvelope		"User payload missing"
//	@Failure		409		{object}	ErrorEnvelope		"Status Conflict"
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}/unfollow [put]
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

// activateUser godoc
//
//	@Summary		Activates/Registers a user
//	@Description	Activates/Registers a user by invitation token
//	@Tags			users
//	@Produce		json
//	@Param			token	path		string			true	"Invitation Token"
//	@Success		204		{object}	MessageEnvelope	"User activated"
//	@Failure		404		{object}	ErrorEnvelope
//	@Failure		500		{object}	ErrorEnvelope
//
//	@Security		ApiKeyAuth
//	@Router			/users/activate/{token} [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])

	err := app.store.Users.ActivateAndClean(r.Context(), string(hashToken))
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		case errors.Is(err, store.ErrorInvalidToken):
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}
	if err := app.jsonResponse(w, http.StatusCreated, "User activated"); err != nil {
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
