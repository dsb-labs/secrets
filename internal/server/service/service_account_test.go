package service_test

import (
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/davidsbond/keeper/internal/server/database"
	"github.com/davidsbond/keeper/internal/server/service"
)

func TestAccountService_Create(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		Account      service.Account
		ExpectsError bool
		Setup        func(accounts *MockAccountRepository, databases *MockDatabaseManager)
	}{
		{
			Name:         "error if account already exists",
			ExpectsError: true,
			Account: service.Account{
				Email:       "test@test.com",
				Password:    "test",
				DisplayName: "test",
			},
			Setup: func(accounts *MockAccountRepository, databases *MockDatabaseManager) {
				accounts.EXPECT().Create(mock.Anything).Return(database.ErrAccountExists).Once()
			},
		},
		{
			Name:         "error if insertion fails",
			ExpectsError: true,
			Account: service.Account{
				Email:       "test@test.com",
				Password:    "test",
				DisplayName: "test",
			},
			Setup: func(accounts *MockAccountRepository, databases *MockDatabaseManager) {
				accounts.EXPECT().Create(mock.Anything).Return(io.EOF).Once()
			},
		},
		{
			Name: "returns restore key on success",
			Account: service.Account{
				Email:       "test@test.com",
				Password:    "test",
				DisplayName: "test",
			},
			Setup: func(accounts *MockAccountRepository, databases *MockDatabaseManager) {
				accounts.EXPECT().Create(mock.Anything).Return(nil).Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			accounts := NewMockAccountRepository(t)
			databases := NewMockDatabaseManager(t)

			if tc.Setup != nil {
				tc.Setup(accounts, databases)
			}

			svc := service.NewAccountService(accounts, databases)
			restoreKey, err := svc.Create(tc.Account)

			if tc.ExpectsError {
				require.Error(t, err)
				require.Nil(t, restoreKey)
				return
			}

			require.NoError(t, err)
			require.Len(t, restoreKey, 32)
		})
	}
}

func TestAccountService_Get(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		ID           uuid.UUID
		Expected     service.Account
		ExpectsError bool
		Setup        func(accounts *MockAccountRepository)
	}{
		{
			Name:         "error if account does not exist",
			ID:           uuid.NameSpaceDNS,
			ExpectsError: true,
			Setup: func(accounts *MockAccountRepository) {
				accounts.EXPECT().
					FindByID(uuid.NameSpaceDNS).
					Return(database.Account{}, database.ErrAccountNotFound).
					Once()
			},
		},
		{
			Name:         "error querying account",
			ID:           uuid.NameSpaceDNS,
			ExpectsError: true,
			Setup: func(accounts *MockAccountRepository) {
				accounts.EXPECT().
					FindByID(uuid.NameSpaceDNS).
					Return(database.Account{}, io.EOF).
					Once()
			},
		},
		{
			Name: "success",
			ID:   uuid.NameSpaceDNS,
			Expected: service.Account{
				Email:       "test@test.com",
				Password:    "REDACTED",
				DisplayName: "Test McTest",
			},
			Setup: func(accounts *MockAccountRepository) {
				expected := database.Account{
					ID:           uuid.NameSpaceDNS,
					Email:        "test@test.com",
					PasswordHash: []byte("secret-password"),
					DisplayName:  "Test McTest",
				}

				accounts.EXPECT().
					FindByID(uuid.NameSpaceDNS).
					Return(expected, nil).
					Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			accounts := NewMockAccountRepository(t)

			if tc.Setup != nil {
				tc.Setup(accounts)
			}

			svc := service.NewAccountService(accounts, nil)
			actual, err := svc.Get(tc.ID)

			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.EqualValues(t, tc.Expected, actual)
		})
	}
}
