package api_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/davidsbond/passwords/internal/server/api"
	"github.com/davidsbond/passwords/internal/server/service"
	"github.com/davidsbond/passwords/internal/server/token"
)

func TestPasswordAPI_Create(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		Request      api.CreatePasswordRequest
		Expected     api.CreatePasswordResponse
		Token        token.Token
		ExpectedCode int
		ExpectsError bool
		Setup        func(svc *MockPasswordService)
	}{
		{
			Name:         "error if no token",
			Token:        token.Token{},
			ExpectsError: true,
			ExpectedCode: http.StatusUnauthorized,
			Request: api.CreatePasswordRequest{
				Username: "test",
				Password: "test",
			},
		},
		{
			Name:  "error if missing username",
			Token: token.TestToken(t, "test"),
			Request: api.CreatePasswordRequest{
				Username: "",
				Password: "password",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusBadRequest,
		},
		{
			Name:  "error if missing password",
			Token: token.TestToken(t, "test"),
			Request: api.CreatePasswordRequest{
				Username: "test@test.com",
				Password: "",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusBadRequest,
		},
		{
			Name:  "error if lifetime has expired",
			Token: token.TestToken(t, "test"),
			Request: api.CreatePasswordRequest{
				Username: "test",
				Password: "test",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusUnauthorized,
			Setup: func(svc *MockPasswordService) {
				svc.EXPECT().Create(mock.Anything).Return(service.ErrReauthenticate).Once()
			},
		},
		{
			Name:  "error if creation fails",
			Token: token.TestToken(t, "test"),
			Request: api.CreatePasswordRequest{
				Username: "test",
				Password: "test",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusInternalServerError,
			Setup: func(svc *MockPasswordService) {
				svc.EXPECT().Create(mock.Anything).Return(io.EOF).Once()
			},
		},
		{
			Name:  "success",
			Token: token.TestToken(t, "test"),
			Request: api.CreatePasswordRequest{
				Username: "test",
				Password: "test",
			},
			Expected:     api.CreatePasswordResponse{},
			ExpectedCode: http.StatusCreated,
			Setup: func(svc *MockPasswordService) {
				svc.EXPECT().Create(mock.Anything).Return(nil).Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			svc := NewMockPasswordService(t)
			if tc.Setup != nil {
				tc.Setup(svc)
			}

			w := httptest.NewRecorder()
			r := request(t, http.MethodPost, "/api/v1/password", tc.Request).
				WithContext(token.ToContext(t.Context(), tc.Token))

			api.NewPasswordAPI(svc).Create(w, r)

			require.Equal(t, tc.ExpectedCode, w.Code)
			if tc.ExpectsError {
				assertAPIError(t, w)
				return
			}

			assertResponse(t, w, tc.Expected)
		})
	}
}

func TestPasswordAPI_List(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		Expected     api.ListPasswordsResponse
		Token        token.Token
		ExpectedCode int
		ExpectsError bool
		Setup        func(svc *MockPasswordService)
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
			Setup: func(svc *MockPasswordService) {
				svc.EXPECT().List(mock.Anything).Return(nil, service.ErrReauthenticate).Once()
			},
		},
		{
			Name:         "error if listing fails",
			Token:        token.TestToken(t, "test"),
			ExpectsError: true,
			ExpectedCode: http.StatusInternalServerError,
			Setup: func(svc *MockPasswordService) {
				svc.EXPECT().List(mock.Anything).Return(nil, io.EOF).Once()
			},
		},
		{
			Name:  "success",
			Token: token.TestToken(t, "test"),
			Expected: api.ListPasswordsResponse{
				Passwords: []api.Password{
					{
						Username: "test",
						Password: "test",
						Domains:  []string{"test"},
					},
				},
			},
			ExpectedCode: http.StatusOK,
			Setup: func(svc *MockPasswordService) {
				expected := []service.Password{
					{
						Username: "test",
						Password: "test",
						Domains:  []string{"test"},
					},
				}

				svc.EXPECT().List(mock.Anything).Return(expected, nil).Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			svc := NewMockPasswordService(t)
			if tc.Setup != nil {
				tc.Setup(svc)
			}

			w := httptest.NewRecorder()
			r := request(t, http.MethodGet, "/api/v1/password", nil).
				WithContext(token.ToContext(t.Context(), tc.Token))

			api.NewPasswordAPI(svc).List(w, r)

			require.Equal(t, tc.ExpectedCode, w.Code)
			if tc.ExpectsError {
				assertAPIError(t, w)
				return
			}

			assertResponse(t, w, tc.Expected)
		})
	}
}
