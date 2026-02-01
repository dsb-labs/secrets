package service_test

import (
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/davidsbond/keeper/internal/server/database"
	"github.com/davidsbond/keeper/internal/server/service"
)

func TestToolService_Export(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		UserID       uuid.UUID
		Expected     service.Export
		ExpectsError bool
		Setup        func(logins *MockLoginRepository, notes *MockNoteRepository, loginProvider *MockRepositoryProvider[service.LoginRepository], noteProvider *MockRepositoryProvider[service.NoteRepository])
	}{
		{
			Name:         "error if login database lifetime has expired",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(_ *MockLoginRepository, _ *MockNoteRepository, loginProvider *MockRepositoryProvider[service.LoginRepository], _ *MockRepositoryProvider[service.NoteRepository]) {
				loginProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(nil, database.ErrClosed).Once()
			},
		},
		{
			Name:         "error getting login database",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(_ *MockLoginRepository, _ *MockNoteRepository, loginProvider *MockRepositoryProvider[service.LoginRepository], _ *MockRepositoryProvider[service.NoteRepository]) {
				loginProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(nil, io.EOF).Once()
			},
		},
		{
			Name:         "error if note database lifetime has expired",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(logins *MockLoginRepository, _ *MockNoteRepository, loginProvider *MockRepositoryProvider[service.LoginRepository], noteProvider *MockRepositoryProvider[service.NoteRepository]) {
				loginProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(logins, nil).Once()
				noteProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(nil, database.ErrClosed).Once()
			},
		},
		{
			Name:         "error getting note database",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(logins *MockLoginRepository, _ *MockNoteRepository, loginProvider *MockRepositoryProvider[service.LoginRepository], noteProvider *MockRepositoryProvider[service.NoteRepository]) {
				loginProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(logins, nil).Once()
				noteProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(nil, io.EOF).Once()
			},
		},
		{
			Name:         "login database lifetime expired on list",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(logins *MockLoginRepository, notes *MockNoteRepository, loginProvider *MockRepositoryProvider[service.LoginRepository], noteProvider *MockRepositoryProvider[service.NoteRepository]) {
				loginProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(logins, nil).Once()
				noteProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(notes, nil).Once()

				logins.EXPECT().
					List().
					Return(nil, database.ErrClosed).Once()
			},
		},
		{
			Name:         "error when listing logins",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(logins *MockLoginRepository, notes *MockNoteRepository, loginProvider *MockRepositoryProvider[service.LoginRepository], noteProvider *MockRepositoryProvider[service.NoteRepository]) {
				loginProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(logins, nil).Once()
				noteProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(notes, nil).Once()

				logins.EXPECT().
					List().
					Return(nil, io.EOF).Once()
			},
		},
		{
			Name:         "note database lifetime expired on list",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(logins *MockLoginRepository, notes *MockNoteRepository, loginProvider *MockRepositoryProvider[service.LoginRepository], noteProvider *MockRepositoryProvider[service.NoteRepository]) {
				loginProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(logins, nil).Once()
				noteProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(notes, nil).Once()

				logins.EXPECT().
					List().
					Return(nil, nil).Once()

				notes.EXPECT().
					List().
					Return(nil, database.ErrClosed).Once()
			},
		},
		{
			Name:         "error when listing notes",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(logins *MockLoginRepository, notes *MockNoteRepository, loginProvider *MockRepositoryProvider[service.LoginRepository], noteProvider *MockRepositoryProvider[service.NoteRepository]) {
				loginProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(logins, nil).Once()
				noteProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(notes, nil).Once()

				logins.EXPECT().
					List().
					Return(nil, nil).Once()

				notes.EXPECT().
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
			},
			UserID: uuid.NameSpaceDNS,
			Setup: func(logins *MockLoginRepository, notes *MockNoteRepository, loginProvider *MockRepositoryProvider[service.LoginRepository], noteProvider *MockRepositoryProvider[service.NoteRepository]) {
				loginProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(logins, nil).Once()
				noteProvider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(notes, nil).Once()

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

				logins.EXPECT().
					List().
					Return(expectedLogins, nil).Once()

				notes.EXPECT().
					List().
					Return(expectedNotes, nil).Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			logins := NewMockLoginRepository(t)
			notes := NewMockNoteRepository(t)
			loginProvider := NewMockRepositoryProvider[service.LoginRepository](t)
			noteProvider := NewMockRepositoryProvider[service.NoteRepository](t)

			if tc.Setup != nil {
				tc.Setup(logins, notes, loginProvider, noteProvider)
			}

			actual, err := service.NewToolService(loginProvider, noteProvider).Export(tc.UserID)
			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.EqualValues(t, tc.Expected, actual)
		})
	}
}
