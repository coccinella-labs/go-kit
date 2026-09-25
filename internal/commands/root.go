package commands

import (
	"github.com/coccinella-labs/go-kit/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate completion scripts",
		Long:  `Generate shell completion scripts for kit.`,
	}

	cmd.AddCommand(newCompletionBashCmd())
	cmd.AddCommand(newCompletionZshCmd())
	cmd.AddCommand(newCompletionFishCmd())

	return cmd
}

func newCompletionBashCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bash",
		Short: "Generate bash completion script",
		Long:  `Generate bash completion script for kit.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenBashCompletion(cmd.OutOrStdout())
		},
	}
}

func newCompletionZshCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "zsh",
		Short: "Generate zsh completion script",
		Long:  `Generate zsh completion script for kit.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenZshCompletion(cmd.OutOrStdout())
		},
	}
}

func newCompletionFishCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "fish",
		Short: "Generate fish completion script",
		Long:  `Generate fish completion script for kit.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenFishCompletion(cmd.OutOrStdout(), false)
		},
	}
}

var (
	cfgFile string
	verbose bool
)

func NewRootCmd(version string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "kit",
		Short:   "A lightweight developer CLI.",
		Long:    ``,
		Version: version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initializeConfig()
		},
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kit/kit.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	rootCmd.AddCommand(newVersionCmd(version))
	rootCmd.AddCommand(newAuthCmd())
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newGithubCmd())
	rootCmd.AddCommand(newContextCmd())
	rootCmd.AddCommand(newCompletionCmd())
	rootCmd.AddCommand(newDoctorCmd())
	rootCmd.AddCommand(newUpdateCmd())

	return rootCmd
}

func initializeConfig() error {
	if cfgFile != "" {
		config.ConfigPath = cfgFile
		config.ConfigName = ""
		config.ConfigType = ""
	}
	return nil
}

func GetConfig() (*viper.Viper, error) {
	return config.GetConfig()
}

func Execute(version string) error {
	return NewRootCmd(version).Execute()
}
