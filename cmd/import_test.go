package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"testing"

	"github.com/simontheleg/konf-go/konf"
	"github.com/simontheleg/konf-go/testhelper"
	"github.com/spf13/afero"
)

func TestImport(t *testing.T) {
	fm := testhelper.FilesystemManager{}
	var determineConfigsCalled bool
	var writeConfigCalledCount int
	var deleteOriginalConfigCalled bool
	// using just a wrapper here instead of a full mock, makes testing it slightly easier
	var wrapDetermineConfig = func(r io.Reader) ([]*konf.Konfig, error) {
		determineConfigsCalled = true
		return konf.KonfsFromKubeconfig(r)
	}
	var mockWriteConfig = func(afero.Fs, *konf.Konfig) (string, error) { writeConfigCalledCount++; return "", nil }
	var mockDeleteOriginalConfig = func(afero.Fs, string) error { deleteOriginalConfigCalled = true; return nil }

	type ExpCalls struct {
		DetermineConfigs     bool
		WriteConfig          int
		DeleteOriginalConfig bool
	}
	tt := map[string]struct {
		Args      []string
		FsCreator func() afero.Fs
		ExpErr    error
		MoveFlag  bool
		ExpCalls
	}{
		"single context": {
			[]string{"./konf/store/dev-eu_dev-eu-1.yaml"},
			testhelper.FSWithFiles(fm.StoreDir, fm.SingleClusterSingleContextEU),
			nil,
			false,
			ExpCalls{DetermineConfigs: true, WriteConfig: 1},
		},
		"empty context": {
			[]string{"./konf/store/no-context.yaml"},
			testhelper.FSWithFiles(fm.StoreDir, fm.KonfWithoutContext),
			fmt.Errorf("no contexts found in file \"./konf/store/no-context.yaml\""),
			false,
			ExpCalls{DetermineConfigs: true, WriteConfig: 0},
		},
		"move flag provided": {
			[]string{"./konf/store/dev-eu_dev-eu-1.yaml"},
			testhelper.FSWithFiles(fm.StoreDir, fm.SingleClusterSingleContextEU),
			nil,
			true,
			ExpCalls{DetermineConfigs: true, WriteConfig: 1, DeleteOriginalConfig: true},
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			determineConfigsCalled = false
			writeConfigCalledCount = 0
			deleteOriginalConfigCalled = false
			fs := tc.FsCreator()

			icmd := newImportCmd()
			icmd.fs = fs
			icmd.determineConfigs = wrapDetermineConfig
			icmd.writeConfig = mockWriteConfig
			icmd.deleteOriginalConfig = mockDeleteOriginalConfig
			icmd.move = tc.MoveFlag
			cmd := icmd.cmd

			// TODO unfortunately I was not able to use ExecuteC here as this would run
			// the cobra.OnInitialize, which sets the filesystem to OS. It should be investigated
			// if there is another way
			err := cmd.RunE(cmd, tc.Args)
			if !testhelper.EqualError(tc.ExpErr, err) {
				t.Errorf("Exp error %q, got %q", tc.ExpErr, err)
			}

			if tc.ExpCalls.DetermineConfigs != determineConfigsCalled {
				t.Errorf("Exp DetermineConfigsCalled to be %t, but got %t", tc.ExpCalls.DetermineConfigs, determineConfigsCalled)
			}

			if tc.ExpCalls.WriteConfig != writeConfigCalledCount {
				t.Errorf("Exp WriteConfigCalled to be %d, but got %d", tc.ExpCalls.WriteConfig, writeConfigCalledCount)
			}

			if tc.ExpCalls.DeleteOriginalConfig != deleteOriginalConfigCalled {
				t.Errorf("Exp DeleteOriginalConfigCalled to be %t, but got %t", tc.ExpCalls.DeleteOriginalConfig, deleteOriginalConfigCalled)
			}

		})
	}
}

func TestDeleteOriginalConfig(t *testing.T) {
	fpath := "/dir/original-file.yaml"

	f := afero.NewMemMapFs()
	afero.WriteFile(f, fpath, nil, 0664)

	if err := deleteOriginalConfig(f, fpath); err != nil {
		t.Fatalf("Could not delete original kubeconfig %q: '%v'", fpath, err)
	}

	if _, err := f.Stat(fpath); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Expected error to be FileNotFound, but got %v", err)
	}
}
