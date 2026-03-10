package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/davidsbond/keeper/pkg/keeper"
)

func TestClient_CreateAccount(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip()
		return
	}

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
	t.Parallel()

	if testing.Short() {
		t.Skip()
		return
	}

	client := setupTest(t)
	ctx := t.Context()

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

	t.Run("error if not authenticated", func(t *testing.T) {
		_, err = client.GetAccount(ctx)
		require.Error(t, err)
		assert.True(t, keeper.IsUnauthorized(err))
	})

	err = client.Login(ctx, email, password)
	require.NoError(t, err)

	t.Run("gets current account", func(t *testing.T) {
		account, err := client.GetAccount(ctx)
		require.NoError(t, err)
		assert.Equal(t, email, account.Email)
		assert.Equal(t, displayName, account.DisplayName)
	})
}

func TestClient_ChangePassword(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip()
		return
	}

	client := setupTest(t)
	ctx := t.Context()

	const (
		email       = "test@test.com"
		password    = "test"
		newPassword = "test2"
		displayName = "Test McTest"
	)

	err := client.CreateAccount(ctx, keeper.Account{
		Email:       email,
		DisplayName: displayName,
		Password:    password,
	})
	require.NoError(t, err)

	t.Run("error if not authenticated", func(t *testing.T) {
		err = client.ChangePassword(ctx, email, newPassword)
		require.Error(t, err)
		assert.True(t, keeper.IsUnauthorized(err))
	})

	err = client.Login(ctx, email, password)
	require.NoError(t, err)

	t.Run("changes password", func(t *testing.T) {
		err = client.ChangePassword(ctx, password, newPassword)
		require.NoError(t, err)
	})

	t.Run("is now unauthenticated", func(t *testing.T) {
		_, err = client.ListLogins(ctx, "")
		require.Error(t, err)
		assert.True(t, keeper.IsUnauthorized(err))
	})

	t.Run("authenticates with new password", func(t *testing.T) {
		err = client.Login(ctx, email, newPassword)
		require.NoError(t, err)
	})
}

func TestClient_DeleteAccount(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip()
		return
	}

	client := setupTest(t)
	ctx := t.Context()

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

	t.Run("error if not authenticated", func(t *testing.T) {
		err = client.DeleteAccount(ctx)
		require.Error(t, err)
		assert.True(t, keeper.IsUnauthorized(err))
	})

	err = client.Login(ctx, email, password)
	require.NoError(t, err)

	t.Run("deletes account", func(t *testing.T) {
		err = client.DeleteAccount(ctx)
		require.NoError(t, err)
	})

	t.Run("can no longer authenticate", func(t *testing.T) {
		err = client.Login(ctx, email, password)
		require.Error(t, err)
		assert.True(t, keeper.IsNotFound(err))
	})
}
