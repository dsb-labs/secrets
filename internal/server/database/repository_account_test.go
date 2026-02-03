package database_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/davidsbond/keeper/internal/server/database"
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
				DisplayName:  "test",
			},
		},
		{
			Name:         "error if email is in use",
			ExpectsError: true,
			Account: database.Account{
				ID:           uuid.New(),
				Email:        "test@test.com",
				PasswordHash: []byte("hash"),
				DisplayName:  "test",
			},
		},
		{
			Name: "creates second account",
			Account: database.Account{
				ID:           uuid.New(),
				Email:        "test1@test.com",
				PasswordHash: []byte("hash"),
				DisplayName:  "test",
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

func TestAccountRepository_FindByEmail(t *testing.T) {
	t.Parallel()

	db := testDB(t)

	tt := []struct {
		Name         string
		Email        string
		ExpectsError bool
		Expected     database.Account
		Setup        func(accounts *database.AccountRepository)
	}{
		{
			Name:  "account exists",
			Email: "test@test.com",
			Setup: func(accounts *database.AccountRepository) {
				require.NoError(t, accounts.Create(database.Account{
					ID:           uuid.NameSpaceDNS,
					Email:        "test@test.com",
					PasswordHash: []byte("hash"),
					DisplayName:  "test",
				}))
			},
			Expected: database.Account{
				ID:           uuid.NameSpaceDNS,
				Email:        "test@test.com",
				PasswordHash: []byte("hash"),
				DisplayName:  "test",
			},
		},
		{
			Name:         "error if account does not exist",
			Email:        "test1@test.com",
			ExpectsError: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			accounts := database.NewAccountRepository(db)
			if tc.Setup != nil {
				tc.Setup(accounts)
			}

			actual, err := accounts.FindByEmail(tc.Email)
			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.EqualValues(t, tc.Expected, actual)
		})
	}
}

func TestAccountRepository_FindByID(t *testing.T) {
	t.Parallel()

	db := testDB(t)

	tt := []struct {
		Name         string
		ID           uuid.UUID
		ExpectsError bool
		Expected     database.Account
		Setup        func(accounts *database.AccountRepository)
	}{
		{
			Name: "account exists",
			ID:   uuid.NameSpaceDNS,
			Setup: func(accounts *database.AccountRepository) {
				require.NoError(t, accounts.Create(database.Account{
					ID:           uuid.NameSpaceDNS,
					Email:        "test@test.com",
					PasswordHash: []byte("hash"),
					DisplayName:  "test",
				}))
			},
			Expected: database.Account{
				ID:           uuid.NameSpaceDNS,
				Email:        "test@test.com",
				PasswordHash: []byte("hash"),
				DisplayName:  "test",
			},
		},
		{
			Name:         "error if account does not exist",
			ID:           uuid.NameSpaceURL,
			ExpectsError: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			accounts := database.NewAccountRepository(db)
			if tc.Setup != nil {
				tc.Setup(accounts)
			}

			actual, err := accounts.FindByID(tc.ID)
			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.EqualValues(t, tc.Expected, actual)
		})
	}
}

func TestAccountRepository_Delete(t *testing.T) {
	t.Parallel()

	db := testDB(t)

	tt := []struct {
		Name         string
		ID           uuid.UUID
		ExpectsError bool
		Setup        func(accounts *database.AccountRepository)
	}{
		{
			Name:         "account does not exist",
			ID:           uuid.NameSpaceDNS,
			ExpectsError: true,
		},
		{
			Name: "account exists",
			ID:   uuid.NameSpaceDNS,
			Setup: func(accounts *database.AccountRepository) {
				require.NoError(t, accounts.Create(database.Account{
					ID:           uuid.NameSpaceDNS,
					Email:        "test@test.com",
					PasswordHash: []byte("password"),
					DisplayName:  "Test McTest",
				}))
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			accounts := database.NewAccountRepository(db)
			if tc.Setup != nil {
				tc.Setup(accounts)
			}

			err := accounts.Delete(tc.ID)
			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
