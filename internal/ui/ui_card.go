package ui

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/davidsbond/x/filter"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/google/uuid"

	"github.com/dsb-labs/secrets/internal/server/service"
	"github.com/dsb-labs/secrets/internal/server/token"
	cardview "github.com/dsb-labs/secrets/internal/ui/view/card"
	statusview "github.com/dsb-labs/secrets/internal/ui/view/status"
)

type (
	// The CardHandler type is responsible for serving web interface pages regarding user payment cards.
	CardHandler struct {
		accounts AccountService
		cards    CardService
	}

	// The CardService interface describes types that manage user payment card records.
	CardService interface {
		// Create should store a new card record for the given user.
		Create(accountID uuid.UUID, card service.Card) error
		// List should return all cards associated with the given user id.
		List(accountID uuid.UUID, filters ...filter.Filter[service.Card]) ([]service.Card, error)
		// Get should return the card record associated with the given user and card identifiers.
		Get(accountID uuid.UUID, cardID uuid.UUID) (service.Card, error)
		// Delete should remove the card record associated with the given user and card identifiers.
		Delete(accountID uuid.UUID, cardID uuid.UUID) error
	}
)

// NewCardHandler returns a new instance of the CardHandler type that will serve card management UIs using
// the provided service implementations.
func NewCardHandler(accounts AccountService, cards CardService) *CardHandler {
	return &CardHandler{accounts: accounts, cards: cards}
}

// Register HTTP endpoints onto the provided http.ServeMux.
func (h *CardHandler) Register(mux *http.ServeMux) {
	mux.Handle("GET /cards", requireToken(h.List))
	mux.Handle("GET /cards/new", requireToken(h.Create))
	mux.Handle("POST /cards", requireToken(h.CreateCallback))
	mux.Handle("GET /cards/{id}", requireToken(h.Detail))
	mux.Handle("POST /cards/{id}/delete", requireToken(h.Delete))
}

// List renders the card list view.
func (h *CardHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	query := r.URL.Query().Get("query")
	results, err := h.cards.List(tkn.ID(), service.CardsByName(query))
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		redirectToLogin(w, r)
		return
	case err != nil:
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	now := time.Now()
	items := make([]cardview.Item, len(results))
	for i, c := range results {
		items[i] = cardview.Item{
			ID:           c.ID.String(),
			Name:         c.Name,
			MaskedNumber: maskCardNumber(c.Number),
			Expiry:       fmt.Sprintf("%02d/%02d", int(c.ExpiryMonth), c.ExpiryYear%100),
			Issuer:       c.Issuer,
			Expired:      c.ExpiryYear < now.Year() || (c.ExpiryYear == now.Year() && c.ExpiryMonth < now.Month()),
		}
	}

	render(ctx, w, http.StatusOK, cardview.List, cardview.ViewModel{
		DisplayName: account.DisplayName,
		Cards:       items,
		Query:       query,
	})
}

// Create renders the card creation form.
func (h *CardHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	render(ctx, w, http.StatusOK, cardview.Create, cardview.CreateViewModel{
		DisplayName: account.DisplayName,
	})
}

// The CreateCardForm type represents the form values submitted when calling CardHandler.CreateCallback.
type CreateCardForm struct {
	// The user-supplied name for the card.
	Name string `form:"name"`
	// The cardholder's name.
	HolderName string `form:"holderName"`
	// The card number.
	Number string `form:"number"`
	// The month the card expires.
	ExpiryMonth time.Month `form:"expiryMonth"`
	// The year the card expires.
	ExpiryYear int `form:"expiryYear"`
	// The card's CVV.
	CVV string `form:"cvv"`
}

// Validate the form.
func (f CreateCardForm) Validate() error {
	return validation.ValidateStruct(&f,
		validation.Field(&f.Name, validation.Required),
		validation.Field(&f.Number, validation.Required, is.CreditCard),
		validation.Field(&f.ExpiryMonth, validation.Required, validation.Min(time.January), validation.Max(time.December)),
		validation.Field(&f.ExpiryYear, validation.Required),
		validation.Field(&f.CVV, validation.Required, validation.Length(3, 4)),
	)
}

// CreateCallback handles the card creation form submission, redirecting to the card list on success.
func (h *CardHandler) CreateCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	form, err := decode[CreateCardForm](r)

	expiryMonth := ""
	if form.ExpiryMonth != 0 {
		expiryMonth = strconv.Itoa(int(form.ExpiryMonth))
	}
	expiryYear := ""
	if form.ExpiryYear != 0 {
		expiryYear = strconv.Itoa(form.ExpiryYear)
	}

	model := cardview.CreateViewModel{
		DisplayName: account.DisplayName,
		Name:        form.Name,
		HolderName:  form.HolderName,
		Number:      form.Number,
		ExpiryMonth: expiryMonth,
		ExpiryYear:  expiryYear,
		CVV:         form.CVV,
	}

	var ve validation.Errors
	switch {
	case errors.As(err, &ve):
		model.Validation.Errors = validationErrors(ve)
		render(ctx, w, http.StatusUnprocessableEntity, cardview.Create, model)
		return
	case err != nil:
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	cardID := uuid.New()
	err = h.cards.Create(tkn.ID(), service.Card{
		ID:          cardID,
		Name:        form.Name,
		HolderName:  form.HolderName,
		Number:      form.Number,
		ExpiryMonth: form.ExpiryMonth,
		ExpiryYear:  form.ExpiryYear,
		CVV:         form.CVV,
		CreatedAt:   time.Now(),
	})
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		redirectToLogin(w, r)
		return
	case err != nil:
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	redirect(w, r, "/cards")
}

// Detail renders the card detail view for a single card record.
func (h *CardHandler) Detail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	cardID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		render(ctx, w, http.StatusNotFound, statusview.NotFound, statusview.NotFoundViewModel{})
		return
	}

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	card, err := h.cards.Get(tkn.ID(), cardID)
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		redirectToLogin(w, r)
		return
	case errors.Is(err, service.ErrCardNotFound):
		render(ctx, w, http.StatusNotFound, statusview.NotFound, statusview.NotFoundViewModel{})
		return
	case err != nil:
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	now := time.Now()
	render(ctx, w, http.StatusOK, cardview.Detail, cardview.DetailViewModel{
		DisplayName: account.DisplayName,
		ID:          card.ID.String(),
		Name:        card.Name,
		HolderName:  card.HolderName,
		Number:      card.Number,
		Expiry:      fmt.Sprintf("%02d/%d", int(card.ExpiryMonth), card.ExpiryYear),
		CVV:         card.CVV,
		CreatedAt:   card.CreatedAt.Format("2 January 2006 at 15:04"),
		Issuer:      card.Issuer,
		Expired:     card.ExpiryYear < now.Year() || (card.ExpiryYear == now.Year() && card.ExpiryMonth < now.Month()),
	})
}

// Delete handles a card deletion request, redirecting to the card list on success.
func (h *CardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	cardID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		render(ctx, w, http.StatusNotFound, statusview.NotFound, statusview.NotFoundViewModel{})
		return
	}

	err = h.cards.Delete(tkn.ID(), cardID)
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		redirectToLogin(w, r)
		return
	case errors.Is(err, service.ErrCardNotFound):
		render(ctx, w, http.StatusNotFound, statusview.NotFound, statusview.NotFoundViewModel{})
		return
	case err != nil:
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	redirect(w, r, "/cards")
}

// maskCardNumber replaces all but the last four digits of a card number with bullets.
func maskCardNumber(number string) string {
	if len(number) < 4 {
		return number
	}
	return "•••• •••• •••• " + number[len(number)-4:]
}
