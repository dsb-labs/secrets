package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/davidsbond/keeper/pkg/keeper"
)

func TestClient_CreateCard(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip()
		return
	}

	client := setupTest(t)
	ctx := t.Context()

	t.Run("error if not authenticated", func(t *testing.T) {
		err := client.CreateCard(ctx, keeper.Card{})
		require.Error(t, err)
		assert.True(t, keeper.IsUnauthorized(err))
	})

	const (
		email       = "test@test.com"
		password    = "test"
		displayName = "Test McTest"
	)

	err := client.CreateAccount(ctx, keeper.Account{
		Email:       email,
		DisplayName: displayName,
		Password:    password,
	})
	require.NoError(t, err)

	err = client.Login(ctx, email, password)
	require.NoError(t, err)

	t.Run("creates card", func(t *testing.T) {
		card := keeper.Card{
			HolderName:  "Test McTest",
			Number:      "4111 1111 1111 1111",
			ExpiryMonth: time.March,
			ExpiryYear:  2027,
			CVV:         "123",
		}

		err = client.CreateCard(ctx, card)
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

		err = client.CreateCard(ctx, card)
		require.Error(t, err)
		assert.True(t, keeper.IsBadRequest(err))
	})
}
