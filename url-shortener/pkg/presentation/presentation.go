// Package presentation initializes the server, open telemetry and router.
package presentation

import (
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alexedwards/scs/goredisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/helpers"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/presentation/rest"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/usecase"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/auth"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/healthcheck"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v3"
	"github.com/go-chi/httprate"
	"github.com/redis/go-redis/v9"
)

const (
	requestTimeout = time.Second * 15
)

// PrepareServer initializes infrastructure and usecases layers, then sets up the router.
func PrepareServer() (http.Handler, error) {
	infrastructure, err := infrastructure.New()
	if err != nil {
		slog.Error("error initializing infrastructure layer", "err", err)

		return nil, err
	}

	healthcheckUC := healthcheck.New(
		infrastructure.DB,
		infrastructure.TasksQueue,
	)
	authUC := auth.New(
		infrastructure.DB,
		infrastructure.Hibp,
		infrastructure.Otp,
		infrastructure.TasksQueue,
	)

	r := setupRoutes(healthcheckUC, authUC)

	return r, nil
}

func setupRoutes(
	healthUC *healthcheck.UsecaseImpl,
	userUC *auth.UsecaseImpl,
) *chi.Mux {
	debug := helpers.MustGetEnvVar("ENVIRONMENT") == "dev"

	logFormat := httplog.SchemaOTEL.Concise(debug)

	serviceName := "url-shortener"

	redisURL := helpers.MustGetEnvVar("REDIS_URL")

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		panic(err)
	}

	client := redis.NewClient(opts)

	sessionManager := scs.New()
	sessionManager.Lifetime = time.Minute * 15
	sessionManager.Store = goredisstore.New(client)
	sessionManager.Cookie = scs.SessionCookie{
		Name:        "urlshortener-session",
		HttpOnly:    true,
		Secure:      true,
		SameSite:    http.SameSiteLaxMode,
		Partitioned: false,
		Persist:     true,
		Path:        "/",
		Domain:      "",
	}

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

	handlers := rest.New(usecases, sessionManager)

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
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // Max request body size of 1MB
			next.ServeHTTP(w, r)
		})
	})
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   strings.Split(helpers.MustGetEnvVar("CORS_ALLOWED_ORIGINS"), ","),
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS", "PUT"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(middleware.Timeout(requestTimeout))

	// Public routes
	r.Group(func(r chi.Router) {
		r.Use(httprate.LimitByIP(10, time.Minute*1))

		r.Get("/health", handlers.HealthCheck)
	})

	r.Group(func(r chi.Router) {
		r.Route("/api", func(r chi.Router) {
			r.Route("/auth", func(r chi.Router) {
				r.Route("/register", func(r chi.Router) {
					r.Use(sessionManager.LoadAndSave)
					r.Use(httprate.LimitByIP(5, time.Minute*1))
					r.Post("/", handlers.CreateUserEmailPassword)
				})
				r.Route("/login", func(r chi.Router) {
					r.Use(sessionManager.LoadAndSave)
					r.Use(httprate.LimitByIP(10, time.Minute*1))
					r.Post("/", handlers.Login)
				})
				r.Route("/validate-password", func(r chi.Router) {
					r.Use(httprate.LimitByIP(10, time.Minute*1))
					r.Post("/", handlers.ValidatePassword)
				})
				r.Route("/refresh", func(r chi.Router) {
					r.Use(httprate.LimitByIP(5, time.Minute*10))
					r.Post("/", handlers.RefreshAccessToken)
				})
				r.Route("/verify-otp", func(r chi.Router) {
					r.Use(sessionManager.LoadAndSave)
					r.Use(httprate.Limit(
						5,
						15*time.Minute,
						httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
							return sessionManager.GetString(r.Context(), "user_id"), nil
						})))
					r.Post("/", handlers.VerifyOTP)
				})
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

			if debug {
				r.Route("/debug", func(r chi.Router) {
					r.Mount("/", middleware.Profiler())
				})
			}
		})
	})

	return r
}
