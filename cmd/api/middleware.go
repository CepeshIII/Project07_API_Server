package main

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/CepeshIII/Project07_API_Server/internal/httputils"
	"github.com/CepeshIII/Project07_API_Server/internal/store"
)

func (app *application) AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// read the auth header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.unauthorizedBasicErrorResponse(w, r, errors.New("authorization header is missing"))
				return
			}

			// parse it -> get the base64
			parts := strings.Split(authHeader, " ")
			if len(parts) <= 1 || parts[0] != "Basic" {
				app.unauthorizedBasicErrorResponse(w, r, errors.New("authorization header is malformed"))
				return
			}

			// decode it
			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				app.unauthorizedBasicErrorResponse(w, r, err)
				return
			}

			// check the credentials
			creds := strings.SplitN(string(decoded), ":", 2)
			if len(creds) != 2 {
				app.unauthorizedBasicErrorResponse(w, r, errors.New("invalid credentials"))
				return
			}

			payload := LoginUserPayload{
				Password: creds[1],
				Username: creds[0],
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
					app.unauthorizedBasicErrorResponse(w, r, errors.New("Invalid Username or password"))
					return
				}

				app.internalServerError(w, r, err)
				return
			}

			// check the user password
			if !user.Password.CheckHash(payload.Password) {
				app.unauthorizedBasicErrorResponse(w, r, errors.New("Invalid Username or password"))
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

			// response := &RegisterUserResponse{
			// 	User: user,
			// 	// Token: token,
			// }

			http.SetCookie(w, &cookie)

			// set new accessToken
			if err := app.setAccessToken(w, session.UserID); err != nil {
				app.internalServerError(w, r, err)
				return
			}

			// if err := app.jsonResponse(w, http.StatusOK, response); err != nil {
			// 	app.internalServerError(w, r, err)
			// 	return
			// }

			next.ServeHTTP(w, r)
		})
	}
}

func (app *application) basicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// read the auth header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.unauthorizedBasicErrorResponse(w, r, errors.New("authorization header is missing"))
				return
			}

			// parse it -> get the base64
			parts := strings.Split(authHeader, " ")
			if len(parts) <= 1 || parts[0] != "Basic" {
				app.unauthorizedBasicErrorResponse(w, r, errors.New("authorization header is malformed"))
				return
			}

			// decode it
			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				app.unauthorizedBasicErrorResponse(w, r, err)
				return
			}

			username := app.config.auth.basic.username
			pass := app.config.auth.basic.password

			// check the credentials
			creds := strings.SplitN(string(decoded), ":", 2)
			if len(creds) != 2 || creds[0] != username || creds[1] != pass {
				app.unauthorizedBasicErrorResponse(w, r, errors.New("invalid credentials"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
