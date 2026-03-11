package keeper

import (
	"context"
	"net/http"
	"path"
	"time"

	"github.com/davidsbond/x/convert"

	"github.com/davidsbond/keeper/internal/server/api"
)

type (
	// The Card type represents a single payment card.
	Card struct {
		// The unique identifier of the card.
		ID string
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

// CreateCard attempts to create a new card record for the authenticated user, returning its identifier on success.
func (c *Client) CreateCard(ctx context.Context, card Card) (string, error) {
	request, err := c.buildRequest(ctx, http.MethodPost, "/api/v1/card", api.CreateCardRequest{
		HolderName:  card.HolderName,
		Number:      card.Number,
		ExpiryMonth: card.ExpiryMonth,
		ExpiryYear:  card.ExpiryYear,
		CVV:         card.CVV,
	})
	if err != nil {
		return "", err
	}

	response, err := doRequest[api.CreateCardResponse](c.client, request)
	if err != nil {
		return "", err
	}

	return response.ID, nil
}

// ListCards attempts to return all card records stored for the authenticated user.
func (c *Client) ListCards(ctx context.Context) ([]Card, error) {
	request, err := c.buildRequest(ctx, http.MethodGet, "/api/v1/card", nil)
	if err != nil {
		return nil, err
	}

	response, err := doRequest[api.ListCardsResponse](c.client, request)
	if err != nil {
		return nil, err
	}

	return convert.Slice(response.Cards, func(in api.Card) Card {
		return Card{
			ID:          in.ID,
			HolderName:  in.HolderName,
			Number:      in.Number,
			ExpiryMonth: in.ExpiryMonth,
			ExpiryYear:  in.ExpiryYear,
			CVV:         in.CVV,
		}
	}), nil
}

// DeleteCard attempts to delete the card record with the specified id for the authenticated user.
func (c *Client) DeleteCard(ctx context.Context, id string) error {
	request, err := c.buildRequest(ctx, http.MethodDelete, path.Join("/api/v1/card", id), nil)
	if err != nil {
		return err
	}

	if _, err = doRequest[api.DeleteCardResponse](c.client, request); err != nil {
		return err
	}

	return nil
}

// GetCard attempts to obtain the card record with the specified id for the authenticated user.
func (c *Client) GetCard(ctx context.Context, id string) (Card, error) {
	request, err := c.buildRequest(ctx, http.MethodGet, path.Join("/api/v1/card", id), nil)
	if err != nil {
		return Card{}, err
	}

	response, err := doRequest[api.GetCardResponse](c.client, request)
	if err != nil {
		return Card{}, err
	}

	return Card{
		ID:          response.Card.ID,
		HolderName:  response.Card.HolderName,
		Number:      response.Card.Number,
		ExpiryMonth: response.Card.ExpiryMonth,
		ExpiryYear:  response.Card.ExpiryYear,
		CVV:         response.Card.CVV,
	}, nil
}
