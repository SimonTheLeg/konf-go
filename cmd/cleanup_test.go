package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"syscall"
	"testing"

	"github.com/simontheleg/konf-go/konf"
	"github.com/simontheleg/konf-go/store"
	"github.com/simontheleg/konf-go/testhelper"
	"github.com/simontheleg/konf-go/utils"
	"github.com/spf13/afero"
)

func TestSelfClean(t *testing.T) {
	activeDir := "./konf/active"
	storeDir := "./konf/store"

	ppid := os.Getppid()

	tt := map[string]struct {
		Fs          afero.Fs
		ExpError    error
		ExpFiles    []string
		NotExpFiles []string
	}{
		"PID FS": {
			ppidFS(activeDir),
			nil,
			[]string{activeDir + "/abc", activeDir + "/1234"},
			[]string{activeDir + "/" + fmt.Sprint(ppid) + ".yaml"},
		},
		"PID file deleted by external source": {
			ppidFileMissing(activeDir),
			nil,
			[]string{activeDir + "/abc", activeDir + "/1234"},
			[]string{},
		},
		// Unfortunately it was not possible with afero memFS to test what happens if
		// someone changes the active dir permissions and we cannot delete it anymore.
		// Apparently in the memFS afero can just delete these files, regardless of
		// permissions :D
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {

			sm := &store.Storemanager{Fs: tc.Fs, Activedir: activeDir, Storedir: storeDir}
			err := selfClean(sm)

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

func ppidFS(activeDir string) afero.Fs {
	ppid := os.Getppid()
	fs := ppidFileMissing(activeDir)
	sm := testhelper.SampleKonfManager{}
	afero.WriteFile(fs, activeDir+"/"+fmt.Sprint(ppid), []byte(sm.SingleClusterSingleContextEU()), utils.KonfPerm)
	return fs
}

func ppidFileMissing(activeDir string) afero.Fs {
	fs := afero.NewMemMapFs()
	sm := testhelper.SampleKonfManager{}
	afero.WriteFile(fs, activeDir+"/abc", []byte("I am not even a kubeconfig, what am I doing here?"), utils.KonfPerm)
	afero.WriteFile(fs, activeDir+"/1234", []byte(sm.SingleClusterSingleContextEU()), utils.KonfPerm)
	return fs
}

func TestCleanLeftOvers(t *testing.T) {

	tt := map[string]struct {
		Setup  func(t *testing.T, sm *store.Storemanager) ([]*exec.Cmd, []*exec.Cmd)
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
			sm := &store.Storemanager{Fs: afero.NewMemMapFs(), Activedir: "./konf/active", Storedir: "./konf/store"}
			cmdsRunning, cmdsStopped := tc.Setup(t, sm)

			t.Cleanup(func() {
				cleanUpRunningCmds(t, cmdsRunning)
			})

			err := cleanLeftOvers(sm)

			if !errors.Is(err, tc.ExpErr) {
				t.Errorf("Want error '%s', got '%s'", tc.ExpErr, err)
			}

			for _, cmd := range cmdsRunning {
				id := konf.IDFromProcessID(cmd.Process.Pid)
				fpath := sm.ActivePathFromID(id)
				_, err := sm.Fs.Stat(fpath)

				if err != nil {
					if errors.Is(err, fs.ErrNotExist) {
						t.Errorf("Exp file '%s' to be present, but it is not", fpath)
					} else {
						t.Fatalf("Unexpected error occurred: '%s'", err)
					}
				}
			}

			for _, cmd := range cmdsStopped {
				id := konf.IDFromProcessID(cmd.Process.Pid)
				fpath := sm.ActivePathFromID(id)
				_, err := sm.Fs.Stat(fpath)

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

func mixedFSWithAllProcs(t *testing.T, sm *store.Storemanager) (cmdsRunning []*exec.Cmd, cmdsStopped []*exec.Cmd) {
	// we are simulating other instances of konf here
	numOfConfs := 3

	skm := testhelper.SampleKonfManager{}

	for i := 1; i <= numOfConfs; i++ {
		// set sleep to an extremely high number as the argument "infinity" does not exist in all versions of the util
		cmd := exec.Command("sleep", "315360000") // aka 10 years. Should be long enough for the unit test to finish ;)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		err := cmd.Start()
		if err != nil {
			t.Fatal(err)
		}
		pid := cmd.Process.Pid
		cmdsRunning = append(cmdsRunning, cmd)
		afero.WriteFile(sm.Fs, sm.ActivePathFromID(konf.IDFromProcessID(pid)), []byte(skm.SingleClusterSingleContextEU()), utils.KonfPerm)
	}

	return cmdsRunning, nil
}

// returns a state where there are more fs files than cmds
func mixedFSIncompleteProcs(t *testing.T, sm *store.Storemanager) (cmdsRunning []*exec.Cmd, cmdsStopped []*exec.Cmd) {
	cmdsRunning, cmdsStopped = mixedFSWithAllProcs(t, sm)

	cmdToKill := cmdsRunning[0]
	origPID := cmdToKill.Process.Pid
	err := cmdToKill.Process.Kill()
	if err != nil {
		t.Fatal(err)
	}

	// we need to call release here, as otherwise our process will have received the signal, but
	// the kernel will still be waiting for the parent to send wait, thus turning our process into a
	// zombie. Zombies are unfortunately a problem as they still have a PID and therefore mess with
	// funcs like cleanLeftOvers
	// err = cmdToKill.Process.Release()
	_, err = cmdToKill.Process.Wait()
	if err != nil {
		t.Fatal(err)
	}

	// Release will set the PID to -1. Therefore we need to set the PID back, so
	// we can use the original PID in our tests
	cmdToKill.Process.Pid = origPID

	cmdsStopped = append(cmdsStopped, cmdsRunning[0])
	cmdsRunning = cmdsRunning[1:]

	return cmdsRunning, cmdsStopped
}

func mixedFSDirtyDir(t *testing.T, sm *store.Storemanager) (cmdsRunning []*exec.Cmd, cmdsStopped []*exec.Cmd) {
	cmdsRunning, cmdsStopped = mixedFSIncompleteProcs(t, sm)

	id := konf.KonfID("/not-a-valid-process-id")
	afero.WriteFile(sm.Fs, sm.ActivePathFromID(id), []byte{}, utils.KonfPerm)

	return cmdsRunning, cmdsStopped

}

func emptyFS(t *testing.T, sm *store.Storemanager) (cmdsRunning []*exec.Cmd, cmdsStopped []*exec.Cmd) {
	// no op
	return nil, nil
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
