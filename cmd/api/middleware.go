package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/CepeshIII/Project07_API_Server/internal/store"
	"github.com/go-chi/chi/v5"
)

// AuthMiddleware authenticates requests using HTTP Basic Authentication,
// establishes user sessions, and injects session/access tokens into cookies.
func (app *application) AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			payload, err := app.extractAndValidateBasicAuth(w, r)
			if err != nil {
				app.unauthorizedBasicErrorResponse(w, r, err)
				return
			}

			user, err := app.verifyUserCredentials(w, r, payload)
			if err != nil {
				return
			}

			if err := app.createAndSetSessionAndToken(w, r, user); err != nil {
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (app *application) basicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := app.extractAndValidateBasicAuth(w, r)
			if err != nil {
				app.unauthorizedBasicErrorResponse(w, r, err)
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
			// This should not happen if the middleware is working correctly
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

		postID, err := getPostIDFromCtx(r)
		if err != nil {
			// This should not happen if the middleware is working correctly
			app.internalServerError(w, r, err)

			return
		}

		post, err := app.store.Posts.GetPostModel(ctx, postID)
		if err != nil {
			app.internalServerError(w, r, err)

			return
		}

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
		targetUser, err := getTargetUserFromCtx(r)
		if err != nil {
			// This should not happen if the middleware is working correctly
			app.internalServerError(w, r, err)

			return
		}

		// get authorization user ID from context
		userID, err := getAuthUserIDFromCtx(r)
		if err != nil {
			// This should not happen if the middleware is working correctly
			app.internalServerError(w, r, errAuthUserMissing)

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

		user, err := app.getUser(ctx, id)
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
				app.statusUnauthorizedError(w, r, errAccessTokenExpired)

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
		id, err := getPostIDFromCtx(r)
		if err != nil {
			// This should not happen if the middleware is working correctly
			app.internalServerError(w, r, err)

			return
		}

		ctx := r.Context()

		post, err := app.store.Posts.GetPost(ctx, id)
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

func (app *application) postIDContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "postsID")

		id, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			app.internalServerError(w, r, err)

			return
		}

		ctx := context.WithValue(r.Context(), postIDCtxKey, id)
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
