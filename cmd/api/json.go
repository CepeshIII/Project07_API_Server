package main

import (
	"encoding/json"
	"net/http"

	"github.com/CepeshIII/Project07_API_Server/internal/store"
	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

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
	Data []store.Follower `json:"data"`
}

type PostsEnvelope struct {
	Data []store.Post `json:"data"`
}

type MessageEnvelope struct {
	Data string `json:"data"`
}

type CreateCommentPayload struct {
	UserID  int64  `json:"user_id" example:"1"`
	Content string `json:"content"`
}

type ErrorEnvelope struct {
	Error string `json:"error"`
}

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1048578
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(data)
}

func writeJSONError(w http.ResponseWriter, status int, message string) error {
	return writeJSON(w, status, &ErrorEnvelope{Error: message})
}

func (app *application) jsonResponse(w http.ResponseWriter, status int, data any) error {
	return writeJSON(w, status, &Envelope{Data: data})
}
