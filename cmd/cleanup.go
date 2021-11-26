package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// TODO current idea: Make cleanup idempotent
// For this cleanup would need to be able to list all processes that are running

// cleanupCmd represents the cleanup command
var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "cleanup inactive kubeconfigs",
	Long: `This command cleans up any unused active configs (stored in konfDir/active).
An active config is considered unused when no process points to it anymore`,
	RunE: func(cmd *cobra.Command, args []string) error {

		fs := afero.NewOsFs()
		err := selfClean(fs)
		if err != nil {
			return err
		}

		return nil
	},
}

// selfClean should just find its parent process and delete that file
// it is required as the idempotent clean would delete all files that
// do not belong to any process anymore, but of course the current process
// is still running at this time
func selfClean(f afero.Fs) error {
	pid := os.Getppid()

	fpath := viper.GetString("konfActiveList") + "/" + fmt.Sprint(pid)
	err := f.Remove(fpath)

	if errors.Is(err, fs.ErrNotExist) {
		log.Printf("current konf '%s' was already deleted, nothing to self-cleanup\n", fpath)
		return nil
	}

	if err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(cleanupCmd)
}
