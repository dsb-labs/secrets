package database

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
)

type (
	// The CardRepository type is responsible for managing the persistence of user's payment cards. This should
	// be instantiated against a user's individual database.
	CardRepository struct {
		db *badger.DB
	}

	// The Card type represents a single payment card as stored in a user's individual database.
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
	}
)

var (
	// ErrCardNotFound is the error given when performing an operation on a card record that does not exist.
	ErrCardNotFound = errors.New("card not found")
)

func (p Card) key() []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("card/")
	buf.Write(p.ID[:])

	return buf.Bytes()
}

// NewCardRepository returns a new instance of the CardRepository type that will persist card data using the provided
// badger.DB database.
func NewCardRepository(db *badger.DB) *CardRepository {
	return &CardRepository{db: db}
}

// Create a new card record.
func (r *CardRepository) Create(card Card) error {
	data, err := json.Marshal(card)
	if err != nil {
		return fmt.Errorf("failed to marshal card %q: %w", card.ID, err)
	}

	return update(r.db, func(txn *badger.Txn) error {
		return txn.Set(card.key(), data)
	})
}

// List all card records.
func (r *CardRepository) List() ([]Card, error) {
	cards := make([]Card, 0)
	err := iterate(r.db, "card/", func(card Card) error {
		cards = append(cards, card)
		return nil
	})

	return cards, err
}

// Delete a card record, returns ErrCardNotFound if the card record does not exist.
func (r *CardRepository) Delete(id uuid.UUID) error {
	return update(r.db, func(txn *badger.Txn) error {
		key := Card{ID: id}.key()

		if _, err := txn.Get(key); errors.Is(err, badger.ErrKeyNotFound) {
			return ErrCardNotFound
		}

		return txn.Delete(key)
	})
}

// Get a card record by its id, returns ErrCardNotFound if the card record does not exist.
func (r *CardRepository) Get(id uuid.UUID) (Card, error) {
	return view(r.db, func(txn *badger.Txn) (Card, error) {
		card := Card{
			ID: id,
		}

		item, err := txn.Get(card.key())
		switch {
		case errors.Is(err, badger.ErrKeyNotFound):
			return Card{}, ErrCardNotFound
		case err != nil:
			return Card{}, err
		}

		err = item.Value(func(val []byte) error {
			return json.Unmarshal(val, &card)
		})

		return card, err
	})
}
