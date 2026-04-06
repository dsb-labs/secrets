package secrets_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsb-labs/secrets/pkg/secrets"
)

func TestClient_CreateCard(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("error if not authenticated", func(t *testing.T) {
		_, err := client.CreateCard(ctx, secrets.Card{})
		require.Error(t, err)
		assert.True(t, secrets.IsUnauthorized(err))
	})

	setupAccount(t, client)

	t.Run("creates card", func(t *testing.T) {
		card := secrets.Card{
			HolderName:  "Test McTest",
			Number:      "4111 1111 1111 1111",
			ExpiryMonth: time.March,
			ExpiryYear:  2027,
			CVV:         "123",
			Name:        "test",
		}

		id, err := client.CreateCard(ctx, card)
		require.NoError(t, err)
		assert.NoError(t, uuid.Validate(id))
	})

	t.Run("error if card is invalid", func(t *testing.T) {
		card := secrets.Card{
			HolderName:  "Test McTest",
			Number:      "not a number",
			ExpiryMonth: time.March,
			ExpiryYear:  2027,
			CVV:         "123",
		}

		_, err := client.CreateCard(ctx, card)
		require.Error(t, err)
		assert.True(t, secrets.IsBadRequest(err))
	})
}

func TestClient_ListCards(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("error if not authenticated", func(t *testing.T) {
		_, err := client.ListCards(ctx, secrets.CardListOptions{})
		require.Error(t, err)
		assert.True(t, secrets.IsUnauthorized(err))
	})

	setupAccount(t, client)

	t.Run("lists no cards", func(t *testing.T) {
		cards, err := client.ListCards(ctx, secrets.CardListOptions{})
		require.NoError(t, err)
		assert.Len(t, cards, 0)
	})

	expected := secrets.Card{
		HolderName:  "Test McTest",
		Number:      "4111 1111 1111 1111",
		ExpiryMonth: time.March,
		ExpiryYear:  2027,
		CVV:         "123",
		Name:        "test",
	}

	cardID, err := client.CreateCard(ctx, expected)
	require.NoError(t, err)

	t.Run("lists cards", func(t *testing.T) {
		cards, err := client.ListCards(ctx, secrets.CardListOptions{})
		require.NoError(t, err)
		if assert.Len(t, cards, 1) {
			actual := cards[0]
			assert.EqualValues(t, cardID, actual.ID)
			assert.EqualValues(t, expected.CVV, actual.CVV)
			assert.EqualValues(t, expected.HolderName, actual.HolderName)
			assert.EqualValues(t, expected.Number, actual.Number)
			assert.EqualValues(t, expected.ExpiryMonth, actual.ExpiryMonth)
			assert.EqualValues(t, expected.ExpiryYear, actual.ExpiryYear)
		}
	})
}

func TestClient_DeleteCard(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("error if not authenticated", func(t *testing.T) {
		err := client.DeleteCard(ctx, uuid.NameSpaceDNS.String())
		require.Error(t, err)
		assert.True(t, secrets.IsUnauthorized(err))
	})

	setupAccount(t, client)

	cardID, err := client.CreateCard(ctx, secrets.Card{
		HolderName:  "Test McTest",
		Number:      "4111 1111 1111 1111",
		ExpiryMonth: time.March,
		ExpiryYear:  2027,
		CVV:         "123",
		Name:        "test",
	})
	require.NoError(t, err)

	t.Run("error if card does not exist", func(t *testing.T) {
		err = client.DeleteCard(ctx, uuid.NameSpaceDNS.String())
		require.Error(t, err)
		assert.True(t, secrets.IsNotFound(err))
	})

	t.Run("deletes card", func(t *testing.T) {
		err = client.DeleteCard(ctx, cardID)
		require.NoError(t, err)
	})

	t.Run("card does not exist", func(t *testing.T) {
		_, err = client.GetCard(ctx, cardID)
		require.Error(t, err)
		assert.True(t, secrets.IsNotFound(err))
	})
}

func TestClient_GetCard(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("error if not authenticated", func(t *testing.T) {
		_, err := client.GetCard(ctx, uuid.NameSpaceDNS.String())
		require.Error(t, err)
		assert.True(t, secrets.IsUnauthorized(err))
	})

	setupAccount(t, client)

	expected := secrets.Card{
		HolderName:  "Test McTest",
		Number:      "4111 1111 1111 1111",
		ExpiryMonth: time.March,
		ExpiryYear:  2027,
		CVV:         "123",
		Name:        "test",
	}

	id, err := client.CreateCard(ctx, expected)
	require.NoError(t, err)

	t.Run("error if card does not exist", func(t *testing.T) {
		_, err = client.GetCard(ctx, uuid.NameSpaceDNS.String())
		require.Error(t, err)
		assert.True(t, secrets.IsNotFound(err))
	})

	t.Run("gets card", func(t *testing.T) {
		actual, err := client.GetCard(ctx, id)
		require.NoError(t, err)

		assert.EqualValues(t, id, actual.ID)
		assert.EqualValues(t, expected.CVV, actual.CVV)
		assert.EqualValues(t, expected.HolderName, actual.HolderName)
		assert.EqualValues(t, expected.Number, actual.Number)
		assert.EqualValues(t, expected.ExpiryMonth, actual.ExpiryMonth)
		assert.EqualValues(t, expected.ExpiryYear, actual.ExpiryYear)
	})
}
