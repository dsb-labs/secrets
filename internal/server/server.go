// Package server provides types and functions for running the keeper server.
package server

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/a-h/templ"
	"github.com/davidsbond/x/closer"
	"github.com/davidsbond/x/lifetime"
	"github.com/davidsbond/x/syncmap"
	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"

	"github.com/davidsbond/keeper/internal/server/api"
	"github.com/davidsbond/keeper/internal/server/database"
	"github.com/davidsbond/keeper/internal/server/service"
	"github.com/davidsbond/keeper/internal/server/token"
	"github.com/davidsbond/keeper/internal/server/ui/layout"
	"github.com/davidsbond/keeper/internal/server/ui/view"
)

// Run the server using the provided configuration. This function blocks until the provided context is cancelled or
// an error occurs.
func Run(ctx context.Context, config Config) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid server configuration: %w", err)
	}

	closers := closer.NewCollection()
	defer closers.Close()

	// Store individual account databases in a concurrent map keyed by the user's unique identifier with each database
	// wrapped in a lifetime.Lifetime that will automatically close it once the configured TTL has elapsed.
	databaseState := syncmap.New[uuid.UUID, *lifetime.Lifetime[*badger.DB]]()
	databaseManager := database.NewManager(config.Database.Path, databaseState, config.Database.TTL)
	closers.Add(databaseManager)

	masterKey, err := base64.StdEncoding.DecodeString(config.Database.MasterKey)
	if err != nil {
		return err
	}

	masterDB, err := database.Open(filepath.Join(config.Database.Path, "master"), masterKey)
	if err != nil {
		return err
	}
	closers.Add(masterDB)

	signingKey, err := base64.StdEncoding.DecodeString(config.JWT.SigningKey)
	if err != nil {
		return err
	}

	tokenGenerator := token.NewGenerator(token.GeneratorConfig{
		Issuer:     config.JWT.Issuer,
		TTL:        config.JWT.TTL,
		SigningKey: signingKey,
		Audience:   config.JWT.Audience,
	})

	tokenParser := token.NewParser(token.ParserConfig{
		Issuer:     config.JWT.Issuer,
		Audience:   config.JWT.Audience,
		SigningKey: signingKey,
	})

	mux := http.NewServeMux()

	accounts := database.NewAccountRepository(masterDB)
	logins := database.NewRepositoryProvider(databaseState, service.LoginRepositoryProvider)
	notes := database.NewRepositoryProvider(databaseState, service.NoteRepositoryProvider)
	cards := database.NewRepositoryProvider(databaseState, service.CardRepositoryProvider)

	api.NewAuthAPI(service.NewAuthService(accounts, databaseManager, tokenGenerator)).Register(mux)
	api.NewAccountAPI(service.NewAccountService(accounts, databaseManager)).Register(mux)
	api.NewLoginAPI(service.NewLoginService(logins)).Register(mux)
	api.NewNoteAPI(service.NewNoteService(notes)).Register(mux)
	api.NewToolAPI(service.NewToolService(logins, notes, cards)).Register(mux)
	api.NewCardAPI(service.NewCardService(cards)).Register(mux)

	mux.Handle("/ui/", templ.Handler(layout.Main("Test", view.Login())))

	server := &http.Server{
		Addr:    config.HTTP.Bind,
		Handler: token.Middleware(tokenParser, mux),
	}

	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return server.ListenAndServe()
	})

	group.Go(func() error {
		<-ctx.Done()
		sCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		return server.Shutdown(sCtx)
	})

	err = group.Wait()
	switch {
	case errors.Is(err, http.ErrServerClosed):
		return nil
	default:
		return err
	}
}
