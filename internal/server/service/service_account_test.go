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

func TestAccountService_Delete(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		ID           uuid.UUID
		ExpectsError bool
		Setup        func(accounts *MockAccountRepository, databases *MockDatabaseManager)
	}{
		{
			Name:         "error if account does not exist",
			ID:           uuid.NameSpaceDNS,
			ExpectsError: true,
			Setup: func(accounts *MockAccountRepository, databases *MockDatabaseManager) {
				accounts.EXPECT().
					Delete(uuid.NameSpaceDNS).
					Return(database.ErrAccountNotFound).
					Once()
			},
		},
		{
			Name:         "error deleting account",
			ID:           uuid.NameSpaceDNS,
			ExpectsError: true,
			Setup: func(accounts *MockAccountRepository, databases *MockDatabaseManager) {
				accounts.EXPECT().
					Delete(uuid.NameSpaceDNS).
					Return(io.EOF).
					Once()
			},
		},
		{
			Name:         "error deleting database",
			ID:           uuid.NameSpaceDNS,
			ExpectsError: true,
			Setup: func(accounts *MockAccountRepository, databases *MockDatabaseManager) {
				accounts.EXPECT().
					Delete(uuid.NameSpaceDNS).
					Return(nil).
					Once()

				databases.EXPECT().
					Delete(uuid.NameSpaceDNS).
					Return(io.EOF).
					Once()
			},
		},
		{
			Name: "success",
			ID:   uuid.NameSpaceDNS,
			Setup: func(accounts *MockAccountRepository, databases *MockDatabaseManager) {
				accounts.EXPECT().
					Delete(uuid.NameSpaceDNS).
					Return(nil).
					Once()

				databases.EXPECT().
					Delete(uuid.NameSpaceDNS).
					Return(nil).
					Once()
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
			err := svc.Delete(tc.ID)
			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestAccountService_ChangePassword(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		ID           uuid.UUID
		OldPassword  string
		NewPassword  string
		ExpectsError bool
		Setup        func(accounts *MockAccountRepository, databases *MockDatabaseManager)
	}{
		{
			Name:         "error if account does not exist",
			ID:           uuid.NameSpaceDNS,
			ExpectsError: true,
			OldPassword:  "test",
			NewPassword:  "test1",
			Setup: func(accounts *MockAccountRepository, databases *MockDatabaseManager) {
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
			OldPassword:  "test",
			NewPassword:  "test1",
			Setup: func(accounts *MockAccountRepository, databases *MockDatabaseManager) {
				accounts.EXPECT().
					FindByID(uuid.NameSpaceDNS).
					Return(database.Account{}, io.EOF).
					Once()
			},
		},
		{
			Name:         "old password does not match",
			ID:           uuid.NameSpaceDNS,
			ExpectsError: true,
			OldPassword:  "test",
			NewPassword:  "test1",
			Setup: func(accounts *MockAccountRepository, databases *MockDatabaseManager) {
				account := database.Account{
					ID:           uuid.NameSpaceDNS,
					PasswordHash: mustBcrypt(t, "boop"),
				}

				accounts.EXPECT().
					FindByID(uuid.NameSpaceDNS).
					Return(account, nil).
					Once()
			},
		},
		{
			Name:         "account not found when updating",
			ID:           uuid.NameSpaceDNS,
			ExpectsError: true,
			OldPassword:  "test",
			NewPassword:  "test1",
			Setup: func(accounts *MockAccountRepository, databases *MockDatabaseManager) {
				account := database.Account{
					ID:           uuid.NameSpaceDNS,
					PasswordHash: mustBcrypt(t, "test"),
				}

				accounts.EXPECT().
					FindByID(uuid.NameSpaceDNS).
					Return(account, nil).
					Once()

				accounts.EXPECT().
					Update(mock.Anything).
					Return(database.ErrAccountNotFound).
					Once()
			},
		},
		{
			Name:         "error when updating account",
			ID:           uuid.NameSpaceDNS,
			ExpectsError: true,
			OldPassword:  "test",
			NewPassword:  "test1",
			Setup: func(accounts *MockAccountRepository, databases *MockDatabaseManager) {
				account := database.Account{
					ID:           uuid.NameSpaceDNS,
					PasswordHash: mustBcrypt(t, "test"),
				}

				accounts.EXPECT().
					FindByID(uuid.NameSpaceDNS).
					Return(account, nil).
					Once()

				accounts.EXPECT().
					Update(mock.Anything).
					Return(io.EOF).
					Once()
			},
		},
		{
			Name:         "error when rotating key",
			ID:           uuid.NameSpaceDNS,
			ExpectsError: true,
			OldPassword:  "test",
			NewPassword:  "test1",
			Setup: func(accounts *MockAccountRepository, databases *MockDatabaseManager) {
				account := database.Account{
					ID:           uuid.NameSpaceDNS,
					PasswordHash: mustBcrypt(t, "test"),
				}

				accounts.EXPECT().
					FindByID(uuid.NameSpaceDNS).
					Return(account, nil).
					Once()

				accounts.EXPECT().
					Update(mock.Anything).
					Return(nil).
					Once()

				databases.EXPECT().
					RotateKey(uuid.NameSpaceDNS, mock.Anything, mock.Anything).
					Return(io.EOF).
					Once()
			},
		},
		{
			Name:        "success",
			ID:          uuid.NameSpaceDNS,
			OldPassword: "test",
			NewPassword: "test1",
			Setup: func(accounts *MockAccountRepository, databases *MockDatabaseManager) {
				account := database.Account{
					ID:           uuid.NameSpaceDNS,
					PasswordHash: mustBcrypt(t, "test"),
				}

				accounts.EXPECT().
					FindByID(uuid.NameSpaceDNS).
					Return(account, nil).
					Once()

				accounts.EXPECT().
					Update(mock.Anything).
					Return(nil).
					Once()

				databases.EXPECT().
					RotateKey(uuid.NameSpaceDNS, mock.Anything, mock.Anything).
					Return(nil).
					Once()
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
			restoreKey, err := svc.ChangePassword(tc.ID, tc.OldPassword, tc.NewPassword)
			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, restoreKey)
		})
	}
}

func TestAccountService_Restore(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		Email        string
		RestoreKey   []byte
		NewPassword  string
		ExpectsError bool
		Setup        func(accounts *MockAccountRepository, databases *MockDatabaseManager)
	}{
		{
			Name:         "error if account does not exist",
			Email:        "test@test.com",
			ExpectsError: true,
			RestoreKey:   []byte("test"),
			NewPassword:  "test1",
			Setup: func(accounts *MockAccountRepository, databases *MockDatabaseManager) {
				accounts.EXPECT().
					FindByEmail("test@test.com").
					Return(database.Account{}, database.ErrAccountNotFound).
					Once()
			},
		},
		{
			Name:         "error querying account",
			Email:        "test@test.com",
			ExpectsError: true,
			RestoreKey:   []byte("test"),
			NewPassword:  "test1",
			Setup: func(accounts *MockAccountRepository, databases *MockDatabaseManager) {
				accounts.EXPECT().
					FindByEmail("test@test.com").
					Return(database.Account{}, io.EOF).
					Once()
			},
		},
		{
			Name:         "restore key does not match",
			Email:        "test@test.com",
			ExpectsError: true,
			RestoreKey:   []byte("test"),
			NewPassword:  "test1",
			Setup: func(accounts *MockAccountRepository, databases *MockDatabaseManager) {
				account := database.Account{
					ID:    uuid.NameSpaceDNS,
					Email: "test@test.com",
				}

				accounts.EXPECT().
					FindByEmail("test@test.com").
					Return(account, nil).
					Once()

				databases.EXPECT().
					RotateKey(uuid.NameSpaceDNS, []byte("test"), mock.Anything).
					Return(database.ErrInvalidKey).
					Once()
			},
		},
		{
			Name:         "error when rotating key",
			Email:        "test@test.com",
			ExpectsError: true,
			RestoreKey:   []byte("test"),
			NewPassword:  "test1",
			Setup: func(accounts *MockAccountRepository, databases *MockDatabaseManager) {
				account := database.Account{
					ID:    uuid.NameSpaceDNS,
					Email: "test@test.com",
				}

				accounts.EXPECT().
					FindByEmail("test@test.com").
					Return(account, nil).
					Once()

				databases.EXPECT().
					RotateKey(uuid.NameSpaceDNS, []byte("test"), mock.Anything).
					Return(io.EOF).
					Once()
			},
		},
		{
			Name:         "account not found when updating",
			Email:        "test@test.com",
			ExpectsError: true,
			RestoreKey:   []byte("test"),
			NewPassword:  "test1",
			Setup: func(accounts *MockAccountRepository, databases *MockDatabaseManager) {
				account := database.Account{
					ID:    uuid.NameSpaceDNS,
					Email: "test@test.com",
				}

				accounts.EXPECT().
					FindByEmail("test@test.com").
					Return(account, nil).
					Once()

				databases.EXPECT().
					RotateKey(uuid.NameSpaceDNS, []byte("test"), mock.Anything).
					Return(nil).
					Once()

				accounts.EXPECT().
					Update(mock.Anything).
					Return(database.ErrAccountNotFound).
					Once()
			},
		},
		{
			Name:         "account not found when updating",
			Email:        "test@test.com",
			ExpectsError: true,
			RestoreKey:   []byte("test"),
			NewPassword:  "test1",
			Setup: func(accounts *MockAccountRepository, databases *MockDatabaseManager) {
				account := database.Account{
					ID:    uuid.NameSpaceDNS,
					Email: "test@test.com",
				}

				accounts.EXPECT().
					FindByEmail("test@test.com").
					Return(account, nil).
					Once()

				databases.EXPECT().
					RotateKey(uuid.NameSpaceDNS, []byte("test"), mock.Anything).
					Return(nil).
					Once()

				accounts.EXPECT().
					Update(mock.Anything).
					Return(io.EOF).
					Once()
			},
		},
		{
			Name:         "success",
			Email:        "test@test.com",
			RestoreKey:   []byte("test"),
			NewPassword:  "test1",
			Setup: func(accounts *MockAccountRepository, databases *MockDatabaseManager) {
				account := database.Account{
					ID:    uuid.NameSpaceDNS,
					Email: "test@test.com",
				}

				accounts.EXPECT().
					FindByEmail("test@test.com").
					Return(account, nil).
					Once()

				databases.EXPECT().
					RotateKey(uuid.NameSpaceDNS, []byte("test"), mock.Anything).
					Return(nil).
					Once()

				accounts.EXPECT().
					Update(mock.Anything).
					Return(nil).
					Once()
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
			restoreKey, err := svc.Restore(tc.Email, tc.RestoreKey, tc.NewPassword)
			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, restoreKey)
		})
	}
}
