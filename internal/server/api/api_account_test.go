package api_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/davidsbond/keeper/internal/server/api"
	"github.com/davidsbond/keeper/internal/server/service"
	"github.com/davidsbond/keeper/internal/server/token"
)

func TestAccountAPI_Create(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		Request      api.CreateAccountRequest
		Expected     api.CreateAccountResponse
		ExpectedCode int
		ExpectsError bool
		Setup        func(svc *MockAccountService)
	}{
		{
			Name: "error if missing email address",
			Request: api.CreateAccountRequest{
				Email:       "",
				Password:    "password",
				DisplayName: "test",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusBadRequest,
		},
		{
			Name: "error if missing password",
			Request: api.CreateAccountRequest{
				Email:       "test@test.com",
				Password:    "",
				DisplayName: "test",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusBadRequest,
		},
		{
			Name: "error if email is invalid",
			Request: api.CreateAccountRequest{
				Email:       "not-an-email",
				Password:    "password",
				DisplayName: "test",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusBadRequest,
		},
		{
			Name: "error if missing display name",
			Request: api.CreateAccountRequest{
				Email:       "test@test.com",
				Password:    "password",
				DisplayName: "",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusBadRequest,
		},
		{
			Name: "error if account already exists",
			Request: api.CreateAccountRequest{
				Email:       "test@test.com",
				Password:    "password",
				DisplayName: "test",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusConflict,
			Setup: func(svc *MockAccountService) {
				svc.EXPECT().Create(mock.Anything).Return(nil, service.ErrAccountExists).Once()
			},
		},
		{
			Name: "error if creation fails",
			Request: api.CreateAccountRequest{
				Email:       "test@test.com",
				Password:    "password",
				DisplayName: "test",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusInternalServerError,
			Setup: func(svc *MockAccountService) {
				svc.EXPECT().Create(mock.Anything).Return(nil, io.EOF).Once()
			},
		},
		{
			Name: "returns restore key on success",
			Request: api.CreateAccountRequest{
				Email:       "test@test.com",
				Password:    "password",
				DisplayName: "test",
			},
			ExpectedCode: http.StatusCreated,
			Expected: api.CreateAccountResponse{
				RestoreKey: []byte("restore-key"),
			},
			Setup: func(svc *MockAccountService) {
				svc.EXPECT().Create(mock.Anything).Return([]byte("restore-key"), nil).Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			svc := NewMockAccountService(t)
			if tc.Setup != nil {
				tc.Setup(svc)
			}

			w := httptest.NewRecorder()
			r := request(t, http.MethodPost, "/api/v1/account", tc.Request)

			api.NewAccountAPI(svc).Create(w, r)

			require.Equal(t, tc.ExpectedCode, w.Code)
			if tc.ExpectsError {
				assertAPIError(t, w)
				return
			}

			assertResponse(t, w, tc.Expected)
		})
	}
}

func TestAccountAPI_Get(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		Expected     api.GetAccountResponse
		Token        token.Token
		ExpectedCode int
		ExpectsError bool
		Setup        func(svc *MockAccountService)
	}{
		{
			Name:         "error if account does not exist",
			Token:        token.TestToken(t, "test"),
			ExpectsError: true,
			ExpectedCode: http.StatusNotFound,
			Setup: func(svc *MockAccountService) {
				svc.EXPECT().
					Get(mock.Anything).
					Return(service.Account{}, service.ErrAccountNotFound).
					Once()
			},
		},
		{
			Name:         "error if get fails",
			Token:        token.TestToken(t, "test"),
			ExpectsError: true,
			ExpectedCode: http.StatusInternalServerError,
			Setup: func(svc *MockAccountService) {
				svc.EXPECT().
					Get(mock.Anything).
					Return(service.Account{}, io.EOF).Once()
			},
		},
		{
			Name:         "success",
			Token:        token.TestToken(t, "test"),
			ExpectedCode: http.StatusOK,
			Expected: api.GetAccountResponse{
				Account: api.Account{
					DisplayName: "Test McTest",
					Email:       "test@test.com",
				},
			},
			Setup: func(svc *MockAccountService) {
				expected := service.Account{
					Password:    "REDACTED",
					DisplayName: "Test McTest",
					Email:       "test@test.com",
				}

				svc.EXPECT().
					Get(mock.Anything).
					Return(expected, nil).Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			svc := NewMockAccountService(t)
			if tc.Setup != nil {
				tc.Setup(svc)
			}

			w := httptest.NewRecorder()
			r := request(t, http.MethodGet, "/api/v1/account", nil).
				WithContext(token.ToContext(t.Context(), tc.Token))

			api.NewAccountAPI(svc).Get(w, r)

			require.Equal(t, tc.ExpectedCode, w.Code)
			if tc.ExpectsError {
				assertAPIError(t, w)
				return
			}

			assertResponse(t, w, tc.Expected)
		})
	}
}

func TestAccountAPI_Delete(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		Expected     api.DeleteAccountResponse
		Token        token.Token
		ExpectedCode int
		ExpectsError bool
		Setup        func(svc *MockAccountService)
	}{
		{
			Name:         "error if account does not exist",
			Token:        token.TestToken(t, "test"),
			ExpectsError: true,
			ExpectedCode: http.StatusNotFound,
			Setup: func(svc *MockAccountService) {
				svc.EXPECT().
					Delete(mock.Anything).
					Return(service.ErrAccountNotFound).
					Once()
			},
		},
		{
			Name:         "error if delete fails",
			Token:        token.TestToken(t, "test"),
			ExpectsError: true,
			ExpectedCode: http.StatusInternalServerError,
			Setup: func(svc *MockAccountService) {
				svc.EXPECT().
					Delete(mock.Anything).
					Return(io.EOF).Once()
			},
		},
		{
			Name:         "success",
			Token:        token.TestToken(t, "test"),
			ExpectedCode: http.StatusOK,
			Setup: func(svc *MockAccountService) {
				svc.EXPECT().
					Delete(mock.Anything).
					Return(nil).Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			svc := NewMockAccountService(t)
			if tc.Setup != nil {
				tc.Setup(svc)
			}

			w := httptest.NewRecorder()
			r := request(t, http.MethodDelete, "/api/v1/account", nil).
				WithContext(token.ToContext(t.Context(), tc.Token))

			api.NewAccountAPI(svc).Delete(w, r)

			require.Equal(t, tc.ExpectedCode, w.Code)
			if tc.ExpectsError {
				assertAPIError(t, w)
				return
			}

			assertResponse(t, w, tc.Expected)
		})
	}
}

func TestAccountAPI_ChangePassword(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		Token        token.Token
		Request      api.UpdatePasswordRequest
		Expected     api.UpdatePasswordResponse
		ExpectedCode int
		ExpectsError bool
		Setup        func(svc *MockAccountService)
	}{
		{
			Name:         "error if missing old password",
			Token:        token.TestToken(t, "test"),
			ExpectsError: true,
			ExpectedCode: http.StatusBadRequest,
			Request: api.UpdatePasswordRequest{
				NewPassword: "test",
				OldPassword: "",
			},
		},
		{
			Name:         "error if missing new password",
			Token:        token.TestToken(t, "test"),
			ExpectsError: true,
			ExpectedCode: http.StatusBadRequest,
			Request: api.UpdatePasswordRequest{
				NewPassword: "",
				OldPassword: "test",
			},
		},
		{
			Name:         "error if old password matches new password",
			Token:        token.TestToken(t, "test"),
			ExpectsError: true,
			ExpectedCode: http.StatusBadRequest,
			Request: api.UpdatePasswordRequest{
				NewPassword: "test",
				OldPassword: "test",
			},
		},
		{
			Name:         "error if account does not exist",
			Token:        token.TestToken(t, "test"),
			ExpectsError: true,
			ExpectedCode: http.StatusNotFound,
			Request: api.UpdatePasswordRequest{
				NewPassword: "new",
				OldPassword: "old",
			},
			Setup: func(svc *MockAccountService) {
				svc.EXPECT().
					ChangePassword(mock.Anything, "old", "new").
					Return(nil, service.ErrAccountNotFound).
					Once()
			},
		},
		{
			Name:         "error if old password is incorrect",
			Token:        token.TestToken(t, "test"),
			ExpectsError: true,
			ExpectedCode: http.StatusBadRequest,
			Request: api.UpdatePasswordRequest{
				NewPassword: "new",
				OldPassword: "old",
			},
			Setup: func(svc *MockAccountService) {
				svc.EXPECT().
					ChangePassword(mock.Anything, "old", "new").
					Return(nil, service.ErrInvalidPassword).
					Once()
			},
		},
		{
			Name:         "error changing password",
			Token:        token.TestToken(t, "test"),
			ExpectsError: true,
			ExpectedCode: http.StatusInternalServerError,
			Request: api.UpdatePasswordRequest{
				NewPassword: "new",
				OldPassword: "old",
			},
			Setup: func(svc *MockAccountService) {
				svc.EXPECT().
					ChangePassword(mock.Anything, "old", "new").
					Return(nil, io.EOF).
					Once()
			},
		},
		{
			Name:         "success",
			Token:        token.TestToken(t, "test"),
			ExpectedCode: http.StatusOK,
			Expected: api.UpdatePasswordResponse{
				RestoreKey: []byte("test-key"),
			},
			Request: api.UpdatePasswordRequest{
				NewPassword: "new",
				OldPassword: "old",
			},
			Setup: func(svc *MockAccountService) {
				svc.EXPECT().
					ChangePassword(mock.Anything, "old", "new").
					Return([]byte("test-key"), nil).
					Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			svc := NewMockAccountService(t)
			if tc.Setup != nil {
				tc.Setup(svc)
			}

			w := httptest.NewRecorder()
			r := request(t, http.MethodPut, "/api/v1/account/password", tc.Request).
				WithContext(token.ToContext(t.Context(), tc.Token))

			api.NewAccountAPI(svc).UpdatePassword(w, r)

			require.Equal(t, tc.ExpectedCode, w.Code)
			if tc.ExpectsError {
				assertAPIError(t, w)
				return
			}

			assertResponse(t, w, tc.Expected)
		})
	}
}
