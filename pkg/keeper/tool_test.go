package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/davidsbond/keeper/pkg/keeper"
)

func TestClient_Export(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("error if not authenticated", func(t *testing.T) {
		_, err := client.Export(ctx)
		require.Error(t, err)
		assert.True(t, keeper.IsUnauthorized(err))
	})

	setupAccount(t, client)

	t.Run("initial export has nothing", func(t *testing.T) {
		export, err := client.Export(ctx)
		require.NoError(t, err)
		assert.Len(t, export.Logins, 0)
		assert.Len(t, export.Cards, 0)
		assert.Len(t, export.Notes, 0)
	})

	expectedCard := keeper.Card{
		HolderName:  "Test McTest",
		Number:      "4111 1111 1111 1111",
		ExpiryMonth: time.March,
		ExpiryYear:  2027,
		CVV:         "123",
	}

	expectedLogin := keeper.Login{
		Username: "test",
		Password: "test",
		Domains:  []string{"test.com"},
	}

	expectedNote := keeper.Note{
		Name:    "test",
		Content: "test",
	}

	noteID, err := client.CreateNote(ctx, expectedNote)
	require.NoError(t, err)

	loginID, err := client.CreateLogin(ctx, expectedLogin)
	require.NoError(t, err)

	cardID, err := client.CreateCard(ctx, expectedCard)
	require.NoError(t, err)

	t.Run("export contains everything", func(t *testing.T) {
		export, err := client.Export(ctx)
		require.NoError(t, err)
		require.Len(t, export.Logins, 1)
		require.Len(t, export.Cards, 1)
		require.Len(t, export.Notes, 1)

		actualCard := export.Cards[0]
		actualLogin := export.Logins[0]
		actualNote := export.Notes[0]

		t.Run("note matches", func(t *testing.T) {
			assert.EqualValues(t, noteID, actualNote.ID)
			assert.EqualValues(t, expectedNote.Content, actualNote.Content)
			assert.EqualValues(t, expectedNote.Name, actualNote.Name)
		})

		t.Run("card matches", func(t *testing.T) {
			assert.EqualValues(t, cardID, actualCard.ID)
			assert.EqualValues(t, expectedCard.CVV, actualCard.CVV)
			assert.EqualValues(t, expectedCard.HolderName, actualCard.HolderName)
			assert.EqualValues(t, expectedCard.Number, actualCard.Number)
			assert.EqualValues(t, expectedCard.ExpiryMonth, actualCard.ExpiryMonth)
			assert.EqualValues(t, expectedCard.ExpiryYear, actualCard.ExpiryYear)
		})

		t.Run("login matches", func(t *testing.T) {
			assert.EqualValues(t, loginID, actualLogin.ID)
			assert.EqualValues(t, expectedLogin.Username, actualLogin.Username)
			assert.EqualValues(t, expectedLogin.Password, actualLogin.Password)
			assert.EqualValues(t, expectedLogin.Domains, actualLogin.Domains)
		})
	})
}
