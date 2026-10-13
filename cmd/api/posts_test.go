package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CepeshIII/Project07_API_Server/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestCreatePostHandler(t *testing.T) {
	authUserID := int64(1)
	title := "Test title"
	content := "Test content"

	tests := []struct {
		name                string
		expectedStatus      int
		payload             any
		contextSetup        func(r *http.Request) *http.Request
		mockSetup           func(m *store.MockPostStore)
		dataValidationSetup func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:           "Success create post",
			expectedStatus: http.StatusCreated,
			payload: CreatePostPayload{
				Title:   title,
				Content: content,
				Tags:    []string{"go", "api"},
			},
			contextSetup: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), authUserIDCtxKey, authUserID))
			},
			mockSetup: func(m *store.MockPostStore) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(p *store.PostModel) bool {
					return p.Title == title && p.Content == content && p.UserID == authUserID
				})).Return(nil)
			},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var envelope MessageEnvelope
				decodeAndValidate(t, rr.Body.String(), &envelope)
			},
		},
		{
			name:           "Auth user missing from context",
			expectedStatus: http.StatusInternalServerError,
			payload: CreatePostPayload{
				Title:   title,
				Content: content,
			},
			contextSetup: func(r *http.Request) *http.Request {
				return r // no auth user in context
			},
			mockSetup: func(m *store.MockPostStore) {},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var envelope ErrorEnvelope
				decodeAndValidate(t, rr.Body.String(), &envelope)
			},
		},
		{
			name:           "Invalid payload (missing title)",
			expectedStatus: http.StatusBadRequest,
			payload:        CreatePostPayload{Content: "Only content"},
			contextSetup: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), authUserIDCtxKey, authUserID))
			},
			mockSetup: func(m *store.MockPostStore) {},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var envelope ErrorEnvelope
				decodeAndValidate(t, rr.Body.String(), &envelope)
			},
		},
		{
			name:           "Internal database error",
			expectedStatus: http.StatusInternalServerError,
			payload: CreatePostPayload{
				Title:   title,
				Content: content,
			},
			contextSetup: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), authUserIDCtxKey, authUserID))
			},
			mockSetup: func(m *store.MockPostStore) {
				m.On("Create", mock.Anything, mock.Anything).Return(store.ErrUnexpected)
			},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var envelope ErrorEnvelope
				decodeAndValidate(t, rr.Body.String(), &envelope)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPosts := new(store.MockPostStore)
			tt.mockSetup(mockPosts)

			app := &application{
				store:     store.Storage{Posts: mockPosts},
				logger:    zap.NewNop().Sugar(),
				validator: validator.New(validator.WithRequiredStructEnabled()),
			}

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/posts", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = tt.contextSetup(req)

			rr := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Post("/posts", app.createPostHandler)
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockPosts.AssertExpectations(t)

			if tt.dataValidationSetup != nil {
				tt.dataValidationSetup(t, rr)
			}
		})
	}
}

func TestGetPostHandler(t *testing.T) {
	samplePost := &store.Post{
		ID:      1,
		Title:   "Sample Post",
		Content: "Sample Content",
	}

	tests := []struct {
		name                string
		contextSetup        func(r *http.Request) *http.Request
		mockSetup           func(m *store.MockCommentStore)
		expectedStatus      int
		dataValidationSetup func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "Success get post with comments",
			contextSetup: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), postCtxKey, samplePost))
			},
			mockSetup: func(m *store.MockCommentStore) {
				m.On("GetByPostID", mock.Anything, samplePost.ID).Return([]store.CommentWithUser{}, nil)
			},
			expectedStatus: http.StatusOK,
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var envelope PostWithCommentsEnvelope
				decodeAndValidate(t, rr.Body.String(), &envelope)
			},
		},
		{
			name: "Post missing from context",
			contextSetup: func(r *http.Request) *http.Request {
				return r // no post in context
			},
			mockSetup:      func(m *store.MockCommentStore) {},
			expectedStatus: http.StatusInternalServerError,
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var envelope ErrorEnvelope
				decodeAndValidate(t, rr.Body.String(), &envelope)
			},
		},
		{
			name: "Internal error fetching comments",
			contextSetup: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), postCtxKey, samplePost))
			},
			mockSetup: func(m *store.MockCommentStore) {
				m.On("GetByPostID", mock.Anything, samplePost.ID).Return(nil, store.ErrUnexpected)
			},
			expectedStatus: http.StatusInternalServerError,
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var envelope ErrorEnvelope
				decodeAndValidate(t, rr.Body.String(), &envelope)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockComments := new(store.MockCommentStore)
			tt.mockSetup(mockComments)

			app := &application{
				store:     store.Storage{Comments: mockComments},
				logger:    zap.NewNop().Sugar(),
				validator: validator.New(validator.WithRequiredStructEnabled()),
			}

			req := httptest.NewRequest(http.MethodGet, "/posts/1", nil)
			req = tt.contextSetup(req)
			rr := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Get("/posts/{id}", app.getPostHandler)
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockComments.AssertExpectations(t)

			if tt.dataValidationSetup != nil {
				tt.dataValidationSetup(t, rr)
			}
		})
	}
}

func TestGetAllPostsHandler(t *testing.T) {
	tests := []struct {
		name                string
		mockSetup           func(m *store.MockPostStore)
		expectedStatus      int
		dataValidationSetup func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "Success get all posts feed",
			mockSetup: func(m *store.MockPostStore) {
				m.On("GetAllPosts", mock.Anything, mock.Anything).Return([]*store.PostFeedItem{}, nil)
			},
			expectedStatus: http.StatusOK,
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var envelope PostsPostWithMetadataEnvelope
				decodeAndValidate(t, rr.Body.String(), &envelope)
			},
		},
		{
			name: "Database error returns internal server error",
			mockSetup: func(m *store.MockPostStore) {
				m.On("GetAllPosts", mock.Anything, mock.Anything).Return(nil, store.ErrUnexpected)
			},
			expectedStatus: http.StatusInternalServerError,
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var envelope ErrorEnvelope
				decodeAndValidate(t, rr.Body.String(), &envelope)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPosts := new(store.MockPostStore)
			tt.mockSetup(mockPosts)

			app := &application{
				store:     store.Storage{Posts: mockPosts},
				logger:    zap.NewNop().Sugar(),
				validator: validator.New(validator.WithRequiredStructEnabled()),
			}

			req := httptest.NewRequest(http.MethodGet, "/posts", nil)
			rr := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Get("/posts", app.getAllPostsHandler)
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockPosts.AssertExpectations(t)

			if tt.dataValidationSetup != nil {
				tt.dataValidationSetup(t, rr)
			}
		})
	}
}

func TestCreateCommentHandler(t *testing.T) {
	authUserID := int64(1)
	postID := int64(10)

	tests := []struct {
		name                string
		payload             any
		contextSetup        func(r *http.Request) *http.Request
		mockSetup           func(m *store.MockCommentStore)
		expectedStatus      int
		dataValidationSetup func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:           "Success add comment",
			expectedStatus: http.StatusCreated,
			payload:        CreateCommentPayload{Content: "Great post!"},
			contextSetup: func(r *http.Request) *http.Request {
				ctx := r.Context()
				ctx = context.WithValue(ctx, authUserIDCtxKey, authUserID)
				r = r.WithContext(context.WithValue(ctx, postIDCtxKey, postID))

				return r
			},
			mockSetup: func(m *store.MockCommentStore) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(c *store.Comment) bool {
					return c.Content == "Great post!" && c.UserID == authUserID && c.PostID == postID
				})).Return(nil)
			},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var envelope MessageEnvelope
				decodeAndValidate(t, rr.Body.String(), &envelope)
			},
		},
		{
			name:           "Auth user missing",
			expectedStatus: http.StatusInternalServerError,
			payload:        CreateCommentPayload{Content: "Great post!"},
			contextSetup: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), postIDCtxKey, postID))
			},
			mockSetup: func(m *store.MockCommentStore) {},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var envelope ErrorEnvelope
				decodeAndValidate(t, rr.Body.String(), &envelope)
			},
		},
		{
			name:           "Invalid payload (empty content)",
			expectedStatus: http.StatusBadRequest,
			payload:        CreateCommentPayload{Content: ""},
			contextSetup: func(r *http.Request) *http.Request {
				ctx := r.Context()
				ctx = context.WithValue(ctx, authUserIDCtxKey, authUserID)
				r = r.WithContext(context.WithValue(ctx, postIDCtxKey, postID))

				return r
			},
			mockSetup: func(m *store.MockCommentStore) {},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var envelope ErrorEnvelope
				decodeAndValidate(t, rr.Body.String(), &envelope)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockComments := new(store.MockCommentStore)
			tt.mockSetup(mockComments)

			app := &application{
				store:     store.Storage{Comments: mockComments},
				logger:    zap.NewNop().Sugar(),
				validator: validator.New(validator.WithRequiredStructEnabled()),
			}

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/posts/%d/comments", postID), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = tt.contextSetup(req)
			rr := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Post("/posts/{id}/comments", app.createCommentHandler)
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockComments.AssertExpectations(t)

			if tt.dataValidationSetup != nil {
				tt.dataValidationSetup(t, rr)
			}
		})
	}
}

func TestGetPostCommentsHandler(t *testing.T) {
	postID := int64(10)

	tests := []struct {
		name                string
		contextSetup        func(r *http.Request) *http.Request
		mockSetup           func(m *store.MockCommentStore)
		expectedStatus      int
		dataValidationSetup func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "Success get post comments",
			contextSetup: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), postIDCtxKey, postID))
			},
			mockSetup: func(m *store.MockCommentStore) {
				m.On("GetByPostID", mock.Anything, postID).Return([]store.CommentWithUser{}, nil)
			},
			expectedStatus: http.StatusOK,
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var envelope CommentsEnvelope
				decodeAndValidate(t, rr.Body.String(), &envelope)
			},
		},
		{
			name: "Post ID missing from context",
			contextSetup: func(r *http.Request) *http.Request {
				return r // no post ID in context
			},
			mockSetup:      func(m *store.MockCommentStore) {},
			expectedStatus: http.StatusInternalServerError,
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var envelope ErrorEnvelope
				decodeAndValidate(t, rr.Body.String(), &envelope)
			},
		},
		{
			name: "Internal error fetching comments",
			contextSetup: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), postIDCtxKey, postID))
			},
			mockSetup: func(m *store.MockCommentStore) {
				m.On("GetByPostID", mock.Anything, postID).Return(nil, store.ErrUnexpected)
			},
			expectedStatus: http.StatusInternalServerError,
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var envelope ErrorEnvelope
				decodeAndValidate(t, rr.Body.String(), &envelope)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockComments := new(store.MockCommentStore)
			tt.mockSetup(mockComments)

			app := &application{
				store:     store.Storage{Comments: mockComments},
				logger:    zap.NewNop().Sugar(),
				validator: validator.New(validator.WithRequiredStructEnabled()),
			}

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/posts/%d/comments", postID), nil)
			req = tt.contextSetup(req)
			rr := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Get("/posts/{id}/comments", app.getPostCommentsHandler)
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockComments.AssertExpectations(t)

			if tt.dataValidationSetup != nil {
				tt.dataValidationSetup(t, rr)
			}
		})
	}
}

func TestUpdatePostHandler(t *testing.T) {
	postID := int64(10)

	tests := []struct {
		name                string
		payload             any
		contextSetup        func(r *http.Request) *http.Request
		mockSetup           func(m *store.MockPostStore)
		expectedStatus      int
		dataValidationSetup func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:           "Success update post",
			expectedStatus: http.StatusCreated,
			payload:        UpdatePostPayload{Title: "Updated Title", Content: "Updated Content"},
			contextSetup: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), postIDCtxKey, postID))
			},
			mockSetup: func(m *store.MockPostStore) {
				m.On("Update", mock.Anything, postID, mock.MatchedBy(func(req *store.UpdatePostRequest) bool {
					return req.Title == "Updated Title" && req.Content == "Updated Content"
				})).Return(nil)
			},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var envelope MessageEnvelope
				decodeAndValidate(t, rr.Body.String(), &envelope)
			},
		},
		{
			name:           "Post ID missing from context",
			expectedStatus: http.StatusInternalServerError,
			payload:        UpdatePostPayload{Title: "Updated Title", Content: "Updated Content"},
			contextSetup: func(r *http.Request) *http.Request {
				return r // no post ID in context
			},
			mockSetup: func(m *store.MockPostStore) {},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var envelope ErrorEnvelope
				decodeAndValidate(t, rr.Body.String(), &envelope)
			},
		},
		{
			name:           "Post not found in database",
			expectedStatus: http.StatusNotFound,
			payload:        UpdatePostPayload{Title: "Updated Title", Content: "Updated Content"},
			contextSetup: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), postIDCtxKey, postID))
			},
			mockSetup: func(m *store.MockPostStore) {
				m.On("Update", mock.Anything, postID, mock.Anything).Return(store.ErrNotFound)
			},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var envelope ErrorEnvelope
				decodeAndValidate(t, rr.Body.String(), &envelope)
			},
		},
		{
			name:           "Database error returns internal server error",
			expectedStatus: http.StatusInternalServerError,
			payload:        UpdatePostPayload{Title: "Updated Title", Content: "Updated Content"},
			contextSetup: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), postIDCtxKey, postID))
			},
			mockSetup: func(m *store.MockPostStore) {
				m.On("Update", mock.Anything, postID, mock.Anything).Return(store.ErrUnexpected)
			},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var envelope ErrorEnvelope
				decodeAndValidate(t, rr.Body.String(), &envelope)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPosts := new(store.MockPostStore)
			tt.mockSetup(mockPosts)

			app := &application{
				store:     store.Storage{Posts: mockPosts},
				logger:    zap.NewNop().Sugar(),
				validator: validator.New(validator.WithRequiredStructEnabled()),
			}

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/posts/%d", postID), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = tt.contextSetup(req)
			rr := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Patch("/posts/{id}", app.updatePostHandler)
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockPosts.AssertExpectations(t)

			if tt.dataValidationSetup != nil {
				tt.dataValidationSetup(t, rr)
			}
		})
	}
}

func TestDeletePostHandler(t *testing.T) {
	postID := int64(10)

	tests := []struct {
		name                string
		contextSetup        func(r *http.Request) *http.Request
		mockSetup           func(mComments *store.MockCommentStore, mPosts *store.MockPostStore)
		expectedStatus      int
		dataValidationSetup func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "Success delete post and comments",
			contextSetup: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), postIDCtxKey, postID))
			},
			mockSetup: func(mComments *store.MockCommentStore, mPosts *store.MockPostStore) {
				mComments.On("DeleteByPostID", mock.Anything, postID).Return(nil)
				mPosts.On("Delete", mock.Anything, postID).Return(nil)
			},
			expectedStatus: http.StatusOK,
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var envelope MessageEnvelope
				decodeAndValidate(t, rr.Body.String(), &envelope)
			},
		},
		{
			name: "Post ID missing from context",
			contextSetup: func(r *http.Request) *http.Request {
				return r // no post ID in context
			},
			mockSetup:      func(mComments *store.MockCommentStore, mPosts *store.MockPostStore) {},
			expectedStatus: http.StatusInternalServerError,
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var envelope ErrorEnvelope
				decodeAndValidate(t, rr.Body.String(), &envelope)
			},
		},
		{
			name: "Comments delete not found returns 404",
			contextSetup: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), postIDCtxKey, postID))
			},
			mockSetup: func(mComments *store.MockCommentStore, mPosts *store.MockPostStore) {
				mComments.On("DeleteByPostID", mock.Anything, postID).Return(store.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var envelope ErrorEnvelope
				decodeAndValidate(t, rr.Body.String(), &envelope)
			},
		},
		{
			name: "Comments delete db error returns 500",
			contextSetup: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), postIDCtxKey, postID))
			},
			mockSetup: func(mComments *store.MockCommentStore, mPosts *store.MockPostStore) {
				mComments.On("DeleteByPostID", mock.Anything, postID).Return(store.ErrUnexpected)
			},
			expectedStatus: http.StatusInternalServerError,
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var envelope ErrorEnvelope
				decodeAndValidate(t, rr.Body.String(), &envelope)
			},
		},
		{
			name: "Post delete not found returns 404",
			contextSetup: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), postIDCtxKey, postID))
			},
			mockSetup: func(mComments *store.MockCommentStore, mPosts *store.MockPostStore) {
				mComments.On("DeleteByPostID", mock.Anything, postID).Return(nil)
				mPosts.On("Delete", mock.Anything, postID).Return(store.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var envelope ErrorEnvelope
				decodeAndValidate(t, rr.Body.String(), &envelope)
			},
		},
		{
			name: "Post delete db error returns 500",
			contextSetup: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), postIDCtxKey, postID))
			},
			mockSetup: func(mComments *store.MockCommentStore, mPosts *store.MockPostStore) {
				mComments.On("DeleteByPostID", mock.Anything, postID).Return(nil)
				mPosts.On("Delete", mock.Anything, postID).Return(store.ErrUnexpected)
			},
			expectedStatus: http.StatusInternalServerError,
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var envelope ErrorEnvelope
				decodeAndValidate(t, rr.Body.String(), &envelope)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockComments := new(store.MockCommentStore)
			mockPosts := new(store.MockPostStore)
			tt.mockSetup(mockComments, mockPosts)

			app := &application{
				store: store.Storage{
					Comments: mockComments,
					Posts:    mockPosts,
				},
				logger:    zap.NewNop().Sugar(),
				validator: validator.New(validator.WithRequiredStructEnabled()),
			}

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/posts/%d", postID), nil)
			req = tt.contextSetup(req)
			rr := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Delete("/posts/{id}", app.deletePostHandler)
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockComments.AssertExpectations(t)
			mockPosts.AssertExpectations(t)

			if tt.dataValidationSetup != nil {
				tt.dataValidationSetup(t, rr)
			}
		})
	}
}
