package secrets_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsb-labs/secrets/pkg/secrets"
)

func TestClient_CreateAccount(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("creates new account", func(t *testing.T) {
		restoreKey, err := client.CreateAccount(ctx, secrets.Account{
			Email:       "test@test.com",
			DisplayName: "Test McTest",
			Password:    "test",
		})

		require.NoError(t, err)
		require.NotNil(t, restoreKey)
	})

	t.Run("error if account already exists", func(t *testing.T) {
		_, err := client.CreateAccount(ctx, secrets.Account{
			Email:       "test@test.com",
			DisplayName: "Test McTest",
			Password:    "test",
		})

		require.Error(t, err)
		assert.True(t, secrets.IsConflict(err))
	})
}

func TestClient_GetAccount(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("error if not authenticated", func(t *testing.T) {
		_, err := client.GetAccount(ctx)
		require.Error(t, err)
		assert.True(t, secrets.IsUnauthorized(err))
	})

	setupAccount(t, client)

	t.Run("gets current account", func(t *testing.T) {
		account, err := client.GetAccount(ctx)
		require.NoError(t, err)
		assert.EqualValues(t, "test@test.com", account.Email)
		assert.EqualValues(t, "Test McTest", account.DisplayName)
	})
}

func TestClient_ChangePassword(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("error if not authenticated", func(t *testing.T) {
		_, err := client.ChangePassword(ctx, "test", "test2")
		require.Error(t, err)
		assert.True(t, secrets.IsUnauthorized(err))
	})

	setupAccount(t, client)

	t.Run("changes password", func(t *testing.T) {
		restoreKey, err := client.ChangePassword(ctx, "test", "test2")
		require.NoError(t, err)
		require.NotNil(t, restoreKey)
	})

	t.Run("is now unauthenticated", func(t *testing.T) {
		_, err := client.ListLogins(ctx, secrets.LoginListOptions{})
		require.Error(t, err)
		assert.True(t, secrets.IsUnauthorized(err))
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
		assert.True(t, secrets.IsUnauthorized(err))
	})

	setupAccount(t, client)

	t.Run("deletes account", func(t *testing.T) {
		err := client.DeleteAccount(ctx)
		require.NoError(t, err)
	})

	t.Run("can no longer authenticate", func(t *testing.T) {
		err := client.Login(ctx, "test@test.com", "test")
		require.Error(t, err)
		assert.True(t, secrets.IsNotFound(err))
	})
}

func TestClient_RestoreAccount(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	restoreKey := setupAccount(t, client)

	t.Run("error if not account does not exist", func(t *testing.T) {
		_, err := client.RestoreAccount(ctx, "nope@nope.com", restoreKey, "something")
		require.Error(t, err)
		assert.True(t, secrets.IsNotFound(err))
	})

	t.Run("error if restore key is invalid", func(t *testing.T) {
		_, err := client.RestoreAccount(ctx, "test@test.com", bytes.Repeat([]byte{0}, 32), "something")
		require.Error(t, err)
		assert.True(t, secrets.IsBadRequest(err))
	})

	t.Run("changes password", func(t *testing.T) {
		newKey, err := client.RestoreAccount(ctx, "test@test.com", restoreKey, "test2")
		require.NoError(t, err)
		require.NotNil(t, newKey)
	})

	t.Run("authenticates with new password", func(t *testing.T) {
		err := client.Login(ctx, "test@test.com", "test2")
		require.NoError(t, err)
	})
}
