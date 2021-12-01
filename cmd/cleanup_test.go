package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"syscall"
	"testing"

	"github.com/simontheleg/konfig/utils"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

func TestSelfClean(t *testing.T) {
	utils.InitViper()
	konfActiveList := viper.GetString("konfActiveList")

	ppid := os.Getppid()

	tt := map[string]struct {
		Fs          afero.Fs
		ExpError    error
		ExpFiles    []string
		NotExpFiles []string
	}{
		"PID FS": {
			ppidFS(),
			nil,
			[]string{viper.GetString("konfActiveList") + "/abc", konfActiveList + "/1234"},
			[]string{konfActiveList + "/" + fmt.Sprint(ppid)},
		},
		"PID file deleted by external source": {
			ppidFileMissing(),
			nil,
			[]string{konfActiveList + "/abc", konfActiveList + "/1234"},
			[]string{},
		},
		// Unfortunately it was not possible with afero to test what happens if
		// someone changes the active dir permissions an we cannot delete it anymore
		// apparently in the memFS afero can just delete these files
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {

			err := selfClean(tc.Fs)

			if !utils.EqualError(err, tc.ExpError) {
				t.Errorf("Want error '%s', got '%s'", tc.ExpError, err)
			}

			for _, s := range tc.ExpFiles {
				if _, err := tc.Fs.Stat(s); err != nil {
					t.Errorf("Exp file '%s' to exist, but it does not", s)
				}
			}

			for _, s := range tc.NotExpFiles {
				_, err := tc.Fs.Stat(s)

				if err == nil {
					t.Errorf("Exp file '%s' to be deleted, but it still exists", s)
				}

				if err != nil && !errors.Is(err, fs.ErrNotExist) {
					t.Fatalf("An unexpected error has occured")
				}
			}
		})
	}
}

func ppidFS() afero.Fs {
	konfActiveList := viper.GetString("konfActiveList")
	ppid := os.Getppid()
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, konfActiveList+"/"+fmt.Sprint(ppid), []byte(singleClusterSingleContext), 0644)
	afero.WriteFile(fs, konfActiveList+"/abc", []byte("I am not event a kubeconfig, what am I doing here?"), 0644)
	afero.WriteFile(fs, konfActiveList+"/1234", []byte(singleClusterSingleContext), 0644)
	return fs
}

func ppidFileMissing() afero.Fs {
	konfActiveList := viper.GetString("konfActiveList")
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, konfActiveList+"/abc", []byte("I am not event a kubeconfig, what am I doing here?"), 0644)
	afero.WriteFile(fs, konfActiveList+"/1234", []byte(singleClusterSingleContext), 0644)
	return fs
}

func TestCleanLeftOvers(t *testing.T) {
	utils.InitViper()
	konfActiveList := viper.GetString("konfActiveList")

	tt := map[string]struct {
		Setup  func(t *testing.T) (afero.Fs, []*exec.Cmd, []*exec.Cmd)
		ExpErr error
	}{
		"all procs still running": {
			mixedFSWithAllProcs,
			nil,
		},
		"some procs have stopped": {
			mixedFSIncompleteProcs,
			nil,
		},
		"dirty dir": {
			mixedFSDirtyDir,
			nil,
		},
		"dir does not exist": {
			emptyFS,
			fs.ErrNotExist,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			f, cmdsRunning, cmdsStopped := tc.Setup(t)

			t.Cleanup(func() {
				cleanUpRunningCmds(t, cmdsRunning)
			})

			err := cleanLeftOvers(f)

			if !errors.Is(err, tc.ExpErr) {
				t.Errorf("Want error '%s', got '%s'", tc.ExpErr, err)
			}

			for _, cmd := range cmdsRunning {
				fpath := konfActiveList + "/" + fmt.Sprint(cmd.Process.Pid)
				_, err := f.Stat(fpath)

				if err != nil {
					if errors.Is(err, fs.ErrNotExist) {
						t.Errorf("Exp file '%s' to be present, but it is not", fpath)
					} else {
						t.Fatalf("Unexpected error occured: '%s'", err)
					}
				}
			}

			for _, cmd := range cmdsStopped {
				fpath := konfActiveList + "/" + fmt.Sprint(cmd.Process.Pid)
				_, err := f.Stat(fpath)

				if !errors.Is(err, fs.ErrNotExist) {
					t.Fatalf("Unexpected error occured: '%s'", err)
				}

				if err == nil {
					t.Errorf("Exp file '%s' to be deleted, but it is still present", fpath)
				}
			}

		})
	}

}

func mixedFSWithAllProcs(t *testing.T) (fs afero.Fs, cmdsRunning []*exec.Cmd, cmdsStopped []*exec.Cmd) {
	// we are simulating other instances of konf here
	numOfConfs := 3

	fs = afero.NewMemMapFs()

	for i := 1; i <= numOfConfs; i++ {
		// set sleep to an extremely high number as the argument "infinity" does not exist in all versions of the util
		cmd := exec.Command("sleep", "315360000") // aka 10 years. Should be long enough for the unit test to finish ;)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		err := cmd.Start()
		if err != nil {
			t.Fatalf(err.Error())
		}
		pid := cmd.Process.Pid
		cmdsRunning = append(cmdsRunning, cmd)
		afero.WriteFile(fs, viper.GetString("konfActiveList")+"/"+fmt.Sprint(pid), []byte(singleClusterSingleContext), 0644)
	}

	return fs, cmdsRunning, nil
}

// returns a state where there are more fs files than cmds
func mixedFSIncompleteProcs(t *testing.T) (fs afero.Fs, cmdsRunning []*exec.Cmd, cmdsStopped []*exec.Cmd) {
	fs, cmdsRunning, cmdsStopped = mixedFSWithAllProcs(t)

	cmdToKill := cmdsRunning[0]
	origPID := cmdToKill.Process.Pid
	err := cmdToKill.Process.Kill()
	if err != nil {
		t.Fatalf(err.Error())
	}

	// we need to call release here, as otherwise our process will have received the signal, but
	// the kernel will still be waiting for the parent to send wait, thus turning our process into a
	// zombie. Zombies are unfortunately a problem as they still have a PID and therefore mess with
	// funcs like cleanLeftOvers
	// err = cmdToKill.Process.Release()
	_, err = cmdToKill.Process.Wait()
	if err != nil {
		t.Fatalf(err.Error())
	}

	// Release will set the PID to -1. Therefore we need to set the PID back, so
	// we can use the original PID in our tests
	cmdToKill.Process.Pid = origPID

	cmdsStopped = append(cmdsStopped, cmdsRunning[0])
	cmdsRunning = cmdsRunning[1:]

	return fs, cmdsRunning, cmdsStopped
}

func mixedFSDirtyDir(t *testing.T) (fs afero.Fs, cmdsRunning []*exec.Cmd, cmdsStopped []*exec.Cmd) {
	fs, cmdsRunning, cmdsStopped = mixedFSIncompleteProcs(t)

	afero.WriteFile(fs, viper.GetString("konfActiveList")+"/not-a-valid-process-id", []byte{}, 0644)

	return fs, cmdsRunning, cmdsStopped

}

func emptyFS(t *testing.T) (fs afero.Fs, cmdsRunning []*exec.Cmd, cmdsStopped []*exec.Cmd) {
	fs = afero.NewMemMapFs()

	return fs, nil, nil
}

func cleanUpRunningCmds(t *testing.T, cmds []*exec.Cmd) {
	rogueProcesses := []*exec.Cmd{}
	for _, cmd := range cmds {

		err := cmd.Process.Kill()
		if err != nil {
			rogueProcesses = append(rogueProcesses, cmd)
		}

		_, err = cmd.Process.Wait()
		if err != nil {
			rogueProcesses = append(rogueProcesses, cmd)
		}
	}
	if len(rogueProcesses) != 0 {
		t.Fatalf("Cleanup went wrong, please manually check the following processes: %v", rogueProcesses)
	}
}
