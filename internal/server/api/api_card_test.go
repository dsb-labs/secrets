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

	"github.com/davidsbond/keeper/internal/server/api"
	"github.com/davidsbond/keeper/internal/server/service"
	"github.com/davidsbond/keeper/internal/server/token"
)

func TestCardAPI_Create(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		Request      api.CreateCardRequest
		Expected     api.CreateCardResponse
		Token        token.Token
		ExpectedCode int
		ExpectsError bool
		Setup        func(svc *MockCardService)
	}{
		{
			Name:  "error if missing card number",
			Token: token.TestToken(t, "test"),
			Request: api.CreateCardRequest{
				HolderName:  "test",
				Number:      "",
				ExpiryMonth: time.January,
				ExpiryYear:  2025,
				CVV:         "123",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusBadRequest,
		},
		{
			Name:  "error if card number is invalid",
			Token: token.TestToken(t, "test"),
			Request: api.CreateCardRequest{
				HolderName:  "test",
				Number:      "not-a-card-number",
				ExpiryMonth: time.January,
				ExpiryYear:  2025,
				CVV:         "123",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusBadRequest,
		},
		{
			Name:  "error if month is out of range",
			Token: token.TestToken(t, "test"),
			Request: api.CreateCardRequest{
				HolderName:  "test",
				Number:      "4111 1111 1111 1111",
				ExpiryMonth: time.Month(100),
				ExpiryYear:  2025,
				CVV:         "123",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusBadRequest,
		},
		{
			Name:  "error if cvv is invalid",
			Token: token.TestToken(t, "test"),
			Request: api.CreateCardRequest{
				HolderName:  "test",
				Number:      "4111 1111 1111 1111",
				ExpiryMonth: time.January,
				ExpiryYear:  2025,
				CVV:         "12345678",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusBadRequest,
		},
		{
			Name:  "error if missing cvv",
			Token: token.TestToken(t, "test"),
			Request: api.CreateCardRequest{
				HolderName:  "test",
				Number:      "4111 1111 1111 1111",
				ExpiryMonth: time.January,
				ExpiryYear:  2025,
				CVV:         "",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusBadRequest,
		},
		{
			Name:  "error if lifetime has expired",
			Token: token.TestToken(t, "test"),
			Request: api.CreateCardRequest{
				HolderName:  "test",
				Number:      "4111 1111 1111 1111",
				ExpiryMonth: time.January,
				ExpiryYear:  2025,
				CVV:         "123",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusUnauthorized,
			Setup: func(svc *MockCardService) {
				svc.EXPECT().Create(mock.Anything, mock.Anything).Return(service.ErrReauthenticate).Once()
			},
		},
		{
			Name:  "error if creation fails",
			Token: token.TestToken(t, "test"),
			Request: api.CreateCardRequest{
				HolderName:  "test",
				Number:      "4111 1111 1111 1111",
				ExpiryMonth: time.January,
				ExpiryYear:  2025,
				CVV:         "123",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusInternalServerError,
			Setup: func(svc *MockCardService) {
				svc.EXPECT().Create(mock.Anything, mock.Anything).Return(io.EOF).Once()
			},
		},
		{
			Name:  "success",
			Token: token.TestToken(t, "test"),
			Request: api.CreateCardRequest{
				HolderName:  "test",
				Number:      "4111 1111 1111 1111",
				ExpiryMonth: time.January,
				ExpiryYear:  2025,
				CVV:         "123",
			},
			Expected:     api.CreateCardResponse{},
			ExpectedCode: http.StatusCreated,
			Setup: func(svc *MockCardService) {
				svc.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			svc := NewMockCardService(t)
			if tc.Setup != nil {
				tc.Setup(svc)
			}

			w := httptest.NewRecorder()
			r := request(t, http.MethodPost, "/api/v1/card", tc.Request).
				WithContext(token.ToContext(t.Context(), tc.Token))

			api.NewCardAPI(svc).Create(w, r)

			require.Equal(t, tc.ExpectedCode, w.Code)
			if tc.ExpectsError {
				assertAPIError(t, w)
				return
			}

			assertResponse(t, w, tc.Expected)
		})
	}
}

func TestCardAPI_List(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		Expected     api.ListCardsResponse
		Token        token.Token
		ExpectedCode int
		ExpectsError bool
		Setup        func(svc *MockCardService)
	}{
		{
			Name:         "error if lifetime has expired",
			Token:        token.TestToken(t, "test"),
			ExpectsError: true,
			ExpectedCode: http.StatusUnauthorized,
			Setup: func(svc *MockCardService) {
				svc.EXPECT().List(mock.Anything, mock.Anything).Return(nil, service.ErrReauthenticate).Once()
			},
		},
		{
			Name:         "error if listing fails",
			Token:        token.TestToken(t, "test"),
			ExpectsError: true,
			ExpectedCode: http.StatusInternalServerError,
			Setup: func(svc *MockCardService) {
				svc.EXPECT().List(mock.Anything, mock.Anything).Return(nil, io.EOF).Once()
			},
		},
		{
			Name:  "success",
			Token: token.TestToken(t, "test"),
			Expected: api.ListCardsResponse{
				Cards: []api.Card{
					{
						ID:          uuid.NameSpaceDNS.String(),
						HolderName:  "test",
						Number:      "4111 1111 1111 1111",
						ExpiryMonth: time.January,
						ExpiryYear:  2025,
						CVV:         "123",
					},
				},
			},
			ExpectedCode: http.StatusOK,
			Setup: func(svc *MockCardService) {
				expected := []service.Card{
					{
						ID:          uuid.NameSpaceDNS,
						HolderName:  "test",
						Number:      "4111 1111 1111 1111",
						ExpiryMonth: time.January,
						ExpiryYear:  2025,
						CVV:         "123",
					},
				}

				svc.EXPECT().List(mock.Anything, mock.Anything).Return(expected, nil).Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			svc := NewMockCardService(t)
			if tc.Setup != nil {
				tc.Setup(svc)
			}

			w := httptest.NewRecorder()
			r := request(t, http.MethodGet, "/api/v1/card", nil).
				WithContext(token.ToContext(t.Context(), tc.Token))

			api.NewCardAPI(svc).List(w, r)

			require.Equal(t, tc.ExpectedCode, w.Code)
			if tc.ExpectsError {
				assertAPIError(t, w)
				return
			}

			assertResponse(t, w, tc.Expected)
		})
	}
}

func TestCardAPI_Delete(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		Expected     api.DeleteCardResponse
		Token        token.Token
		ID           string
		ExpectedCode int
		ExpectsError bool
		Setup        func(svc *MockCardService)
	}{
		{
			Name:         "error if card id is not uuid",
			Token:        token.TestToken(t, "test"),
			ID:           "not-a-uuid",
			ExpectsError: true,
			ExpectedCode: http.StatusBadRequest,
		},
		{
			Name:         "error if lifetime has expired",
			Token:        token.TestToken(t, "test"),
			ID:           uuid.NameSpaceDNS.String(),
			ExpectsError: true,
			ExpectedCode: http.StatusUnauthorized,
			Setup: func(svc *MockCardService) {
				svc.EXPECT().Delete(mock.Anything, uuid.NameSpaceDNS).Return(service.ErrReauthenticate).Once()
			},
		},
		{
			Name:         "error if card does not exist",
			Token:        token.TestToken(t, "test"),
			ID:           uuid.NameSpaceDNS.String(),
			ExpectsError: true,
			ExpectedCode: http.StatusNotFound,
			Setup: func(svc *MockCardService) {
				svc.EXPECT().Delete(mock.Anything, uuid.NameSpaceDNS).Return(service.ErrCardNotFound).Once()
			},
		},
		{
			Name:         "error if deletion fails",
			Token:        token.TestToken(t, "test"),
			ID:           uuid.NameSpaceDNS.String(),
			ExpectsError: true,
			ExpectedCode: http.StatusInternalServerError,
			Setup: func(svc *MockCardService) {
				svc.EXPECT().Delete(mock.Anything, uuid.NameSpaceDNS).Return(io.EOF).Once()
			},
		},
		{
			Name:         "success",
			Token:        token.TestToken(t, "test"),
			ExpectedCode: http.StatusOK,
			ID:           uuid.NameSpaceDNS.String(),
			Setup: func(svc *MockCardService) {
				svc.EXPECT().Delete(mock.Anything, uuid.NameSpaceDNS).Return(nil).Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			svc := NewMockCardService(t)
			if tc.Setup != nil {
				tc.Setup(svc)
			}

			w := httptest.NewRecorder()
			r := request(t, http.MethodDelete, "/api/v1/card", nil).
				WithContext(token.ToContext(t.Context(), tc.Token))
			r.SetPathValue("id", tc.ID)

			api.NewCardAPI(svc).Delete(w, r)

			require.Equal(t, tc.ExpectedCode, w.Code)
			if tc.ExpectsError {
				assertAPIError(t, w)
				return
			}

			assertResponse(t, w, tc.Expected)
		})
	}
}

func TestCardAPI_Get(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		Expected     api.GetCardResponse
		Token        token.Token
		ID           string
		ExpectedCode int
		ExpectsError bool
		Setup        func(svc *MockCardService)
	}{
		{
			Name:         "error if card id is not uuid",
			Token:        token.TestToken(t, "test"),
			ID:           "not-a-uuid",
			ExpectsError: true,
			ExpectedCode: http.StatusBadRequest,
		},
		{
			Name:         "error if lifetime has expired",
			Token:        token.TestToken(t, "test"),
			ID:           uuid.NameSpaceDNS.String(),
			ExpectsError: true,
			ExpectedCode: http.StatusUnauthorized,
			Setup: func(svc *MockCardService) {
				svc.EXPECT().
					Get(mock.Anything, uuid.NameSpaceDNS).
					Return(service.Card{}, service.ErrReauthenticate).Once()
			},
		},
		{
			Name:         "error if card does not exist",
			Token:        token.TestToken(t, "test"),
			ID:           uuid.NameSpaceDNS.String(),
			ExpectsError: true,
			ExpectedCode: http.StatusNotFound,
			Setup: func(svc *MockCardService) {
				svc.EXPECT().
					Get(mock.Anything, uuid.NameSpaceDNS).
					Return(service.Card{}, service.ErrCardNotFound).Once()
			},
		},
		{
			Name:         "error if get fails",
			Token:        token.TestToken(t, "test"),
			ID:           uuid.NameSpaceDNS.String(),
			ExpectsError: true,
			ExpectedCode: http.StatusInternalServerError,
			Setup: func(svc *MockCardService) {
				svc.EXPECT().
					Get(mock.Anything, uuid.NameSpaceDNS).
					Return(service.Card{}, io.EOF).Once()
			},
		},
		{
			Name:         "success",
			Token:        token.TestToken(t, "test"),
			ExpectedCode: http.StatusOK,
			ID:           uuid.NameSpaceDNS.String(),
			Expected: api.GetCardResponse{
				Card: api.Card{
					ID:          uuid.NameSpaceDNS.String(),
					HolderName:  "test",
					Number:      "4111 1111 1111 1111",
					ExpiryMonth: time.January,
					ExpiryYear:  2025,
					CVV:         "123",
				},
			},
			Setup: func(svc *MockCardService) {
				svc.EXPECT().
					Get(mock.Anything, uuid.NameSpaceDNS).
					Return(service.Card{
						ID:          uuid.NameSpaceDNS,
						HolderName:  "test",
						Number:      "4111 1111 1111 1111",
						ExpiryMonth: time.January,
						ExpiryYear:  2025,
						CVV:         "123",
					}, nil).Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			svc := NewMockCardService(t)
			if tc.Setup != nil {
				tc.Setup(svc)
			}

			w := httptest.NewRecorder()
			r := request(t, http.MethodGet, "/api/v1/card", nil).
				WithContext(token.ToContext(t.Context(), tc.Token))
			r.SetPathValue("id", tc.ID)

			api.NewCardAPI(svc).Get(w, r)

			require.Equal(t, tc.ExpectedCode, w.Code)
			if tc.ExpectsError {
				assertAPIError(t, w)
				return
			}

			assertResponse(t, w, tc.Expected)
		})
	}
}
