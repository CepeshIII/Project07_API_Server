package main

import (
	"errors"
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
)

func getTargetUserFromCtx(r *http.Request) *store.User {
	user, _ := r.Context().Value(targetUserCtxKey).(*store.User)
	return user
}

func getAuthUserIDFromCtx(r *http.Request) (int64, error) {
	value := r.Context().Value(authUserIDCtxKey)
	if value == nil {
		return 0, errors.New("authenticated user ID missing from context")
	}

	idString := fmt.Sprint(value)
	userID, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func getPostFromCtx(r *http.Request) *store.Post {
	post, _ := r.Context().Value(postCtxKey).(*store.Post)
	return post
}
