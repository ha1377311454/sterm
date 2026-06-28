package cmd

import (
	"fmt"
	"os"

	"github.com/ha1377311454/sterm/internal/config"
	"github.com/ha1377311454/sterm/internal/ui"
	"github.com/spf13/cobra"
)

var runtimeOptions config.Options

var rootCmd = &cobra.Command{
	Use:   "sterm",
	Short: "SSH connection manager with file transfer",
	Long:  "sterm — a terminal SSH manager with SFTP support and customizable themes",
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := ui.NewAppWithOptions(runtimeOptions)
		if err != nil {
			return err
		}
		return app.Run()
	},
}

func init() {
	rootCmd.Flags().StringVar(&runtimeOptions.ConfigDir, "config-dir", "", "config directory (default: platform-specific user config dir)")
	rootCmd.Flags().StringVar(&runtimeOptions.KeyFile, "key-file", "", "AES encryption key file for stored passwords")
	rootCmd.Flags().StringArrayVar(&runtimeOptions.ThemeDirs, "theme-dir", nil, "custom theme directory, can be specified multiple times")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
