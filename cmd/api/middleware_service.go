package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/CepeshIII/Project07_API_Server/internal/credential"
	"github.com/CepeshIII/Project07_API_Server/internal/httputils"
	"github.com/CepeshIII/Project07_API_Server/internal/store"
)

func (app *application) extractAndValidateBasicAuth(w http.ResponseWriter, r *http.Request) (LoginUserPayload, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return LoginUserPayload{}, errAuthHeaderMissing
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) <= 1 || parts[0] != "Basic" {
		return LoginUserPayload{}, errAuthHeaderMalformed
	}

	decoded, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return LoginUserPayload{}, fmt.Errorf("failed to decode base64 authorization string: %w", err)
	}

	creds := strings.SplitN(string(decoded), ":", 2) //nolint:mnd
	if len(creds) != 2 {                             //nolint:mnd
		return LoginUserPayload{}, errInvalidCredentials
	}

	payload := LoginUserPayload{
		Password: creds[1],
		Username: creds[0],
	}

	if err := app.validator.Struct(payload); err != nil {
		app.clearSessionCookie(w)
		return LoginUserPayload{}, fmt.Errorf("failed to validate login payload: %w", err)
	}

	return payload, nil
}

func (app *application) verifyUserCredentials(w http.ResponseWriter, r *http.Request, payload LoginUserPayload) (*store.User, error) {
	ctx := r.Context()

	user, err := app.store.Users.GetByUsername(ctx, payload.Username)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.clearSessionCookie(w)
			app.unauthorizedBasicErrorResponse(w, r, errInvalidUsernameOrPassword)

			return nil, errInvalidUsernameOrPassword
		}

		app.internalServerError(w, r, err)

		return nil, err
	}

	if !user.Password.CheckHash(payload.Password) {
		app.unauthorizedBasicErrorResponse(w, r, errInvalidUsernameOrPassword)
		return nil, errInvalidUsernameOrPassword
	}

	return user, nil
}

func (app *application) createAndSetSessionAndToken(w http.ResponseWriter, r *http.Request, user *store.User) error {
	ctx := r.Context()
	token, tokenHash := credential.New()

	session := store.SessionData{
		UserID:    user.ID,
		Token:     token,
		TokenHash: tokenHash,
		UserAgent: r.Header.Get("User-Agent"),
		IPAdress:  httputils.GetClientIP(r),
		IsRevoke:  false,
		ExpiresAt: time.Now().Add(app.config.Auth.Tokens.SessionTokenExp),
	}

	if err := app.store.Sessions.CreateSession(ctx, &session); err != nil {
		app.internalServerError(w, r, err)
		return err
	}

	cookie := http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  session.ExpiresAt,
	}

	http.SetCookie(w, &cookie)

	accessToken, accessTokenExp, err := app.generateAccessToken(session.UserID, "user")
	if err != nil {
		app.internalServerError(w, r, err)
		return err
	}

	setAccessTokenCookie(w, accessToken, accessTokenExp)

	return nil
}
