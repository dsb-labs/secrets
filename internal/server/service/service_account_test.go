package service_test

import (
	"io"
	"testing"

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
