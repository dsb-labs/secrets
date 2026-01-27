package api_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/davidsbond/passwords/internal/server/api"
	"github.com/davidsbond/passwords/internal/server/service"
	"github.com/davidsbond/passwords/internal/server/token"
)

func TestLoginAPI_Create(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		Request      api.CreateLoginRequest
		Expected     api.CreateLoginResponse
		Token        token.Token
		ExpectedCode int
		ExpectsError bool
		Setup        func(svc *MockLoginService)
	}{
		{
			Name:         "error if no token",
			Token:        token.Token{},
			ExpectsError: true,
			ExpectedCode: http.StatusUnauthorized,
			Request: api.CreateLoginRequest{
				Username: "test",
				Password: "test",
			},
		},
		{
			Name:  "error if missing username",
			Token: token.TestToken(t, "test"),
			Request: api.CreateLoginRequest{
				Username: "",
				Password: "password",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusBadRequest,
		},
		{
			Name:  "error if missing password",
			Token: token.TestToken(t, "test"),
			Request: api.CreateLoginRequest{
				Username: "test@test.com",
				Password: "",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusBadRequest,
		},
		{
			Name:  "error if lifetime has expired",
			Token: token.TestToken(t, "test"),
			Request: api.CreateLoginRequest{
				Username: "test",
				Password: "test",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusUnauthorized,
			Setup: func(svc *MockLoginService) {
				svc.EXPECT().Create(mock.Anything, mock.Anything).Return(service.ErrReauthenticate).Once()
			},
		},
		{
			Name:  "error if creation fails",
			Token: token.TestToken(t, "test"),
			Request: api.CreateLoginRequest{
				Username: "test",
				Password: "test",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusInternalServerError,
			Setup: func(svc *MockLoginService) {
				svc.EXPECT().Create(mock.Anything, mock.Anything).Return(io.EOF).Once()
			},
		},
		{
			Name:  "success",
			Token: token.TestToken(t, "test"),
			Request: api.CreateLoginRequest{
				Username: "test",
				Password: "test",
			},
			Expected:     api.CreateLoginResponse{},
			ExpectedCode: http.StatusCreated,
			Setup: func(svc *MockLoginService) {
				svc.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			svc := NewMockLoginService(t)
			if tc.Setup != nil {
				tc.Setup(svc)
			}

			w := httptest.NewRecorder()
			r := request(t, http.MethodPost, "/api/v1/login", tc.Request).
				WithContext(token.ToContext(t.Context(), tc.Token))

			api.NewLoginAPI(svc).Create(w, r)

			require.Equal(t, tc.ExpectedCode, w.Code)
			if tc.ExpectsError {
				assertAPIError(t, w)
				return
			}

			assertResponse(t, w, tc.Expected)
		})
	}
}

func TestLoginAPI_List(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		Expected     api.ListLoginsResponse
		Token        token.Token
		ExpectedCode int
		ExpectsError bool
		Setup        func(svc *MockLoginService)
	}{
		{
			Name:         "error if no token",
			Token:        token.Token{},
			ExpectsError: true,
			ExpectedCode: http.StatusUnauthorized,
		},
		{
			Name:         "error if lifetime has expired",
			Token:        token.TestToken(t, "test"),
			ExpectsError: true,
			ExpectedCode: http.StatusUnauthorized,
			Setup: func(svc *MockLoginService) {
				svc.EXPECT().List(mock.Anything, mock.Anything).Return(nil, service.ErrReauthenticate).Once()
			},
		},
		{
			Name:         "error if listing fails",
			Token:        token.TestToken(t, "test"),
			ExpectsError: true,
			ExpectedCode: http.StatusInternalServerError,
			Setup: func(svc *MockLoginService) {
				svc.EXPECT().List(mock.Anything, mock.Anything).Return(nil, io.EOF).Once()
			},
		},
		{
			Name:  "success",
			Token: token.TestToken(t, "test"),
			Expected: api.ListLoginsResponse{
				Logins: []api.Login{
					{
						ID:       uuid.NameSpaceDNS.String(),
						Username: "test",
						Password: "test",
						Domains:  []string{"test"},
					},
				},
			},
			ExpectedCode: http.StatusOK,
			Setup: func(svc *MockLoginService) {
				expected := []service.Login{
					{
						ID:       uuid.NameSpaceDNS,
						Username: "test",
						Password: "test",
						Domains:  []string{"test"},
					},
				}

				svc.EXPECT().List(mock.Anything, mock.Anything).Return(expected, nil).Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			svc := NewMockLoginService(t)
			if tc.Setup != nil {
				tc.Setup(svc)
			}

			w := httptest.NewRecorder()
			r := request(t, http.MethodGet, "/api/v1/login", nil).
				WithContext(token.ToContext(t.Context(), tc.Token))

			api.NewLoginAPI(svc).List(w, r)

			require.Equal(t, tc.ExpectedCode, w.Code)
			if tc.ExpectsError {
				assertAPIError(t, w)
				return
			}

			assertResponse(t, w, tc.Expected)
		})
	}
}
