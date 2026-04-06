package secrets_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsb-labs/secrets/pkg/secrets"
)

func TestClient_CreateNote(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("error if not authenticated", func(t *testing.T) {
		_, err := client.CreateNote(ctx, secrets.Note{})
		require.Error(t, err)
		assert.True(t, secrets.IsUnauthorized(err))
	})

	setupAccount(t, client)

	t.Run("creates note", func(t *testing.T) {
		note := secrets.Note{
			Name:    "test",
			Content: "test",
		}

		id, err := client.CreateNote(ctx, note)
		require.NoError(t, err)
		require.NoError(t, uuid.Validate(id))
	})

	t.Run("error if note is invalid", func(t *testing.T) {
		note := secrets.Note{}
		_, err := client.CreateNote(ctx, note)
		require.Error(t, err)
		assert.True(t, secrets.IsBadRequest(err))
	})
}

func TestClient_ListNotes(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("error if not authenticated", func(t *testing.T) {
		_, err := client.ListNotes(ctx, secrets.NoteListOptions{})
		require.Error(t, err)
		assert.True(t, secrets.IsUnauthorized(err))
	})

	setupAccount(t, client)

	t.Run("lists no notes", func(t *testing.T) {
		notes, err := client.ListNotes(ctx, secrets.NoteListOptions{})
		require.NoError(t, err)
		assert.Len(t, notes, 0)
	})

	expected := secrets.Note{
		Name:    "test",
		Content: "test",
	}

	noteID, err := client.CreateNote(ctx, expected)
	require.NoError(t, err)

	t.Run("lists notes", func(t *testing.T) {
		notes, err := client.ListNotes(ctx, secrets.NoteListOptions{})
		require.NoError(t, err)
		if assert.Len(t, notes, 1) {
			actual := notes[0]
			assert.EqualValues(t, noteID, actual.ID)
			assert.EqualValues(t, expected.Content, actual.Content)
			assert.EqualValues(t, expected.Name, actual.Name)
		}
	})

	t.Run("lists notes by query", func(t *testing.T) {
		notes, err := client.ListNotes(ctx, secrets.NoteListOptions{Query: "test"})
		require.NoError(t, err)
		if assert.Len(t, notes, 1) {
			actual := notes[0]
			assert.EqualValues(t, noteID, actual.ID)
			assert.EqualValues(t, expected.Content, actual.Content)
			assert.EqualValues(t, expected.Name, actual.Name)
		}
	})
}

func TestClient_DeleteNote(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("error if not authenticated", func(t *testing.T) {
		err := client.DeleteNote(ctx, uuid.NameSpaceDNS.String())
		require.Error(t, err)
		assert.True(t, secrets.IsUnauthorized(err))
	})

	setupAccount(t, client)

	noteID, err := client.CreateNote(ctx, secrets.Note{
		Name:    "test",
		Content: "test",
	})
	require.NoError(t, err)

	t.Run("error if note does not exist", func(t *testing.T) {
		err = client.DeleteNote(ctx, uuid.NameSpaceDNS.String())
		require.Error(t, err)
		assert.True(t, secrets.IsNotFound(err))
	})

	t.Run("deletes note", func(t *testing.T) {
		err = client.DeleteNote(ctx, noteID)
		require.NoError(t, err)
	})

	t.Run("note does not exist", func(t *testing.T) {
		_, err = client.GetNote(ctx, noteID)
		require.Error(t, err)
		assert.True(t, secrets.IsNotFound(err))
	})
}

func TestClient_GetNote(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("error if not authenticated", func(t *testing.T) {
		_, err := client.GetNote(ctx, uuid.NameSpaceDNS.String())
		require.Error(t, err)
		assert.True(t, secrets.IsUnauthorized(err))
	})

	setupAccount(t, client)

	expected := secrets.Note{
		Name:    "test",
		Content: "test",
	}

	noteID, err := client.CreateNote(ctx, expected)
	require.NoError(t, err)

	t.Run("error if note does not exist", func(t *testing.T) {
		_, err = client.GetNote(ctx, uuid.NameSpaceDNS.String())
		require.Error(t, err)
		assert.True(t, secrets.IsNotFound(err))
	})

	t.Run("gets note", func(t *testing.T) {
		actual, err := client.GetNote(ctx, noteID)
		require.NoError(t, err)

		assert.EqualValues(t, noteID, actual.ID)
		assert.EqualValues(t, expected.Content, actual.Content)
		assert.EqualValues(t, expected.Name, actual.Name)
	})
}
