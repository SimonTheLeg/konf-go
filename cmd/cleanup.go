package cmd

import (
	"errors"
	"io/fs"
	"os"
	"strconv"

	"github.com/mitchellh/go-ps"
	"github.com/simontheleg/konf-go/config"
	log "github.com/simontheleg/konf-go/log"
	"github.com/simontheleg/konf-go/utils"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// cleanupCmd represents the cleanup command
var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Cleanup inactive kubeconfigs",
	Long: `This command cleans up any unused active configs (stored in konfDir/active).
An active config is considered unused when no process points to it anymore`,
	RunE: func(cmd *cobra.Command, args []string) error {

		fs := afero.NewOsFs()
		err := cleanLeftOvers(fs)
		if err != nil {
			return err
		}

		err = selfClean(fs)
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

	fpath := utils.IDFromProcessID(pid).ActivePath()
	err := f.Remove(fpath)

	if errors.Is(err, fs.ErrNotExist) {
		log.Info("current konf '%s' was already deleted, nothing to self-cleanup\n", fpath)
		return nil
	}

	if err != nil {
		return err
	}

	return nil
}

// cleanLeftOvers should look through the list of all processes that are available
// and clean up any files that are not in use any more. It's main purpose is to clean-up
// any leftovers that can occur if a previous session was not cleaned up nicely. This is
// necessary as we cannot tell a user that a selfClean has failed if they close the shell
// session before
func cleanLeftOvers(f afero.Fs) error {
	konfs, err := afero.ReadDir(f, config.ActiveDir())

	if err != nil {
		return err
	}

	for _, konf := range konfs {
		// We need to trim of the .yaml file extension to get to the PID
		konfID := utils.IDFromFileInfo(konf)
		pid, err := strconv.Atoi(string(konfID))
		if err != nil {
			log.Warn("file '%s' could not be converted into an int, and therefore cannot be a valid process id. Skip for cleanup", konf.Name())
			continue
		}

		p, err := ps.FindProcess(pid)
		if err != nil {
			return err
		}

		if p == nil {
			err := f.Remove(konfID.ActivePath())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(cleanupCmd)
}
