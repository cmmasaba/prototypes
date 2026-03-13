// Package presentation initializes the server, open telemetry and router.
package presentation

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/presentation/rest"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/usecase"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/healthcheck"
	"github.com/go-chi/chi"
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

func isRunningInDebug() bool {
	debug, ok := os.LookupEnv("DEBUG")
	if !ok {
		debug = "false"
	}

	status, err := strconv.ParseBool(debug)
	if err != nil {
		slog.Warn("error parsing debug flag, defaulting to false", "err", err)

		return false
	}

	return status
}

// PrepareServer initalizes infrastructure and usecases layers, then sets up the router.
func PrepareServer() (http.Handler, error) {
	database, err := repository.New()
	if err != nil {
		slog.Error("error initializing db", "err", err)

		return nil, err
	}

	infrastructure, err := infrastructure.New(database)
	if err != nil {
		slog.Error("error initializing infrastructure layer", "err", err)

		return nil, err
	}

	healthcheckUC := healthcheck.New(infrastructure)

	allUsecases := usecase.New(healthcheckUC)

	r := setupRoutes(allUsecases)

	return r, nil
}

func setupRoutes(usecases *usecase.Usecase) *chi.Mux {
	debugStatus := isRunningInDebug()
	logFormat := httplog.SchemaOTEL.Concise(debugStatus)
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

	handlers := rest.New(usecases)

	r := chi.NewRouter()

	r.Use(middleware.CleanPath)
	r.Use(middleware.RequestID)
	r.Use(httplog.RequestLogger(logger, &httplog.Options{
		Level:  slog.LevelInfo,
		Schema: httplog.SchemaOTEL,
		Skip: func(req *http.Request, respStatus int) bool {
			return respStatus == 404 || respStatus == 405
		},
		LogRequestBody:  func(_ *http.Request) bool { return debugStatus },
		LogResponseBody: func(_ *http.Request) bool { return debugStatus },
	}))
	r.Use(telemetry.MetricsMiddleware(serviceName))
	r.Use(middleware.Heartbeat("/ping"))
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
	r.Mount("/debug", middleware.Profiler())
	r.Group(func(r chi.Router) {
		r.Use(httprate.LimitByIP(100, time.Minute*1))

		r.Get("/health", handlers.HealthCheck)
	})

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Route("/api", func(r chi.Router) {
			r.Route("/links", func(r chi.Router) {
				r.Post("", func(w http.ResponseWriter, r *http.Request) {})
				r.Get("/{code}", func(w http.ResponseWriter, r *http.Request) {})
				r.Delete("/{code}", func(w http.ResponseWriter, r *http.Request) {})
			})

			r.Route("/clicks", func(r chi.Router) {
				r.Post("", func(w http.ResponseWriter, r *http.Request) {})
				r.Get("", func(w http.ResponseWriter, r *http.Request) {})
			})

			r.Route("/users", func(r chi.Router) {
				r.Post("", func(w http.ResponseWriter, r *http.Request) {})
				r.Get("/{slug}", func(w http.ResponseWriter, r *http.Request) {})
			})
		})
	})

	return r
}
