// go: build integration
package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/dto"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/helpers"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/cache"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/services/otp"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/services/tasks"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/presentation"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/auth"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/healthcheck"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

type stubCache struct{}

func (c stubCache) Get(_ context.Context, _ string) ([]byte, error) {
	return nil, nil
}

func (c stubCache) Set(_ context.Context, _ string, _ any, _ time.Duration) error {
	return nil
}

type stubHIBP struct{}

func (sh *stubHIBP) CheckPasswordIsBreached(_ context.Context, _ string) (bool, error) {
	return false, nil
}

type stubTasks struct {
	mu    sync.Mutex
	calls []tasks.EmailDeliveryPayload
}

func (st *stubTasks) NewEmailDeliveryTask(_ context.Context, task tasks.EmailDeliveryPayload, _ tasks.Priority) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	st.calls = append(st.calls, task)

	return nil
}

func (st *stubTasks) getLatestOTPCode(t *testing.T) string {
	t.Helper()
	st.mu.Lock()
	defer st.mu.Unlock()

	require.NotEmpty(t, st.calls, "no email tasks captured")
	last := st.calls[len(st.calls)-1]
	code, ok := last.Opts["otpCode"]
	require.True(t, ok, "no otpCode in email task opts")

	return code
}

func (st *stubTasks) reset() {
	st.mu.Lock()
	defer st.mu.Unlock()

	st.calls = nil
}

func (st *stubTasks) PingQueue(_ context.Context) bool {
	return true
}

var (
	testServer *httptest.Server
	repo       *repository.Repository
	taskQ      *stubTasks
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	dbName := helpers.MustGetEnvVar("POSTGRES_DB")
	dbUser := helpers.MustGetEnvVar("POSTGRES_USER")
	dbPassword := helpers.MustGetEnvVar("POSTGRES_PASSWORD")
	cleanup := func(containers ...testcontainers.Container) {
		for _, container := range containers {
			if err := testcontainers.TerminateContainer(container); err != nil {
				slog.Error("terminate container failed", "err", err)
			}
		}
	}

	postgresCtr, err := postgres.Run(
		ctx,
		"postgres:18-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		postgres.BasicWaitStrategies(),
		postgres.WithOrderedInitScripts(
			filepath.Join("..", "db", "migrations", "000001_initial.up.sql"),
		),
		postgres.WithSQLDriver("pgx"),
	)
	if err != nil {
		slog.Error("start postgres container failed, terminating...", "err", err)
		cleanup(postgresCtr)

		return
	}

	dbConnString, err := postgresCtr.ConnectionString(ctx)
	if err != nil {
		slog.Error("failed to get db connection string", "err", err)
		cleanup(postgresCtr)

		return
	}

	repo, err = repository.New(dbConnString, stubCache{})
	if err != nil {
		slog.Error("failed to initialize repository", "err", err)
		cleanup(postgresCtr)

		return
	}

	redisCtr, err := tcredis.Run(ctx,
		"redis:7",
		tcredis.WithSnapshotting(10, 1),
		tcredis.WithLogLevel(tcredis.LogLevelVerbose),
	)
	if err != nil {
		slog.Error("start redis container failed, terminating...", "err", err)
		cleanup(postgresCtr, redisCtr)

		return
	}

	redisConnString, err := redisCtr.ConnectionString(ctx)
	if err != nil {
		slog.Error("failed to get redis connection string", "err", err)
		cleanup(postgresCtr, redisCtr)

		return
	}

	opts, err := redis.ParseURL(redisConnString)
	if err != nil {
		cleanup(postgresCtr, redisCtr)

		return
	}

	redisClient := redis.NewClient(opts)
	otpProvider := otp.New(repo)
	hibp := &stubHIBP{}
	taskQ = &stubTasks{}
	cache := cache.New(redisClient)

	auth := auth.New(repo, hibp, otpProvider, taskQ, cache)
	health := healthcheck.New(repo, taskQ)

	router := presentation.SetupRoutes(redisClient, health, auth)

	testServer = httptest.NewTLSServer(router)

	exitCode := m.Run()

	cleanup(postgresCtr, redisCtr)
	testServer.Close()
	os.Exit(exitCode)
}

// newTestClient creates a new http client with its own cookie jar for each test.
func newTestClient(t *testing.T) *http.Client {
	t.Helper()

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	return &http.Client{
		Transport: testServer.Client().Transport,
		Jar:       jar,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func postJSON(t *testing.T, client *http.Client, path string, body any, extraHeaders map[string]string) *http.Response {
	t.Helper()

	jsonBytes, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, testServer.URL+path, bytes.NewReader(jsonBytes))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req) // nolint: gosec
	require.NoError(t, err)

	return resp
}

func decodeJSON[T any](t *testing.T, resp *http.Response) T { // nolint: ireturn
	t.Helper()

	defer resp.Body.Close()

	var result T

	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	return result
}

func registerUser(t *testing.T, client *http.Client, email, password string) dto.OTPRequiredResponse {
	t.Helper()

	resp := postJSON(t, client, "/api/auth/register", dto.EmailPasswordUserInput{ // nolint: bodyclose
		Email:    email,
		Password: password,
	}, nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	return decodeJSON[dto.OTPRequiredResponse](t, resp)
}

func loginUser(t *testing.T, client *http.Client, email, password string) dto.OTPRequiredResponse {
	t.Helper()

	resp := postJSON(t, client, "/api/auth/login", dto.LoginInput{ // nolint: bodyclose
		Email:    email,
		Password: password,
	}, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	return decodeJSON[dto.OTPRequiredResponse](t, resp)
}

func verifyOTP(t *testing.T, client *http.Client, code string, purpose dto.OTPPurpose) dto.AuthResponse {
	t.Helper()

	resp := postJSON(t, client, "/api/auth/verify-otp", dto.VerifyOTPInput{ // nolint: bodyclose
		Purpose: purpose,
		Value:   code,
	}, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	return decodeJSON[dto.AuthResponse](t, resp)
}

func fullSignUpFlow(t *testing.T, client *http.Client, email, password string) dto.AuthResponse {
	t.Helper()

	registerUser(t, client, email, password)
	otpCode := taskQ.getLatestOTPCode(t)

	return verifyOTP(t, client, otpCode, dto.EmailVerification)
}

func TestSignUpAndVerifyOTP(t *testing.T) {
	tests := []struct {
		name, email, password string
		client                *http.Client
	}{
		{
			name:     "happy case: register new user and verify otp - bearer mode",
			client:   newTestClient(t),
			email:    gofakeit.Email(),
			password: gofakeit.Password(true, true, true, true, false, 14),
		},
		{
			name:     "happy case: register new user and verify otp - cookie mode",
			client:   newTestClient(t),
			email:    gofakeit.Email(),
			password: gofakeit.Password(true, true, true, true, false, 14),
		},
		{
			name:     "happy case: duplicate email doesn't leak email existence",
			client:   newTestClient(t),
			email:    gofakeit.Email(),
			password: gofakeit.Password(true, true, true, true, false, 14),
		},
		{
			name:     "sad case: invalid email format returns 400",
			client:   newTestClient(t),
			email:    gofakeit.BeerName(),
			password: gofakeit.Password(true, true, true, true, false, 14),
		},
		{
			name:     "sad case: invalid otp code returns 400",
			client:   newTestClient(t),
			email:    gofakeit.Email(),
			password: gofakeit.Password(true, true, true, true, false, 14),
		},
		{
			name:     "sad case: verify otp without session returns 401",
			client:   newTestClient(t),
			email:    gofakeit.Email(),
			password: gofakeit.Password(true, true, true, true, false, 14),
		},
	}

	for _, tt := range tests {
		taskQ.reset()

		if tt.name == "happy case: register new user and verify otp - bearer mode" {
			regResp := registerUser(t, tt.client, tt.email, tt.password)
			assert.Equal(t, regResp.Status, "otp_required")
			assert.Equal(t, tt.email, regResp.User.Email)

			otpCode := taskQ.getLatestOTPCode(t)
			assert.Equal(t, 6, len(otpCode), "otp code must be 6 digits")

			authResp := verifyOTP(t, tt.client, otpCode, dto.EmailVerification)
			assert.NotEmpty(t, authResp.AccessToken)
			assert.NotEmpty(t, authResp.RefreshToken)
			assert.Equal(t, int64(900), authResp.ExpiresIn)
			assert.Equal(t, tt.email, authResp.User.Email)
			assert.NotEmpty(t, authResp.User.PublicID)
		}

		if tt.name == "happy case: register new user and verify otp - cookie mode" {
			registerUser(t, tt.client, tt.email, tt.password)
			otpCode := taskQ.getLatestOTPCode(t)

			resp := postJSON(t, tt.client, "/api/auth/verify-otp", dto.VerifyOTPInput{ // nolint: bodyclose
				Purpose: dto.EmailVerification,
				Value:   otpCode,
			}, map[string]string{
				"X-Auth-Mode": "cookie",
			})
			require.Equal(t, http.StatusOK, resp.StatusCode)

			cookies := resp.Cookies()
			cookieNames := make(map[string]string)

			for _, c := range cookies {
				cookieNames[c.Name] = c.Value
			}

			assert.Contains(t, cookieNames, "access_token", "access_token cookie should be set")
			assert.Contains(t, cookieNames, "refresh_token", "refresh_token cookie should be set")
			assert.Contains(t, cookieNames, "csrf_token", "csrf_token cookie should be set")

			authResp := decodeJSON[dto.AuthResponse](t, resp)
			assert.Empty(t, authResp.AccessToken, "access_token should be empty for cookie mode")
			assert.Empty(t, authResp.RefreshToken, "refresh_token should be empty for cookie mode")
		}

		if tt.name == "happy case: duplicate email doesn't leak email existence" {
			fullSignUpFlow(t, tt.client, tt.email, tt.password)

			client := newTestClient(t)

			taskQ.reset()

			resp := postJSON(t, client, "/api/auth/register", dto.EmailPasswordUserInput{ // nolint: bodyclose
				Email:    tt.email,
				Password: tt.password,
			}, nil)
			require.Equal(t, http.StatusCreated, resp.StatusCode)

			result := decodeJSON[dto.OTPRequiredResponse](t, resp)
			assert.Equal(t, "otp_required", result.Status)

			taskQ.mu.Lock()
			last := taskQ.calls[len(taskQ.calls)-1]
			taskQ.mu.Unlock()
			assert.Equal(t, dto.TypeSecurityAlertEmail, last.EmailType)
		}

		if tt.name == "sad case: invalid email format returns 400" {
			resp := postJSON(t, tt.client, "/api/auth/register", dto.EmailPasswordUserInput{
				Email:    tt.email,
				Password: tt.password,
			}, nil)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

			resp.Body.Close()
		}

		if tt.name == "sad case: invalid otp code returns 400" {
			registerUser(t, tt.client, tt.email, tt.password)

			resp := postJSON(t, tt.client, "/api/auth/verify-otp", dto.VerifyOTPInput{
				Purpose: dto.EmailVerification,
				Value:   "00000",
			}, nil)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

			resp.Body.Close()
		}

		if tt.name == "sad case: verify otp without session return 401" {
			resp := postJSON(t, tt.client, "/api/auth/verify-otp", dto.VerifyOTPInput{
				Purpose: dto.EmailVerification,
				Value:   "00000",
			}, nil)
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

			resp.Body.Close()
		}
	}
}

func TestLoginAndVerifyOTP(t *testing.T) {
	client := newTestClient(t)
	email := gofakeit.Email()
	password := gofakeit.Password(true, true, true, true, false, 14)

	fullSignUpFlow(t, client, email, password)

	tests := []struct {
		client                *http.Client
		email, name, password string
	}{
		{
			name:     "happy case: successful login",
			client:   newTestClient(t),
			email:    email,
			password: password,
		},
		{
			name:     "sad case: bad password returns 401",
			client:   newTestClient(t),
			email:    email,
			password: gofakeit.Password(true, true, true, true, false, 14),
		},
		{
			name:     "sad case: non-existent email returns 401",
			client:   newTestClient(t),
			email:    gofakeit.Email(),
			password: password,
		},
	}

	for _, tt := range tests {
		taskQ.reset()

		if tt.name == "happy case: successful login" {
			loginResp := loginUser(t, tt.client, tt.email, tt.password)
			assert.Equal(t, "otp_required", loginResp.Status)

			otpCode := taskQ.getLatestOTPCode(t)
			authResp := verifyOTP(t, tt.client, otpCode, dto.Login)

			assert.NotEmpty(t, authResp.AccessToken)
			assert.NotEmpty(t, authResp.RefreshToken)
			assert.Equal(t, email, authResp.User.Email)
		}

		if tt.name == "sad case: bad password returns 401" {
			resp := postJSON(t, tt.client, "/api/auth/login", dto.LoginInput{
				Email:    tt.email,
				Password: tt.password,
			}, nil)
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

			resp.Body.Close()
		}

		if tt.name == "sad case: non-existent email returns 401" {
			resp := postJSON(t, tt.client, "/api/auth/login", dto.LoginInput{
				Email:    tt.email,
				Password: tt.password,
			}, nil)
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

			resp.Body.Close()
		}
	}
}

func TestTokenRefresh(t *testing.T) {
	tests := []struct {
		name, email, password string
		client                *http.Client
	}{
		{
			name:     "happy case: token refresh successful",
			email:    gofakeit.Email(),
			client:   newTestClient(t),
			password: gofakeit.Password(true, true, true, true, false, 14),
		},
		{
			name:     "sad case: old refresh token revoked after use",
			email:    gofakeit.Email(),
			client:   newTestClient(t),
			password: gofakeit.Password(true, true, true, true, false, 14),
		},
		{
			name:     "happy case: refresh token via cookie auth mode",
			email:    gofakeit.Email(),
			client:   newTestClient(t),
			password: gofakeit.Password(true, true, true, true, false, 14),
		},
		{
			name:     "sad case: invalid refresh token returns 401",
			email:    gofakeit.Email(),
			client:   newTestClient(t),
			password: gofakeit.Password(true, true, true, true, false, 14),
		},
	}

	for _, tt := range tests {
		taskQ.reset()

		if tt.name == "happy case: token refresh successful" {
			authResp := fullSignUpFlow(t, tt.client, tt.email, tt.password)
			oldToken := authResp.RefreshToken
			resp := postJSON(t, tt.client, "/api/auth/refresh", dto.RefreshAccessTokenInput{ //nolint: bodyclose
				Token: oldToken,
			}, nil)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			refreshResp := decodeJSON[dto.RefreshAccessTokenResponse](t, resp)
			require.NotEmpty(t, refreshResp.AccessToken)
			require.NotEmpty(t, refreshResp.RefreshToken)
			require.NotEqual(t, oldToken, refreshResp.RefreshToken)
		}

		if tt.name == "happy case: old refresh token revoked after use" {
			authResp := fullSignUpFlow(t, tt.client, tt.email, tt.password)
			oldToken := authResp.RefreshToken
			resp := postJSON(t, tt.client, "/api/auth/refresh", dto.RefreshAccessTokenInput{ //nolint: bodyclose
				Token: oldToken,
			}, nil)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			resp.Body.Close()

			resp = postJSON(t, tt.client, "/api/auth/refresh", dto.RefreshAccessTokenInput{ //nolint: bodyclose
				Token: oldToken,
			}, nil)
			require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			resp.Body.Close()
		}

		if tt.name == "happy case: refresh token via cookie auth mode" {
			registerUser(t, tt.client, tt.email, tt.password)
			otpCode := taskQ.getLatestOTPCode(t)
			resp := postJSON(t, tt.client, "/api/auth/verify-otp", dto.VerifyOTPInput{
				Purpose: dto.EmailVerification,
				Value:   otpCode,
			}, map[string]string{"X-Auth-Mode": "cookie"})
			require.Equal(t, http.StatusOK, resp.StatusCode)
			resp.Body.Close()

			resp = postJSON( // nolint: bodyclose
				t,
				tt.client,
				"/api/auth/refresh",
				nil,
				map[string]string{"X-Auth-Mode": "cookie"},
			)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			refreshResp := decodeJSON[dto.RefreshAccessTokenResponse](t, resp)
			assert.Empty(t, refreshResp.AccessToken)
			assert.Empty(t, refreshResp.RefreshToken)
		}

		if tt.name == "sad case: invalid refresh token returns 401" {
			resp := postJSON(t, tt.client, "/api/auth/refresh", dto.RefreshAccessTokenInput{
				Token: gofakeit.BeerName(),
			}, nil)
			require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			resp.Body.Close()
		}
	}
}

func TestLogout(t *testing.T) {
	tests := []struct {
		name, email, password string
		client                *http.Client
	}{
		{
			name:     "happy case: logout clears cookies and refresh tokens",
			email:    gofakeit.Email(),
			client:   newTestClient(t),
			password: gofakeit.Password(true, true, true, true, false, 14),
		},
		{
			name:     "sad case: unauthenticated request returns 401",
			email:    "",
			client:   newTestClient(t),
			password: "",
		},
	}

	for _, tt := range tests {
		if tt.name == "happy case: logout clears cookies and refresh tokens" {
			authResp := fullSignUpFlow(t, tt.client, tt.email, tt.password)

			resp := postJSON(t, tt.client, "/api/auth/logout", nil, map[string]string{
				"Authorization": "Bearer " + authResp.AccessToken,
			})
			require.Equal(t, http.StatusOK, resp.StatusCode)

			var logoutResp map[string]string

			_ = json.NewDecoder(resp.Body).Decode(&logoutResp)
			resp.Body.Close()
			assert.Equal(t, "logged_out", logoutResp["status"])

			for _, c := range resp.Cookies() {
				if c.Name == "access_token" || c.Name == "refresh_token" || c.Name == "csrf_token" {
					assert.Equal(t, -1, c.MaxAge, "cookie %s should be cleared", c.Name)
				}
			}

			resp = postJSON(t, tt.client, "/api/auth/refresh", dto.RefreshAccessTokenInput{
				Token: authResp.RefreshToken,
			}, nil)
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			resp.Body.Close()
		}

		if tt.name == "sad case: unauthenticated request returns 401" {
			resp := postJSON(t, tt.client, "/api/auth/logout", nil, nil)
			require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			resp.Body.Close()
		}
	}
}
