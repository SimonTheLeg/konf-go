package cmd

import (
	"io"

	"github.com/simontheleg/konf-go/config"
	"github.com/simontheleg/konf-go/log"
	"github.com/simontheleg/konf-go/utils"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var (
	konfDir string
	silent  bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "konf",
	Short: "Root Command",
	Long: `
konf is a lightweight kubeconfig manager

Before switchting between kubeconfigs make sure to import them via 'konf import'
Afterwards switch between different kubeconfigs via 'konf set'
	`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	initPersistentFlags()

	if err := initConfig(); err != nil {
		return err
	}

	// addCommands needs to be run after the config has been initialized!
	initCommands()

	// make sure the default directories exist for the sub-commands
	if err := utils.EnsureDir(afero.NewOsFs()); err != nil {
		return err
	}

	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}

// initialize flags that are valid for all commands
func initPersistentFlags() {
	rootCmd.PersistentFlags().StringVar(&konfDir, "konf-dir", "", "konfs directory for kubeconfigs and tracking active konfs (default is $HOME/.kube/konfs)")
	rootCmd.PersistentFlags().BoolVar(&silent, "silent", false, "suppress log output if set to true (default is false)")
}

func initConfig() error {
	conf, err := config.ConfFromHomeDir()
	if err != nil {
		return err
	}

	if konfDir != "" {
		conf.KonfDir = konfDir
	}
	if silent {
		conf.Silent = silent
		log.InitLogger(io.Discard, io.Discard)
	}

	config.InitWithOverrides(conf)
	return nil
}

func initCommands() {
	rootCmd.AddCommand(cleanupCmd)
	rootCmd.AddCommand(newCompletionCmd().cmd)
	rootCmd.AddCommand(newDeleteCommand().cmd)
	rootCmd.AddCommand(newImportCmd().cmd)
	rootCmd.AddCommand(newNamespaceCmd().cmd)
	rootCmd.AddCommand(newSetCommand().cmd)
	rootCmd.AddCommand(newShellwrapperCmd().cmd)
	rootCmd.AddCommand(newVersionCommand().cmd)
}
