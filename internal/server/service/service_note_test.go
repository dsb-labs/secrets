package service_test

import (
	"io"
	"testing"

	"github.com/davidsbond/x/filter"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/davidsbond/keeper/internal/server/database"
	"github.com/davidsbond/keeper/internal/server/service"
)

func TestNoteService_Create(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		UserID       uuid.UUID
		Note         service.Note
		ExpectsError bool
		Setup        func(notes *MockNoteRepository, provider *MockRepositoryProvider[service.NoteRepository])
	}{
		{
			Name:         "error if database lifetime has expired",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Note: service.Note{
				Name:    "test",
				Content: "test",
			},
			Setup: func(notes *MockNoteRepository, provider *MockRepositoryProvider[service.NoteRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(nil, database.ErrClosed).Once()
			},
		},
		{
			Name:         "error when getting user database fails",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Note: service.Note{
				Name:    "test",
				Content: "test",
			},
			Setup: func(notes *MockNoteRepository, provider *MockRepositoryProvider[service.NoteRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(nil, io.EOF).Once()
			},
		},
		{
			Name:         "error when lifetime has expired when creating",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Note: service.Note{
				Name:    "test",
				Content: "test",
			},
			Setup: func(notes *MockNoteRepository, provider *MockRepositoryProvider[service.NoteRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(notes, nil).Once()

				notes.EXPECT().Create(mock.Anything).Return(database.ErrClosed).Once()
			},
		},
		{
			Name:         "error when creating record",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Note: service.Note{
				Name:    "test",
				Content: "test",
			},
			Setup: func(notes *MockNoteRepository, provider *MockRepositoryProvider[service.NoteRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(notes, nil).Once()

				notes.EXPECT().Create(mock.Anything).Return(io.EOF).Once()
			},
		},
		{
			Name:   "success",
			UserID: uuid.NameSpaceDNS,
			Note: service.Note{
				Name:    "test",
				Content: "test",
			},
			Setup: func(notes *MockNoteRepository, provider *MockRepositoryProvider[service.NoteRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(notes, nil).Once()

				notes.EXPECT().Create(mock.Anything).Return(nil).Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			notes := NewMockNoteRepository(t)
			provider := NewMockRepositoryProvider[service.NoteRepository](t)

			if tc.Setup != nil {
				tc.Setup(notes, provider)
			}

			svc := service.NewNoteService(provider)
			err := svc.Create(tc.UserID, tc.Note)

			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestNoteService_List(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		UserID       uuid.UUID
		Expected     []service.Note
		ExpectsError bool
		Setup        func(notes *MockNoteRepository, provider *MockRepositoryProvider[service.NoteRepository])
		Filters      []filter.Filter[service.Note]
	}{
		{
			Name:         "error if database lifetime has expired",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(notes *MockNoteRepository, provider *MockRepositoryProvider[service.NoteRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(nil, database.ErrClosed).Once()
			},
		},
		{
			Name:         "error when getting user database fails",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(notes *MockNoteRepository, provider *MockRepositoryProvider[service.NoteRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(nil, io.EOF).Once()
			},
		},
		{
			Name:         "error if database lifetime expired has on list",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(notes *MockNoteRepository, provider *MockRepositoryProvider[service.NoteRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(notes, nil).Once()

				notes.EXPECT().List().Return(nil, database.ErrClosed).Once()
			},
		},
		{
			Name:         "error when listing notes",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(notes *MockNoteRepository, provider *MockRepositoryProvider[service.NoteRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(notes, nil).Once()

				notes.EXPECT().List().Return(nil, io.EOF).Once()
			},
		},
		{
			Name:   "success",
			UserID: uuid.NameSpaceDNS,
			Expected: []service.Note{
				{
					ID:      uuid.NameSpaceDNS,
					Name:    "test",
					Content: "test",
				},
			},
			Setup: func(notes *MockNoteRepository, provider *MockRepositoryProvider[service.NoteRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(notes, nil).Once()

				expected := []database.Note{
					{
						ID:      uuid.NameSpaceDNS,
						Name:    "test",
						Content: "test",
					},
				}

				notes.EXPECT().List().Return(expected, nil).Once()
			},
		},
		{
			Name:   "uses filters",
			UserID: uuid.NameSpaceDNS,
			Expected: []service.Note{
				{
					ID:      uuid.NameSpaceDNS,
					Name:    "test",
					Content: "test",
				},
			},
			Setup: func(notes *MockNoteRepository, provider *MockRepositoryProvider[service.NoteRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(notes, nil).Once()

				expected := []database.Note{
					{
						ID:      uuid.NameSpaceDNS,
						Name:    "test",
						Content: "test",
					},
					{
						ID:      uuid.NameSpaceURL,
						Name:    "bing",
						Content: "bong",
					},
				}

				notes.EXPECT().List().Return(expected, nil).Once()
			},
			Filters: []filter.Filter[service.Note]{
				service.NotesByQuery("tes"),
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			notes := NewMockNoteRepository(t)
			provider := NewMockRepositoryProvider[service.NoteRepository](t)

			if tc.Setup != nil {
				tc.Setup(notes, provider)
			}

			actual, err := service.NewNoteService(provider).List(tc.UserID, tc.Filters...)
			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.EqualValues(t, tc.Expected, actual)
		})
	}
}

func TestNoteService_Delete(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		UserID       uuid.UUID
		NoteID       uuid.UUID
		ExpectsError bool
		Setup        func(notes *MockNoteRepository, provider *MockRepositoryProvider[service.NoteRepository])
	}{
		{
			Name:         "error if database lifetime has expired",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(notes *MockNoteRepository, provider *MockRepositoryProvider[service.NoteRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(nil, database.ErrClosed).Once()
			},
		},
		{
			Name:         "error when getting user database fails",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(notes *MockNoteRepository, provider *MockRepositoryProvider[service.NoteRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(nil, io.EOF).Once()
			},
		},
		{
			Name:         "error if database lifetime expired has on delete",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			NoteID:       uuid.NameSpaceDNS,
			Setup: func(notes *MockNoteRepository, provider *MockRepositoryProvider[service.NoteRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(notes, nil).Once()

				notes.EXPECT().Delete(uuid.NameSpaceDNS).Return(database.ErrClosed).Once()
			},
		},
		{
			Name:         "error when deleting note",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			NoteID:       uuid.NameSpaceDNS,
			Setup: func(notes *MockNoteRepository, provider *MockRepositoryProvider[service.NoteRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(notes, nil).Once()

				notes.EXPECT().Delete(uuid.NameSpaceDNS).Return(io.EOF).Once()
			},
		},
		{
			Name:         "error if note does not exist",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			NoteID:       uuid.NameSpaceDNS,
			Setup: func(notes *MockNoteRepository, provider *MockRepositoryProvider[service.NoteRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(notes, nil).Once()

				notes.EXPECT().Delete(uuid.NameSpaceDNS).Return(database.ErrNoteNotFound).Once()
			},
		},
		{
			Name:   "success",
			UserID: uuid.NameSpaceDNS,
			NoteID: uuid.NameSpaceDNS,
			Setup: func(notes *MockNoteRepository, provider *MockRepositoryProvider[service.NoteRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(notes, nil).Once()

				notes.EXPECT().Delete(uuid.NameSpaceDNS).Return(nil).Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			notes := NewMockNoteRepository(t)
			provider := NewMockRepositoryProvider[service.NoteRepository](t)

			if tc.Setup != nil {
				tc.Setup(notes, provider)
			}

			err := service.NewNoteService(provider).Delete(tc.UserID, tc.NoteID)
			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
