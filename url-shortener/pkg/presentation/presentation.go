// Package presentation initializes the server, open telemetry and router.
package presentation

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/services/hibp"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/presentation/rest"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/usecase"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/auth"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/healthcheck"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v3"
	"github.com/go-chi/httprate"
)

const (
	throttleLimit  = 100
	backlogLimit   = 100
	backlogTimeout = time.Second * 60
	requestTimeout = time.Second * 15
)

// PrepareServer initializes infrastructure and usecases layers, then sets up the router.
func PrepareServer() (http.Handler, error) {
	database, err := repository.New()
	if err != nil {
		slog.Error("error initializing db", "err", err)

		return nil, err
	}

	pwned, err := hibp.New()
	if err != nil {
		slog.Error("error initializing HIBP api")
	}

	infrastructure, err := infrastructure.New(database, pwned)
	if err != nil {
		slog.Error("error initializing infrastructure layer", "err", err)

		return nil, err
	}

	healthcheckUC := healthcheck.New(infrastructure)
	authUC := auth.New(infrastructure)

	r := setupRoutes(healthcheckUC, authUC)

	return r, nil
}

func setupRoutes(
	healthUC *healthcheck.UsecaseImplHealth,
	userUC *auth.UsecaseImplUser,
) *chi.Mux {
	debug := os.Getenv("ENVIRONMENT") == "dev"
	logFormat := httplog.SchemaOTEL.Concise(debug)
	serviceName := "url-shortener"

	multiHandler := telemetry.NewMultiHandler(
		telemetry.NewHandler(serviceName),
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			ReplaceAttr: logFormat.ReplaceAttr,
		}),
	).WithAttrs([]slog.Attr{
		{Key: "version", Value: slog.StringValue(os.Getenv("VERSION"))},
		{Key: "environment", Value: slog.StringValue(os.Getenv("ENVIRONMENT"))},
		{Key: "app", Value: slog.StringValue(serviceName)},
	})

	logger := slog.New(multiHandler)
	slog.SetDefault(logger)

	usecases := usecase.New(healthUC, userUC)

	handlers := rest.New(usecases)

	r := chi.NewRouter()

	r.Use(middleware.CleanPath)
	r.Use(middleware.RequestID)
	r.Use(httplog.RequestLogger(logger, &httplog.Options{
		Level:  slog.LevelInfo,
		Schema: httplog.SchemaOTEL,
		Skip: func(_ *http.Request, respStatus int) bool {
			return respStatus == 404 || respStatus == 405
		},
		LogRequestBody:  func(_ *http.Request) bool { return debug },
		LogResponseBody: func(_ *http.Request) bool { return debug },
	}))
	r.Use(telemetry.MetricsMiddleware(serviceName))
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS", "PUT"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(middleware.ThrottleBacklog(throttleLimit, backlogLimit, backlogTimeout))
	r.Use(middleware.Timeout(requestTimeout))

	// Public routes
	r.Group(func(r chi.Router) {
		r.Use(httprate.LimitByIP(100, time.Minute*1))

		r.Get("/health", handlers.HealthCheck)
	})

	r.Group(func(r chi.Router) {
		r.Route("/api", func(r chi.Router) {
			r.Route("/auth", func(r chi.Router) {
				r.Post("/register", handlers.CreateUserEmailPassword)
				r.Post("/login", handlers.Login)
				r.Post("/validate-password", handlers.ValidatePassword)
				r.Post("/refresh", handlers.RefreshAccessToken)
			})

			r.Route("/links", func(r chi.Router) {
				r.Use(AuthMiddleware(userUC))

				r.Post("/", func(_ http.ResponseWriter, _ *http.Request) {})
				r.Get("/{code}", func(_ http.ResponseWriter, _ *http.Request) {})
				r.Delete("/{code}", func(_ http.ResponseWriter, _ *http.Request) {})
			})

			r.Route("/clicks", func(r chi.Router) {
				r.Use(AuthMiddleware(userUC))

				r.Post("/", func(_ http.ResponseWriter, _ *http.Request) {})
				r.Get("/", func(_ http.ResponseWriter, _ *http.Request) {})
			})

			r.Route("/users", func(r chi.Router) {
				r.Use(AuthMiddleware(userUC))

				r.Post("/", func(_ http.ResponseWriter, _ *http.Request) {})
				r.Get("/{slug}", func(_ http.ResponseWriter, _ *http.Request) {})
			})

			r.Route("/debug", func(r chi.Router) {
				r.Use(AuthMiddleware(userUC))

				r.Mount("/", middleware.Profiler())
			})
		})
	})

	return r
}
