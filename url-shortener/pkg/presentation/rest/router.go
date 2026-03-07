package rest

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

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

func router() *chi.Mux {
	debugStatus := isRunningInDebug()
	logFormat := httplog.SchemaOTEL.Concise(debugStatus)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: logFormat.ReplaceAttr,
	})).With(
		slog.String("app", "url-shortener"),
		slog.String("version", os.Getenv("VERSION")),
		slog.String("env", os.Getenv("ENVIRONMENT")),
	)

	r := chi.NewRouter()

	r.Use(middleware.CleanPath)
	r.Use(httplog.RequestLogger(logger, &httplog.Options{
		Level:  slog.LevelInfo,
		Schema: httplog.SchemaOTEL,
		Skip: func(req *http.Request, respStatus int) bool {
			return respStatus == 404 || respStatus == 405
		},
		LogRequestBody:  func(_ *http.Request) bool { return debugStatus },
		LogResponseBody: func(_ *http.Request) bool { return debugStatus },
	}))
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

		r.Get("/health", health)
	})

	// Authenticated routes
	r.Group(func(r chi.Router) {
		// r.Use(httprate.(100, time.Minute*1))
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
