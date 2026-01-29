package service_test

import (
	"io"
	"testing"

	"github.com/davidsbond/x/filter"
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
		Filters      []filter.Filter[service.Login]
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
			Name:         "error if database lifetime expired has on list",
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
					ID:       uuid.NameSpaceDNS,
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
		{
			Name:   "uses filters",
			UserID: uuid.NameSpaceDNS,
			Expected: []service.Login{
				{
					ID:       uuid.NameSpaceDNS,
					Username: "test@test.com",
					Password: "test",
					Domains:  []string{"https://account.google.com"},
				},
			},
			Setup: func(logins *MockLoginRepository, provider *MockRepositoryProvider[service.LoginRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(logins, nil).Once()

				expected := []database.Login{
					{
						ID:       uuid.NameSpaceURL,
						Username: "test@test.com",
						Password: "test",
						Domains:  []string{"https://facebook.com"},
					},
					{
						ID:       uuid.NameSpaceDNS,
						Username: "test@test.com",
						Password: "test",
						Domains:  []string{"https://account.google.com"},
					},
				}

				logins.EXPECT().List().Return(expected, nil).Once()
			},
			Filters: []filter.Filter[service.Login]{
				service.LoginByDomain("google.com"),
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

			actual, err := service.NewLoginService(provider).List(tc.UserID, tc.Filters...)
			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.EqualValues(t, tc.Expected, actual)
		})
	}
}

func TestLoginService_Delete(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		UserID       uuid.UUID
		LoginID      uuid.UUID
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
			Name:         "error if database lifetime expired has on delete",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			LoginID:      uuid.NameSpaceDNS,
			Setup: func(logins *MockLoginRepository, provider *MockRepositoryProvider[service.LoginRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(logins, nil).Once()

				logins.EXPECT().Delete(uuid.NameSpaceDNS).Return(database.ErrClosed).Once()
			},
		},
		{
			Name:         "error when deleting login",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			LoginID:      uuid.NameSpaceDNS,
			Setup: func(logins *MockLoginRepository, provider *MockRepositoryProvider[service.LoginRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(logins, nil).Once()

				logins.EXPECT().Delete(uuid.NameSpaceDNS).Return(io.EOF).Once()
			},
		},
		{
			Name:         "error if login does not exist",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			LoginID:      uuid.NameSpaceDNS,
			Setup: func(logins *MockLoginRepository, provider *MockRepositoryProvider[service.LoginRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(logins, nil).Once()

				logins.EXPECT().Delete(uuid.NameSpaceDNS).Return(database.ErrLoginNotFound).Once()
			},
		},
		{
			Name:    "success",
			UserID:  uuid.NameSpaceDNS,
			LoginID: uuid.NameSpaceDNS,
			Setup: func(logins *MockLoginRepository, provider *MockRepositoryProvider[service.LoginRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(logins, nil).Once()

				logins.EXPECT().Delete(uuid.NameSpaceDNS).Return(nil).Once()
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

			err := service.NewLoginService(provider).Delete(tc.UserID, tc.LoginID)
			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestLoginService_Get(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		UserID       uuid.UUID
		LoginID      uuid.UUID
		Expected     service.Login
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
			Name:         "error if database lifetime expired has on delete",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			LoginID:      uuid.NameSpaceDNS,
			Setup: func(logins *MockLoginRepository, provider *MockRepositoryProvider[service.LoginRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(logins, nil).Once()

				logins.EXPECT().
					Get(uuid.NameSpaceDNS).
					Return(database.Login{}, database.ErrClosed).Once()
			},
		},
		{
			Name:         "error when querying login",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			LoginID:      uuid.NameSpaceDNS,
			Setup: func(logins *MockLoginRepository, provider *MockRepositoryProvider[service.LoginRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(logins, nil).Once()

				logins.EXPECT().
					Get(uuid.NameSpaceDNS).
					Return(database.Login{}, io.EOF).Once()
			},
		},
		{
			Name:         "error if login does not exist",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			LoginID:      uuid.NameSpaceDNS,
			Setup: func(logins *MockLoginRepository, provider *MockRepositoryProvider[service.LoginRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(logins, nil).Once()

				logins.EXPECT().
					Get(uuid.NameSpaceDNS).
					Return(database.Login{}, database.ErrLoginNotFound).Once()
			},
		},
		{
			Name:    "success",
			UserID:  uuid.NameSpaceDNS,
			LoginID: uuid.NameSpaceDNS,
			Expected: service.Login{
				ID:       uuid.NameSpaceDNS,
				Username: "test",
				Password: "test",
				Domains:  []string{"test"},
			},
			Setup: func(logins *MockLoginRepository, provider *MockRepositoryProvider[service.LoginRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(logins, nil).Once()

				expected := database.Login{
					ID:       uuid.NameSpaceDNS,
					Username: "test",
					Password: "test",
					Domains:  []string{"test"},
				}

				logins.EXPECT().
					Get(uuid.NameSpaceDNS).
					Return(expected, nil).Once()
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

			actual, err := service.NewLoginService(provider).Get(tc.UserID, tc.LoginID)
			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.Expected, actual)
		})
	}
}
