package api_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/dsb-labs/secrets/internal/server/api"
	"github.com/dsb-labs/secrets/internal/server/service"
	"github.com/dsb-labs/secrets/internal/server/token"
)

func TestToolAPI_Export(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		Expected     api.ExportResponse
		Token        token.Token
		ExpectedCode int
		ExpectsError bool
		Setup        func(svc *MockToolService)
	}{
		{
			Name:         "error if lifetime has expired",
			Token:        token.TestToken(t, "test"),
			ExpectsError: true,
			ExpectedCode: http.StatusUnauthorized,
			Setup: func(svc *MockToolService) {
				svc.EXPECT().
					Export(mock.Anything).
					Return(service.Export{}, service.ErrReauthenticate).Once()
			},
		},
		{
			Name:         "error if export fails",
			Token:        token.TestToken(t, "test"),
			ExpectsError: true,
			ExpectedCode: http.StatusInternalServerError,
			Setup: func(svc *MockToolService) {
				svc.EXPECT().
					Export(mock.Anything).
					Return(service.Export{}, io.EOF).Once()
			},
		},
		{
			Name:         "success",
			Token:        token.TestToken(t, "test"),
			ExpectedCode: http.StatusOK,
			Expected: api.ExportResponse{
				Logins: []api.Login{
					{
						ID:       uuid.NameSpaceDNS.String(),
						Username: "test",
						Password: "test",
						Domains:  []string{"test"},
					},
				},
				Notes: []api.Note{
					{
						ID:      uuid.NameSpaceURL.String(),
						Name:    "test",
						Content: "test",
					},
				},
				Cards: []api.Card{
					{
						ID:          uuid.NameSpaceOID.String(),
						HolderName:  "test",
						Number:      "4111 1111 1111 1111",
						ExpiryMonth: time.January,
						ExpiryYear:  2025,
						CVV:         "123",
					},
				},
			},
			Setup: func(svc *MockToolService) {
				expected := service.Export{
					Logins: []service.Login{
						{
							ID:       uuid.NameSpaceDNS,
							Username: "test",
							Password: "test",
							Domains:  []string{"test"},
						},
					},
					Notes: []service.Note{
						{
							ID:      uuid.NameSpaceURL,
							Name:    "test",
							Content: "test",
						},
					},
					Cards: []service.Card{
						{
							ID:          uuid.NameSpaceOID,
							HolderName:  "test",
							Number:      "4111 1111 1111 1111",
							ExpiryMonth: time.January,
							ExpiryYear:  2025,
							CVV:         "123",
						},
					},
				}

				svc.EXPECT().
					Export(mock.Anything).
					Return(expected, nil).Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			svc := NewMockToolService(t)
			if tc.Setup != nil {
				tc.Setup(svc)
			}

			w := httptest.NewRecorder()
			r := request(t, http.MethodGet, "/api/v1/export", nil).
				WithContext(token.ToContext(t.Context(), tc.Token))

			api.NewToolAPI(svc).Export(w, r)

			require.Equal(t, tc.ExpectedCode, w.Code)
			if tc.ExpectsError {
				assertAPIError(t, w)
				return
			}

			assertResponse(t, w, tc.Expected)
		})
	}
}
