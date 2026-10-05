package main

// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/CepeshIII/Project07_API_Server/internal/store"
// 	"github.com/go-chi/chi/v5"
// 	"github.com/go-playground/validator/v10"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// 	"go.uber.org/zap"
// )

// // Ensure validator is initialized for tests if not done elsewhere
// func init() {
// 	Validate = validator.New()
// }

// func TestCreatePostHandler(t *testing.T) {
// 	authUserID := int64(1)

// 	tests := []struct {
// 		name           string
// 		payload        any
// 		contextSetup   func(r *http.Request) *http.Request
// 		mockSetup      func(m *store.MockPostStore)
// 		expectedStatus int
// 	}{
// 		{
// 			name: "Success create post",
// 			payload: CreatePostPayload{
// 				Title:   "Test Title",
// 				Content: "Test Content",
// 				Tags:    []string{"go", "api"},
// 			},
// 			contextSetup: func(r *http.Request) *http.Request {
// 				return r.WithContext(context.WithValue(r.Context(), authUserIDCtxKey, authUserID))
// 			},
// 			mockSetup: func(m *store.MockPostStore) {
// 				m.On("Create", mock.Anything, mock.MatchedBy(func(p *store.PostModel) bool {
// 					return p.Title == "Test Title" && p.Content == "Test Content" && p.UserID == authUserID
// 				})).Return(nil)
// 			},
// 			expectedStatus: http.StatusCreated,
// 		},
// 		{
// 			name: "Auth user missing from context",
// 			payload: CreatePostPayload{
// 				Title:   "Test Title",
// 				Content: "Test Content",
// 			},
// 			contextSetup: func(r *http.Request) *http.Request {
// 				return r // no auth user in context
// 			},
// 			mockSetup:      func(m *store.MockPostStore) {},
// 			expectedStatus: http.StatusBadRequest,
// 		},
// 		{
// 			name:    "Invalid payload (missing title)",
// 			payload: CreatePostPayload{Content: "Only content"},
// 			contextSetup: func(r *http.Request) *http.Request {
// 				return r.WithContext(context.WithValue(r.Context(), authUserIDCtxKey, authUserID))
// 			},
// 			mockSetup:      func(m *store.MockPostStore) {},
// 			expectedStatus: http.StatusBadRequest,
// 		},
// 		{
// 			name: "Internal database error",
// 			payload: CreatePostPayload{
// 				Title:   "Test Title",
// 				Content: "Test Content",
// 			},
// 			contextSetup: func(r *http.Request) *http.Request {
// 				return r.WithContext(context.WithValue(r.Context(), authUserIDCtxKey, authUserID))
// 			},
// 			mockSetup: func(m *store.MockPostStore) {
// 				m.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))
// 			},
// 			expectedStatus: http.StatusInternalServerError,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockPosts := new(store.MockPostStore)
// 			tt.mockSetup(mockPosts)

// 			app := &application{
// 				store:  store.Storage{Posts: mockPosts},
// 				logger: zap.NewNop().Sugar(),
// 			}

// 			body, _ := json.Marshal(tt.payload)
// 			req := httptest.NewRequest(http.MethodPost, "/posts", bytes.NewBuffer(body))
// 			req.Header.Set("Content-Type", "application/json")
// 			req = tt.contextSetup(req)

// 			rr := httptest.NewRecorder()

// 			r := chi.NewRouter()
// 			r.Post("/posts", app.createPostHandler)
// 			r.ServeHTTP(rr, req)

// 			assert.Equal(t, tt.expectedStatus, rr.Code)
// 			mockPosts.AssertExpectations(t)
// 		})
// 	}
// }

// func TestGetPostHandler(t *testing.T) {
// 	samplePost := &store.Post{
// 		ID:      1,
// 		Title:   "Sample Post",
// 		Content: "Sample Content",
// 	}

// 	tests := []struct {
// 		name           string
// 		contextSetup   func(r *http.Request) *http.Request
// 		mockSetup      func(m *store.MockCommentStore)
// 		expectedStatus int
// 	}{
// 		{
// 			name: "Success get post with comments",
// 			contextSetup: func(r *http.Request) *http.Request {
// 				return r.WithContext(context.WithValue(r.Context(), postCtxKey, samplePost))
// 			},
// 			mockSetup: func(m *store.MockCommentStore) {
// 				m.On("GetByPostID", mock.Anything, samplePost.ID).Return([]store.CommentWithUser{}, nil)
// 			},
// 			expectedStatus: http.StatusOK,
// 		},
// 		{
// 			name: "Post missing from context",
// 			contextSetup: func(r *http.Request) *http.Request {
// 				return r // no post in context
// 			},
// 			mockSetup:      func(m *store.MockCommentStore) {},
// 			expectedStatus: http.StatusInternalServerError,
// 		},
// 		{
// 			name: "Internal error fetching comments",
// 			contextSetup: func(r *http.Request) *http.Request {
// 				return r.WithContext(context.WithValue(r.Context(), postCtxKey, samplePost))
// 			},
// 			mockSetup: func(m *store.MockCommentStore) {
// 				m.On("GetByPostID", mock.Anything, samplePost.ID).Return(nil, errors.New("db error"))
// 			},
// 			expectedStatus: http.StatusInternalServerError,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockComments := new(store.MockCommentStore)
// 			tt.mockSetup(mockComments)

// 			app := &application{
// 				store:  store.Storage{Comments: mockComments},
// 				logger: zap.NewNop().Sugar(),
// 			}

// 			req := httptest.NewRequest(http.MethodGet, "/posts/1", nil)
// 			req = tt.contextSetup(req)
// 			rr := httptest.NewRecorder()

// 			r := chi.NewRouter()
// 			r.Get("/posts/{id}", app.getPostHandler)
// 			r.ServeHTTP(rr, req)

// 			assert.Equal(t, tt.expectedStatus, rr.Code)
// 			mockComments.AssertExpectations(t)
// 		})
// 	}
// }

// func TestGetAllPostsHandler(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		mockSetup      func(m *store.MockPostStore)
// 		expectedStatus int
// 	}{
// 		{
// 			name: "Success get all posts feed",
// 			mockSetup: func(m *store.MockPostStore) {
// 				m.On("GetAllPosts", mock.Anything, mock.Anything).Return([]store.PostFeedItem{}, nil)
// 			},
// 			expectedStatus: http.StatusOK,
// 		},
// 		{
// 			name: "Database error returns internal server error",
// 			mockSetup: func(m *store.MockPostStore) {
// 				m.On("GetAllPosts", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
// 			},
// 			expectedStatus: http.StatusInternalServerError,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockPosts := new(store.MockPostStore)
// 			tt.mockSetup(mockPosts)

// 			app := &application{
// 				store:  store.Storage{Posts: mockPosts},
// 				logger: zap.NewNop().Sugar(),
// 			}

// 			req := httptest.NewRequest(http.MethodGet, "/posts", nil)
// 			rr := httptest.NewRecorder()

// 			r := chi.NewRouter()
// 			r.Get("/posts", app.getAllPostsHandler)
// 			r.ServeHTTP(rr, req)

// 			assert.Equal(t, tt.expectedStatus, rr.Code)
// 			mockPosts.AssertExpectations(t)
// 		})
// 	}
// }

// func TestCreateCommentHandler(t *testing.T) {
// 	authUserID := int64(1)
// 	postID := int64(10)

// 	tests := []struct {
// 		name           string
// 		payload        any
// 		contextSetup   func(r *http.Request) *http.Request
// 		mockSetup      func(m *store.MockCommentStore)
// 		expectedStatus int
// 	}{
// 		{
// 			name:    "Success add comment",
// 			payload: CreateCommentPayload{Content: "Great post!"},
// 			contextSetup: func(r *http.Request) *http.Request {
// 				r = r.WithContext(context.WithValue(r.Context(), authUserIDCtxKey, authUserID))
// 				r = r.WithContext(context.WithValue(r.Context(), postIDCtxKey, postID))
// 				return r
// 			},
// 			mockSetup: func(m *store.MockCommentStore) {
// 				m.On("Create", mock.Anything, mock.MatchedBy(func(c *store.Comment) bool {
// 					return c.Content == "Great post!" && c.UserID == authUserID && c.PostID == postID
// 				})).Return(nil)
// 			},
// 			expectedStatus: http.StatusCreated,
// 		},
// 		{
// 			name:    "Auth user missing",
// 			payload: CreateCommentPayload{Content: "Great post!"},
// 			contextSetup: func(r *http.Request) *http.Request {
// 				return r.WithContext(context.WithValue(r.Context(), postIDCtxKey, postID))
// 			},
// 			mockSetup:      func(m *store.MockCommentStore) {},
// 			expectedStatus: http.StatusBadRequest,
// 		},
// 		{
// 			name:    "Invalid payload (empty content)",
// 			payload: CreateCommentPayload{Content: ""},
// 			contextSetup: func(r *http.Request) *http.Request {
// 				r = r.WithContext(context.WithValue(r.Context(), authUserIDCtxKey, authUserID))
// 				r = r.WithContext(context.WithValue(r.Context(), postIDCtxKey, postID))
// 				return r
// 			},
// 			mockSetup:      func(m *store.MockCommentStore) {},
// 			expectedStatus: http.StatusBadRequest,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockComments := new(store.MockCommentStore)
// 			tt.mockSetup(mockComments)

// 			app := &application{
// 				store:  store.Storage{Comments: mockComments},
// 				logger: zap.NewNop().Sugar(),
// 			}

// 			body, _ := json.Marshal(tt.payload)
// 			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/posts/%d/comments", postID), bytes.NewBuffer(body))
// 			req.Header.Set("Content-Type", "application/json")
// 			req = tt.contextSetup(req)
// 			rr := httptest.NewRecorder()

// 			r := chi.NewRouter()
// 			r.Post("/posts/{id}/comments", app.createCommentHandler)
// 			r.ServeHTTP(rr, req)

// 			assert.Equal(t, tt.expectedStatus, rr.Code)
// 			mockComments.AssertExpectations(t)
// 		})
// 	}
// }

// func TestUpdatePostHandler(t *testing.T) {
// 	postID := int64(1)

// 	tests := []struct {
// 		name           string
// 		payload        any
// 		mockSetup      func(m *store.MockPostStore)
// 		expectedStatus int
// 	}{
// 		{
// 			name:    "Success update post",
// 			payload: UpdatePostPayload{Title: "Updated Title"},
// 			mockSetup: func(m *store.MockPostStore) {
// 				m.On("Update", mock.Anything, postID, mock.Anything).Return(nil)
// 			},
// 			expectedStatus: http.StatusCreated,
// 		},
// 		{
// 			name:    "Post not found on update",
// 			payload: UpdatePostPayload{Title: "Updated Title"},
// 			mockSetup: func(m *store.MockPostStore) {
// 				m.On("Update", mock.Anything, postID, mock.Anything).Return(store.ErrNotFound)
// 			},
// 			expectedStatus: http.StatusNotFound,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockPosts := new(store.MockPostStore)
// 			tt.mockSetup(mockPosts)

// 			app := &application{
// 				store:  store.Storage{Posts: mockPosts},
// 				logger: zap.NewNop().Sugar(),
// 			}

// 			body, _ := json.Marshal(tt.payload)
// 			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/posts/%d", postID), bytes.NewBuffer(body))
// 			req.Header.Set("Content-Type", "application/json")

// 			// Inject postID into context via chi router simulation or context setup
// 			req = req.WithContext(context.WithValue(req.Context(), postIDCtxKey, postID))
// 			rr := httptest.NewRecorder()

// 			r := chi.NewRouter()
// 			r.Patch("/posts/{id}", app.updatePostHandler)
// 			r.ServeHTTP(rr, req)

// 			assert.Equal(t, tt.expectedStatus, rr.Code)
// 			mockPosts.AssertExpectations(t)
// 		})
// 	}
// }

// func TestDeletePostHandler(t *testing.T) {
// 	postID := int64(1)

// 	tests := []struct {
// 		name           string
// 		mockSetup      func(cm *store.MockCommentStore, pm *store.MockPostStore)
// 		expectedStatus int
// 	}{
// 		{
// 			name: "Success delete post and comments",
// 			mockSetup: func(cm *store.MockCommentStore, pm *store.MockPostStore) {
// 				cm.On("DeleteByPostID", mock.Anything, postID).Return(nil)
// 				pm.On("Delete", mock.Anything, postID).Return(nil)
// 			},
// 			expectedStatus: http.StatusOK,
// 		},
// 		{
// 			name: "Post not found when deleting",
// 			mockSetup: func(cm *store.MockCommentStore, pm *store.MockPostStore) {
// 				cm.On("DeleteByPostID", mock.Anything, postID).Return(nil)
// 				pm.On("Delete", mock.Anything, postID).Return(store.ErrNotFound)
// 			},
// 			expectedStatus: http.StatusNotFound,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockComments := new(store.MockCommentStore)
// 			mockPosts := new(store.MockPostStore)
// 			tt.mockSetup(mockComments, mockPosts)

// 			app := &application{
// 				store: store.Storage{
// 					Comments: mockComments,
// 					Posts:    mockPosts,
// 				},
// 				logger: zap.NewNop().Sugar(),
// 			}

// 			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/posts/%d", postID), nil)
// 			req = req.WithContext(context.WithValue(req.Context(), postIDCtxKey, postID))
// 			rr := httptest.NewRecorder()

// 			r := chi.NewRouter()
// 			r.Delete("/posts/{id}", app.deletePostHandler)
// 			r.ServeHTTP(rr, req)

// 			assert.Equal(t, tt.expectedStatus, rr.Code)
// 			mockComments.AssertExpectations(t)
// 			mockPosts.AssertExpectations(t)
// 		})
// 	}
// }
