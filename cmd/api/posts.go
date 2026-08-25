package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/CepeshIII/Project07_API_Server/internal/store"

	"github.com/go-chi/chi/v5"
)

type postKey string

const postCtx postKey = "post"

type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,max=200"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

// CreatePost godoc
//
//	@Summary		Create a new post
//	@Description	Creates a new post
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			post	body		CreatePostPayload	true	"Create post request"
//
//	@Success		201		{object}	store.Post
//	@Failure		400		{object}	ErrorEnvelope
//	@Failure		500		{object}	ErrorEnvelope
//	@Security		ApiKeyAuth
//	@Router			/posts/ [post]
func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := getUserIDFromCtx(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var payload CreatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		UserID:  userId,
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

// GetPost godoc
//
//	@Summary		Fetches a post
//	@Description	Fetches a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Post ID"
//	@Success		200	{object}	PostWithCommentsEnvelope
//	@Failure		400	{object}	ErrorEnvelope
//	@Failure		404	{object}	ErrorEnvelope
//	@Failure		500	{object}	ErrorEnvelope
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [get]
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

// GetPostComments godoc
//
//	@Summary		Fetches a post comments
//	@Description	Fetches a post comments by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Post ID"
//	@Success		200	{object}	CommentsEnvelope
//	@Failure		400	{object}	ErrorEnvelope
//	@Failure		404	{object}	ErrorEnvelope
//	@Failure		500	{object}	ErrorEnvelope
//	@Security		ApiKeyAuth
//	@Router			/posts/{id}/comments [get]
func (app *application) getPostCommentsHandler(w http.ResponseWriter, r *http.Request) {
	// Get post from database
	post := getPostFromCtx(r)

	ctx := r.Context()
	comments, err := app.store.Comments.GetByPostID(ctx, post.ID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			comments = make([]store.CommentWithUser, 0)
			// app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusOK, comments); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// AddComment godoc
//
//	@Summary		Add a comment to post
//	@Description	Add a comment to post by postId
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int						true	"Post ID"
//	@Param			id	body		CreateCommentPayload	true	"Add comment request"
//	@Success		201	{object}	MessageEnvelope
//	@Failure		400	{object}	ErrorEnvelope
//	@Failure		404	{object}	ErrorEnvelope
//	@Failure		500	{object}	ErrorEnvelope
//	@Security		ApiKeyAuth
//	@Router			/posts/{id}/comments [post]
func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := getUserIDFromCtx(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

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
		UserID:  userId,
		Content: payload.Content,
	}

	err = app.store.Comments.Create(ctx, comment)
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

// UpdatePost godoc
//
//	@Summary		Update a post
//	@Description	Updates an existing post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Post ID"
//	@Param			post	body		UpdatePostPayload	true	"Update post request"
//	@Success		201		{object}	store.Post
//	@Failure		400		{object}	ErrorEnvelope
//	@Failure		404		{object}	ErrorEnvelope
//	@Failure		500		{object}	ErrorEnvelope
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [patch]
func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := getUserIDFromCtx(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Get post from database
	post := getPostFromCtx(r)

	if post.UserID != userId {
		app.statusForbiddenResponse(w, r, errors.New("You do not have permission to edit this post"))
		return
	}

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

// DeletePost godoc
//
//	@Summary		Delete a post
//	@Description	Deletes a post and all its comments by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Post ID"
//	@Success		200	{object}	MessageEnvelope
//	@Failure		400	{object}	ErrorEnvelope
//	@Failure		404	{object}	ErrorEnvelope
//	@Failure		500	{object}	ErrorEnvelope
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [delete]
func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := getUserIDFromCtx(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Get post from database
	post := getPostFromCtx(r)

	if post.UserID != userId {
		app.statusForbiddenResponse(w, r, errors.New("You do not have permission to delete this post"))
		return
	}

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
	if err := app.jsonResponse(w, http.StatusOK, "OK"); err != nil {
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
