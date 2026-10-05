package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/CepeshIII/Project07_API_Server/internal/store"
)

type contextKey string

const (
	targetUserCtxKey contextKey = "user"
	authUserIDCtxKey contextKey = "userID"
	postCtxKey       contextKey = "post"
	postIDCtxKey     contextKey = "postID"
)

func getTargetUserFromCtx(r *http.Request) (*store.User, error) {
	user, ok := r.Context().Value(targetUserCtxKey).(*store.User)
	if !ok {
		return nil, errTargetUserMissing
	}
	return user, nil
}

func getAuthUserIDFromCtx(r *http.Request) (int64, error) {
	value := r.Context().Value(authUserIDCtxKey)
	if value == nil {
		return 0, errAuthUserMissing
	}

	idString := fmt.Sprint(value)
	userID, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func getPostFromCtx(r *http.Request) (*store.Post, error) {
	post, ok := r.Context().Value(postCtxKey).(*store.Post)
	if !ok {
		return nil, errPostMissing
	}
	return post, nil
}

func getPostIDFromCtx(r *http.Request) (int64, error) {
	post, ok := r.Context().Value(postIDCtxKey).(int64)
	if !ok {
		return 0, errPostIDMissing
	}
	return post, nil
}
