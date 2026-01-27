package main

import (
	"context"
	"os"
	"os/signal"
	"runtime/debug"

	"github.com/spf13/cobra"

	"github.com/davidsbond/passwords/cmd/auth"
	"github.com/davidsbond/passwords/cmd/login"
	"github.com/davidsbond/passwords/cmd/serve"
)

//go:generate go tool mockery
//go:generate go tool go-licenses save --one_output --force --ignore "github.com/davidsbond/passwords" --save_path licenses .
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	cmd := &cobra.Command{
		Use:   "passwords",
		Short: "A simple password manager",
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
	)

	if err := cmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
