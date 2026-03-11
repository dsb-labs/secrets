package keeper_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/davidsbond/keeper/pkg/keeper"
)

func TestClient_CreateLogin(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("error if not authenticated", func(t *testing.T) {
		_, err := client.CreateLogin(ctx, keeper.Login{})
		require.Error(t, err)
		assert.True(t, keeper.IsUnauthorized(err))
	})

	setupAccount(t, client)

	t.Run("creates login", func(t *testing.T) {
		login := keeper.Login{
			Username: "test",
			Password: "test",
			Domains:  []string{"test.com"},
		}

		id, err := client.CreateLogin(ctx, login)
		require.NoError(t, err)
		require.NoError(t, uuid.Validate(id))
	})

	t.Run("error if login is invalid", func(t *testing.T) {
		login := keeper.Login{}
		_, err := client.CreateLogin(ctx, login)
		require.Error(t, err)
		assert.True(t, keeper.IsBadRequest(err))
	})
}

func TestClient_ListLogins(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("error if not authenticated", func(t *testing.T) {
		_, err := client.ListLogins(ctx, "")
		require.Error(t, err)
		assert.True(t, keeper.IsUnauthorized(err))
	})

	setupAccount(t, client)

	t.Run("lists no logins", func(t *testing.T) {
		logins, err := client.ListLogins(ctx, "")
		require.NoError(t, err)
		assert.Len(t, logins, 0)
	})

	expected := keeper.Login{
		Username: "test",
		Password: "test",
		Domains:  []string{"test.com"},
	}

	loginID, err := client.CreateLogin(ctx, expected)
	require.NoError(t, err)

	t.Run("lists logins", func(t *testing.T) {
		logins, err := client.ListLogins(ctx, "")
		require.NoError(t, err)
		if assert.Len(t, logins, 1) {
			actual := logins[0]
			assert.EqualValues(t, loginID, actual.ID)
			assert.EqualValues(t, expected.Username, actual.Username)
			assert.EqualValues(t, expected.Password, actual.Password)
			assert.EqualValues(t, expected.Domains, actual.Domains)
		}
	})

	t.Run("lists logins by domain", func(t *testing.T) {
		logins, err := client.ListLogins(ctx, "test.com")
		require.NoError(t, err)
		if assert.Len(t, logins, 1) {
			actual := logins[0]
			assert.EqualValues(t, loginID, actual.ID)
			assert.EqualValues(t, expected.Username, actual.Username)
			assert.EqualValues(t, expected.Password, actual.Password)
			assert.EqualValues(t, expected.Domains, actual.Domains)
		}
	})
}

func TestClient_DeleteLogin(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("error if not authenticated", func(t *testing.T) {
		err := client.DeleteLogin(ctx, uuid.NameSpaceDNS.String())
		require.Error(t, err)
		assert.True(t, keeper.IsUnauthorized(err))
	})

	setupAccount(t, client)

	loginID, err := client.CreateLogin(ctx, keeper.Login{
		Username: "test",
		Password: "test",
		Domains:  []string{"test.com"},
	})
	require.NoError(t, err)

	t.Run("error if login does not exist", func(t *testing.T) {
		err = client.DeleteLogin(ctx, uuid.NameSpaceDNS.String())
		require.Error(t, err)
		assert.True(t, keeper.IsNotFound(err))
	})

	t.Run("deletes login", func(t *testing.T) {
		err = client.DeleteLogin(ctx, loginID)
		require.NoError(t, err)
	})

	t.Run("login does not exist", func(t *testing.T) {
		_, err = client.GetLogin(ctx, loginID)
		require.Error(t, err)
		assert.True(t, keeper.IsNotFound(err))
	})
}

func TestClient_GetLogin(t *testing.T) {
	client := setupTest(t)
	ctx := t.Context()

	t.Run("error if not authenticated", func(t *testing.T) {
		_, err := client.GetLogin(ctx, uuid.NameSpaceDNS.String())
		require.Error(t, err)
		assert.True(t, keeper.IsUnauthorized(err))
	})

	setupAccount(t, client)

	expected := keeper.Login{
		Username: "test",
		Password: "test",
		Domains:  []string{"test.com"},
	}

	loginID, err := client.CreateLogin(ctx, expected)
	require.NoError(t, err)

	t.Run("error if login does not exist", func(t *testing.T) {
		_, err = client.GetLogin(ctx, uuid.NameSpaceDNS.String())
		require.Error(t, err)
		assert.True(t, keeper.IsNotFound(err))
	})

	t.Run("gets login", func(t *testing.T) {
		actual, err := client.GetLogin(ctx, loginID)
		require.NoError(t, err)

		assert.EqualValues(t, loginID, actual.ID)
		assert.EqualValues(t, expected.Username, actual.Username)
		assert.EqualValues(t, expected.Password, actual.Password)
		assert.EqualValues(t, expected.Domains, actual.Domains)
	})
}
