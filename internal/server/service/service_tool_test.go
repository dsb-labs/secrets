package service_test

import (
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/davidsbond/keeper/internal/server/database"
	"github.com/davidsbond/keeper/internal/server/service"
)

type (
	ExportMocks struct {
		logins        *MockLoginRepository
		loginProvider *MockRepositoryProvider[service.LoginRepository]
		notes         *MockNoteRepository
		noteProvider  *MockRepositoryProvider[service.NoteRepository]
		cards         *MockCardRepository
		cardProvider  *MockRepositoryProvider[service.CardRepository]
	}
)

func TestToolService_Export(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		UserID       uuid.UUID
		Expected     service.Export
		ExpectsError bool
		Setup        func(mocks *ExportMocks)
	}{
		{
			Name:         "error if login database lifetime has expired",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(mocks *ExportMocks) {
				mocks.loginProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(nil, database.ErrClosed).Once()
			},
		},
		{
			Name:         "error getting login database",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(mocks *ExportMocks) {
				mocks.loginProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(nil, io.EOF).Once()
			},
		},
		{
			Name:         "error if note database lifetime has expired",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(mocks *ExportMocks) {
				mocks.loginProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(mocks.logins, nil).Once()
				mocks.noteProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(nil, database.ErrClosed).Once()
			},
		},
		{
			Name:         "error getting note database",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(mocks *ExportMocks) {
				mocks.loginProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(mocks.logins, nil).Once()
				mocks.noteProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(nil, io.EOF).Once()
			},
		},
		{
			Name:         "error if card database lifetime has expired",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(mocks *ExportMocks) {
				mocks.loginProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(mocks.logins, nil).Once()
				mocks.noteProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(mocks.notes, nil).Once()
				mocks.cardProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(nil, database.ErrClosed).Once()
			},
		},
		{
			Name:         "error getting card database",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(mocks *ExportMocks) {
				mocks.loginProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(mocks.logins, nil).Once()
				mocks.noteProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(mocks.notes, nil).Once()
				mocks.cardProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(nil, io.EOF).Once()
			},
		},
		{
			Name:         "login database lifetime expired on list",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(mocks *ExportMocks) {
				mocks.loginProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(mocks.logins, nil).Once()
				mocks.noteProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(mocks.notes, nil).Once()
				mocks.cardProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(mocks.cards, nil).Once()

				mocks.logins.EXPECT().
					List().
					Return(nil, database.ErrClosed).Once()
			},
		},
		{
			Name:         "error when listing logins",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(mocks *ExportMocks) {
				mocks.loginProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(mocks.logins, nil).Once()
				mocks.noteProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(mocks.notes, nil).Once()
				mocks.cardProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(mocks.cards, nil).Once()

				mocks.logins.EXPECT().
					List().
					Return(nil, io.EOF).Once()
			},
		},
		{
			Name:         "note database lifetime expired on list",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(mocks *ExportMocks) {
				mocks.loginProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(mocks.logins, nil).Once()
				mocks.noteProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(mocks.notes, nil).Once()
				mocks.cardProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(mocks.cards, nil).Once()

				mocks.logins.EXPECT().
					List().
					Return(nil, nil).Once()

				mocks.notes.EXPECT().
					List().
					Return(nil, database.ErrClosed).Once()
			},
		},
		{
			Name:         "error when listing notes",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(mocks *ExportMocks) {
				mocks.loginProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(mocks.logins, nil).Once()
				mocks.noteProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(mocks.notes, nil).Once()
				mocks.cardProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(mocks.cards, nil).Once()

				mocks.logins.EXPECT().
					List().
					Return(nil, nil).Once()

				mocks.notes.EXPECT().
					List().
					Return(nil, io.EOF).Once()
			},
		},
		{
			Name:         "card database lifetime expired on list",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(mocks *ExportMocks) {
				mocks.loginProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(mocks.logins, nil).Once()
				mocks.noteProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(mocks.notes, nil).Once()
				mocks.cardProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(mocks.cards, nil).Once()

				mocks.logins.EXPECT().
					List().
					Return(nil, nil).Once()

				mocks.notes.EXPECT().
					List().
					Return(nil, nil).Once()

				mocks.cards.EXPECT().
					List().
					Return(nil, database.ErrClosed).Once()
			},
		},
		{
			Name:         "error when listing cards",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(mocks *ExportMocks) {
				mocks.loginProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(mocks.logins, nil).Once()
				mocks.noteProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(mocks.notes, nil).Once()
				mocks.cardProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(mocks.cards, nil).Once()

				mocks.logins.EXPECT().
					List().
					Return(nil, nil).Once()

				mocks.notes.EXPECT().
					List().
					Return(nil, nil).Once()

				mocks.cards.EXPECT().
					List().
					Return(nil, io.EOF).Once()
			},
		},
		{
			Name: "success",
			Expected: service.Export{
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
						ID:          uuid.NameSpaceURL,
						HolderName:  "test",
						Number:      "test",
						ExpiryMonth: time.January,
						ExpiryYear:  2025,
						CVV:         "123",
					},
				},
			},
			UserID: uuid.NameSpaceDNS,
			Setup: func(mocks *ExportMocks) {
				mocks.loginProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(mocks.logins, nil).Once()
				mocks.noteProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(mocks.notes, nil).Once()
				mocks.cardProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(mocks.cards, nil).Once()

				expectedLogins := []database.Login{
					{
						ID:       uuid.NameSpaceDNS,
						Username: "test",
						Password: "test",
						Domains:  []string{"test"},
					},
				}

				expectedNotes := []database.Note{
					{
						ID:      uuid.NameSpaceURL,
						Name:    "test",
						Content: "test",
					},
				}

				expectedCards := []database.Card{
					{
						ID:          uuid.NameSpaceURL,
						HolderName:  "test",
						Number:      "test",
						ExpiryMonth: time.January,
						ExpiryYear:  2025,
						CVV:         "123",
					},
				}

				mocks.logins.EXPECT().
					List().
					Return(expectedLogins, nil).Once()

				mocks.notes.EXPECT().
					List().
					Return(expectedNotes, nil).Once()

				mocks.cards.EXPECT().
					List().
					Return(expectedCards, nil).Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			mocks := &ExportMocks{
				logins:        NewMockLoginRepository(t),
				loginProvider: NewMockRepositoryProvider[service.LoginRepository](t),
				notes:         NewMockNoteRepository(t),
				noteProvider:  NewMockRepositoryProvider[service.NoteRepository](t),
				cards:         NewMockCardRepository(t),
				cardProvider:  NewMockRepositoryProvider[service.CardRepository](t),
			}

			if tc.Setup != nil {
				tc.Setup(mocks)
			}

			actual, err := service.NewToolService(mocks.loginProvider, mocks.noteProvider, mocks.cardProvider).Export(tc.UserID)
			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.EqualValues(t, tc.Expected, actual)
		})
	}
}
