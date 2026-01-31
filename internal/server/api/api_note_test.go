package api_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/davidsbond/keeper/internal/server/api"
	"github.com/davidsbond/keeper/internal/server/service"
	"github.com/davidsbond/keeper/internal/server/token"
)

func TestNoteAPI_Create(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		Request      api.CreateNoteRequest
		Expected     api.CreateNoteResponse
		Token        token.Token
		ExpectedCode int
		ExpectsError bool
		Setup        func(svc *MockNoteService)
	}{
		{
			Name:         "error if no token",
			Token:        token.Token{},
			ExpectsError: true,
			ExpectedCode: http.StatusUnauthorized,
			Request: api.CreateNoteRequest{
				Name:    "test",
				Content: "test",
			},
		},
		{
			Name:  "error if missing name",
			Token: token.TestToken(t, "test"),
			Request: api.CreateNoteRequest{
				Name:    "",
				Content: "content",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusBadRequest,
		},
		{
			Name:  "error if missing content",
			Token: token.TestToken(t, "test"),
			Request: api.CreateNoteRequest{
				Name:    "test",
				Content: "",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusBadRequest,
		},
		{
			Name:  "error if lifetime has expired",
			Token: token.TestToken(t, "test"),
			Request: api.CreateNoteRequest{
				Name:    "test",
				Content: "test",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusUnauthorized,
			Setup: func(svc *MockNoteService) {
				svc.EXPECT().Create(mock.Anything, mock.Anything).Return(service.ErrReauthenticate).Once()
			},
		},
		{
			Name:  "error if creation fails",
			Token: token.TestToken(t, "test"),
			Request: api.CreateNoteRequest{
				Name:    "test",
				Content: "test",
			},
			ExpectsError: true,
			ExpectedCode: http.StatusInternalServerError,
			Setup: func(svc *MockNoteService) {
				svc.EXPECT().Create(mock.Anything, mock.Anything).Return(io.EOF).Once()
			},
		},
		{
			Name:  "success",
			Token: token.TestToken(t, "test"),
			Request: api.CreateNoteRequest{
				Name:    "test",
				Content: "test",
			},
			Expected:     api.CreateNoteResponse{},
			ExpectedCode: http.StatusCreated,
			Setup: func(svc *MockNoteService) {
				svc.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			svc := NewMockNoteService(t)
			if tc.Setup != nil {
				tc.Setup(svc)
			}

			w := httptest.NewRecorder()
			r := request(t, http.MethodPost, "/api/v1/note", tc.Request).
				WithContext(token.ToContext(t.Context(), tc.Token))

			api.NewNoteAPI(svc).Create(w, r)

			require.Equal(t, tc.ExpectedCode, w.Code)
			if tc.ExpectsError {
				assertAPIError(t, w)
				return
			}

			assertResponse(t, w, tc.Expected)
		})
	}
}

func TestNoteAPI_List(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		Expected     api.ListNotesResponse
		Token        token.Token
		ExpectedCode int
		ExpectsError bool
		Setup        func(svc *MockNoteService)
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
			Setup: func(svc *MockNoteService) {
				svc.EXPECT().List(mock.Anything).Return(nil, service.ErrReauthenticate).Once()
			},
		},
		{
			Name:         "error if listing fails",
			Token:        token.TestToken(t, "test"),
			ExpectsError: true,
			ExpectedCode: http.StatusInternalServerError,
			Setup: func(svc *MockNoteService) {
				svc.EXPECT().List(mock.Anything).Return(nil, io.EOF).Once()
			},
		},
		{
			Name:  "success",
			Token: token.TestToken(t, "test"),
			Expected: api.ListNotesResponse{
				Notes: []api.Note{
					{
						ID:      uuid.NameSpaceDNS.String(),
						Name:    "test",
						Content: "test",
					},
				},
			},
			ExpectedCode: http.StatusOK,
			Setup: func(svc *MockNoteService) {
				expected := []service.Note{
					{
						ID:      uuid.NameSpaceDNS,
						Name:    "test",
						Content: "test",
					},
				}

				svc.EXPECT().List(mock.Anything).Return(expected, nil).Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			svc := NewMockNoteService(t)
			if tc.Setup != nil {
				tc.Setup(svc)
			}

			w := httptest.NewRecorder()
			r := request(t, http.MethodGet, "/api/v1/note", nil).
				WithContext(token.ToContext(t.Context(), tc.Token))

			api.NewNoteAPI(svc).List(w, r)

			require.Equal(t, tc.ExpectedCode, w.Code)
			if tc.ExpectsError {
				assertAPIError(t, w)
				return
			}

			assertResponse(t, w, tc.Expected)
		})
	}
}

func TestNoteAPI_Delete(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		Expected     api.DeleteNoteResponse
		Token        token.Token
		ID           string
		ExpectedCode int
		ExpectsError bool
		Setup        func(svc *MockNoteService)
	}{
		{
			Name:         "error if no token",
			Token:        token.Token{},
			ExpectsError: true,
			ExpectedCode: http.StatusUnauthorized,
		},
		{
			Name:         "error if note id is not uuid",
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
			Setup: func(svc *MockNoteService) {
				svc.EXPECT().Delete(mock.Anything, uuid.NameSpaceDNS).Return(service.ErrReauthenticate).Once()
			},
		},
		{
			Name:         "error if note does not exist",
			Token:        token.TestToken(t, "test"),
			ID:           uuid.NameSpaceDNS.String(),
			ExpectsError: true,
			ExpectedCode: http.StatusNotFound,
			Setup: func(svc *MockNoteService) {
				svc.EXPECT().Delete(mock.Anything, uuid.NameSpaceDNS).Return(service.ErrNoteNotFound).Once()
			},
		},
		{
			Name:         "error if deletion fails",
			Token:        token.TestToken(t, "test"),
			ID:           uuid.NameSpaceDNS.String(),
			ExpectsError: true,
			ExpectedCode: http.StatusInternalServerError,
			Setup: func(svc *MockNoteService) {
				svc.EXPECT().Delete(mock.Anything, uuid.NameSpaceDNS).Return(io.EOF).Once()
			},
		},
		{
			Name:         "success",
			Token:        token.TestToken(t, "test"),
			ExpectedCode: http.StatusOK,
			ID:           uuid.NameSpaceDNS.String(),
			Setup: func(svc *MockNoteService) {
				svc.EXPECT().Delete(mock.Anything, uuid.NameSpaceDNS).Return(nil).Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			svc := NewMockNoteService(t)
			if tc.Setup != nil {
				tc.Setup(svc)
			}

			w := httptest.NewRecorder()
			r := request(t, http.MethodDelete, "/api/v1/note", nil).
				WithContext(token.ToContext(t.Context(), tc.Token))
			r.SetPathValue("id", tc.ID)

			api.NewNoteAPI(svc).Delete(w, r)

			require.Equal(t, tc.ExpectedCode, w.Code)
			if tc.ExpectsError {
				assertAPIError(t, w)
				return
			}

			assertResponse(t, w, tc.Expected)
		})
	}
}
