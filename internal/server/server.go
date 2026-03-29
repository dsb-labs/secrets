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

	"filippo.io/csrf"
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
	"github.com/davidsbond/keeper/internal/ui"
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

	accountRepo := database.NewAccountRepository(masterDB)
	loginRepo := database.NewRepositoryProvider(databaseState, service.LoginRepositoryProvider)
	noteRepo := database.NewRepositoryProvider(databaseState, service.NoteRepositoryProvider)
	cardRepo := database.NewRepositoryProvider(databaseState, service.CardRepositoryProvider)

	authSvc := service.NewAuthService(accountRepo, databaseManager, tokenGenerator)
	accountSvc := service.NewAccountService(accountRepo, databaseManager)
	loginSvc := service.NewLoginService(loginRepo)
	noteSvc := service.NewNoteService(noteRepo)
	cardSvc := service.NewCardService(cardRepo)
	toolSvc := service.NewToolService(loginRepo, noteRepo, cardRepo)

	// API handlers.
	api.NewAuthAPI(authSvc).Register(mux)
	api.NewAccountAPI(accountSvc).Register(mux)
	api.NewLoginAPI(loginSvc).Register(mux)
	api.NewNoteAPI(noteSvc).Register(mux)
	api.NewToolAPI(toolSvc).Register(mux)
	api.NewCardAPI(cardSvc).Register(mux)

	// UI handlers.
	ui.NewAuthHandler(authSvc).Register(mux)
	ui.NewDashboardHandler(accountSvc).Register(mux)
	ui.NewAccountHandler(accountSvc).Register(mux)
	ui.NewLoginHandler(accountSvc, loginSvc).Register(mux)
	ui.NewNoteHandler(accountSvc, noteSvc).Register(mux)
	ui.NewCardHandler(accountSvc, cardSvc).Register(mux)
	ui.NewToolHandler(accountSvc, toolSvc).Register(mux)
	ui.NewAssetHandler().Register(mux)
	ui.NewNotFoundHandler().Register(mux)

	protection := csrf.New()

	middlewares := []func(handler http.Handler) http.Handler{
		protection.Handler,
		token.Middleware(tokenParser),
	}

	server := &http.Server{
		Addr:    config.HTTP.Bind,
		Handler: mux,
	}

	for _, middleware := range middlewares {
		server.Handler = middleware(server.Handler)
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
