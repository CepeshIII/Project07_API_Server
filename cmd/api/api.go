package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/CepeshIII/Project07_API_Server/docs"
	"github.com/CepeshIII/Project07_API_Server/internal/auth"
	"github.com/CepeshIII/Project07_API_Server/internal/env"
	"github.com/CepeshIII/Project07_API_Server/internal/mailer"
	"github.com/CepeshIII/Project07_API_Server/internal/store"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger" // http-swagger middleware

	"go.uber.org/zap"
)

type application struct {
	config config
	store  store.Storage
	logger *zap.SugaredLogger
	mailer mailer.Client
	auth   auth.Authenticator
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

type jwtAuthConfig struct {
	jwtSecret string
	iss       string
}

type config struct {
	addr        string
	db          dbConfig
	env         string
	apiURL      string
	mail        mailConfig
	frontendURL string
	loggerEnv   string
	auth        authConfig
}

type authConfig struct {
	basic   basicAuthConfig
	jwtAuth jwtAuthConfig
	tokens  tokensConfig
}

type basicAuthConfig struct {
	username string
	password string
}

type tokensConfig struct {
	accessTokenExp  time.Duration
	sessionTokenExp time.Duration
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

	// MUST BE FIRST: Handle CORS Preflights (OPTIONS)
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://*", "http://*"}, // Add your frontend URL
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum cache duration for preflight requests
	}))

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

		r.With(app.AuthMiddleware()).Get("/health", app.healthCheckHandler)

		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL(docsURL),
		))

		r.Route("/posts", func(r chi.Router) {

			r.Group(func(r chi.Router) {
				r.Use(app.acssesTokenMiddleware)

				r.Post("/", app.createPostHandler)
			})

			r.Route("/{postsID}", func(r chi.Router) {
				r.Use(app.postContextMiddleware)

				r.Get("/", app.getPostHandler)
				r.Get("/comments", app.getPostCommentsHandler)

				r.Group(func(r chi.Router) {
					r.Use(app.acssesTokenMiddleware)

					r.Patch("/", app.updatePostHandler)
					r.Post("/comments", app.createCommentHandler)
					r.Delete("/", app.deletePostHandler)
				})
			})

		})

		r.Route("/users", func(r chi.Router) {

			r.Put("/activate/{token}", app.activateUserHandler)
			r.Options("/activate/{token}", app.healthCheckHandler)

			r.Group(func(r chi.Router) {
				r.Get("/feed", app.getUserFeedHandler)
			})

			r.Route("/{userID}", func(r chi.Router) {
				r.Use(app.userContextMiddleware)

				r.Get("/", app.getUserHandler)
				r.Get("/followers", app.getFollowersHandler)

				r.Group(func(r chi.Router) {
					r.Use(app.acssesTokenMiddleware)

					r.Put("/follow", app.followUserHandler)
					r.Put("/unfollow", app.unfollowUserHandler)
					r.Delete("/", app.deleteUserHandler)
				})
			})

		})

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", app.registerUserHandler)
			r.Post("/refresh", app.refreshAccessTokenHandler)
			r.Post("/login", app.loginUserHandler)
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
