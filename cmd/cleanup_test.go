package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"syscall"
	"testing"

	"github.com/simontheleg/konf-go/config"
	"github.com/simontheleg/konf-go/testhelper"
	"github.com/simontheleg/konf-go/utils"
	"github.com/spf13/afero"
)

func TestSelfClean(t *testing.T) {
	activeDir := config.ActiveDir()

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
			[]string{activeDir + "/abc", utils.KonfID("1234").ActivePath()},
			[]string{activeDir + "/" + fmt.Sprint(ppid) + ".yaml"},
		},
		"PID file deleted by external source": {
			ppidFileMissing(),
			nil,
			[]string{activeDir + "/abc", utils.KonfID("1234").ActivePath()},
			[]string{},
		},
		// Unfortunately it was not possible with afero to test what happens if
		// someone changes the active dir permissions an we cannot delete it anymore
		// apparently in the memFS afero can just delete these files
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {

			err := selfClean(tc.Fs)

			if !testhelper.EqualError(err, tc.ExpError) {
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
					t.Fatalf("An unexpected error has occurred")
				}
			}
		})
	}
}

func ppidFS() afero.Fs {
	ppid := os.Getppid()
	fs := ppidFileMissing()
	sm := testhelper.SampleKonfManager{}
	afero.WriteFile(fs, utils.IDFromProcessID(ppid).ActivePath(), []byte(sm.SingleClusterSingleContextEU()), utils.KonfPerm)
	return fs
}

func ppidFileMissing() afero.Fs {
	fs := afero.NewMemMapFs()
	sm := testhelper.SampleKonfManager{}
	afero.WriteFile(fs, config.ActiveDir()+"/abc", []byte("I am not even a kubeconfig, what am I doing here?"), utils.KonfPerm)
	afero.WriteFile(fs, utils.KonfID("1234").ActivePath(), []byte(sm.SingleClusterSingleContextEU()), utils.KonfPerm)
	return fs
}

func TestCleanLeftOvers(t *testing.T) {

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
				fpath := utils.IDFromProcessID(cmd.Process.Pid).ActivePath()
				_, err := f.Stat(fpath)

				if err != nil {
					if errors.Is(err, fs.ErrNotExist) {
						t.Errorf("Exp file '%s' to be present, but it is not", fpath)
					} else {
						t.Fatalf("Unexpected error occurred: '%s'", err)
					}
				}
			}

			for _, cmd := range cmdsStopped {
				fpath := utils.IDFromProcessID(cmd.Process.Pid).ActivePath()
				_, err := f.Stat(fpath)

				if !errors.Is(err, fs.ErrNotExist) {
					t.Fatalf("Unexpected error occurred: '%s'", err)
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
	sm := testhelper.SampleKonfManager{}

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
		afero.WriteFile(fs, utils.IDFromProcessID(pid).ActivePath(), []byte(sm.SingleClusterSingleContextEU()), utils.KonfPerm)
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

	afero.WriteFile(fs, utils.KonfID("/not-a-valid-process-id").ActivePath(), []byte{}, utils.KonfPerm)

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
