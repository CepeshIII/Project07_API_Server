package main

import (
	"errors"
	"net/http"
)

var (
	errTargetUserMissing = errors.New("target user missing from context")
	errAuthUserMissing   = errors.New("authenticated user missing from context")

	errPostMissing   = errors.New("post missing from context")
	errPostIDMissing = errors.New("post ID missing from context")

	errInvalidUsernameOrPassword = errors.New("invalid Username or password")
	errServerEncounteredProblem  = errors.New("the server encountered a problem")

	errSessionTokenExpired      = errors.New("session token has expired")
	errSessionNotFoundOrInvalid = errors.New("session not found or invalid")
	errSessionRevoked           = errors.New("session has been revoked")
	errSessionExpired           = errors.New("session has expired")

	errAuthHeaderMissing   = errors.New("authorization header is missing")
	errAuthHeaderMalformed = errors.New("authorization header is malformed")
	errInvalidCredentials  = errors.New("invalid credentials")
	errAccessTokenExpired  = errors.New("access token has expired")

	errCannotUnfollowYourself = errors.New("you cannot unfollow yourself")
	errCannotFollowYourself   = errors.New("you cannot follow yourself")

	errInvalidUserID = errors.New("invalid user ID")
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("internal server error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	if err := writeJSONError(w, http.StatusInternalServerError, errServerEncounteredProblem.Error()); err != nil {
		app.logger.Errorw(err.Error())
	}
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw("bad request", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	if err := writeJSONError(w, http.StatusBadRequest, err.Error()); err != nil {
		app.logger.Errorw(err.Error())
	}
}

func (app *application) statusUnauthorizedError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw("unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	if err := writeJSONError(w, http.StatusUnauthorized, err.Error()); err != nil {
		app.logger.Errorw(err.Error())
	}
}

func (app *application) unauthorizedBasicErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw("unauthorized basic error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)

	if err := writeJSONError(w, http.StatusUnauthorized, err.Error()); err != nil {
		app.logger.Errorw(err.Error())
	}
}

func (app *application) forbiddenResponse(w http.ResponseWriter, r *http.Request) {
	app.logger.Warnw("forbidden access", "method", r.Method, "path", r.URL.Path)

	if err := writeJSONError(w, http.StatusForbidden, "forbidden"); err != nil {
		app.logger.Errorw(err.Error())
	}
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw("resource not found", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	if err := writeJSONError(w, http.StatusNotFound, "resource not found"); err != nil {
		app.logger.Errorw(err.Error())
	}
}

func (app *application) conflictResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw("conflict error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	if err := writeJSONError(w, http.StatusConflict, err.Error()); err != nil {
		app.logger.Errorw(err.Error())
	}
}

func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request, retryAfter string) {
	app.logger.Warnw("rate limit exceeded", "method", r.Method, "path", r.URL.Path)

	w.Header().Set("Retry-After", retryAfter)

	if err := writeJSONError(w, http.StatusTooManyRequests, "rate limit exceeded, retry after: "+retryAfter); err != nil {
		app.logger.Errorw(err.Error())
	}
}
