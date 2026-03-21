package account

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/davidsbond/keeper/internal/cli"
	"github.com/davidsbond/keeper/pkg/keeper"
)

func restore() *cobra.Command {
	return &cobra.Command{
		Use:   "restore [email]",
		Short: "Restore an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client := cli.ClientFromContext(ctx)

			email := args[0]
			rawKey, err := cli.PromptPassword("Restore key")
			if err != nil {
				return fmt.Errorf("failed to read restore key: %w", err)
			}

			restoreKey, err := keeper.ParseRestoreKey(rawKey)
			if err != nil {
				return fmt.Errorf("failed to decode restore key: %w", err)
			}

			newPassword, err := cli.PromptPassword("New password")
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}

			confirm, err := cli.PromptPassword("Confirm new password")
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}

			if newPassword != confirm {
				return errors.New("new passwords do not match")
			}

			restoreKey, err = client.RestoreAccount(ctx, email, restoreKey, newPassword)
			if err != nil {
				return fmt.Errorf("failed to change password: %w", err)
			}

			fmt.Printf("Account restored. Please store the following key for account recovery:\n\n%s\n", restoreKey)

			return nil
		},
	}
}
