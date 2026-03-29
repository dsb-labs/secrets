package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/davidsbond/x/convert"
	"github.com/google/uuid"

	"github.com/davidsbond/keeper/internal/server/database"
	"github.com/davidsbond/keeper/internal/server/export/bitwarden"
)

type (
	// The ToolService type provides common user tool implementations, like data export/import.
	ToolService struct {
		logins RepositoryProvider[LoginRepository]
		notes  RepositoryProvider[NoteRepository]
		cards  RepositoryProvider[CardRepository]
	}

	// The Export type represents a user's entire dataset.
	Export struct {
		// The user's logins.
		Logins []Login
		// The user's notes.
		Notes []Note
		// The user's payment cards.
		Cards []Card
	}

	// The ImportSource type is used to denote where imported data has come from for conversion.
	ImportSource uint

	// The ImportResult type describes the result of an import.
	ImportResult struct {
		// The number of logins imported.
		Logins int
		// The number of notes imported.
		Notes int
		// The number of cards imported.
		Cards int
		// Any errors that occurred during import for individual items.
		Errors []string
	}
)

// Constants for import sources.
const (
	ImportSourceKeeper ImportSource = iota
	ImportSourceBitwarden
)

var (
	// ErrInvalidImportSource is the error given when calling ToolService.Import with an invalid ImportSource value.
	ErrInvalidImportSource = errors.New("invalid import source")
)

// NewToolService returns a new instance of the ToolService type that will query logins, cards and notes from the given
// repository provider implementations.
func NewToolService(logins RepositoryProvider[LoginRepository], notes RepositoryProvider[NoteRepository], cards RepositoryProvider[CardRepository]) *ToolService {
	return &ToolService{
		notes:  notes,
		logins: logins,
		cards:  cards,
	}
}

// Export the specified user's data.  Returns ErrReauthenticate if the underlying individual user database's lifetime
// has expired and the caller must reauthenticate.
func (svc *ToolService) Export(userID uuid.UUID) (Export, error) {
	loginRepo, err := svc.logins.For(userID)
	switch {
	case errors.Is(err, database.ErrClosed):
		return Export{}, ErrReauthenticate
	case err != nil:
		return Export{}, fmt.Errorf("failed to get database for user: %w", err)
	}

	noteRepo, err := svc.notes.For(userID)
	switch {
	case errors.Is(err, database.ErrClosed):
		return Export{}, ErrReauthenticate
	case err != nil:
		return Export{}, fmt.Errorf("failed to get database for user: %w", err)
	}

	cardRepo, err := svc.cards.For(userID)
	switch {
	case errors.Is(err, database.ErrClosed):
		return Export{}, ErrReauthenticate
	case err != nil:
		return Export{}, fmt.Errorf("failed to get database for user: %w", err)
	}

	logins, err := loginRepo.List()
	switch {
	case errors.Is(err, database.ErrClosed):
		return Export{}, ErrReauthenticate
	case err != nil:
		return Export{}, fmt.Errorf("failed to list logins: %w", err)
	}

	notes, err := noteRepo.List()
	switch {
	case errors.Is(err, database.ErrClosed):
		return Export{}, ErrReauthenticate
	case err != nil:
		return Export{}, fmt.Errorf("failed to list notes: %w", err)
	}

	cards, err := cardRepo.List()
	switch {
	case errors.Is(err, database.ErrClosed):
		return Export{}, ErrReauthenticate
	case err != nil:
		return Export{}, fmt.Errorf("failed to list cards: %w", err)
	}

	return Export{
		Notes: convert.Slice(notes, func(in database.Note) Note {
			return Note{
				ID:      in.ID,
				Name:    in.Name,
				Content: in.Content,
			}
		}),
		Logins: convert.Slice(logins, func(in database.Login) Login {
			return Login{
				ID:       in.ID,
				Username: in.Username,
				Password: in.Password,
				Domains:  in.Domains,
			}
		}),
		Cards: convert.Slice(cards, func(in database.Card) Card {
			return Card{
				ID:          in.ID,
				HolderName:  in.HolderName,
				Number:      in.Number,
				ExpiryMonth: in.ExpiryMonth,
				ExpiryYear:  in.ExpiryYear,
				CVV:         in.CVV,
			}
		}),
	}, nil
}

// Import data for a user, converting the data from its original source (identified by the ImportSource enum) to a
// keeper Export type. UUIDs will not be preserved and an ImportResult is returned indicating the total number of
// items imported along with any individual errors.
func (svc *ToolService) Import(userID uuid.UUID, source ImportSource, data io.Reader) (ImportResult, error) {
	switch source {
	case ImportSourceKeeper:
		return svc.importKeeper(userID, data)
	case ImportSourceBitwarden:
		return svc.importBitwarden(userID, data)
	default:
		return ImportResult{}, ErrInvalidImportSource
	}
}

func (svc *ToolService) importKeeper(userID uuid.UUID, data io.Reader) (ImportResult, error) {
	var export Export
	if err := json.NewDecoder(data).Decode(&export); err != nil {
		return ImportResult{}, fmt.Errorf("failed to decode export data: %w", err)
	}

	result := ImportResult{
		Logins: len(export.Logins),
		Notes:  len(export.Notes),
		Cards:  len(export.Cards),
	}

	return result, svc.performImport(userID, export)
}

func (svc *ToolService) importBitwarden(userID uuid.UUID, data io.Reader) (ImportResult, error) {
	var bw bitwarden.Bitwarden
	if err := json.NewDecoder(data).Decode(&bw); err != nil {
		return ImportResult{}, fmt.Errorf("failed to decode bitwarden data: %w", err)
	}

	var export Export
	var result ImportResult

	for _, item := range bw.Items {
		switch item.Type {
		default:
			continue
		case bitwarden.ItemTypeLogin:
			export.Logins = append(export.Logins, Login{
				Username:  item.Login.Username,
				Password:  item.Login.Password,
				CreatedAt: item.CreationDate,
				Domains: convert.Slice(item.Login.Uris, func(have bitwarden.URI) string {
					return have.URI
				}),
			})

			result.Logins++
		case bitwarden.ItemTypeSecureNote:
			fields := convert.Slice(item.Fields, func(have bitwarden.Field) string {
				return have.Name + "\n" + have.Value
			})

			export.Notes = append(export.Notes, Note{
				Name:      item.Name,
				Content:   strings.Join(append([]string{item.Notes}, fields...), "\n"),
				CreatedAt: item.CreationDate,
			})

			result.Notes++
		case bitwarden.ItemTypeCard:
			month, err := strconv.Atoi(item.Card.ExpMonth)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to import card %q, expiry month is invalid: %v", item.Name, err))
				continue
			}

			year, err := strconv.Atoi(item.Card.ExpYear)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to import card %q, expiry year is invalid: %v", item.Name, err))
				continue
			}

			export.Cards = append(export.Cards, Card{
				HolderName:  item.Card.CardholderName,
				Number:      item.Card.Number,
				ExpiryMonth: time.Month(month),
				ExpiryYear:  year,
				CVV:         item.Card.Code,
				CreatedAt:   item.CreationDate,
				Name:        item.Name,
				Issuer:      cardIssuer(item.Card.Number),
			})

			result.Cards++
		}
	}

	return result, svc.performImport(userID, export)
}

func (svc *ToolService) performImport(userID uuid.UUID, export Export) error {
	loginRepo, err := svc.logins.For(userID)
	switch {
	case errors.Is(err, database.ErrClosed):
		return ErrReauthenticate
	case err != nil:
		return fmt.Errorf("failed to get database for user: %w", err)
	}

	noteRepo, err := svc.notes.For(userID)
	switch {
	case errors.Is(err, database.ErrClosed):
		return ErrReauthenticate
	case err != nil:
		return fmt.Errorf("failed to get database for user: %w", err)
	}

	cardRepo, err := svc.cards.For(userID)
	switch {
	case errors.Is(err, database.ErrClosed):
		return ErrReauthenticate
	case err != nil:
		return fmt.Errorf("failed to get database for user: %w", err)
	}

	for _, login := range export.Logins {
		err = loginRepo.Create(database.Login{
			ID:        uuid.New(),
			Username:  login.Username,
			Password:  login.Password,
			Domains:   login.Domains,
			CreatedAt: login.CreatedAt,
		})
		switch {
		case errors.Is(err, database.ErrClosed):
			return ErrReauthenticate
		case err != nil:
			return fmt.Errorf("failed to create login: %w", err)
		}
	}

	for _, note := range export.Notes {
		err = noteRepo.Create(database.Note{
			ID:        uuid.New(),
			Name:      note.Name,
			Content:   note.Content,
			CreatedAt: note.CreatedAt,
		})
		switch {
		case errors.Is(err, database.ErrClosed):
			return ErrReauthenticate
		case err != nil:
			return fmt.Errorf("failed to create note: %w", err)
		}
	}

	for _, card := range export.Cards {
		err = cardRepo.Create(database.Card{
			ID:          uuid.New(),
			HolderName:  card.HolderName,
			Number:      card.Number,
			ExpiryMonth: card.ExpiryMonth,
			ExpiryYear:  card.ExpiryYear,
			CVV:         card.CVV,
			CreatedAt:   card.CreatedAt,
			Name:        card.Name,
			Issuer:      card.Issuer,
		})
		switch {
		case errors.Is(err, database.ErrClosed):
			return ErrReauthenticate
		case err != nil:
			return fmt.Errorf("failed to create card: %w", err)
		}
	}

	return nil
}
