package main

import (
	"net/http"

	"github.com/CepeshIII/Project07_API_Server/internal/store"
)

func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	fq := store.PaginatedFeedQuery{
		Limit:  app.config.Pagination.DefaultLimit,
		Offset: 0,
		Sort:   "desc",
		Tags:   []string{},
		Query:  "",
	}

	fq, err := fq.Parse(r)
	if err != nil {
		app.badRequestResponse(w, r, err)

		return
	}

	if err := app.validator.Struct(fq); err != nil {
		app.badRequestResponse(w, r, err)

		return
	}

	// The user targeted in the URL path: /users/{userID}/follow
	targetUser, err := getTargetUserFromCtx(r)
	if err != nil {
		// This should not happen if the middleware is working correctly
		app.internalServerError(w, r, err)

		return
	}

	ctx := r.Context()

	posts, err := app.store.Posts.GetUserFeed(ctx, targetUser.ID, fq)
	if err != nil {
		app.internalServerError(w, r, err)

		return
	}

	if err := app.jsonResponse(w, http.StatusOK, posts); err != nil {
		app.internalServerError(w, r, err)

		return
	}
}
