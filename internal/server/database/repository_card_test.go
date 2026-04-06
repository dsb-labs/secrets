package database_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsb-labs/secrets/internal/server/database"
)

func TestCardRepository_Create(t *testing.T) {
	t.Parallel()

	db := testDB(t)

	tt := []struct {
		Name         string
		Card         database.Card
		ExpectsError bool
	}{
		{
			Name: "creates card",
			Card: database.Card{
				ID:          uuid.New(),
				HolderName:  "Test McTest",
				Number:      "000000000000",
				ExpiryMonth: time.January,
				ExpiryYear:  2025,
				CVV:         "123",
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			err := database.NewCardRepository(db).Create(tc.Card)
			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestCardRepository_List(t *testing.T) {
	t.Parallel()

	db := testDB(t)

	tt := []struct {
		Name     string
		Expected []database.Card
		Setup    func(cards *database.CardRepository)
	}{
		{
			Name: "lists cards",
			Expected: []database.Card{
				{
					ID:          uuid.NameSpaceDNS,
					HolderName:  "Test McTest",
					Number:      "000000000000",
					ExpiryMonth: time.January,
					ExpiryYear:  2025,
					CVV:         "123",
				},
			},
			Setup: func(cards *database.CardRepository) {
				expected := database.Card{
					ID:          uuid.NameSpaceDNS,
					HolderName:  "Test McTest",
					Number:      "000000000000",
					ExpiryMonth: time.January,
					ExpiryYear:  2025,
					CVV:         "123",
				}

				require.NoError(t, cards.Create(expected))
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			cards := database.NewCardRepository(db)
			if tc.Setup != nil {
				tc.Setup(cards)
			}

			actual, err := cards.List()
			require.NoError(t, err)
			assert.EqualValues(t, tc.Expected, actual)
		})
	}
}

func TestCardRepository_Delete(t *testing.T) {
	t.Parallel()

	db := testDB(t)

	tt := []struct {
		Name         string
		ID           uuid.UUID
		ExpectsError bool
		Setup        func(cards *database.CardRepository)
	}{
		{
			Name: "deletes card",
			ID:   uuid.NameSpaceDNS,
			Setup: func(cards *database.CardRepository) {
				expected := database.Card{
					ID:          uuid.NameSpaceDNS,
					HolderName:  "Test McTest",
					Number:      "000000000000",
					ExpiryMonth: time.January,
					ExpiryYear:  2025,
					CVV:         "123",
				}

				require.NoError(t, cards.Create(expected))
			},
		},
		{
			Name:         "error if card does not exist",
			ID:           uuid.NameSpaceDNS,
			ExpectsError: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			cards := database.NewCardRepository(db)
			if tc.Setup != nil {
				tc.Setup(cards)
			}

			err := cards.Delete(tc.ID)
			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestCardRepository_Get(t *testing.T) {
	t.Parallel()

	db := testDB(t)

	tt := []struct {
		Name         string
		ID           uuid.UUID
		ExpectsError bool
		Expected     database.Card
		Setup        func(cards *database.CardRepository)
	}{
		{
			Name: "gets card",
			ID:   uuid.NameSpaceDNS,
			Expected: database.Card{
				ID:          uuid.NameSpaceDNS,
				HolderName:  "Test McTest",
				Number:      "000000000000",
				ExpiryMonth: time.January,
				ExpiryYear:  2025,
				CVV:         "123",
			},
			Setup: func(cards *database.CardRepository) {
				expected := database.Card{
					ID:          uuid.NameSpaceDNS,
					HolderName:  "Test McTest",
					Number:      "000000000000",
					ExpiryMonth: time.January,
					ExpiryYear:  2025,
					CVV:         "123",
				}

				require.NoError(t, cards.Create(expected))
			},
		},
		{
			Name:         "error if card does not exist",
			ID:           uuid.NameSpaceURL,
			ExpectsError: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			cards := database.NewCardRepository(db)
			if tc.Setup != nil {
				tc.Setup(cards)
			}

			actual, err := cards.Get(tc.ID)
			if tc.ExpectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.EqualValues(t, tc.Expected, actual)
		})
	}
}
