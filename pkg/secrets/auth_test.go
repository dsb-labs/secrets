package secrets_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsb-labs/secrets/pkg/secrets"
)

func TestClient_Login(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("client should have no token", func(t *testing.T) {
		require.Empty(t, client.Token())
	})

	const (
		email       = "test@test.com"
		password    = "test"
		displayName = "Test McTest"
	)

	t.Run("error if account does not exist", func(t *testing.T) {
		err := client.Login(ctx, email, password)
		require.Error(t, err)
		assert.True(t, secrets.IsNotFound(err))
	})

	_, err := client.CreateAccount(ctx, secrets.Account{
		Email:       email,
		DisplayName: displayName,
		Password:    password,
	})
	require.NoError(t, err)

	t.Run("error if password is incorrect", func(t *testing.T) {
		err = client.Login(ctx, email, "nope")
		require.Error(t, err)
		assert.True(t, secrets.IsBadRequest(err))
	})

	t.Run("sets token on success", func(t *testing.T) {
		err = client.Login(ctx, email, password)
		require.NoError(t, err)
		assert.NotEmpty(t, client.Token())
	})
}

func TestClient_Logout(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("error if not authenticated", func(t *testing.T) {
		err := client.Logout(ctx)
		require.Error(t, err)
		assert.True(t, secrets.IsUnauthorized(err))
	})

	setupAccount(t, client)

	t.Run("clears token on success", func(t *testing.T) {
		err := client.Logout(ctx)
		require.NoError(t, err)
		assert.Empty(t, client.Token())
	})

	t.Run("is deauthenticated", func(t *testing.T) {
		_, err := client.ListLogins(ctx, secrets.LoginListOptions{})
		require.Error(t, err)
		assert.True(t, secrets.IsUnauthorized(err))
	})
}
