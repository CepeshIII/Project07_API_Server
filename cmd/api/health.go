package main

import (
	"net/http"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "ok",
		"env":     app.config.env,
		"version": version,
	}

	err := app.jsonResponse(w, http.StatusOK, data)
	// err := errors.New("Server has been fall")
	if err != nil {
		app.internalServerError(w, r, err)
	}

	// w.Header().Set("Content-Type", "application/json")
	// w.Write([]byte(`{"status": "ok"}`))
	// app.store.Posts.Create(r.Context(), &store.Post{})
	// app.store.Users.Create(r.Context(), &store.User{})
}
