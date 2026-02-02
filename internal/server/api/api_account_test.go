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
