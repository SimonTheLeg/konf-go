package cmd

import (
	"io"
	"os"

	"github.com/simontheleg/konf-go/config"
	"github.com/simontheleg/konf-go/log"
	"github.com/simontheleg/konf-go/utils"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
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
func Execute() {
	rootCmd.SetOut(os.Stderr)
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(wrapInit)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.konfig.yaml)")
	rootCmd.PersistentFlags().StringVar(&konfDir, "konfDir", "", "konfs directory for kubeconfigs and tracking active konfs (default is $HOME/.kube/konfs)")
	rootCmd.PersistentFlags().BoolVar(&silent, "silent", false, "suppress log output if set to true (default is false)")

}

// wrapInit is required as cobra.OnInitialize only accepts func() as interface
func wrapInit() {
	err := config.Init(cfgFile, konfDir)
	cobra.CheckErr(err)

	if silent {
		log.InitLogger(io.Discard, io.Discard)
	}

	err = utils.EnsureDir(afero.NewOsFs())
	cobra.CheckErr(err)
}
