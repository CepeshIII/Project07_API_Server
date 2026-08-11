package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/CepeshIII/Project07_API_Server/docs"
	"github.com/CepeshIII/Project07_API_Server/internal/env"
	"github.com/CepeshIII/Project07_API_Server/internal/mailer"
	"github.com/CepeshIII/Project07_API_Server/internal/store"

	httpSwagger "github.com/swaggo/http-swagger" // http-swagger middleware

	"go.uber.org/zap"
)

type application struct {
	config config
	store  store.Storage
	logger *zap.SugaredLogger
	mailer mailer.Client
}

type mailConfig struct {
	exp        time.Duration
	fromEmail  string
	maxRetries int
	sendGrid   sendGridConfig
	mailtrap   mailtrapConfig
}

type sendGridConfig struct {
	apikey string
}

type mailtrapConfig struct {
	apikey string
}

type config struct {
	addr        string
	db          dbConfig
	env         string
	apiURL      string
	mail        mailConfig
	frontendURL string
	loggerEnv   string
}

type dbConfig struct {
	address      string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

func DefaultDBConfig() dbConfig {
	addr := env.GetString("DB_ADDR",
		"postgres://admin:adminpassword@localhost:5433/social?sslmode=disable")

	maxOpenConns := env.GetInt("DB_MAX_OPEN_CONNS", 30)
	maxIdleConns := env.GetInt("DB_MAX_IDLE_CONNS", 30)
	maxIdleTime := env.GetString("DB_MAX_IDLE_TIME", "15m")

	db := dbConfig{
		address:      addr,
		maxOpenConns: maxOpenConns,
		maxIdleConns: maxIdleConns,
		maxIdleTime:  maxIdleTime,
	}

	return db
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromRemoteAddr) // pick one ClientIPFrom* based on your infra, see below
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)

		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL(docsURL),
		))

		r.Route("/posts", func(r chi.Router) {
			r.Post("/", app.createPostHandler)

			r.Route("/{postsID}", func(r chi.Router) {
				r.Use(app.postContextMiddleware)

				r.Get("/", app.getPostHandler)
				r.Get("/comments", app.getPostCommentsHandler)

				r.Patch("/", app.updatePostHandler)

				r.Post("/comments", app.createCommentHandler)

				r.Delete("/", app.deletePostHandler)
			})

		})

		r.Route("/users", func(r chi.Router) {
			r.Put("/activate/{token}", app.activateUserHandler)
			r.Route("/{userID}", func(r chi.Router) {
				r.Use(app.userContextMiddleware)

				r.Get("/", app.getUserHandler)

				r.Get("/followers", app.getFollowersHandler)

				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
				r.Delete("/", app.deleteUserHandler)
			})

			r.Group(func(r chi.Router) {
				r.Get("/feed", app.getUserFeedHandler)

			})
		})

		r.Route("/authentication", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler)
		})

	})

	return r
}

func (app *application) run(mux http.Handler) error {
	// Docs
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = "/v1"

	srv := http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	app.logger.Info(fmt.Sprintf("Server has start at %s\n", app.config.addr))

	err := srv.ListenAndServe()

	app.logger.Info(fmt.Sprintf("Server has finish at %s\n", app.config.addr))

	return err
}
