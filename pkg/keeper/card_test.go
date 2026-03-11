package keeper_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/davidsbond/keeper/pkg/keeper"
)

func TestClient_CreateCard(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("error if not authenticated", func(t *testing.T) {
		err := client.CreateCard(ctx, keeper.Card{})
		require.Error(t, err)
		assert.True(t, keeper.IsUnauthorized(err))
	})

	setupAccount(t, client)

	t.Run("creates card", func(t *testing.T) {
		card := keeper.Card{
			HolderName:  "Test McTest",
			Number:      "4111 1111 1111 1111",
			ExpiryMonth: time.March,
			ExpiryYear:  2027,
			CVV:         "123",
		}

		err := client.CreateCard(ctx, card)
		require.NoError(t, err)
	})

	t.Run("error if card is invalid", func(t *testing.T) {
		card := keeper.Card{
			HolderName:  "Test McTest",
			Number:      "not a number",
			ExpiryMonth: time.March,
			ExpiryYear:  2027,
			CVV:         "123",
		}

		err := client.CreateCard(ctx, card)
		require.Error(t, err)
		assert.True(t, keeper.IsBadRequest(err))
	})
}

func TestClient_ListCards(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("error if not authenticated", func(t *testing.T) {
		_, err := client.ListCards(ctx)
		require.Error(t, err)
		assert.True(t, keeper.IsUnauthorized(err))
	})

	setupAccount(t, client)

	t.Run("lists no cards", func(t *testing.T) {
		cards, err := client.ListCards(ctx)
		require.NoError(t, err)
		assert.Len(t, cards, 0)
	})

	expected := keeper.Card{
		HolderName:  "Test McTest",
		Number:      "4111 1111 1111 1111",
		ExpiryMonth: time.March,
		ExpiryYear:  2027,
		CVV:         "123",
	}

	err := client.CreateCard(ctx, expected)
	require.NoError(t, err)

	t.Run("lists cards", func(t *testing.T) {
		cards, err := client.ListCards(ctx)
		require.NoError(t, err)
		if assert.Len(t, cards, 1) {
			actual := cards[0]
			assert.Equal(t, expected.CVV, actual.CVV)
			assert.Equal(t, expected.HolderName, actual.HolderName)
			assert.Equal(t, expected.Number, actual.Number)
			assert.Equal(t, expected.ExpiryMonth, actual.ExpiryMonth)
			assert.Equal(t, expected.ExpiryYear, actual.ExpiryYear)
		}
	})
}

func TestClient_DeleteCard(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("error if not authenticated", func(t *testing.T) {
		err := client.DeleteCard(ctx, uuid.NameSpaceDNS.String())
		require.Error(t, err)
		assert.True(t, keeper.IsUnauthorized(err))
	})

	setupAccount(t, client)

	err := client.CreateCard(ctx, keeper.Card{
		HolderName:  "Test McTest",
		Number:      "4111 1111 1111 1111",
		ExpiryMonth: time.March,
		ExpiryYear:  2027,
		CVV:         "123",
	})
	require.NoError(t, err)

	cards, err := client.ListCards(ctx)
	require.NoError(t, err)
	require.Len(t, cards, 1)

	t.Run("error if card does not exist", func(t *testing.T) {
		err = client.DeleteCard(ctx, uuid.NameSpaceDNS.String())
		require.Error(t, err)
		assert.True(t, keeper.IsNotFound(err))
	})

	t.Run("deletes card", func(t *testing.T) {
		err = client.DeleteCard(ctx, cards[0].ID)
		require.NoError(t, err)
	})

	t.Run("card does not exist", func(t *testing.T) {
		_, err = client.GetCard(ctx, cards[0].ID)
		require.Error(t, err)
		assert.True(t, keeper.IsNotFound(err))
	})
}

func TestClient_GetCard(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("error if not authenticated", func(t *testing.T) {
		_, err := client.GetCard(ctx, uuid.NameSpaceDNS.String())
		require.Error(t, err)
		assert.True(t, keeper.IsUnauthorized(err))
	})

	setupAccount(t, client)

	err := client.CreateCard(ctx, keeper.Card{
		HolderName:  "Test McTest",
		Number:      "4111 1111 1111 1111",
		ExpiryMonth: time.March,
		ExpiryYear:  2027,
		CVV:         "123",
	})
	require.NoError(t, err)

	cards, err := client.ListCards(ctx)
	require.NoError(t, err)
	require.Len(t, cards, 1)
	expected := cards[0]
	
	t.Run("error if card does not exist", func(t *testing.T) {
		_, err = client.GetCard(ctx, uuid.NameSpaceDNS.String())
		require.Error(t, err)
		assert.True(t, keeper.IsNotFound(err))
	})
	
	t.Run("gets card", func(t *testing.T) {
		actual, err := client.GetCard(ctx, expected.ID)
		require.NoError(t, err)
		
		assert.Equal(t, expected.CVV, actual.CVV)
		assert.Equal(t, expected.HolderName, actual.HolderName)
		assert.Equal(t, expected.Number, actual.Number)
		assert.Equal(t, expected.ExpiryMonth, actual.ExpiryMonth)
		assert.Equal(t, expected.ExpiryYear, actual.ExpiryYear)
	})
}
