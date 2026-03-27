package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/davidsbond/x/convert"
	"github.com/davidsbond/x/filter"
	"github.com/google/uuid"

	"github.com/davidsbond/keeper/internal/server/database"
)

type (
	// The CardService type responsible for managing individual user card records.
	CardService struct {
		cards RepositoryProvider[CardRepository]
	}

	// The CardRepository interface describes types that persist card records.
	CardRepository interface {
		// Create should store a new payment card record.
		Create(database.Card) error
		// List should return all payment card records.
		List() ([]database.Card, error)
		// Delete should remove a payment card record, returning database.ErrCardNotFound if it does not exist.
		Delete(uuid.UUID) error
		// Get should return the payment card record associated with the given id, returning database.ErrCardNotFound if it
		// does not exist.
		Get(uuid.UUID) (database.Card, error)
	}

	// The Card type represents a single payment card record.
	Card struct {
		// The card's unique identifier.
		ID uuid.UUID
		// The cardholder's name.
		HolderName string
		// The card number.
		Number string
		// The month the card expires.
		ExpiryMonth time.Month
		// The year the card expires.
		ExpiryYear int
		// The card's CVV.
		CVV string
		// When the card was created.
		CreatedAt time.Time
	}
)

var (
	// ErrCardNotFound is the error given when trying to perform an operation against a card record that does not
	// exist.
	ErrCardNotFound = errors.New("card not found")
)

// NewCardService returns a new instance of the CardService type that will manage individual user cards using
// CardRepository implementations provided by the given RepositoryProvider implementation.
func NewCardService(cards RepositoryProvider[CardRepository]) *CardService {
	return &CardService{
		cards: cards,
	}
}

// Create a new payment card record for the specified user. Returns ErrReauthenticate if the underlying individual user
// database's lifetime has expired and the caller must reauthenticate.
func (svc *CardService) Create(userID uuid.UUID, card Card) error {
	repo, err := svc.cards.For(userID)
	switch {
	case errors.Is(err, database.ErrClosed):
		return ErrReauthenticate
	case err != nil:
		return fmt.Errorf("failed to get database for user: %w", err)
	}

	record := database.Card{
		ID:          card.ID,
		HolderName:  card.HolderName,
		Number:      card.Number,
		ExpiryMonth: card.ExpiryMonth,
		ExpiryYear:  card.ExpiryYear,
		CVV:         card.CVV,
		CreatedAt:   card.CreatedAt,
	}

	err = repo.Create(record)
	switch {
	case errors.Is(err, database.ErrClosed):
		return ErrReauthenticate
	case err != nil:
		return fmt.Errorf("failed to create card record: %w", err)
	default:
		return nil
	}
}

// List all payment card records for the specified user. Returns ErrReauthenticate if the underlying individual user
// database's lifetime has expired and the caller must reauthenticate.
func (svc *CardService) List(userID uuid.UUID, filters ...filter.Filter[Card]) ([]Card, error) {
	repo, err := svc.cards.For(userID)
	switch {
	case errors.Is(err, database.ErrClosed):
		return nil, ErrReauthenticate
	case err != nil:
		return nil, fmt.Errorf("failed to get database for user: %w", err)
	}

	results, err := repo.List()
	switch {
	case errors.Is(err, database.ErrClosed):
		return nil, ErrReauthenticate
	case err != nil:
		return nil, fmt.Errorf("failed to list card records: %w", err)
	}

	cards := convert.Slice(results, func(in database.Card) Card {
		return Card{
			ID:          in.ID,
			HolderName:  in.HolderName,
			Number:      in.Number,
			ExpiryMonth: in.ExpiryMonth,
			ExpiryYear:  in.ExpiryYear,
			CVV:         in.CVV,
			CreatedAt:   in.CreatedAt,
		}
	})

	if len(filters) == 0 {
		return cards, nil
	}

	return filter.All(cards, filters...), nil
}

// Delete a payment card record for the given user. Returns ErrReauthenticate if the underlying individual user
// database's lifetime has expired and the caller must reauthenticate. Returns ErrCardNotFound if the specified card
// record does not exist.
func (svc *CardService) Delete(userID uuid.UUID, cardID uuid.UUID) error {
	repo, err := svc.cards.For(userID)
	switch {
	case errors.Is(err, database.ErrClosed):
		return ErrReauthenticate
	case err != nil:
		return fmt.Errorf("failed to get database for user: %w", err)
	}

	err = repo.Delete(cardID)
	switch {
	case errors.Is(err, database.ErrClosed):
		return ErrReauthenticate
	case errors.Is(err, database.ErrCardNotFound):
		return ErrCardNotFound
	case err != nil:
		return fmt.Errorf("failed to delete card record: %w", err)
	}

	return nil
}

// Get a payment card record associated with the given user and card identifiers. Returns ErrReauthenticate if the
// underlying  individual user database's lifetime has expired and the caller must reauthenticate. Returns
// ErrCardNotFound if the specified payment card record does not exist.
func (svc *CardService) Get(userID uuid.UUID, cardID uuid.UUID) (Card, error) {
	repo, err := svc.cards.For(userID)
	switch {
	case errors.Is(err, database.ErrClosed):
		return Card{}, ErrReauthenticate
	case err != nil:
		return Card{}, fmt.Errorf("failed to get database for user: %w", err)
	}

	result, err := repo.Get(cardID)
	switch {
	case errors.Is(err, database.ErrClosed):
		return Card{}, ErrReauthenticate
	case errors.Is(err, database.ErrCardNotFound):
		return Card{}, ErrCardNotFound
	case err != nil:
		return Card{}, fmt.Errorf("failed to get card record: %w", err)
	}

	return Card{
		ID:          result.ID,
		HolderName:  result.HolderName,
		Number:      result.Number,
		ExpiryMonth: result.ExpiryMonth,
		ExpiryYear:  result.ExpiryYear,
		CVV:         result.CVV,
		CreatedAt:   result.CreatedAt, 
	}, nil
}
