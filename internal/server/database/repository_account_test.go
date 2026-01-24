package database_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/davidsbond/passwords/internal/server/database"
)

func TestAccountRepository_Create(t *testing.T) {
	t.Parallel()

	db := testDB(t)

	tt := []struct {
		Name         string
		Account      database.Account
		ExpectsError bool
	}{
		{
			Name: "creates account",
			Account: database.Account{
				ID:           uuid.New(),
				Email:        "test@test.com",
				PasswordHash: []byte("hash"),
			},
		},
		{
			Name:         "error if email is in use",
			ExpectsError: true,
			Account: database.Account{
				ID:           uuid.New(),
				Email:        "test@test.com",
				PasswordHash: []byte("hash"),
			},
		},
		{
			Name: "creates second account",
			Account: database.Account{
				ID:           uuid.New(),
				Email:        "test1@test.com",
				PasswordHash: []byte("hash"),
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			err := database.NewAccountRepository(db).Create(tc.Account)
			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
