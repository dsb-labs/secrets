package service_test

import (
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/davidsbond/keeper/internal/server/database"
	"github.com/davidsbond/keeper/internal/server/service"
	"github.com/davidsbond/keeper/internal/server/token"
)

func TestAuthService_Login(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		Email        string
		Password     string
		ExpectsError bool
		Setup        func(accounts *MockAccountRepository, databases *MockDatabaseManager, tokens *MockTokenGenerator)
	}{
		{
			Name:         "error if account does not exist",
			ExpectsError: true,
			Email:        "test@test.com",
			Password:     "test",
			Setup: func(accounts *MockAccountRepository, _ *MockDatabaseManager, _ *MockTokenGenerator) {
				accounts.EXPECT().
					FindByEmail("test@test.com").
					Return(database.Account{}, database.ErrAccountNotFound).Once()
			},
		},
		{
			Name:         "error looking up account",
			ExpectsError: true,
			Email:        "test@test.com",
			Password:     "test",
			Setup: func(accounts *MockAccountRepository, _ *MockDatabaseManager, _ *MockTokenGenerator) {
				accounts.EXPECT().
					FindByEmail("test@test.com").
					Return(database.Account{}, io.EOF).Once()
			},
		},
		{
			Name:         "error comparing password",
			ExpectsError: true,
			Email:        "test@test.com",
			Password:     "test",
			Setup: func(accounts *MockAccountRepository, _ *MockDatabaseManager, _ *MockTokenGenerator) {
				accounts.EXPECT().
					FindByEmail("test@test.com").
					Return(database.Account{PasswordHash: []byte("something")}, nil).Once()
			},
		},
		{
			Name:         "error unlocking database",
			ExpectsError: true,
			Email:        "test@test.com",
			Password:     "test",
			Setup: func(accounts *MockAccountRepository, databases *MockDatabaseManager, _ *MockTokenGenerator) {
				accounts.EXPECT().
					FindByEmail("test@test.com").
					Return(database.Account{
						ID:           uuid.NameSpaceDNS,
						PasswordHash: mustBcrypt(t, "test"),
					}, nil).Once()

				databases.EXPECT().
					Unlock(uuid.NameSpaceDNS, mock.Anything).
					Return(io.EOF).Once()
			},
		},
		{
			Name:         "error generating token",
			ExpectsError: true,
			Email:        "test@test.com",
			Password:     "test",
			Setup: func(accounts *MockAccountRepository, databases *MockDatabaseManager, tokens *MockTokenGenerator) {
				accounts.EXPECT().
					FindByEmail("test@test.com").
					Return(database.Account{
						ID:           uuid.NameSpaceDNS,
						PasswordHash: mustBcrypt(t, "test"),
					}, nil).Once()

				databases.EXPECT().
					Unlock(uuid.NameSpaceDNS, mock.Anything).
					Return(nil).Once()

				tokens.EXPECT().Generate(uuid.NameSpaceDNS).Return(token.Token{}, io.EOF).Once()
			},
		},
		{
			Name:     "success",
			Email:    "test@test.com",
			Password: "test",
			Setup: func(accounts *MockAccountRepository, databases *MockDatabaseManager, tokens *MockTokenGenerator) {
				accounts.EXPECT().
					FindByEmail("test@test.com").
					Return(database.Account{
						ID:           uuid.NameSpaceDNS,
						PasswordHash: mustBcrypt(t, "test"),
					}, nil).Once()

				databases.EXPECT().
					Unlock(uuid.NameSpaceDNS, mock.Anything).
					Return(nil).Once()

				tokens.EXPECT().Generate(uuid.NameSpaceDNS).Return(token.TestToken(t, "test"), nil).Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			accounts := NewMockAccountRepository(t)
			databases := NewMockDatabaseManager(t)
			tokens := NewMockTokenGenerator(t)

			if tc.Setup != nil {
				tc.Setup(accounts, databases, tokens)
			}

			svc := service.NewAuthService(accounts, databases, tokens)
			tkn, err := svc.Login(tc.Email, tc.Password)

			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.True(t, tkn.Valid())
		})
	}
}

func TestAuthService_Logout(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		UserID       uuid.UUID
		ExpectsError bool
		Setup        func(databases *MockDatabaseManager)
	}{
		{
			Name:         "error locking database",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(databases *MockDatabaseManager) {
				databases.EXPECT().
					Lock(uuid.NameSpaceDNS).
					Return(io.EOF).Once()
			},
		},
		{
			Name:   "success",
			UserID: uuid.NameSpaceDNS,
			Setup: func(databases *MockDatabaseManager) {
				databases.EXPECT().
					Lock(uuid.NameSpaceDNS).
					Return(nil).Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			databases := NewMockDatabaseManager(t)
			if tc.Setup != nil {
				tc.Setup(databases)
			}

			err := service.NewAuthService(nil, databases, nil).Logout(tc.UserID)
			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func mustBcrypt(t *testing.T, password string) []byte {
	t.Helper()

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	return bytes
}
