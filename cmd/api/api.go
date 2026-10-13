// Package main is the entry point for the Project07 API Server,
// handling server initialization, routing, and configuration setup.
package main

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"

	"github.com/CepeshIII/Project07_API_Server/docs"
	"github.com/CepeshIII/Project07_API_Server/internal/auth"
	config1 "github.com/CepeshIII/Project07_API_Server/internal/config"
	"github.com/CepeshIII/Project07_API_Server/internal/mailer"
	"github.com/CepeshIII/Project07_API_Server/internal/ratelimiter"
	"github.com/CepeshIII/Project07_API_Server/internal/store"
	"github.com/CepeshIII/Project07_API_Server/internal/store/cache"
	httpSwagger "github.com/swaggo/http-swagger" // http-swagger middleware

	"go.uber.org/zap"
)

type application struct {
	config       config1.Config
	store        store.Storage
	logger       *zap.SugaredLogger
	mailer       mailer.Client
	auth         auth.Authenticator
	cacheStorage cache.Storage
	rateLimiter  ratelimiter.Limiter
	validator    *validator.Validate
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromRemoteAddr) // pick one ClientIPFrom* based on your infra, see below
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(app.rateLimiterMiddleWare)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(app.config.Server.RequestTimeout))

	docsURL := app.config.Addr + "/swagger/doc.json"

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)
		r.With(app.basicAuthMiddleware()).Get("/metrics", expvar.Handler().ServeHTTP)

		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL(docsURL),
		))

		r.Route("/posts", func(r chi.Router) {
			r.Get("/", app.getAllPostsHandler)

			r.Group(func(r chi.Router) {
				r.Use(app.accessTokenMiddleware)

				r.Post("/", app.createPostHandler)
			})

			r.Route("/{postsID}", func(r chi.Router) {
				r.Use(app.postIDContextMiddleware)
				// r.Use(app.postContextMiddleware)

				r.With(app.postContextMiddleware).Get("/", app.getPostHandler)
				r.Get("/comments", app.getPostCommentsHandler)

				r.Group(func(r chi.Router) {
					r.Use(app.accessTokenMiddleware)

					r.Patch("/", app.checkPostOwnership("moderator", app.updatePostHandler))
					r.Delete("/", app.checkPostOwnership("admin", app.deletePostHandler))
					r.Post("/comments", app.createCommentHandler)
				})
			})
		})

		r.Route("/users", func(r chi.Router) {
			r.Put("/activate/{token}", app.activateUserHandler)
			r.Options("/activate/{token}", app.healthCheckHandler)

			r.Route("/{userID}", func(r chi.Router) {
				r.Use(app.userContextMiddleware)

				r.Get("/", app.getUserHandler)
				r.Get("/feed", app.getUserFeedHandler)
				r.Get("/followers", app.getFollowersHandler)

				r.Group(func(r chi.Router) {
					r.Use(app.accessTokenMiddleware)

					r.Put("/follow", app.followUserHandler)
					r.Put("/unfollow", app.unfollowUserHandler)
					r.Delete("/", app.checkUserOwnership("admin", app.deleteUserHandler))
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
	docs.SwaggerInfo.Host = app.config.ExternalURL
	docs.SwaggerInfo.BasePath = "/v1"

	srv := http.Server{
		Addr:         app.config.Addr,
		Handler:      mux,
		WriteTimeout: app.config.Server.WriteTimeout,
		ReadTimeout:  app.config.Server.ReadTimeout,
		IdleTimeout:  app.config.Server.IdleTimeout,
	}

	// Create a channel to communicate the final shutdown status (an error or nil) back to the main function
	shutdown := make(chan error)

	// Call the clean helper method instead of a massive inline block
	app.waitForShutdown(shutdown, &srv)

	app.logger.Info(
		"Server starting at " + app.config.Addr,
	)

	// Start server in a separate goroutine
	go func() {
		app.logger.Infow("Starting server", "addr", srv.Addr)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			shutdown <- err
		}
	}()

	err := <-shutdown
	if err != nil {
		return fmt.Errorf("HTTP server shutdown failed: %w", err)
	}

	app.logger.Info(
		"Server finished at " + app.config.Addr,
	)

	return nil
}

func (app *application) waitForShutdown(shutdown chan error, srv *http.Server) {
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		s := <-quit

		ctx, cancel := context.WithTimeout(
			context.Background(),
			app.config.Server.ShutdownTimeout,
		)
		defer cancel()

		app.logger.Infow("Signal caught", "signal", s.String())

		shutdown <- srv.Shutdown(ctx)
	}()
}
