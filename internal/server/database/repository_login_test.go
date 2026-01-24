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
		Password     database.Login
		ExpectsError bool
	}{
		{
			Name: "creates password",
			Password: database.Login{
				ID:       uuid.New(),
				Username: "test@test.com",
				Password: "password",
				Domains:  []string{"test.com"},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			err := database.NewLoginRepository(db).Create(tc.Password)
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
		Setup    func(passwords *database.LoginRepository)
	}{
		{
			Name: "lists passwords",
			Expected: []database.Login{
				{
					ID:       uuid.NameSpaceDNS,
					Username: "test@test.com",
					Password: "password",
					Domains:  []string{"test.com"},
				},
			},
			Setup: func(passwords *database.LoginRepository) {
				expected := database.Login{
					ID:       uuid.NameSpaceDNS,
					Username: "test@test.com",
					Password: "password",
					Domains:  []string{"test.com"},
				}

				require.NoError(t, passwords.Create(expected))
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			passwords := database.NewLoginRepository(db)
			if tc.Setup != nil {
				tc.Setup(passwords)
			}

			actual, err := passwords.List()
			require.NoError(t, err)
			assert.EqualValues(t, tc.Expected, actual)
		})
	}
}
