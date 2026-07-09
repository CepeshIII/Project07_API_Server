package main

import (
	"context"
	"course/api_server/internal/store"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type postKey string

const postCtx postKey = "post"

type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,max=200"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

type CreateCommentPayload struct {
	UserID  int64  `json:"user_id"`
	Content string `json:"content"`
}

func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	userId := 1

	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		UserID:  int64(userId),
	}

	ctx := r.Context()

	if err := app.store.Posts.Create(ctx, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	// Get post from database
	post := getPostFromCtx(r)

	// Get comments for post from database
	comments, err := app.store.Comments.GetByPostID(r.Context(), post.ID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			break
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	var postWithComments store.PostWithComments
	postWithComments.Post = post
	postWithComments.Comments = comments

	// Write responce
	if err := app.jsonResponse(w, http.StatusOK, postWithComments); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) getPostCommentsHandler(w http.ResponseWriter, r *http.Request) {
	// Get post from database
	post := getPostFromCtx(r)

	ctx := r.Context()
	comments, err := app.store.Comments.GetByPostID(ctx, post.ID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, comments); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	// Get post from database
	post := getPostFromCtx(r)
	ctx := r.Context()

	var payload CreateCommentPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	comment := &store.Comment{
		PostID:  post.ID,
		UserID:  payload.UserID,
		Content: payload.Content,
	}

	err := app.store.Comments.Create(ctx, comment)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, "OK"); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

type UpdatePostPayload struct {
	Title   string   `json:"title" validate:"omitempty,max=200"`
	Content string   `json:"content" validate:"omitempty,max=1000"`
	Tags    []string `json:"tags"`
}

func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	// Get post from database
	post := getPostFromCtx(r)

	// Parse post ID
	id, err := parsePostID(r)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	// Try read updated post data from request
	var payload UpdatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if payload.Title != "" {
		post.Title = payload.Title
	}
	if payload.Content != "" {
		post.Content = payload.Content
	}
	post.Tags = payload.Tags
	post.UserID = int64(1)

	// Try patch post data to database
	ctx := r.Context()
	if err := app.store.Posts.Update(ctx, id, post); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	// Write responce
	if err := app.jsonResponse(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	// Get post from database
	post := getPostFromCtx(r)

	ctx := r.Context()
	// Try delete comments for post from database
	if err := app.store.Comments.DeleteByPostID(ctx, post.ID); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	// Try delete post from database
	if err := app.store.Posts.Delete(ctx, post.ID); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	// Write responce
	if err := app.jsonResponse(w, http.StatusCreated, "OK"); err != nil {
		app.internalServerError(w, r, err)
		return
	}
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

		ctx = context.WithValue(ctx, postCtx, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromCtx(r *http.Request) *store.Post {
	post, _ := r.Context().Value(postCtx).(*store.Post)
	return post
}

func parsePostID(r *http.Request) (int64, error) {
	idParam := chi.URLParam(r, "postsID")
	return strconv.ParseInt(idParam, 10, 64)
}
