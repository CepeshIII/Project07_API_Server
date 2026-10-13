package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/CepeshIII/Project07_API_Server/internal/credential"
	"github.com/CepeshIII/Project07_API_Server/internal/httputils"
	"github.com/CepeshIII/Project07_API_Server/internal/store"
)

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,max=100" default:"user"`
	Email    string `json:"email" validate:"required,email,max=255"  default:"mail@example.com"`
	Password string `json:"password" validate:"required,min=3,max=72"  default:"password"`
}

type LoginUserPayload struct {
	Username string `json:"username" validate:"required,max=100" default:"user"`
	Password string `json:"password" validate:"required,min=3,max=72"  default:"password"`
}

type RegisterUserResponse struct {
	User  *store.User `json:"user"`
	Token string      `json:"verification_token"`
}

// RegisterUserHandler godoc
//
//	@Summary		Registers a user
//	@Description	Registers a user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		RegisterUserPayload	true	"User credentials"
//	@Success		201		{object}	RegisterUserResponseEnvelope
//	@Failure		400		{object}	ErrorEnvelope
//	@Failure		409		{object}	ErrorEnvelope
//	@Failure		500		{object}	ErrorEnvelope
//	@Router			/auth/register  [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := app.validator.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	userWithRole := &store.UserWithRole{
		User: store.User{
			Username: payload.Username,
			Email:    payload.Email,
		},
		Role: store.Role{
			Name: "user",
		},
	}

	// hash the user password
	if err := userWithRole.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()

	plainToken, tokenHash := credential.New()

	// store the user
	if err := app.store.Users.CreateAndInvite(ctx, userWithRole, tokenHash, app.config.Mail.Exp); err != nil {
		if errors.Is(err, store.ErrDuplicateEmail) || errors.Is(err, store.ErrDuplicateUsername) {
			app.conflictResponse(w, r, err)
			return
		}

		app.internalServerError(w, r, err)

		return
	}

	if err := app.sendUserActivation(r.Context(), plainToken, userWithRole); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	response := &RegisterUserResponse{
		User:  &userWithRole.User,
		Token: plainToken,
	}

	// send the response
	if err := app.jsonResponse(w, http.StatusCreated, response); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// LoginUserHandler godoc
//
//	@Summary		Logins a user
//	@Description	Logins a user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		LoginUserPayload	true	"User credentials"
//	@Success		200		{object}	RegisterUserResponseEnvelope
//	@Failure		401		{object}	ErrorEnvelope
//	@Failure		400		{object}	ErrorEnvelope
//	@Failure		500		{object}	ErrorEnvelope
//	@Router			/auth/login  [post]
func (app *application) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload LoginUserPayload

	if err := readJSON(w, r, &payload); err != nil {
		app.clearSessionCookie(w)
		app.badRequestResponse(w, r, err)

		return
	}

	if err := app.validator.Struct(payload); err != nil {
		app.clearSessionCookie(w)
		app.badRequestResponse(w, r, err)

		return
	}

	ctx := r.Context()

	// check the user
	user, err := app.store.Users.GetByUsername(ctx, payload.Username)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.clearSessionCookie(w)
			app.statusUnauthorizedError(w, r, errInvalidUsernameOrPassword)

			return
		}

		app.internalServerError(w, r, err)

		return
	}

	// check the user password
	if !user.Password.CheckHash(payload.Password) {
		app.statusUnauthorizedError(w, r, errInvalidUsernameOrPassword)
		return
	}

	token, tokenHash := credential.New()

	// create new session
	session := store.SessionData{
		UserID:    user.ID,
		TokenHash: tokenHash,
		Token:     token,
		UserAgent: r.Header.Get("User-Agent"),
		IPAdress:  httputils.GetClientIP(r),
		IsRevoke:  false,
		ExpiresAt: time.Now().Add(app.config.Auth.Tokens.SessionTokenExp),
	}

	// store session in DB
	if err := app.store.Sessions.CreateSession(ctx, &session); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	// set new Session Token to coockie
	setSessionTokenCookie(w, session)

	// create new Access Token
	accesstoken, accessTokenExp, err := app.generateAccessToken(user.ID, "user")
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	// set new Access Token to coockie
	setAccessTokenCookie(w, accesstoken, accessTokenExp)

	response := &RegisterUserResponse{
		User: user,
		// Token: token,
	}
	if err := app.jsonResponse(w, http.StatusOK, response); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// RefreshAccessTokenHandler godoc
//
//	@Summary		Refreshes Access Token
//	@Description	Refreshes Access Token
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	MessageEnvelope
//	@Failure		400	{object}	ErrorEnvelope
//	@Failure		401	{object}	ErrorEnvelope
//	@Failure		500	{object}	ErrorEnvelope
//	@Router			/auth/refresh [post]
func (app *application) refreshAccessTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Try to get session token from the cookie
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			app.statusUnauthorizedError(w, r, errSessionTokenExpired)
			return
		}

		app.badRequestResponse(w, r, err)

		return
	}

	sessionToken := cookie.Value
	tokenHash := credential.Hash(sessionToken)

	// Check session token in the database
	session := store.SessionData{
		TokenHash: tokenHash,
	}

	if err := app.store.Sessions.GetSessionByTokenHash(r.Context(), &session); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.clearSessionCookie(w)
			app.statusUnauthorizedError(w, r, errSessionNotFoundOrInvalid)

			return
		}

		app.internalServerError(w, r, err)

		return
	}

	if session.IsRevoke {
		app.clearSessionCookie(w)
		app.statusUnauthorizedError(w, r, errSessionRevoked)

		return
	}

	if time.Now().After(session.ExpiresAt) {
		app.clearSessionCookie(w)
		app.statusUnauthorizedError(w, r, errSessionExpired)

		return
	}

	// create new Access Token
	accesstoken, accessTokenExp, err := app.generateAccessToken(session.UserID, "user")
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	// set new Access Token to coockie
	setAccessTokenCookie(w, accesstoken, accessTokenExp)

	if err := app.jsonResponse(w, http.StatusOK, "Access token refreshed successfully"); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
