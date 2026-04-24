package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/brianvoe/gofakeit"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/dto"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/presentation/rest/mocks"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/auth"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/shortener"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var errInternal = errors.New("an error occurred")

type args struct {
	respWriter *httptest.ResponseRecorder
	req        *http.Request
	session    *scs.SessionManager
}

func primeTestContext(sm *scs.SessionManager, r *http.Request) *http.Request {
	ctx, _ := sm.Load(r.Context(), "")
	return r.WithContext(ctx)
}

func TestHandlers_ShortenURL(t *testing.T) {
	url := "https://example.com"
	ownershipToken := "xasuascca"

	tests := []struct {
		name  string
		setup func(uc *mocks.Mockusecases) args
	}{
		{
			name: "sad case: bad input data",
			setup: func(_ *mocks.Mockusecases) args {
				body := []byte(`bad data`)

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/shorten",
						bytes.NewReader(body),
					),
				}
			},
		},
		{
			name: "sad case: missing url in input data",
			setup: func(_ *mocks.Mockusecases) args {
				body, _ := json.Marshal(dto.ShortenURLInput{})

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/shorten",
						bytes.NewReader(body),
					),
				}
			},
		},
		{
			name: "sad case: invalid uri format",
			setup: func(uc *mocks.Mockusecases) args {
				body, _ := json.Marshal(dto.ShortenURLInput{
					URL: url,
				})

				uc.EXPECT().ShortenURL(mock.Anything, mock.AnythingOfType("*dto.ShortenURLInput")).Return(nil, shortener.ErrInvalidURIFormat)

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/shorten",
						bytes.NewReader(body),
					),
				}
			},
		},
		{
			name: "sad case: input url too long",
			setup: func(uc *mocks.Mockusecases) args {
				body, _ := json.Marshal(dto.ShortenURLInput{
					URL: url,
				})

				uc.EXPECT().ShortenURL(mock.Anything, mock.AnythingOfType("*dto.ShortenURLInput")).Return(nil, shortener.ErrURLTooLong)

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/shorten",
						bytes.NewReader(body),
					),
				}
			},
		},
		{
			name: "sad case: internal server error",
			setup: func(uc *mocks.Mockusecases) args {
				body, _ := json.Marshal(dto.ShortenURLInput{
					URL: url,
				})

				uc.EXPECT().ShortenURL(mock.Anything, mock.AnythingOfType("*dto.ShortenURLInput")).Return(nil, errInternal)

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/shorten",
						bytes.NewReader(body),
					),
				}
			},
		},
		{
			name: "happy case: anonymous request success",
			setup: func(uc *mocks.Mockusecases) args {
				body, _ := json.Marshal(dto.ShortenURLInput{
					URL: url,
				})

				uc.EXPECT().ShortenURL(mock.Anything, mock.AnythingOfType("*dto.ShortenURLInput")).Return(&dto.ShortenURLResponse{
					ShortURL:       gofakeit.URL(),
					OwnershipToken: ownershipToken,
				}, nil)

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/shorten",
						bytes.NewReader(body),
					),
				}
			},
		},
		{
			name: "happy case: authenticated request success",
			setup: func(uc *mocks.Mockusecases) args {
				body, _ := json.Marshal(dto.ShortenURLInput{
					URL:       url,
					ExpiresAt: time.Now().Add(5 * time.Minute),
				})

				uc.EXPECT().ShortenURL(mock.Anything, mock.AnythingOfType("*dto.ShortenURLInput")).Return(&dto.ShortenURLResponse{
					ShortURL: gofakeit.URL(),
				}, nil)

				req := httptest.NewRequestWithContext(
					context.Background(),
					http.MethodPost,
					"/api/shorten",
					bytes.NewReader(body),
				)

				return args{
					respWriter: httptest.NewRecorder(),
					req:        req,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecases := mocks.NewMockusecases(t)
			params := tt.setup(usecases)

			h := New(usecases, params.session)

			h.ShortenURL(params.respWriter, params.req)

			result := params.respWriter.Result()

			var resp dto.ShortenURLResponse

			switch tt.name {
			case "happy case: anonymous request success":
				_ = json.NewDecoder(result.Body).Decode(&resp)

				assert.Equal(t, http.StatusOK, result.StatusCode)
				assert.NotEmpty(t, resp.ShortURL)
				assert.NotEmpty(t, resp.OwnershipToken)
			case "happy case: authenticated request success":
				_ = json.NewDecoder(result.Body).Decode(&resp)

				assert.Equal(t, http.StatusOK, result.StatusCode)
				assert.NotEmpty(t, resp.ShortURL)
				assert.Empty(t, resp.OwnershipToken)
			case "sad case: internal server error":
				assert.Equal(t, http.StatusInternalServerError, result.StatusCode)
			default:
				assert.Equal(t, http.StatusBadRequest, result.StatusCode)
			}
		})
	}
}

func TestHandlers_RedirectToOriginalURL(t *testing.T) {
	tests := []struct {
		name  string
		setup func(uc *mocks.Mockusecases) args
	}{
		{
			name: "sad case: shortened url code too short",
			setup: func(_ *mocks.Mockusecases) args {
				req := httptest.NewRequestWithContext(
					context.Background(),
					http.MethodGet,
					"/api/shorten/",
					nil,
				)

				req.SetPathValue("code", "abcde")

				return args{
					respWriter: httptest.NewRecorder(),
					req:        req,
				}
			},
		},
		{
			name: "sad case: shortened url code too long",
			setup: func(_ *mocks.Mockusecases) args {
				req := httptest.NewRequestWithContext(
					context.Background(),
					http.MethodGet,
					"/api/shorten/",
					nil,
				)

				req.SetPathValue("code", "abcdefghi")

				return args{
					respWriter: httptest.NewRecorder(),
					req:        req,
				}
			},
		},
		{
			name: "sad case: shortened url has expired",
			setup: func(uc *mocks.Mockusecases) args {
				req := httptest.NewRequestWithContext(
					context.Background(),
					http.MethodGet,
					"/api/shorten/",
					nil,
				)

				req.SetPathValue("code", "abcdefg")

				uc.EXPECT().GetOriginalURL(mock.Anything, mock.Anything).Return("", shortener.ErrURLExpired)

				return args{
					respWriter: httptest.NewRecorder(),
					req:        req,
				}
			},
		},
		{
			name: "sad case: shortened url not found",
			setup: func(uc *mocks.Mockusecases) args {
				req := httptest.NewRequestWithContext(
					context.Background(),
					http.MethodGet,
					"/api/shorten/",
					nil,
				)

				req.SetPathValue("code", "abcdefg")

				uc.EXPECT().GetOriginalURL(mock.Anything, mock.Anything).Return("", shortener.ErrURLNotFound)

				return args{
					respWriter: httptest.NewRecorder(),
					req:        req,
				}
			},
		},
		{
			name: "sad case: internal server error",
			setup: func(uc *mocks.Mockusecases) args {
				req := httptest.NewRequestWithContext(
					context.Background(),
					http.MethodGet,
					"/api/shorten/",
					nil,
				)

				req.SetPathValue("code", "abcdefg")

				uc.EXPECT().GetOriginalURL(mock.Anything, mock.Anything).Return("", errInternal)

				return args{
					respWriter: httptest.NewRecorder(),
					req:        req,
				}
			},
		},
		{
			name: "happy case: successful redirect",
			setup: func(uc *mocks.Mockusecases) args {
				req := httptest.NewRequestWithContext(
					context.Background(),
					http.MethodGet,
					"/api/shorten/",
					nil,
				)

				req.SetPathValue("code", "abcdefg")

				uc.EXPECT().GetOriginalURL(mock.Anything, mock.Anything).Return("", nil)

				return args{
					respWriter: httptest.NewRecorder(),
					req:        req,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecases := mocks.NewMockusecases(t)
			params := tt.setup(usecases)

			h := New(usecases, params.session)

			h.RedirectToOriginalURL(params.respWriter, params.req)

			result := params.respWriter.Result()

			switch tt.name {
			case "happy case: successful redirect":
				assert.Equal(t, http.StatusFound, result.StatusCode)
			case "sad case: shortened url has expired":
				assert.Equal(t, http.StatusGone, result.StatusCode)
			case "sad case: internal server error":
				assert.Equal(t, http.StatusInternalServerError, result.StatusCode)
			default:
				assert.Equal(t, http.StatusNotFound, result.StatusCode)
			}
		})
	}
}

func TestHandlers_RequestNewOTP(t *testing.T) {
	tests := []struct {
		name  string
		setup func(uc *mocks.Mockusecases) args
	}{
		{
			name: "sad case: bad request body",
			setup: func(_ *mocks.Mockusecases) args {
				body := []byte(`bad body`)

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/auth/verify-otp",
						bytes.NewReader(body),
					),
				}
			},
		},
		{
			name: "sad case: missing required fields",
			setup: func(_ *mocks.Mockusecases) args {
				body, _ := json.Marshal(dto.RequestOTPInput{})

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/auth/verify-otp",
						bytes.NewReader(body),
					),
				}
			},
		},
		{
			name: "sad case: internal server error",
			setup: func(uc *mocks.Mockusecases) args {
				body, _ := json.Marshal(dto.RequestOTPInput{
					UserPublicID: gofakeit.UUID(),
					Recipient:    gofakeit.Email(),
					Purpose:      dto.EmailVerification,
				})

				uc.EXPECT().RequestNewOTP(mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("dto.OTPPurpose")).Return(errInternal)

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/auth/verify-otp",
						bytes.NewReader(body),
					),
				}
			},
		},
		{
			name: "happy case: successful otp request",
			setup: func(uc *mocks.Mockusecases) args {
				body, _ := json.Marshal(dto.RequestOTPInput{
					UserPublicID: gofakeit.UUID(),
					Recipient:    gofakeit.Email(),
					Purpose:      dto.EmailVerification,
				})

				uc.EXPECT().RequestNewOTP(mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("dto.OTPPurpose")).Return(nil)

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/auth/verify-otp",
						bytes.NewReader(body),
					),
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecases := mocks.NewMockusecases(t)
			params := tt.setup(usecases)

			h := New(usecases, params.session)

			h.RequestNewOTP(params.respWriter, params.req)

			result := params.respWriter.Result()

			switch tt.name {
			case "happy case: successful otp request":
				assert.Equal(t, http.StatusOK, result.StatusCode)
			case "sad case: internal server error":
				assert.Equal(t, http.StatusInternalServerError, result.StatusCode)
			default:
				assert.Equal(t, http.StatusBadRequest, result.StatusCode)
			}
		})
	}
}

func TestHandlers_Logout(t *testing.T) {
	tests := []struct {
		name  string
		setup func(uc *mocks.Mockusecases) args
	}{
		{
			name: "sad case: internal server error",
			setup: func(uc *mocks.Mockusecases) args {
				uc.EXPECT().Logout(mock.Anything).Return(errInternal)

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/auth/logout",
						nil,
					),
				}
			},
		},
		{
			name: "happy case: successful logout",
			setup: func(uc *mocks.Mockusecases) args {
				uc.EXPECT().Logout(mock.Anything).Return(nil)

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/auth/logout",
						nil,
					),
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecases := mocks.NewMockusecases(t)
			params := tt.setup(usecases)

			h := New(usecases, params.session)

			h.Logout(params.respWriter, params.req)

			result := params.respWriter.Result()

			switch tt.name {
			case "happy case: successful logout":
				type logoutBody struct {
					Status string `json:"status"`
				}

				var resp logoutBody

				_ = json.NewDecoder(result.Body).Decode(&resp)

				assert.Equal(t, http.StatusOK, result.StatusCode)
				assert.Equal(t, resp.Status, "logged_out")
			default:
				assert.Equal(t, http.StatusInternalServerError, result.StatusCode)
			}
		})
	}
}

func TestHandlers_GithubOAuthCallback(t *testing.T) {
	tests := []struct {
		name  string
		setup func(uc *mocks.Mockusecases) args
	}{
		{
			name: "sad case: internal server error",
			setup: func(uc *mocks.Mockusecases) args {
				uc.EXPECT().OAuthFlowCallback(mock.Anything, mock.Anything, mock.Anything).Return(nil, "/", errInternal)

				data := url.Values{}

				data.Set("code", "asxvf")

				req := httptest.NewRequestWithContext(
					context.Background(),
					http.MethodPost,
					"/api/auth/oauth/github/callback",
					strings.NewReader(data.Encode()),
				)

				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				return args{
					respWriter: httptest.NewRecorder(),
					req:        req,
				}
			},
		},
		{
			name: "happy case: successful auth with cookies",
			setup: func(uc *mocks.Mockusecases) args {
				createdUser := &dto.AuthResponse{
					User: dto.User{
						PublicID:  gofakeit.UUID(),
						Email:     gofakeit.Email(),
						CreatedAt: time.Now(),
					},
					AccessToken:  "axscg",
					RefreshToken: "askfs",
					ExpiresIn:    900,
				}

				uc.EXPECT().OAuthFlowCallback(mock.Anything, mock.Anything, mock.Anything).Return(createdUser, "/", nil)

				data := url.Values{}

				data.Set("code", "asxvf")

				req := httptest.NewRequestWithContext(
					context.Background(),
					http.MethodPost,
					"/api/auth/oauth/github/callback",
					strings.NewReader(data.Encode()),
				)

				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				req.Header.Set(authModeHeader, "cookie")

				return args{
					respWriter: httptest.NewRecorder(),
					req:        req,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecases := mocks.NewMockusecases(t)
			params := tt.setup(usecases)
			h := New(usecases, params.session)

			h.GithubOAuthCallback(params.respWriter, params.req)

			result := params.respWriter.Result()

			var resp dto.AuthResponse

			switch tt.name {
			case "happy case: successful auth with cookies":
				_ = json.NewDecoder(result.Body).Decode(&resp)

				assert.Equal(t, http.StatusSeeOther, result.StatusCode)
				assert.Empty(t, resp.AccessToken)
				assert.Empty(t, resp.RefreshToken)
			default:
				assert.Equal(t, http.StatusInternalServerError, result.StatusCode)
			}
		})
	}
}

func TestHandlers_GoogleOAuthCallback(t *testing.T) {
	tests := []struct {
		name  string
		setup func(uc *mocks.Mockusecases) args
	}{
		{
			name: "sad case: internal server error",
			setup: func(uc *mocks.Mockusecases) args {
				uc.EXPECT().OAuthFlowCallback(mock.Anything, mock.Anything, mock.Anything).Return(nil, "/api/", errInternal)

				data := url.Values{}

				data.Set("code", "asxvf")

				req := httptest.NewRequestWithContext(
					context.Background(),
					http.MethodPost,
					"/api/auth/oauth/google/callback",
					strings.NewReader(data.Encode()),
				)

				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				return args{
					respWriter: httptest.NewRecorder(),
					req:        req,
				}
			},
		},
		{
			name: "happy case: successful auth with cookies",
			setup: func(uc *mocks.Mockusecases) args {
				createdUser := &dto.AuthResponse{
					User: dto.User{
						PublicID:  gofakeit.UUID(),
						Email:     gofakeit.Email(),
						CreatedAt: time.Now(),
					},
					AccessToken:  "axscg",
					RefreshToken: "askfs",
					ExpiresIn:    900,
				}

				uc.EXPECT().OAuthFlowCallback(mock.Anything, mock.Anything, mock.Anything).Return(createdUser, "/api/", nil)

				data := url.Values{}

				data.Set("code", "asxvf")

				req := httptest.NewRequestWithContext(
					context.Background(),
					http.MethodPost,
					"/api/auth/oauth/google/callback",
					strings.NewReader(data.Encode()),
				)

				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				req.Header.Set(authModeHeader, "cookie")

				return args{
					respWriter: httptest.NewRecorder(),
					req:        req,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecases := mocks.NewMockusecases(t)
			params := tt.setup(usecases)
			h := New(usecases, params.session)

			h.GoogleOAuthCallback(params.respWriter, params.req)

			result := params.respWriter.Result()

			var resp dto.AuthResponse

			switch tt.name {
			case "happy case: successful auth with cookies":
				_ = json.NewDecoder(result.Body).Decode(&resp)

				assert.Equal(t, http.StatusSeeOther, result.StatusCode)
				assert.Empty(t, resp.AccessToken)
				assert.Empty(t, resp.RefreshToken)
			default:
				assert.Equal(t, http.StatusInternalServerError, result.StatusCode)
			}
		})
	}
}

func TestHandlers_InitGithubOAuth(t *testing.T) {
	tests := []struct {
		name  string
		setup func(uc *mocks.Mockusecases) args
	}{
		{
			name: "sad case: internal server error",
			setup: func(uc *mocks.Mockusecases) args {
				uc.EXPECT().InitOAuthFlow(mock.Anything, mock.AnythingOfType("dto.OAuthProvider"), mock.Anything).Return("", errInternal)

				data := url.Values{}

				data.Set("return_to", "/api/")

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/auth/oauth/github",
						strings.NewReader(data.Encode()),
					),
				}
			},
		},
		{
			name: "happy case: redirect to consent page",
			setup: func(uc *mocks.Mockusecases) args {
				uc.EXPECT().InitOAuthFlow(mock.Anything, mock.AnythingOfType("dto.OAuthProvider"), mock.Anything).Return("/oauth/consent", nil)

				data := url.Values{}

				data.Set("return_to", "/api/shorten")

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/auth/oauth/github",
						strings.NewReader(data.Encode()),
					),
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecases := mocks.NewMockusecases(t)
			params := tt.setup(usecases)
			h := New(usecases, params.session)

			h.InitGithubOAuth(params.respWriter, params.req)

			result := params.respWriter.Result()

			switch tt.name {
			case "happy case: redirect to consent page":
				assert.Equal(t, http.StatusTemporaryRedirect, result.StatusCode)
			default:
				assert.Equal(t, http.StatusInternalServerError, result.StatusCode)
			}
		})
	}
}

func TestHandlers_InitGoogleOAuth(t *testing.T) {
	tests := []struct {
		name  string
		setup func(uc *mocks.Mockusecases) args
	}{
		{
			name: "sad case: internal server error",
			setup: func(uc *mocks.Mockusecases) args {
				uc.EXPECT().InitOAuthFlow(mock.Anything, mock.AnythingOfType("dto.OAuthProvider"), mock.Anything).Return("", errInternal)

				data := url.Values{}

				data.Set("return_to", "/api/")

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/auth/oauth/google",
						strings.NewReader(data.Encode()),
					),
				}
			},
		},
		{
			name: "happy case: redirect to consent page",
			setup: func(uc *mocks.Mockusecases) args {
				uc.EXPECT().InitOAuthFlow(mock.Anything, mock.AnythingOfType("dto.OAuthProvider"), mock.Anything).Return("/oauth/consent", nil)

				data := url.Values{}

				data.Set("return_to", "/api/shorten")

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/auth/oauth/google",
						strings.NewReader(data.Encode()),
					),
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecases := mocks.NewMockusecases(t)
			params := tt.setup(usecases)
			h := New(usecases, params.session)

			h.InitGoogleOAuth(params.respWriter, params.req)

			result := params.respWriter.Result()

			switch tt.name {
			case "happy case: redirect to consent page":
				assert.Equal(t, http.StatusTemporaryRedirect, result.StatusCode)
			default:
				assert.Equal(t, http.StatusInternalServerError, result.StatusCode)
			}
		})
	}
}

func TestHandlers_VerifyOTP(t *testing.T) {
	tests := []struct {
		name  string
		setup func(uc *mocks.Mockusecases) args
	}{
		{
			name: "sad case: unable to get user from context",
			setup: func(_ *mocks.Mockusecases) args {
				session := scs.New()

				req := httptest.NewRequestWithContext(
					context.Background(),
					http.MethodPost,
					"/api/auth/verify-otp",
					nil,
				)

				req = primeTestContext(session, req)

				session.Put(req.Context(), "user_id", "")

				return args{
					session:    session,
					respWriter: httptest.NewRecorder(),
					req:        req,
				}
			},
		},
		{
			name: "sad case: bad input",
			setup: func(_ *mocks.Mockusecases) args {
				session := scs.New()

				req := httptest.NewRequestWithContext(
					context.Background(),
					http.MethodPost,
					"/api/auth/verify-otp",
					bytes.NewReader([]byte(`body`)),
				)

				req = primeTestContext(session, req)

				session.Put(req.Context(), "user_id", gofakeit.UUID())

				return args{
					session:    session,
					respWriter: httptest.NewRecorder(),
					req:        req,
				}
			},
		},
		{
			name: "sad case: expired otp code",
			setup: func(uc *mocks.Mockusecases) args {
				session := scs.New()
				body, _ := json.Marshal(dto.VerifyOTPInput{
					Purpose: dto.EmailVerification,
					Value:   "000000",
				})

				uc.EXPECT().VerifyOTP(mock.Anything, mock.AnythingOfType("*dto.VerifyOTPInput")).Return(nil, auth.ErrExpiredOTPCode)

				req := httptest.NewRequestWithContext(
					context.Background(),
					http.MethodPost,
					"/api/auth/verify-otp",
					bytes.NewReader(body),
				)

				req = primeTestContext(session, req)

				session.Put(req.Context(), "user_id", gofakeit.UUID())

				return args{
					session:    session,
					respWriter: httptest.NewRecorder(),
					req:        req,
				}
			},
		},
		{
			name: "sad case: incorrect otp code",
			setup: func(uc *mocks.Mockusecases) args {
				session := scs.New()
				body, _ := json.Marshal(dto.VerifyOTPInput{
					Purpose: dto.EmailVerification,
					Value:   "000000",
				})

				uc.EXPECT().VerifyOTP(mock.Anything, mock.AnythingOfType("*dto.VerifyOTPInput")).Return(nil, auth.ErrIncorrectOTP)

				req := httptest.NewRequestWithContext(
					context.Background(),
					http.MethodPost,
					"/api/auth/verify-otp",
					bytes.NewReader(body),
				)

				req = primeTestContext(session, req)

				session.Put(req.Context(), "user_id", gofakeit.UUID())

				return args{
					session:    session,
					respWriter: httptest.NewRecorder(),
					req:        req,
				}
			},
		},
		{
			name: "sad case: internal server error",
			setup: func(uc *mocks.Mockusecases) args {
				session := scs.New()
				body, _ := json.Marshal(dto.VerifyOTPInput{
					Purpose: dto.EmailVerification,
					Value:   "000000",
				})

				uc.EXPECT().VerifyOTP(mock.Anything, mock.AnythingOfType("*dto.VerifyOTPInput")).Return(nil, errInternal)

				req := httptest.NewRequestWithContext(
					context.Background(),
					http.MethodPost,
					"/api/auth/verify-otp",
					bytes.NewReader(body),
				)

				req = primeTestContext(session, req)

				session.Put(req.Context(), "user_id", gofakeit.UUID())

				return args{
					session:    session,
					respWriter: httptest.NewRecorder(),
					req:        req,
				}
			},
		},
		{
			name: "happy case: successful otp verification",
			setup: func(uc *mocks.Mockusecases) args {
				session := scs.New()
				body, _ := json.Marshal(dto.VerifyOTPInput{
					Purpose: dto.EmailVerification,
					Value:   "000000",
				})

				uc.EXPECT().VerifyOTP(mock.Anything, mock.AnythingOfType("*dto.VerifyOTPInput")).Return(&dto.AuthResponse{
					User:         dto.User{},
					AccessToken:  "axcvg",
					RefreshToken: "asedf",
					ExpiresIn:    900,
				}, nil)

				req := httptest.NewRequestWithContext(
					context.Background(),
					http.MethodPost,
					"/api/auth/verify-otp",
					bytes.NewReader(body),
				)

				req = primeTestContext(session, req)

				session.Put(req.Context(), "user_id", gofakeit.UUID())

				return args{
					session:    session,
					respWriter: httptest.NewRecorder(),
					req:        req,
				}
			},
		},
		{
			name: "happy case: successful otp verification cookie mode",
			setup: func(uc *mocks.Mockusecases) args {
				session := scs.New()
				body, _ := json.Marshal(dto.VerifyOTPInput{
					Purpose: dto.EmailVerification,
					Value:   "000000",
				})

				uc.EXPECT().VerifyOTP(mock.Anything, mock.AnythingOfType("*dto.VerifyOTPInput")).Return(&dto.AuthResponse{
					User:         dto.User{},
					AccessToken:  "axcvg",
					RefreshToken: "asedf",
					ExpiresIn:    900,
				}, nil)

				req := httptest.NewRequestWithContext(
					context.Background(),
					http.MethodPost,
					"/api/auth/verify-otp",
					bytes.NewReader(body),
				)

				req = primeTestContext(session, req)
				req.Header.Set(authModeHeader, "cookie")

				session.Put(req.Context(), "user_id", gofakeit.UUID())

				return args{
					session:    session,
					respWriter: httptest.NewRecorder(),
					req:        req,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecases := mocks.NewMockusecases(t)
			params := tt.setup(usecases)
			h := New(usecases, params.session)

			h.VerifyOTP(params.respWriter, params.req)

			result := params.respWriter.Result()

			var resp dto.AuthResponse

			switch tt.name {
			case "happy case: successful otp verification":
				_ = json.NewDecoder(result.Body).Decode(&resp)

				assert.Equal(t, http.StatusOK, result.StatusCode)
				assert.NotEmpty(t, resp.AccessToken)
				assert.NotEmpty(t, resp.RefreshToken)
			case "happy case: successful otp verification cookie mode":
				_ = json.NewDecoder(result.Body).Decode(&resp)

				assert.Equal(t, http.StatusOK, result.StatusCode)
				assert.Empty(t, resp.AccessToken)
				assert.Empty(t, resp.RefreshToken)
			case "sad case: internal server error":
				assert.Equal(t, http.StatusInternalServerError, result.StatusCode)
			case "sad case: unable to get user from context":
				assert.Equal(t, http.StatusUnauthorized, result.StatusCode)
			default:
				assert.Equal(t, http.StatusBadRequest, result.StatusCode)
			}
		})
	}
}

func TestHandlers_RefreshAccessToken(t *testing.T) {
	tests := []struct {
		name  string
		setup func(uc *mocks.Mockusecases) args
	}{
		{
			name: "sad case: bad input",
			setup: func(_ *mocks.Mockusecases) args {
				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/auth/refresh",
						bytes.NewReader([]byte(`body`)),
					),
				}
			},
		},
		{
			name: "sad case: expired refresh token",
			setup: func(uc *mocks.Mockusecases) args {
				body, _ := json.Marshal(dto.RefreshAccessTokenInput{
					Token: "asxcfs",
				})

				uc.EXPECT().RefreshAccessToken(mock.Anything, mock.Anything).Return(nil, auth.ErrExpiredToken)

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/auth/refresh",
						bytes.NewReader(body),
					),
				}
			},
		},
		{
			name: "sad case: invalid refresh token",
			setup: func(uc *mocks.Mockusecases) args {
				body, _ := json.Marshal(dto.RefreshAccessTokenInput{
					Token: "asxcfs",
				})

				uc.EXPECT().RefreshAccessToken(mock.Anything, mock.Anything).Return(nil, auth.ErrInvalidToken)

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/auth/refresh",
						bytes.NewReader(body),
					),
				}
			},
		},
		{
			name: "sad case: internal server error",
			setup: func(uc *mocks.Mockusecases) args {
				body, _ := json.Marshal(dto.RefreshAccessTokenInput{
					Token: "asxcfs",
				})

				uc.EXPECT().RefreshAccessToken(mock.Anything, mock.Anything).Return(nil, errInternal)

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/auth/refresh",
						bytes.NewReader(body),
					),
				}
			},
		},
		{
			name: "happy case: successful refresh",
			setup: func(uc *mocks.Mockusecases) args {
				body, _ := json.Marshal(dto.RefreshAccessTokenInput{
					Token: "asxcfs",
				})

				uc.EXPECT().RefreshAccessToken(mock.Anything, mock.Anything).Return(&dto.RefreshAccessTokenResponse{
					AccessToken:  "axdf",
					RefreshToken: "aser",
					ExpiresIn:    900,
				}, nil)

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/auth/refresh",
						bytes.NewReader(body),
					),
				}
			},
		},
		{
			name: "happy case: successful refresh cookie mode",
			setup: func(uc *mocks.Mockusecases) args {
				body, _ := json.Marshal(dto.RefreshAccessTokenInput{
					Token: "asxcfs",
				})

				uc.EXPECT().RefreshAccessToken(mock.Anything, mock.Anything).Return(&dto.RefreshAccessTokenResponse{
					AccessToken:  "axdf",
					RefreshToken: "aser",
					ExpiresIn:    900,
				}, nil)

				req := httptest.NewRequestWithContext(
					context.Background(),
					http.MethodPost,
					"/api/auth/refresh",
					bytes.NewReader(body),
				)

				req.Header.Set(authModeHeader, "cookie")
				req.AddCookie(&http.Cookie{
					Name:  refreshTokenCookie,
					Value: "aserf",
				})

				return args{
					respWriter: httptest.NewRecorder(),
					req:        req,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecases := mocks.NewMockusecases(t)
			params := tt.setup(usecases)
			h := New(usecases, params.session)

			h.RefreshAccessToken(params.respWriter, params.req)

			result := params.respWriter.Result()

			var resp dto.RefreshAccessTokenResponse

			switch tt.name {
			case "happy case: successful refresh":
				_ = json.NewDecoder(result.Body).Decode(&resp)

				assert.Equal(t, http.StatusOK, result.StatusCode)
				assert.NotEmpty(t, resp.AccessToken)
				assert.NotEmpty(t, resp.RefreshToken)
			case "happy case: successful refresh cookie mode":
				_ = json.NewDecoder(result.Body).Decode(&resp)

				assert.Equal(t, http.StatusOK, result.StatusCode)
				assert.Empty(t, resp.AccessToken)
				assert.Empty(t, resp.RefreshToken)
			case "sad case: expired refresh token",
				"sad case: invalid refresh token":
				assert.Equal(t, http.StatusUnauthorized, result.StatusCode)
			case "sad case: internal server error":
				assert.Equal(t, http.StatusInternalServerError, result.StatusCode)
			default:
				assert.Equal(t, http.StatusBadRequest, result.StatusCode)
			}
		})
	}
}

func TestHandlers_Login(t *testing.T) {
	tests := []struct {
		name  string
		setup func(uc *mocks.Mockusecases) args
	}{
		{
			name: "sad case: bad input",
			setup: func(_ *mocks.Mockusecases) args {
				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/auth/login",
						bytes.NewReader([]byte(`bad input`)),
					),
				}
			},
		},
		{
			name: "sad case: missing required fields",
			setup: func(_ *mocks.Mockusecases) args {
				body, _ := json.Marshal(dto.LoginInput{})

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/auth/login",
						bytes.NewReader(body),
					),
				}
			},
		},
		{
			name: "sad case: invalid credentials",
			setup: func(uc *mocks.Mockusecases) args {
				body, _ := json.Marshal(dto.LoginInput{
					Email:    gofakeit.Email(),
					Password: "asderfg",
				})

				uc.EXPECT().Login(mock.Anything, mock.AnythingOfType("*dto.LoginInput")).Return(nil, auth.ErrInvalidCredentials)

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/auth/login",
						bytes.NewReader(body),
					),
				}
			},
		},
		{
			name: "sad case: internal server error",
			setup: func(uc *mocks.Mockusecases) args {
				body, _ := json.Marshal(dto.LoginInput{
					Email:    gofakeit.Email(),
					Password: "asderfg",
				})

				uc.EXPECT().Login(mock.Anything, mock.AnythingOfType("*dto.LoginInput")).Return(nil, errInternal)

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/auth/login",
						bytes.NewReader(body),
					),
				}
			},
		},
		{
			name: "happy case: successful login",
			setup: func(uc *mocks.Mockusecases) args {
				body, _ := json.Marshal(dto.LoginInput{
					Email:    gofakeit.Email(),
					Password: "asderfg",
				})

				uc.EXPECT().Login(mock.Anything, mock.AnythingOfType("*dto.LoginInput")).Return(&dto.OTPRequiredResponse{
					User: dto.User{
						PublicID: gofakeit.UUID(),
					},
					Status: "otp_required",
				}, nil)

				req := httptest.NewRequestWithContext(
					context.Background(),
					http.MethodPost,
					"/api/auth/login",
					bytes.NewReader(body),
				)
				session := scs.New()

				req = primeTestContext(session, req)

				return args{
					session:    session,
					respWriter: httptest.NewRecorder(),
					req:        req,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecases := mocks.NewMockusecases(t)
			params := tt.setup(usecases)
			h := New(usecases, params.session)

			h.Login(params.respWriter, params.req)

			result := params.respWriter.Result()

			switch tt.name {
			case "happy case: successful login":
				assert.Equal(t, http.StatusOK, result.StatusCode)
			case "sad case: invalid credentials":
				assert.Equal(t, http.StatusUnauthorized, result.StatusCode)
			case "sad case: internal server error":
				assert.Equal(t, http.StatusInternalServerError, result.StatusCode)
			default:
				assert.Equal(t, http.StatusBadRequest, result.StatusCode)
			}
		})
	}
}

func TestHandlers_ValidatePassword(t *testing.T) {
	tests := []struct {
		name  string
		setup func(uc *mocks.Mockusecases) args
	}{
		{
			name: "sad case: bad input",
			setup: func(_ *mocks.Mockusecases) args {
				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/auth/validate-password",
						bytes.NewReader([]byte(`bad body`)),
					),
				}
			},
		},
		{
			name: "sad case: missing required fields",
			setup: func(_ *mocks.Mockusecases) args {
				body, _ := json.Marshal(dto.ValidatePasswordInput{})

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/auth/validate-password",
						bytes.NewReader(body),
					),
				}
			},
		},
		{
			name: "happy case: successful verification",
			setup: func(uc *mocks.Mockusecases) args {
				body, _ := json.Marshal(dto.ValidatePasswordInput{
					Password: "asdefg",
				})

				uc.EXPECT().ValidatePasswordStrength(mock.Anything, mock.AnythingOfType("*dto.ValidatePasswordInput")).Return(true)

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/auth/validate-password",
						bytes.NewReader(body),
					),
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecases := mocks.NewMockusecases(t)
			params := tt.setup(usecases)
			h := New(usecases, params.session)

			h.ValidatePassword(params.respWriter, params.req)

			result := params.respWriter.Result()

			switch tt.name {
			case "happy case: successful verification":
				assert.Equal(t, http.StatusOK, result.StatusCode)
			default:
				assert.Equal(t, http.StatusBadRequest, result.StatusCode)
			}
		})
	}
}

func TestHandlers_HealthCheck(t *testing.T) {
	tests := []struct {
		name  string
		setup func(uc *mocks.Mockusecases) args
	}{
		{
			name: "check health status",
			setup: func(uc *mocks.Mockusecases) args {
				uc.EXPECT().HealthCheck(mock.Anything).Return(map[string]bool{})

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodGet,
						"/health",
						nil,
					),
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecases := mocks.NewMockusecases(t)
			params := tt.setup(usecases)
			h := New(usecases, params.session)

			h.HealthCheck(params.respWriter, params.req)
		})
	}
}

func TestHandlers_CreateUserEmailPassword(t *testing.T) {
	tests := []struct {
		name  string
		setup func(uc *mocks.Mockusecases) args
	}{
		{
			name: "sad case: bad input",
			setup: func(_ *mocks.Mockusecases) args {
				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/auth/register",
						bytes.NewReader([]byte(`bad input`)),
					),
				}
			},
		},
		{
			name: "sad case: missing required fields",
			setup: func(_ *mocks.Mockusecases) args {
				body, _ := json.Marshal(dto.EmailPasswordUserInput{})

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/auth/register",
						bytes.NewReader(body),
					),
				}
			},
		},
		{
			name: "sad case: invalid email syntax",
			setup: func(uc *mocks.Mockusecases) args {
				body, _ := json.Marshal(dto.EmailPasswordUserInput{
					Email:    gofakeit.Email(),
					Password: "12345678",
				})

				uc.EXPECT().CreateUserEmailPassword(mock.Anything, mock.AnythingOfType("*dto.EmailPasswordUserInput")).Return(nil, auth.ErrInvalidEmailSyntax)

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/auth/register",
						bytes.NewReader(body),
					),
				}
			},
		},
		{
			name: "sad case: password breached",
			setup: func(uc *mocks.Mockusecases) args {
				body, _ := json.Marshal(dto.EmailPasswordUserInput{
					Email:    gofakeit.Email(),
					Password: "12345678",
				})

				uc.EXPECT().CreateUserEmailPassword(mock.Anything, mock.AnythingOfType("*dto.EmailPasswordUserInput")).Return(nil, auth.ErrPasswordBreached)

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/auth/register",
						bytes.NewReader(body),
					),
				}
			},
		},
		{
			name: "sad case: internal server error",
			setup: func(uc *mocks.Mockusecases) args {
				body, _ := json.Marshal(dto.EmailPasswordUserInput{
					Email:    gofakeit.Email(),
					Password: "12345678",
				})

				uc.EXPECT().CreateUserEmailPassword(mock.Anything, mock.AnythingOfType("*dto.EmailPasswordUserInput")).Return(nil, errInternal)

				return args{
					respWriter: httptest.NewRecorder(),
					req: httptest.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						"/api/auth/register",
						bytes.NewReader(body),
					),
				}
			},
		},
		{
			name: "happy case: successfully create user",
			setup: func(uc *mocks.Mockusecases) args {
				email := gofakeit.Email()
				body, _ := json.Marshal(dto.EmailPasswordUserInput{
					Email:    email,
					Password: "12345678",
				})

				uc.EXPECT().CreateUserEmailPassword(mock.Anything, mock.AnythingOfType("*dto.EmailPasswordUserInput")).Return(&dto.OTPRequiredResponse{
					User: dto.User{
						PublicID:  gofakeit.UUID(),
						Email:     email,
						Password:  "12345678",
						CreatedAt: time.Now(),
					},
					Status: "otp_required",
				}, nil)

				req := httptest.NewRequestWithContext(
					context.Background(),
					http.MethodPost,
					"/api/auth/register",
					bytes.NewReader(body),
				)
				session := scs.New()

				req = primeTestContext(session, req)

				return args{
					respWriter: httptest.NewRecorder(),
					req:        req,
					session:    session,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecases := mocks.NewMockusecases(t)
			params := tt.setup(usecases)
			h := New(usecases, params.session)

			h.CreateUserEmailPassword(params.respWriter, params.req)

			result := params.respWriter.Result()

			var resp dto.OTPRequiredResponse

			switch tt.name {
			case "happy case: successfully create user":
				_ = json.NewDecoder(result.Body).Decode(&resp)

				assert.Equal(t, http.StatusCreated, result.StatusCode)
				assert.Equal(t, "otp_required", resp.Status)
				assert.Empty(t, resp.User.Password)
				assert.NotEmpty(t, resp.User.PublicID)
				assert.NotEmpty(t, resp.User.Email)
			case "sad case: internal server error":
				assert.Equal(t, http.StatusInternalServerError, result.StatusCode)
			default:
				assert.Equal(t, http.StatusBadRequest, result.StatusCode)
			}
		})
	}
}
