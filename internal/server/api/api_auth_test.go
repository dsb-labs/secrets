package api_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/davidsbond/keeper/internal/server/api"
	"github.com/davidsbond/keeper/internal/server/service"
	"github.com/davidsbond/keeper/internal/server/token"
)

func TestAuthAPI_Login(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		Request      api.LoginRequest
		Expected     api.LoginResponse
		ExpectedCode int
		ExpectsError bool
		Setup        func(svc *MockAuthService)
	}{
		{
			Name: "error if missing email address",
			Request: api.LoginRequest{
				Email:    "",
				Password: "password",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusBadRequest,
		},
		{
			Name: "error if missing password",
			Request: api.LoginRequest{
				Email:    "test@test.com",
				Password: "",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusBadRequest,
		},
		{
			Name: "error if email is invalid",
			Request: api.LoginRequest{
				Email:    "not-an-email",
				Password: "password",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusBadRequest,
		},
		{
			Name: "error if account does not exist",
			Request: api.LoginRequest{
				Email:    "test@test.com",
				Password: "password",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusNotFound,
			Setup: func(svc *MockAuthService) {
				svc.EXPECT().
					Login("test@test.com", "password").Return(token.Token{}, service.ErrAccountNotFound).
					Once()
			},
		},
		{
			Name: "error if password is invalid",
			Request: api.LoginRequest{
				Email:    "test@test.com",
				Password: "password",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusBadRequest,
			Setup: func(svc *MockAuthService) {
				svc.EXPECT().
					Login("test@test.com", "password").Return(token.Token{}, service.ErrInvalidPassword).
					Once()
			},
		},
		{
			Name: "error if login fails",
			Request: api.LoginRequest{
				Email:    "test@test.com",
				Password: "password",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusInternalServerError,
			Setup: func(svc *MockAuthService) {
				svc.EXPECT().
					Login("test@test.com", "password").Return(token.Token{}, io.EOF).
					Once()
			},
		},
		{
			Name: "returns token on success",
			Request: api.LoginRequest{
				Email:    "test@test.com",
				Password: "password",
			},
			Expected: api.LoginResponse{
				Token: "test",
			},
			ExpectedCode: http.StatusOK,
			Setup: func(svc *MockAuthService) {
				svc.EXPECT().
					Login("test@test.com", "password").Return(token.TestToken(t, "test"), nil).
					Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			svc := NewMockAuthService(t)
			if tc.Setup != nil {
				tc.Setup(svc)
			}

			w := httptest.NewRecorder()
			r := request(t, http.MethodPost, "/api/v1/auth", tc.Request)

			api.NewAuthAPI(svc).Login(w, r)

			require.Equal(t, tc.ExpectedCode, w.Code)
			if tc.ExpectsError {
				assertAPIError(t, w)
				return
			}

			assertResponse(t, w, tc.Expected)

			cookie, err := http.ParseSetCookie(w.Header().Get("Set-Cookie"))
			require.NoError(t, err)
			require.NotNil(t, cookie)
			assert.EqualValues(t, tc.Expected.Token, cookie.Value)
		})
	}
}

func TestAuthAPI_Logout(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		Token        token.Token
		ExpectsError bool
		ExpectedCode int
		Expected     api.LogoutResponse
		Setup        func(svc *MockAuthService)
	}{
		{
			Name:         "error if no valid token",
			Token:        token.Token{},
			ExpectsError: true,
			ExpectedCode: http.StatusUnauthorized,
		},
		{
			Name:         "error service fails",
			Token:        token.TestToken(t, "test"),
			ExpectsError: true,
			ExpectedCode: http.StatusInternalServerError,
			Setup: func(svc *MockAuthService) {
				svc.EXPECT().Logout(mock.Anything).Return(io.EOF).Once()
			},
		},
		{
			Name:         "success",
			Token:        token.TestToken(t, "test"),
			Expected:     api.LogoutResponse{},
			ExpectedCode: http.StatusOK,
			Setup: func(svc *MockAuthService) {
				svc.EXPECT().Logout(mock.Anything).Return(nil).Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			svc := NewMockAuthService(t)
			if tc.Setup != nil {
				tc.Setup(svc)
			}

			w := httptest.NewRecorder()
			r := request(t, http.MethodDelete, "/api/v1/auth", nil).
				WithContext(token.ToContext(t.Context(), tc.Token))

			api.NewAuthAPI(svc).Logout(w, r)

			require.Equal(t, tc.ExpectedCode, w.Code)
			if tc.ExpectsError {
				assertAPIError(t, w)
				return
			}

			assertResponse(t, w, tc.Expected)
			cookie, err := http.ParseSetCookie(w.Header().Get("Set-Cookie"))
			require.NoError(t, err)
			require.NotNil(t, cookie)
			assert.Empty(t, cookie.Value)
			assert.EqualValues(t, -1, cookie.MaxAge)
		})
	}
}
