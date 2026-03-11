package keeper_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/davidsbond/keeper/pkg/keeper"
)

func TestClient_CreateNote(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("error if not authenticated", func(t *testing.T) {
		err := client.CreateNote(ctx, keeper.Note{})
		require.Error(t, err)
		assert.True(t, keeper.IsUnauthorized(err))
	})

	setupAccount(t, client)

	t.Run("creates note", func(t *testing.T) {
		note := keeper.Note{
			Name:    "test",
			Content: "test",
		}

		err := client.CreateNote(ctx, note)
		require.NoError(t, err)
	})

	t.Run("error if note is invalid", func(t *testing.T) {
		note := keeper.Note{}
		err := client.CreateNote(ctx, note)
		require.Error(t, err)
		assert.True(t, keeper.IsBadRequest(err))
	})
}

func TestClient_ListNotes(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("error if not authenticated", func(t *testing.T) {
		_, err := client.ListNotes(ctx, "")
		require.Error(t, err)
		assert.True(t, keeper.IsUnauthorized(err))
	})

	setupAccount(t, client)

	t.Run("lists no notes", func(t *testing.T) {
		notes, err := client.ListNotes(ctx, "")
		require.NoError(t, err)
		assert.Len(t, notes, 0)
	})

	expected := keeper.Note{
		Name:    "test",
		Content: "test",
	}

	err := client.CreateNote(ctx, expected)
	require.NoError(t, err)

	t.Run("lists notes", func(t *testing.T) {
		notes, err := client.ListNotes(ctx, "")
		require.NoError(t, err)
		if assert.Len(t, notes, 1) {
			actual := notes[0]
			assert.Equal(t, expected.Content, actual.Content)
			assert.Equal(t, expected.Name, actual.Name)
		}
	})

	t.Run("lists notes by query", func(t *testing.T) {
		notes, err := client.ListNotes(ctx, "test")
		require.NoError(t, err)
		if assert.Len(t, notes, 1) {
			actual := notes[0]
			assert.Equal(t, expected.Content, actual.Content)
			assert.Equal(t, expected.Name, actual.Name)
		}
	})
}

func TestClient_DeleteNote(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("error if not authenticated", func(t *testing.T) {
		err := client.DeleteNote(ctx, uuid.NameSpaceDNS.String())
		require.Error(t, err)
		assert.True(t, keeper.IsUnauthorized(err))
	})

	setupAccount(t, client)

	err := client.CreateNote(ctx, keeper.Note{
		Name:    "test",
		Content: "test",
	})
	require.NoError(t, err)

	notes, err := client.ListNotes(ctx, "")
	require.NoError(t, err)
	require.Len(t, notes, 1)

	t.Run("error if note does not exist", func(t *testing.T) {
		err = client.DeleteNote(ctx, uuid.NameSpaceDNS.String())
		require.Error(t, err)
		assert.True(t, keeper.IsNotFound(err))
	})

	t.Run("deletes note", func(t *testing.T) {
		err = client.DeleteNote(ctx, notes[0].ID)
		require.NoError(t, err)
	})

	t.Run("note does not exist", func(t *testing.T) {
		_, err = client.GetNote(ctx, notes[0].ID)
		require.Error(t, err)
		assert.True(t, keeper.IsNotFound(err))
	})
}

func TestClient_GetNote(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("error if not authenticated", func(t *testing.T) {
		_, err := client.GetNote(ctx, uuid.NameSpaceDNS.String())
		require.Error(t, err)
		assert.True(t, keeper.IsUnauthorized(err))
	})

	setupAccount(t, client)

	err := client.CreateNote(ctx, keeper.Note{
		Name:    "test",
		Content: "test",
	})
	require.NoError(t, err)

	notes, err := client.ListNotes(ctx, "")
	require.NoError(t, err)
	require.Len(t, notes, 1)
	expected := notes[0]

	t.Run("error if note does not exist", func(t *testing.T) {
		_, err = client.GetNote(ctx, uuid.NameSpaceDNS.String())
		require.Error(t, err)
		assert.True(t, keeper.IsNotFound(err))
	})

	t.Run("gets note", func(t *testing.T) {
		actual, err := client.GetNote(ctx, expected.ID)
		require.NoError(t, err)

		assert.Equal(t, expected.Content, actual.Content)
		assert.Equal(t, expected.Name, actual.Name)
	})
}
