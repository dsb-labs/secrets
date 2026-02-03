package main

import (
	"context"
	"os"
	"os/signal"
	"runtime/debug"

	"github.com/spf13/cobra"

	"github.com/davidsbond/keeper/cmd/account"
	"github.com/davidsbond/keeper/cmd/auth"
	"github.com/davidsbond/keeper/cmd/login"
	"github.com/davidsbond/keeper/cmd/note"
	"github.com/davidsbond/keeper/cmd/serve"
)

//go:generate go tool mockery
//go:generate go tool go-licenses save --one_output --force --ignore "github.com/davidsbond/keeper" --save_path licenses .
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	cmd := &cobra.Command{
		Use:          "keeper",
		Short:        "A simple, self-hostable secret manager",
		SilenceUsage: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		cmd.Version = info.Main.Version
	}

	cmd.AddCommand(
		serve.Command(),
		auth.Command(),
		login.Command(),
		note.Command(),
		account.Command(),
	)

	if err := cmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
