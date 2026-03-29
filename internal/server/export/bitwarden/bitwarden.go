// Package bitwarden provides types for working with exports of Bitwarden vaults. 
package bitwarden

import (
	"time"
)

const (
	// ItemTypeUnknown represents an unrecognised or unsupported vault item type.
	ItemTypeUnknown int = iota
	// ItemTypeLogin represents a login item (credentials for a website or service).
	// When an item has this type its Login sub-object is populated.
	ItemTypeLogin
	// ItemTypeSecureNote represents a secure note item. The note body is stored in the
	// top-level Item.Notes field; the secureNote sub-object in the Bitwarden export is
	// always an empty object and is not modelled here.
	ItemTypeSecureNote
	// ItemTypeCard represents a payment card item. When an item has this type its Card
	// sub-object is populated.
	ItemTypeCard
)

type (
	// Bitwarden is the root object of an unencrypted Bitwarden JSON vault export.
	Bitwarden struct {
		Items []Item `json:"items"`
	}

	// URI holds a single website address associated with a login item. Bitwarden
	// supports per-URI match rules (default, host, startsWith, exact, regex, never)
	// but those are not used during import and are omitted here.
	URI struct {
		URI string `json:"uri"`
	}

	// Login holds the credential fields for an item of type ItemTypeLogin.
	Login struct {
		// Uris is the list of website addresses the credentials belong to.
		Uris []URI `json:"uris"`
		// Username is the account identifier (email address, username, etc.).
		Username string `json:"username"`
		// Password is the secret credential used to authenticate.
		Password string `json:"password"`
	}

	// Card holds the payment card fields for an item of type ItemTypeCard.
	Card struct {
		// CardholderName is the name embossed on the card.
		CardholderName string `json:"cardholderName"`
		// Number is the full card account number.
		Number string `json:"number"`
		// ExpMonth is the card's expiry month as a numeric string (e.g. "1" for January).
		ExpMonth string `json:"expMonth"`
		// ExpYear is the card's four-digit expiry year as a string (e.g. "2025").
		ExpYear string `json:"expYear"`
		// Code is the card verification value (CVV/CVC).
		Code string `json:"code"`
	}

	// Item represents a single entry in a Bitwarden vault export. The Type field
	// determines which sub-object (Login or Card) carries the item's data; all other
	// sub-objects are empty for a given type. Common metadata (name, notes, custom
	// fields) is held directly on the item regardless of type.
	Item struct {
		// CreationDate is when the item was first created in the vault.
		CreationDate time.Time `json:"creationDate"`
		// ID is the item's unique identifier within Bitwarden.
		ID string `json:"id"`
		// Type identifies the kind of vault item; see the ItemType* constants.
		Type int `json:"type"`
		// Name is the user-supplied display name for the item.
		Name string `json:"name"`
		// Notes contains free-form text attached to the item. For secure note items
		// this field holds the entire note body.
		Notes string `json:"notes"`
		// Fields holds any custom fields added to the item. Bitwarden supports text,
		// hidden, boolean, and linked field types; only the name and value are used
		// during import.
		Fields []Field `json:"fields"`
		// Login is populated when Type is ItemTypeLogin.
		Login Login `json:"login"`
		// Card is populated when Type is ItemTypeCard.
		Card Card `json:"card"`
	}

	// Field represents a custom field attached to a vault item. Bitwarden supports
	// four field types (text=0, hidden=1, boolean=2, linked=3); the type is not used
	// during import so it is omitted here.
	Field struct {
		// Name is the label of the custom field.
		Name string `json:"name"`
		// Value is the content of the custom field.
		Value string `json:"value"`
	}
)
