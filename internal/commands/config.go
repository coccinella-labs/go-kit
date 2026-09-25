package commands

import (
	"fmt"

	"github.com/coccinella-labs/go-kit/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration commands",
		Long:  `View and modify kit configuration.`,
	}

	configCmd.AddCommand(newConfigListCmd())
	configCmd.AddCommand(newConfigSetCmd())
	configCmd.AddCommand(newConfigGetCmd())

	return configCmd
}

func newConfigListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configuration values",
		Long:  `Display the current configuration values.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			settings := cfg.AllSettings()
			if len(settings) == 0 {
				cmd.Println("No configuration values set.")
				return nil
			}

			for key, value := range settings {
				fmt.Fprintf(cmd.OutOrStdout(), "%s=%v\n", key, value)
			}
			return nil
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long:  `Set a configuration value in the kit config file.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			cfg, err := config.GetConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			cfg.Set(key, value)

			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			cmd.Printf("Set %s=%s\n", key, value)
			return nil
		},
	}
}

func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Long:  `Get a configuration value from the kit config.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			cfg, err := config.GetConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			value := cfg.Get(key)
			if value == nil {
				return fmt.Errorf("key not found: %s", key)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%v\n", value)
			return nil
		},
	}
}
