package database_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/davidsbond/passwords/internal/server/database"
)

func TestLoginRepository_Create(t *testing.T) {
	t.Parallel()

	db := testDB(t)

	tt := []struct {
		Name         string
		Login        database.Login
		ExpectsError bool
	}{
		{
			Name: "creates login",
			Login: database.Login{
				ID:       uuid.New(),
				Username: "test@test.com",
				Password: "password",
				Domains:  []string{"test.com"},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			err := database.NewLoginRepository(db).Create(tc.Login)
			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestLoginRepository_List(t *testing.T) {
	t.Parallel()

	db := testDB(t)

	tt := []struct {
		Name     string
		Expected []database.Login
		Setup    func(logins *database.LoginRepository)
	}{
		{
			Name: "lists logins",
			Expected: []database.Login{
				{
					ID:       uuid.NameSpaceDNS,
					Username: "test@test.com",
					Password: "password",
					Domains:  []string{"test.com"},
				},
			},
			Setup: func(logins *database.LoginRepository) {
				expected := database.Login{
					ID:       uuid.NameSpaceDNS,
					Username: "test@test.com",
					Password: "password",
					Domains:  []string{"test.com"},
				}

				require.NoError(t, logins.Create(expected))
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			logins := database.NewLoginRepository(db)
			if tc.Setup != nil {
				tc.Setup(logins)
			}

			actual, err := logins.List()
			require.NoError(t, err)
			assert.EqualValues(t, tc.Expected, actual)
		})
	}
}

func TestLoginRepository_Delete(t *testing.T) {
	t.Parallel()

	db := testDB(t)

	tt := []struct {
		Name         string
		ID           uuid.UUID
		ExpectsError bool
		Setup        func(logins *database.LoginRepository)
	}{
		{
			Name: "deletes login",
			ID:   uuid.NameSpaceDNS,
			Setup: func(logins *database.LoginRepository) {
				expected := database.Login{
					ID:       uuid.NameSpaceDNS,
					Username: "test@test.com",
					Password: "password",
					Domains:  []string{"test.com"},
				}

				require.NoError(t, logins.Create(expected))
			},
		},
		{
			Name:         "error if login does not exist",
			ID:           uuid.NameSpaceDNS,
			ExpectsError: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			logins := database.NewLoginRepository(db)
			if tc.Setup != nil {
				tc.Setup(logins)
			}

			err := logins.Delete(tc.ID)
			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestLoginRepository_Get(t *testing.T) {
	t.Parallel()

	db := testDB(t)

	tt := []struct {
		Name         string
		ID           uuid.UUID
		ExpectsError bool
		Expected     database.Login
		Setup        func(logins *database.LoginRepository)
	}{
		{
			Name: "gets login",
			ID:   uuid.NameSpaceDNS,
			Expected: database.Login{
				ID:       uuid.NameSpaceDNS,
				Username: "test@test.com",
				Password: "password",
				Domains:  []string{"test.com"},
			},
			Setup: func(logins *database.LoginRepository) {
				expected := database.Login{
					ID:       uuid.NameSpaceDNS,
					Username: "test@test.com",
					Password: "password",
					Domains:  []string{"test.com"},
				}

				require.NoError(t, logins.Create(expected))
			},
		},
		{
			Name:         "error if login does not exist",
			ID:           uuid.NameSpaceURL,
			ExpectsError: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			logins := database.NewLoginRepository(db)
			if tc.Setup != nil {
				tc.Setup(logins)
			}

			actual, err := logins.Get(tc.ID)
			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.EqualValues(t, tc.Expected, actual)
		})
	}
}
