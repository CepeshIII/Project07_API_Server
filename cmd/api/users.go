package main

import (
	"errors"
	"net/http"

	"github.com/CepeshIII/Project07_API_Server/internal/store"

	"github.com/go-chi/chi/v5"
)

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
	targetUser := getTargetUserFromCtx(r)
	if targetUser == nil {
		// This should not happen if the middleware is working correctly
		app.internalServerError(w, r, errTargetUserMissing)
	}

	if err := app.jsonResponse(w, http.StatusOK, targetUser); err != nil {
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
//	@Param			userID	path		int				true	"User ID"
//	@Success		201		{object}	MessageEnvelope	"User followed"
//	@Failure		400		{object}	ErrorEnvelope	"User not found"
//	@Failure		404		{object}	ErrorEnvelope	"User payload missing"
//	@Failure		409		{object}	ErrorEnvelope	"Status Conflict"
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}/follow [put]
func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	// The user targeted in the URL path: /users/{userID}/follow
	targetUser := getTargetUserFromCtx(r)
	if targetUser == nil {
		// This should not happen if the middleware is working correctly
		app.internalServerError(w, r, errTargetUserMissing)
		return
	}

	// The authenticated user performing the action
	followerID, err := getAuthUserIDFromCtx(r)
	if err != nil {
		// This should not happen if the middleware is working correctly
		app.internalServerError(w, r, errAuthUserMissing)
		return
	}

	// Prevent user from following themselves
	if targetUser.ID == followerID {
		app.badRequestResponse(w, r, errors.New("you cannot follow yourself"))
		return
	}

	// create a relationship where followerID follows targetUser.ID
	err = app.store.Followers.FollowUser(r.Context(), followerID, targetUser.ID)
	if err != nil {
		if errors.Is(err, store.ErrorConflict) {
			app.conflictResponse(w, r, err)
			return
		}

		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, "User followed"); err != nil {
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
	// Get the user that the current user wants to follow from the URL parameter
	currentUser := getTargetUserFromCtx(r)

	if currentUser == nil {
		// This should not happen if the middleware is working correctly
		app.internalServerError(w, r, errTargetUserMissing)
		return
	}

	followers, err := app.store.Followers.GetFollowers(r.Context(), currentUser.ID)
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
//	@Param			userID	path		int				true	"User ID"
//	@Success		201		{object}	MessageEnvelope	"User unfollowed"
//	@Failure		400		{object}	ErrorEnvelope	"User not found"
//	@Failure		404		{object}	ErrorEnvelope	"User payload missing"
//	@Failure		409		{object}	ErrorEnvelope	"Status Conflict"
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}/unfollow [put]
func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	// The user targeted in the URL path: /users/{userID}/unfollow
	targetUser := getTargetUserFromCtx(r)
	if targetUser == nil {
		// This should not happen if the middleware is working correctly
		app.internalServerError(w, r, errTargetUserMissing)
		return
	}

	// The authenticated user performing the action
	followerID, err := getAuthUserIDFromCtx(r)
	if err != nil {
		// This should not happen if the middleware is working correctly
		app.internalServerError(w, r, errAuthUserMissing)
		return
	}

	// Prevent user from unfollowing themselves
	if targetUser.ID == followerID {
		app.badRequestResponse(w, r, errors.New("you cannot unfollow yourself"))
		return
	}

	// delete a relationship where followerID unfollows targetUser.ID
	err = app.store.Followers.UnfollowUser(r.Context(), followerID, targetUser.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.notFoundResponse(w, r, err)
			return
		}

		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, "User unfollowed"); err != nil {
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
//	@Success		200		{object}	MessageEnvelope	"User activated"
//	@Failure		404		{object}	ErrorEnvelope
//	@Failure		500		{object}	ErrorEnvelope
//
//	@Security		ApiKeyAuth
//	@Router			/users/activate/{token} [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the token from the URL path and hash it for comparison with the stored hash in the database
	token := chi.URLParam(r, "token")
	tokenHash := HashToken(token)

	// Activate the user and clean up the invitation token from the database
	err := app.store.Users.ActivateAndClean(r.Context(), string(tokenHash))
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

	// Respond with a success message
	if err := app.jsonResponse(w, http.StatusOK, "User activated"); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// deleteUser godoc
//
//	@Summary		Deletes a user
//	@Description	Deletes a user by user id
//	@Tags			users
//	@Produce		json
//	@Param			userID	path		string			true	"UserID"
//	@Success		200		{object}	MessageEnvelope	"User deleted"
//	@Failure		500		{object}	ErrorEnvelope
//	@Security		ApiKeyAuth
//	@Router			/users/{userID} [delete]
func (app *application) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	targetUser := getTargetUserFromCtx(r)
	if targetUser == nil {
		app.internalServerError(w, r, errTargetUserMissing)
		return
	}

	err := app.store.Users.DeleteUser(r.Context(), targetUser.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err = app.jsonResponse(w, http.StatusOK, "User deleted"); err != nil {
		app.internalServerError(w, r, err)
	}
}
