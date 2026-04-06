package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/davidsbond/x/convert"
	"github.com/davidsbond/x/filter"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/google/uuid"

	"github.com/dsb-labs/secrets/internal/server/service"
	"github.com/dsb-labs/secrets/internal/server/token"
)

type (
	// The CardAPI exposes HTTP endpoints for managing individual payment cards.
	CardAPI struct {
		cards CardService
	}

	// The CardService interface describes types that manage payment cards.
	CardService interface {
		// Create should create a new payment card record for the given user id.
		Create(accountID uuid.UUID, card service.Card) error
		// List should return all cards associated with the given user id.
		List(accountID uuid.UUID, filters ...filter.Filter[service.Card]) ([]service.Card, error)
		// Delete should remove the payment card record associated with the given user and card id. Returning
		// service.ErrCardNotFound if it does not exist.
		Delete(accountID uuid.UUID, cardID uuid.UUID) error
		// Get should return the payment card record associated with the given user and card id. Returning
		// service.ErrCardNotFound if it does not exist.
		Get(accountID uuid.UUID, cardID uuid.UUID) (service.Card, error)
	}

	// The Card type represents a single payment card.
	Card struct {
		// The unique identifier of the card.
		ID string `json:"id"`
		// The cardholder's name.
		HolderName string `json:"holderName"`
		// The card number.
		Number string `json:"number"`
		// The month the card expires.
		ExpiryMonth time.Month `json:"expiryMonth"`
		// The year the card expires.
		ExpiryYear int `json:"expiryYear"`
		// The card's CVV.
		CVV string `json:"cvv"`
		// When the card was created.
		CreatedAt time.Time `json:"createdAt"`
		// A user-supplied name for the card
		Name string `json:"name"`
		// The card issuer
		Issuer string `json:"issuer"`
	}
)

// NewCardAPI returns a new instance of the CardAPI type that manages payment cards via the
// given CardService implementation.
func NewCardAPI(cards CardService) *CardAPI {
	return &CardAPI{cards: cards}
}

// Register the HTTP endpoints onto the given http.ServeMux.
func (api *CardAPI) Register(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/card", requireToken(api.Create))
	mux.Handle("GET /api/v1/card", requireToken(api.List))
	mux.Handle("GET /api/v1/card/{id}", requireToken(api.Get))
	mux.Handle("DELETE /api/v1/card/{id}", requireToken(api.Delete))
}

type (
	// The CreateCardRequest type represents the request body given when calling CardAPI.Create
	CreateCardRequest struct {
		// The cardholder's name.
		HolderName string `json:"holderName"`
		// The card number.
		Number string `json:"number"`
		// The month the card expires.
		ExpiryMonth time.Month `json:"expiryMonth"`
		// The year the card expires.
		ExpiryYear int `json:"expiryYear"`
		// The card's CVV.
		CVV string `json:"cvv"`
		// A user-supplied name for the card
		Name string `json:"name"`
	}

	// The CreateCardResponse type represents the response body returned when calling CardAPI.Create
	CreateCardResponse struct {
		ID string `json:"id"`
	}
)

// Validate the request.
func (r CreateCardRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Number, validation.Required, is.CreditCard),
		validation.Field(&r.ExpiryMonth, validation.Required, validation.Min(time.January), validation.Max(time.December)),
		validation.Field(&r.ExpiryYear, validation.Required),
		validation.Field(&r.CVV, validation.Required, validation.Length(3, 4)),
		validation.Field(&r.Name, validation.Required),
	)
}

// Create handles an inbound HTTP request to store a new payment card record for a user. On success, it responds with
// an http.StatusCreated code and a JSON-encoded CreateCardResponse.
func (api *CardAPI) Create(w http.ResponseWriter, r *http.Request) {
	tkn := token.FromContext(r.Context())
	request, err := decode[CreateCardRequest](r.Body)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "failed to decode request: %v", err)
		return
	}

	card := service.Card{
		ID:          uuid.New(),
		HolderName:  request.HolderName,
		Number:      request.Number,
		ExpiryMonth: request.ExpiryMonth,
		ExpiryYear:  request.ExpiryYear,
		CVV:         request.CVV,
		CreatedAt:   time.Now(),
		Name:        request.Name,
	}

	err = api.cards.Create(tkn.ID(), card)
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		writeError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	case err != nil:
		writeErrorf(w, http.StatusInternalServerError, "failed to create card: %v", err)
		return
	}

	write(w, http.StatusCreated, CreateCardResponse{
		ID: card.ID.String(),
	})
}

type (
	// The ListCardsResponse type represents the response body returned when calling CardAPI.List
	ListCardsResponse struct {
		// The cards stored for the account.
		Cards []Card `json:"cards"`
	}
)

// List handles an inbound HTTP request to list all payment card records for a user. On success, it responds with
// an http.StatusOK code and a JSON-encoded ListCardsResponse.
func (api *CardAPI) List(w http.ResponseWriter, r *http.Request) {
	tkn := token.FromContext(r.Context())
	filters := make([]filter.Filter[service.Card], 0)
	if name := r.URL.Query().Get("name"); name != "" {
		filters = append(filters, service.CardsByName(name))
	}

	results, err := api.cards.List(tkn.ID(), filters...)
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		writeError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	case err != nil:
		writeErrorf(w, http.StatusInternalServerError, "failed to list cards: %v", err)
		return
	}

	write(w, http.StatusOK, ListCardsResponse{
		Cards: convert.Slice(results, func(in service.Card) Card {
			return Card{
				ID:          in.ID.String(),
				HolderName:  in.HolderName,
				Number:      in.Number,
				ExpiryMonth: in.ExpiryMonth,
				ExpiryYear:  in.ExpiryYear,
				CVV:         in.CVV,
				CreatedAt:   in.CreatedAt,
				Name:        in.Name,
				Issuer:      in.Issuer,
			}
		}),
	})
}

type (
	// The DeleteCardResponse type represents the response body returned when calling CardAPI.Delete
	DeleteCardResponse struct{}
)

// Delete handles an inbound HTTP request to delete a payment card record for a user. On success, it responds with
// an http.StatusOK code and a JSON-encoded DeleteCardResponse.
func (api *CardAPI) Delete(w http.ResponseWriter, r *http.Request) {
	tkn := token.FromContext(r.Context())
	cardID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "failed to parse card id: %v", err)
		return
	}

	err = api.cards.Delete(tkn.ID(), cardID)
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		writeError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	case errors.Is(err, service.ErrCardNotFound):
		writeErrorf(w, http.StatusNotFound, "card %q does not exist", cardID)
		return
	case err != nil:
		writeErrorf(w, http.StatusInternalServerError, "failed to delete card: %v", err)
		return
	}

	write(w, http.StatusOK, DeleteCardResponse{})
}

type (
	// The GetCardResponse type represents the response body returned when calling CardAPI.Get
	GetCardResponse struct {
		// The requested card details.
		Card Card `json:"card"`
	}
)

// Get handles an inbound HTTP request to query a payment card record for a user. On success, it responds with
// an http.StatusOK code and a JSON-encoded GetCardResponse.
func (api *CardAPI) Get(w http.ResponseWriter, r *http.Request) {
	tkn := token.FromContext(r.Context())
	cardID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "failed to parse card id: %v", err)
		return
	}

	result, err := api.cards.Get(tkn.ID(), cardID)
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		writeError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	case errors.Is(err, service.ErrCardNotFound):
		writeErrorf(w, http.StatusNotFound, "card %q does not exist", cardID)
		return
	case err != nil:
		writeErrorf(w, http.StatusInternalServerError, "failed to get card: %v", err)
		return
	}

	write(w, http.StatusOK, GetCardResponse{
		Card: Card{
			ID:          result.ID.String(),
			HolderName:  result.HolderName,
			Number:      result.Number,
			ExpiryMonth: result.ExpiryMonth,
			ExpiryYear:  result.ExpiryYear,
			CVV:         result.CVV,
			CreatedAt:   result.CreatedAt,
			Name:        result.Name,
			Issuer:      result.Issuer,
		},
	})
}
