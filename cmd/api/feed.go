package main

import (
	"net/http"

	"github.com/CepeshIII/Project07_API_Server/internal/store"
)

func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {

	fq := store.PaginatedFeedQuery{
		Limit:  20,
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

	if err := Validate.Struct(fq); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	posts, err := app.store.Posts.GetUserFeed(ctx, 2, fq) // Replace 1 with the actual user ID
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, posts); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
