package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/CepeshIII/Project07_API_Server/internal/store"
)

type Envelope struct {
	Data any `json:"data"`
}

type HealthCheckEnvelope struct {
	Data HealthCheckData `json:"data"`
}

type UserEnvelope struct {
	Data store.User `json:"data"`
}

type RegisterUserResponseEnvelope struct {
	Data RegisterUserResponse `json:"data"`
}

type PostEnvelope struct {
	Data store.Post `json:"data"`
}

type PostWithCommentsEnvelope struct {
	Data store.PostWithComments `json:"data"`
}

type CommentsEnvelope struct {
	Data []store.Comment `json:"data"`
}

type FollowersEnvelope struct {
	Data []store.FollowerModel `json:"data"`
}

type PostsEnvelope struct {
	Data []store.Post `json:"data"`
}

type PostsPostWithMetadataEnvelope struct {
	Data []store.PostFeedItem `json:"data"`
}

type MessageEnvelope struct {
	Data string `json:"data"`
}

type CreateCommentPayload struct {
	Content string `json:"content" validate:"required,max=500"`
}

type ErrorEnvelope struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

func readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1048578
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(data); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	return nil
}

func writeJSONError(w http.ResponseWriter, status int, message string) error {
	return writeJSON(w, status, &ErrorEnvelope{Error: message})
}

func (app *application) jsonResponse(w http.ResponseWriter, status int, data any) error {
	return writeJSON(w, status, &Envelope{Data: data})
}
