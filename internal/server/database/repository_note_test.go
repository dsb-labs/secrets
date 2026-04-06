package database_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsb-labs/secrets/internal/server/database"
)

func TestNoteRepository_Create(t *testing.T) {
	t.Parallel()

	db := testDB(t)

	tt := []struct {
		Name         string
		Note         database.Note
		ExpectsError bool
	}{
		{
			Name: "creates note",
			Note: database.Note{
				ID:      uuid.New(),
				Name:    "test",
				Content: "test",
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			err := database.NewNoteRepository(db).Create(tc.Note)
			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestNoteRepository_List(t *testing.T) {
	t.Parallel()

	db := testDB(t)

	tt := []struct {
		Name     string
		Expected []database.Note
		Setup    func(notes *database.NoteRepository)
	}{
		{
			Name: "lists notes",
			Expected: []database.Note{
				{
					ID:      uuid.NameSpaceDNS,
					Name:    "test",
					Content: "test",
				},
			},
			Setup: func(notes *database.NoteRepository) {
				expected := database.Note{
					ID:      uuid.NameSpaceDNS,
					Name:    "test",
					Content: "test",
				}

				require.NoError(t, notes.Create(expected))
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			notes := database.NewNoteRepository(db)
			if tc.Setup != nil {
				tc.Setup(notes)
			}

			actual, err := notes.List()
			require.NoError(t, err)
			assert.EqualValues(t, tc.Expected, actual)
		})
	}
}

func TestNoteRepository_Delete(t *testing.T) {
	t.Parallel()

	db := testDB(t)

	tt := []struct {
		Name         string
		ID           uuid.UUID
		ExpectsError bool
		Setup        func(notes *database.NoteRepository)
	}{
		{
			Name: "deletes note",
			ID:   uuid.NameSpaceDNS,
			Setup: func(notes *database.NoteRepository) {
				expected := database.Note{
					ID:      uuid.NameSpaceDNS,
					Name:    "test",
					Content: "test",
				}

				require.NoError(t, notes.Create(expected))
			},
		},
		{
			Name:         "error if note does not exist",
			ID:           uuid.NameSpaceDNS,
			ExpectsError: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			notes := database.NewNoteRepository(db)
			if tc.Setup != nil {
				tc.Setup(notes)
			}

			err := notes.Delete(tc.ID)
			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestNoteRepository_Get(t *testing.T) {
	t.Parallel()

	db := testDB(t)

	tt := []struct {
		Name         string
		ID           uuid.UUID
		ExpectsError bool
		Expected     database.Note
		Setup        func(notes *database.NoteRepository)
	}{
		{
			Name: "gets note",
			ID:   uuid.NameSpaceDNS,
			Expected: database.Note{
				ID:      uuid.NameSpaceDNS,
				Name:    "test",
				Content: "test",
			},
			Setup: func(notes *database.NoteRepository) {
				expected := database.Note{
					ID:      uuid.NameSpaceDNS,
					Name:    "test",
					Content: "test",
				}

				require.NoError(t, notes.Create(expected))
			},
		},
		{
			Name:         "error if note does not exist",
			ID:           uuid.NameSpaceURL,
			ExpectsError: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			notes := database.NewNoteRepository(db)
			if tc.Setup != nil {
				tc.Setup(notes)
			}

			actual, err := notes.Get(tc.ID)
			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.EqualValues(t, tc.Expected, actual)
		})
	}
}
