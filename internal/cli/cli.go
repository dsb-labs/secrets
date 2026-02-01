// Package cli provides shared functionality required by commands exposed by the CLI.
package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/davidsbond/keeper/internal/cli/config"
	"github.com/davidsbond/keeper/pkg/keeper"
)

type (
	ctxKey struct{}
)

// CreateClient is to be used as a PersistentPreRun function for a cobra root command that adds a keeper.Client
// instance into the command context for child commands to use.
func CreateClient(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}

	apiURL, err := cmd.Flags().GetString("api-url")
	if err != nil {
		return err
	}

	cfg, err := config.Load(configPath)
	switch {
	case errors.Is(err, config.ErrNotFound):
		// If there's no config, it could be a first time run so we can just set up a client
		// without a token as we're probably about to login
		break
	case err != nil:
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := keeper.NewClient(apiURL)
	client.SetToken(cfg.Token)

	cmd.SetContext(ClientToContext(ctx, client))

	return nil
}

// ClientToContext adds the given keeper.Client to the context.Context.
func ClientToContext(ctx context.Context, client *keeper.Client) context.Context {
	return context.WithValue(ctx, ctxKey{}, client)
}

// ClientFromContext returns a keeper.Client from the context.Context, or nil if one is not found.
func ClientFromContext(ctx context.Context) *keeper.Client {
	client, ok := ctx.Value(ctxKey{}).(*keeper.Client)
	if !ok {
		return nil
	}

	return client
}

// PromptPassword creates an "enter password" prompt on stdout/stdin that masks the input text, returning the entered
// password as a string.
func PromptPassword() (string, error) {
	fmt.Print("Enter password: ")
	pwd, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}

	fmt.Println("")
	return string(pwd), nil
}
