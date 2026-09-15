package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
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
				Token:     token,
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
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
				Expires:  session.ExpiresAt,
			}

			// response := &RegisterUserResponse{
			// 	User: user,
			// 	// Token: token,
			// }

			http.SetCookie(w, &cookie)

			// create new Access Token
			accesstoken, accessTokenExp, err := app.generateAccessToken(session.UserID, "user")
			if err != nil {
				app.internalServerError(w, r, err)
				return
			}

			// set new Access Token to coockie
			setAccessTokenCookie(w, accesstoken, accessTokenExp)

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

func (app *application) checkPostOwnership(roleName string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, err := getAuthUserIDFromCtx(r)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		userWithRole, err := app.getUserWithRole(ctx, userID)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundResponse(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		post := getPostFromCtx(r)

		if post.UserID != userWithRole.ID {
			ok, err := app.checkRolePrecedence(ctx, userWithRole, roleName)
			if err != nil {
				app.internalServerError(w, r, err)
				return
			}
			if !ok {
				app.forbiddenResponse(w, r)
				return
			}
		}

		// Pass execution to next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) checkUserOwnership(roleName string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// get target user form context
		targetUser := getTargetUserFromCtx(r)
		if targetUser == nil {
			// This should not happen if the middleware is working correctly
			app.internalServerError(w, r, errors.New("target user missing from context"))
			return
		}

		// get authorization user ID from context
		userID, err := getAuthUserIDFromCtx(r)
		if err != nil {
			// This should not happen if the middleware is working correctly
			app.internalServerError(w, r, errors.New("authenticated user missing from context"))
			return
		}

		// get authorizated user data by ID
		userWithRole, err := app.getUserWithRole(ctx, userID)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		// chech if auth user is owner of target user account
		if targetUser.ID != userWithRole.ID {
			// chech if auth user have permission
			ok, err := app.checkRolePrecedence(ctx, userWithRole, roleName)
			if err != nil {
				app.internalServerError(w, r, err)
				return
			}

			if !ok {
				app.forbiddenResponse(w, r)
				return
			}
		}

		// Pass execution to next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUserID(r)
		if err != nil {
			app.badRequestResponse(w, r, err)
			return
		}

		ctx := r.Context()
		user, err := app.getUser(ctx, (int64)(id))
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundResponse(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		ctx = context.WithValue(ctx, targetUserCtxKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) accessTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to get access token from the cookie
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
		claims, err := app.auth.ValidateToken(cookie.Value)
		if err != nil {
			app.clearAccessCookie(w)
			app.statusUnauthorizedError(w, r, err)
			return
		}

		// Inject authenticated user ID into request context
		ctx := context.WithValue(r.Context(), authUserIDCtxKey, claims.UserID)

		// Pass execution to next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) postContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := parsePostID(r)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		ctx := r.Context()
		post, err := app.store.Posts.GetByID(ctx, id)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundResponse(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		ctx = context.WithValue(ctx, postCtxKey, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) rateLimiterMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		OK, retryAfter := app.rateLimiter.Allow(r.RemoteAddr)

		if !OK {
			app.rateLimitExceededResponse(w, r, fmt.Sprint(retryAfter))
			return
		}

		next.ServeHTTP(w, r)
	})
}
