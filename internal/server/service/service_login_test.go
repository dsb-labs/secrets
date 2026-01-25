package service_test

import (
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/davidsbond/passwords/internal/server/database"
	"github.com/davidsbond/passwords/internal/server/service"
)

func TestLoginService_Create(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		UserID       uuid.UUID
		Login        service.Login
		ExpectsError bool
		Setup        func(logins *MockLoginRepository, provider *MockRepositoryProvider[service.LoginRepository])
	}{
		{
			Name:         "error if database lifetime has expired",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Login: service.Login{
				Username: "test@test.com",
				Password: "test",
			},
			Setup: func(logins *MockLoginRepository, provider *MockRepositoryProvider[service.LoginRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(nil, database.ErrClosed).Once()
			},
		},
		{
			Name:         "error when getting user database fails",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Login: service.Login{
				Username: "test@test.com",
				Password: "test",
			},
			Setup: func(logins *MockLoginRepository, provider *MockRepositoryProvider[service.LoginRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(nil, io.EOF).Once()
			},
		},
		{
			Name:         "error when lifetime has expired when creating",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Login: service.Login{
				Username: "test@test.com",
				Password: "test",
			},
			Setup: func(logins *MockLoginRepository, provider *MockRepositoryProvider[service.LoginRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(logins, nil).Once()

				logins.EXPECT().Create(mock.Anything).Return(database.ErrClosed).Once()
			},
		},
		{
			Name:         "error when creating record",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Login: service.Login{
				Username: "test@test.com",
				Password: "test",
			},
			Setup: func(logins *MockLoginRepository, provider *MockRepositoryProvider[service.LoginRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(logins, nil).Once()

				logins.EXPECT().Create(mock.Anything).Return(io.EOF).Once()
			},
		},
		{
			Name:   "success",
			UserID: uuid.NameSpaceDNS,
			Login: service.Login{
				Username: "test@test.com",
				Password: "test",
			},
			Setup: func(logins *MockLoginRepository, provider *MockRepositoryProvider[service.LoginRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(logins, nil).Once()

				logins.EXPECT().Create(mock.Anything).Return(nil).Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			logins := NewMockLoginRepository(t)
			provider := NewMockRepositoryProvider[service.LoginRepository](t)

			if tc.Setup != nil {
				tc.Setup(logins, provider)
			}

			svc := service.NewLoginService(provider)
			err := svc.Create(tc.UserID, tc.Login)

			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestLoginService_List(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		UserID       uuid.UUID
		Expected     []service.Login
		ExpectsError bool
		Setup        func(logins *MockLoginRepository, provider *MockRepositoryProvider[service.LoginRepository])
	}{
		{
			Name:         "error if database lifetime has expired",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(logins *MockLoginRepository, provider *MockRepositoryProvider[service.LoginRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(nil, database.ErrClosed).Once()
			},
		},
		{
			Name:         "error when getting user database fails",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(logins *MockLoginRepository, provider *MockRepositoryProvider[service.LoginRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(nil, io.EOF).Once()
			},
		},
		{
			Name:         "error if database lifetime has on list",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(logins *MockLoginRepository, provider *MockRepositoryProvider[service.LoginRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(logins, nil).Once()

				logins.EXPECT().List().Return(nil, database.ErrClosed).Once()
			},
		},
		{
			Name:         "error when listing logins",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(logins *MockLoginRepository, provider *MockRepositoryProvider[service.LoginRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(logins, nil).Once()

				logins.EXPECT().List().Return(nil, io.EOF).Once()
			},
		},
		{
			Name:   "success",
			UserID: uuid.NameSpaceDNS,
			Expected: []service.Login{
				{
					Username: "test",
					Password: "test",
					Domains:  []string{"test"},
				},
			},
			Setup: func(logins *MockLoginRepository, provider *MockRepositoryProvider[service.LoginRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(logins, nil).Once()

				expected := []database.Login{
					{
						ID:       uuid.NameSpaceDNS,
						Username: "test",
						Password: "test",
						Domains:  []string{"test"},
					},
				}

				logins.EXPECT().List().Return(expected, nil).Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			logins := NewMockLoginRepository(t)
			provider := NewMockRepositoryProvider[service.LoginRepository](t)

			if tc.Setup != nil {
				tc.Setup(logins, provider)
			}

			actual, err := service.NewLoginService(provider).List(tc.UserID)
			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.EqualValues(t, tc.Expected, actual)
		})
	}
}
