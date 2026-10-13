package main

import (
	"net/http"
)

type HealthCheckData struct {
	Status  string `example:"ok"          json:"status"`
	Env     string `example:"development" json:"env"`
	Version string `example:"1.0.0"       json:"version"`
}

// HealthCheck godoc
//
//	@Summary		Health check
//	@Description	Returns the API health status, environment, and version
//	@Tags			health
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Failure		500	{object}	ErrorEnvelope
//	@Router			/health [get]
func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := HealthCheckData{
		Status:  "ok",
		Env:     app.config.Env,
		Version: version,
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
