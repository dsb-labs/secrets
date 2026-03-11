package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/davidsbond/keeper/pkg/keeper"
)

func TestClient_CreateAccount(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("creates new account", func(t *testing.T) {
		err := client.CreateAccount(ctx, keeper.Account{
			Email:       "test@test.com",
			DisplayName: "Test McTest",
			Password:    "test",
		})

		require.NoError(t, err)
	})

	t.Run("error if account already exists", func(t *testing.T) {
		err := client.CreateAccount(ctx, keeper.Account{
			Email:       "test@test.com",
			DisplayName: "Test McTest",
			Password:    "test",
		})

		require.Error(t, err)
		assert.True(t, keeper.IsConflict(err))
	})
}

func TestClient_GetAccount(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("error if not authenticated", func(t *testing.T) {
		_, err := client.GetAccount(ctx)
		require.Error(t, err)
		assert.True(t, keeper.IsUnauthorized(err))
	})

	setupAccount(t, client)

	t.Run("gets current account", func(t *testing.T) {
		account, err := client.GetAccount(ctx)
		require.NoError(t, err)
		assert.Equal(t, "test@test.com", account.Email)
		assert.Equal(t, "Test McTest", account.DisplayName)
	})
}

func TestClient_ChangePassword(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("error if not authenticated", func(t *testing.T) {
		err := client.ChangePassword(ctx, "test", "test2")
		require.Error(t, err)
		assert.True(t, keeper.IsUnauthorized(err))
	})

	setupAccount(t, client)

	t.Run("changes password", func(t *testing.T) {
		err := client.ChangePassword(ctx, "test", "test2")
		require.NoError(t, err)
	})

	t.Run("is now unauthenticated", func(t *testing.T) {
		_, err := client.ListLogins(ctx, "")
		require.Error(t, err)
		assert.True(t, keeper.IsUnauthorized(err))
	})

	t.Run("authenticates with new password", func(t *testing.T) {
		err := client.Login(ctx, "test@test.com", "test2")
		require.NoError(t, err)
	})
}

func TestClient_DeleteAccount(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("error if not authenticated", func(t *testing.T) {
		err := client.DeleteAccount(ctx)
		require.Error(t, err)
		assert.True(t, keeper.IsUnauthorized(err))
	})

	setupAccount(t, client)

	t.Run("deletes account", func(t *testing.T) {
		err := client.DeleteAccount(ctx)
		require.NoError(t, err)
	})

	t.Run("can no longer authenticate", func(t *testing.T) {
		err := client.Login(ctx, "test@test.com", "test")
		require.Error(t, err)
		assert.True(t, keeper.IsNotFound(err))
	})
}
