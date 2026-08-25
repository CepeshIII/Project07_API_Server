package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/CepeshIII/Project07_API_Server/internal/auth"
	"github.com/CepeshIII/Project07_API_Server/internal/httputils"
	"github.com/CepeshIII/Project07_API_Server/internal/mailer"
	"github.com/CepeshIII/Project07_API_Server/internal/store"
	"github.com/golang-jwt/jwt/v5"
)

const sessionCookieName = "session_token"
const accessCookieName = "access_token"

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

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
	}

	// hash the user password
	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()

	plainToken, tokenHash, err := NewInvitationToken()
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	// store the user
	if err := app.store.Users.CreateAndInvite(ctx, user, tokenHash, app.config.mail.exp); err != nil {
		if errors.Is(err, store.ErrorDuplicateEmail) || errors.Is(err, store.ErrorDuplicateUsername) {
			app.conflictResponse(w, r, err)
			return
		}

		app.internalServerError(w, r, err)
		return
	}

	activationURL := fmt.Sprintf("%s/confirm/%s", app.config.frontendURL, plainToken)

	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: activationURL,
	}

	compiledTemplate, err := mailer.BuildTemplate(mailer.UserWelcomeTemplate, vars)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	var lastSendErr error
	var successSend bool
	for i := range app.config.mail.maxRetries {
		lastSendErr = app.mailer.Send(r.Context(), user.Username, user.Email, compiledTemplate)

		if lastSendErr == nil {
			successSend = true
			break
		}

		app.logger.Warnw("failed to send activation email",
			"attempt", i+1,
			"max_retries", app.config.mail.maxRetries,
			"email", user.Email,
			"error", lastSendErr,
		)

		// exponential backoff before retrying
		time.Sleep(time.Second * time.Duration(i+1))
		continue

	}

	// send the invitation email
	if !successSend {
		// rollback the user creation if email sending fails (SAGA pattern)
		if rollbackErr := app.store.Users.DeleteUserAndInvitation(ctx, user.ID); rollbackErr != nil {
			app.logger.Errorw("error rolling back user creation", "error", rollbackErr)
		}

		app.internalServerError(w, r, lastSendErr)
		return
	}

	response := &RegisterUserResponse{
		User:  user,
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

	if err := Validate.Struct(payload); err != nil {
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
			app.statusUnauthorizedError(w, r, errors.New("Invalid Username or password"))
			return
		}

		app.internalServerError(w, r, err)
		return
	}

	// check the user password
	if !user.Password.CheckHash(payload.Password) {
		app.statusUnauthorizedError(w, r, errors.New("Invalid Username or password"))
		return
	}

	token, tokenHash, err := NewSessionToken()
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	session := store.SessionData{
		UserID:    user.ID,
		TokenHash: tokenHash,
		UserAgent: r.Header.Get("User-Agent"),
		IPAdress:  httputils.GetClientIP(r),
		IsRevoke:  false,
		ExpiresAt: time.Now().Add(app.config.auth.tokens.sessionTokenExp),
	}

	if err := app.store.Sessions.CreateSession(ctx, &session); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	cookie := http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Переконайся, що використовуєш HTTPS (у локальній розробці без TLS браузер може відхилити Secure куку!)
		SameSite: http.SameSiteLaxMode,
		Expires:  session.ExpiresAt,
	}

	response := &RegisterUserResponse{
		User: user,
		// Token: token,
	}

	http.SetCookie(w, &cookie)

	// set new accessToken
	if err := app.setAccessToken(w, session.UserID); err != nil {
		app.internalServerError(w, r, err)
		return
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
			app.statusUnauthorizedError(w, r, errors.New("Session token has expired. Please refresh your session."))
			return
		}
		app.badRequestResponse(w, r, err)
		return
	}
	sessionToken := cookie.Value
	tokenHash := HashToken(sessionToken)

	// Check session token in the database
	session := store.SessionData{
		TokenHash: tokenHash,
	}

	if err := app.store.Sessions.GetSessionByTokenHash(r.Context(), &session); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.clearSessionCookie(w)
			app.statusUnauthorizedError(w, r, errors.New("Session not found or invalid. Please log in again."))
			return
		}
		app.internalServerError(w, r, err)
		return
	}

	if session.IsRevoke {
		app.clearSessionCookie(w)
		app.statusUnauthorizedError(w, r, errors.New("Session has been revoked. Please log in again."))
		return
	}

	if time.Now().After(session.ExpiresAt) {
		app.clearSessionCookie(w)
		app.statusUnauthorizedError(w, r, errors.New("Session has expired. Please log in again."))
		return
	}

	// Send a new access token
	if err := app.setAccessToken(w, session.UserID); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, "Access token refreshed successfully"); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	app.clearAccessCookie(w)
}

func (app *application) clearAccessCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     accessCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func (app *application) acssesTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to get acsees token from the cookie
		cookie, err := r.Cookie(accessCookieName)
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				app.statusUnauthorizedError(w, r, errors.New("Access token has expired. Please refresh your session."))
				return
			}

			app.badRequestResponse(w, r, err)
			return
		}

		// Validate JWT access token and extract claims
		// claims, err := ValidateAccessToken(cookie.Value)

		claims, err := app.auth.ValidateToken(cookie.Value)
		if err != nil {
			app.clearAccessCookie(w)
			app.statusUnauthorizedError(w, r, err)
			return
		}

		// Inject authenticated user ID into request context
		ctx := context.WithValue(r.Context(), userIDCtx, claims.UserID)

		// Pass execution to next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

func (app *application) setAccessToken(w http.ResponseWriter, userID int64) error {
	exp := time.Now().Add(app.config.auth.tokens.accessTokenExp)

	claims := auth.CustomClaims{
		UserID: userID,
		Role:   "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    app.config.auth.jwtAuth.iss,
			Audience: jwt.ClaimStrings{
				app.config.auth.jwtAuth.iss,
			},
		},
	}

	// claims := jwt.MapClaims{
	// 	"sub":  userID,
	// 	"role": "admin",
	// 	"exp":  time.Now().Add(app.config.auth.tokens.accessTokenExp).Unix(),
	// 	"nbf":  time.Now().Unix(),
	// 	"iss":  app.config.auth.jwtAuth.iss,
	// }

	token, err := app.auth.GenerateToken(claims)
	if err != nil {
		return err
	}

	cookie := http.Cookie{
		Name:     accessCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  exp,
	}

	http.SetCookie(w, &cookie)
	return nil
}
