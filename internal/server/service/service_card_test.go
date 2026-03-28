package service_test

import (
	"io"
	"testing"
	"time"

	"github.com/davidsbond/x/filter"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/davidsbond/keeper/internal/server/database"
	"github.com/davidsbond/keeper/internal/server/service"
)

func TestCardService_Create(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		UserID       uuid.UUID
		Card         service.Card
		ExpectsError bool
		Setup        func(cards *MockCardRepository, provider *MockRepositoryProvider[service.CardRepository])
	}{
		{
			Name:         "error if database lifetime has expired",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Card: service.Card{
				ID:          uuid.NameSpaceDNS,
				HolderName:  "Test McTest",
				Number:      "000000000000",
				ExpiryMonth: time.January,
				ExpiryYear:  2025,
				CVV:         "123",
			},
			Setup: func(cards *MockCardRepository, provider *MockRepositoryProvider[service.CardRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(nil, database.ErrClosed).Once()
			},
		},
		{
			Name:         "error when getting user database fails",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Card: service.Card{
				ID:          uuid.NameSpaceDNS,
				HolderName:  "Test McTest",
				Number:      "000000000000",
				ExpiryMonth: time.January,
				ExpiryYear:  2025,
				CVV:         "123",
			},
			Setup: func(cards *MockCardRepository, provider *MockRepositoryProvider[service.CardRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(nil, io.EOF).Once()
			},
		},
		{
			Name:         "error when lifetime has expired when creating",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Card: service.Card{
				ID:          uuid.NameSpaceDNS,
				HolderName:  "Test McTest",
				Number:      "000000000000",
				ExpiryMonth: time.January,
				ExpiryYear:  2025,
				CVV:         "123",
			},
			Setup: func(cards *MockCardRepository, provider *MockRepositoryProvider[service.CardRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(cards, nil).Once()

				cards.EXPECT().Create(mock.Anything).Return(database.ErrClosed).Once()
			},
		},
		{
			Name:         "error when creating record",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Card: service.Card{
				ID:          uuid.NameSpaceDNS,
				HolderName:  "Test McTest",
				Number:      "000000000000",
				ExpiryMonth: time.January,
				ExpiryYear:  2025,
				CVV:         "123",
			},
			Setup: func(cards *MockCardRepository, provider *MockRepositoryProvider[service.CardRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(cards, nil).Once()

				cards.EXPECT().Create(mock.Anything).Return(io.EOF).Once()
			},
		},
		{
			Name:   "success",
			UserID: uuid.NameSpaceDNS,
			Card: service.Card{
				ID:          uuid.NameSpaceDNS,
				HolderName:  "Test McTest",
				Number:      "000000000000",
				ExpiryMonth: time.January,
				ExpiryYear:  2025,
				CVV:         "123",
			},
			Setup: func(cards *MockCardRepository, provider *MockRepositoryProvider[service.CardRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(cards, nil).Once()

				cards.EXPECT().Create(mock.Anything).Return(nil).Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			cards := NewMockCardRepository(t)
			provider := NewMockRepositoryProvider[service.CardRepository](t)

			if tc.Setup != nil {
				tc.Setup(cards, provider)
			}

			svc := service.NewCardService(provider)
			err := svc.Create(tc.UserID, tc.Card)

			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestCardService_List(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		UserID       uuid.UUID
		Expected     []service.Card
		ExpectsError bool
		Setup        func(cards *MockCardRepository, provider *MockRepositoryProvider[service.CardRepository])
		Filters      []filter.Filter[service.Card]
	}{
		{
			Name:         "error if database lifetime has expired",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(cards *MockCardRepository, provider *MockRepositoryProvider[service.CardRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(nil, database.ErrClosed).Once()
			},
		},
		{
			Name:         "error when getting user database fails",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(cards *MockCardRepository, provider *MockRepositoryProvider[service.CardRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(nil, io.EOF).Once()
			},
		},
		{
			Name:         "error if database lifetime expired has on list",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(cards *MockCardRepository, provider *MockRepositoryProvider[service.CardRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(cards, nil).Once()

				cards.EXPECT().List().Return(nil, database.ErrClosed).Once()
			},
		},
		{
			Name:         "error when listing cards",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(cards *MockCardRepository, provider *MockRepositoryProvider[service.CardRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(cards, nil).Once()

				cards.EXPECT().List().Return(nil, io.EOF).Once()
			},
		},
		{
			Name:   "success",
			UserID: uuid.NameSpaceDNS,
			Expected: []service.Card{
				{
					ID:          uuid.NameSpaceDNS,
					HolderName:  "Test McTest",
					Number:      "000000000000",
					ExpiryMonth: time.January,
					ExpiryYear:  2025,
					CVV:         "123",
				},
			},
			Setup: func(cards *MockCardRepository, provider *MockRepositoryProvider[service.CardRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(cards, nil).Once()

				expected := []database.Card{
					{
						ID:          uuid.NameSpaceDNS,
						HolderName:  "Test McTest",
						Number:      "000000000000",
						ExpiryMonth: time.January,
						ExpiryYear:  2025,
						CVV:         "123",
					},
				}

				cards.EXPECT().List().Return(expected, nil).Once()
			},
		},
		{
			Name:   "uses filters",
			UserID: uuid.NameSpaceDNS,
			Expected: []service.Card{
				{
					ID:   uuid.NameSpaceDNS,
					Name: "test",
				},
			},
			Setup: func(cards *MockCardRepository, provider *MockRepositoryProvider[service.CardRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(cards, nil).Once()

				expected := []database.Card{
					{
						ID:   uuid.NameSpaceURL,
						Name: "abcde",
					},
					{
						ID:   uuid.NameSpaceDNS,
						Name: "test",
					},
				}

				cards.EXPECT().List().Return(expected, nil).Once()
			},
			Filters: []filter.Filter[service.Card]{
				service.CardsByName("test"),
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			cards := NewMockCardRepository(t)
			provider := NewMockRepositoryProvider[service.CardRepository](t)

			if tc.Setup != nil {
				tc.Setup(cards, provider)
			}

			actual, err := service.NewCardService(provider).List(tc.UserID, tc.Filters...)
			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.EqualValues(t, tc.Expected, actual)
		})
	}
}

func TestCardService_Delete(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		UserID       uuid.UUID
		CardID       uuid.UUID
		ExpectsError bool
		Setup        func(cards *MockCardRepository, provider *MockRepositoryProvider[service.CardRepository])
	}{
		{
			Name:         "error if database lifetime has expired",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(cards *MockCardRepository, provider *MockRepositoryProvider[service.CardRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(nil, database.ErrClosed).Once()
			},
		},
		{
			Name:         "error when getting user database fails",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(cards *MockCardRepository, provider *MockRepositoryProvider[service.CardRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(nil, io.EOF).Once()
			},
		},
		{
			Name:         "error if database lifetime expired has on delete",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			CardID:       uuid.NameSpaceDNS,
			Setup: func(cards *MockCardRepository, provider *MockRepositoryProvider[service.CardRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(cards, nil).Once()

				cards.EXPECT().Delete(uuid.NameSpaceDNS).Return(database.ErrClosed).Once()
			},
		},
		{
			Name:         "error when deleting card",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			CardID:       uuid.NameSpaceDNS,
			Setup: func(cards *MockCardRepository, provider *MockRepositoryProvider[service.CardRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(cards, nil).Once()

				cards.EXPECT().Delete(uuid.NameSpaceDNS).Return(io.EOF).Once()
			},
		},
		{
			Name:         "error if card does not exist",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			CardID:       uuid.NameSpaceDNS,
			Setup: func(cards *MockCardRepository, provider *MockRepositoryProvider[service.CardRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(cards, nil).Once()

				cards.EXPECT().Delete(uuid.NameSpaceDNS).Return(database.ErrCardNotFound).Once()
			},
		},
		{
			Name:   "success",
			UserID: uuid.NameSpaceDNS,
			CardID: uuid.NameSpaceDNS,
			Setup: func(cards *MockCardRepository, provider *MockRepositoryProvider[service.CardRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(cards, nil).Once()

				cards.EXPECT().Delete(uuid.NameSpaceDNS).Return(nil).Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			cards := NewMockCardRepository(t)
			provider := NewMockRepositoryProvider[service.CardRepository](t)

			if tc.Setup != nil {
				tc.Setup(cards, provider)
			}

			err := service.NewCardService(provider).Delete(tc.UserID, tc.CardID)
			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestCardService_Get(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		UserID       uuid.UUID
		CardID       uuid.UUID
		Expected     service.Card
		ExpectsError bool
		Setup        func(cards *MockCardRepository, provider *MockRepositoryProvider[service.CardRepository])
	}{
		{
			Name:         "error if database lifetime has expired",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(cards *MockCardRepository, provider *MockRepositoryProvider[service.CardRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(nil, database.ErrClosed).Once()
			},
		},
		{
			Name:         "error when getting user database fails",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			Setup: func(cards *MockCardRepository, provider *MockRepositoryProvider[service.CardRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(nil, io.EOF).Once()
			},
		},
		{
			Name:         "error if database lifetime expired has on delete",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			CardID:       uuid.NameSpaceDNS,
			Setup: func(cards *MockCardRepository, provider *MockRepositoryProvider[service.CardRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(cards, nil).Once()

				cards.EXPECT().
					Get(uuid.NameSpaceDNS).
					Return(database.Card{}, database.ErrClosed).Once()
			},
		},
		{
			Name:         "error when querying card",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			CardID:       uuid.NameSpaceDNS,
			Setup: func(cards *MockCardRepository, provider *MockRepositoryProvider[service.CardRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(cards, nil).Once()

				cards.EXPECT().
					Get(uuid.NameSpaceDNS).
					Return(database.Card{}, io.EOF).Once()
			},
		},
		{
			Name:         "error if card does not exist",
			ExpectsError: true,
			UserID:       uuid.NameSpaceDNS,
			CardID:       uuid.NameSpaceDNS,
			Setup: func(cards *MockCardRepository, provider *MockRepositoryProvider[service.CardRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(cards, nil).Once()

				cards.EXPECT().
					Get(uuid.NameSpaceDNS).
					Return(database.Card{}, database.ErrCardNotFound).Once()
			},
		},
		{
			Name:   "success",
			UserID: uuid.NameSpaceDNS,
			CardID: uuid.NameSpaceDNS,
			Expected: service.Card{
				ID:          uuid.NameSpaceDNS,
				HolderName:  "Test McTest",
				Number:      "000000000000",
				ExpiryMonth: time.January,
				ExpiryYear:  2025,
				CVV:         "123",
			},
			Setup: func(cards *MockCardRepository, provider *MockRepositoryProvider[service.CardRepository]) {
				provider.EXPECT().
					For(uuid.NameSpaceDNS).
					Return(cards, nil).Once()

				expected := database.Card{
					ID:          uuid.NameSpaceDNS,
					HolderName:  "Test McTest",
					Number:      "000000000000",
					ExpiryMonth: time.January,
					ExpiryYear:  2025,
					CVV:         "123",
				}

				cards.EXPECT().
					Get(uuid.NameSpaceDNS).
					Return(expected, nil).Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			cards := NewMockCardRepository(t)
			provider := NewMockRepositoryProvider[service.CardRepository](t)

			if tc.Setup != nil {
				tc.Setup(cards, provider)
			}

			actual, err := service.NewCardService(provider).Get(tc.UserID, tc.CardID)
			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.Expected, actual)
		})
	}
}
