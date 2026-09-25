package commands

import (
	"fmt"

	"github.com/coccinella-labs/go-kit/internal/auth"
	"github.com/coccinella-labs/go-kit/internal/exitcode"
	"github.com/spf13/cobra"
)

func newAuthCmd() *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication commands",
		Long:  `Manage authentication tokens and credentials.`,
	}

	authCmd.AddCommand(newAuthLoginCmd())
	authCmd.AddCommand(newAuthStatusCmd())
	authCmd.AddCommand(newAuthLogoutCmd())
	authCmd.AddCommand(newAuthWhoamiCmd())

	return authCmd
}

func newAuthLoginCmd() *cobra.Command {
	var provider string
	var clientID string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with a provider",
		Long:  `Store an authentication token for a service provider.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			token, _ := cmd.Flags().GetString("token")
			if token != "" {
				prov, err := auth.ParseProvider(provider)
				if err != nil {
					return exitcode.New(exitcode.ErrInvalidUsage, err)
				}
				if err := auth.Login(prov, token); err != nil {
					return exitcode.New(exitcode.ErrConfig, fmt.Errorf("failed to login: %w", err))
				}
				cmd.Printf("Successfully authenticated with %s\n", provider)
				return nil
			}

			ctx := cmd.Context()
			scopes := []string{"repo", "read:user"}

			prov, err := auth.ParseProvider(provider)
			if err != nil {
				return exitcode.New(exitcode.ErrInvalidUsage, err)
			}

			accessToken, err := auth.LoginWithOAuth(ctx, prov, clientID, scopes)
			if err != nil {
				return exitcode.New(exitcode.ErrNetwork, fmt.Errorf("oauth login failed: %w", err))
			}

			login, err := auth.ValidateToken(ctx, prov)
			if err != nil {
				return exitcode.New(exitcode.ErrAPI, fmt.Errorf("token validation failed: %w", err))
			}

			_ = accessToken
			cmd.Printf("Successfully authenticated with %s as %s\n", provider, login)
			return nil
		},
	}

	cmd.Flags().StringVarP(&provider, "provider", "p", string(auth.ProviderGitHub), "provider: github")
	cmd.Flags().StringP("token", "t", "", "authentication token (PAT)")
	cmd.Flags().StringVar(&clientID, "client-id", "", "OAuth client ID (defaults to GitHub CLI client ID)")

	return cmd
}

func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		Long:  `Display authentication status for all providers.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			providers := []auth.Provider{auth.ProviderGitHub}

			for _, provider := range providers {
				if auth.IsLoggedIn(provider) {
					cmd.Printf("%s: authenticated\n", provider)
				} else {
					cmd.Printf("%s: not authenticated\n", provider)
				}
			}

			return nil
		},
	}
}

func newAuthLogoutCmd() *cobra.Command {
	var provider string

	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove stored credentials",
		Long:  `Delete the stored authentication token for a provider.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			prov, err := auth.ParseProvider(provider)
			if err != nil {
				return err
			}

			if err := auth.Logout(prov); err != nil {
				return fmt.Errorf("failed to logout: %w", err)
			}
			cmd.Printf("Successfully logged out from %s\n", provider)
			return nil
		},
	}

	cmd.Flags().StringVarP(&provider, "provider", "p", string(auth.ProviderGitHub), "provider: github")

	return cmd
}

func newAuthWhoamiCmd() *cobra.Command {
	var provider string

	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Show current user",
		Long:  `Display the current authenticated user for a provider.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			prov, err := auth.ParseProvider(provider)
			if err != nil {
				return exitcode.New(exitcode.ErrInvalidUsage, err)
			}

			login, err := auth.ValidateToken(cmd.Context(), prov)
			if err != nil {
				return exitcode.New(exitcode.ErrNotAuthenticated, fmt.Errorf("not authenticated or token invalid: run 'kit auth login %s'", provider))
			}

			cmd.Printf("%s\n", login)
			return nil
		},
	}

	cmd.Flags().StringVarP(&provider, "provider", "p", string(auth.ProviderGitHub), "provider: github")

	return cmd
}
