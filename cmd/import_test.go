package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/simontheleg/konf-go/konf"
	"github.com/simontheleg/konf-go/store"
	"github.com/simontheleg/konf-go/testhelper"
	"github.com/spf13/afero"
)

func TestImport(t *testing.T) {
	storeDir := "./konf/store"
	activeDir := "./konf/active"
	fm := testhelper.FilesystemManager{Storedir: storeDir, Activedir: activeDir}
	var determineConfigsCalled int
	var writeConfigCalledCount int
	var deleteOriginalConfigCalled int
	var filesForDirCalled int
	// using just a wrapper here instead of a full mock, makes testing it slightly easier
	var wrapDetermineConfig = func(r io.Reader) ([]*konf.Konfig, error) {
		determineConfigsCalled++
		return konf.KonfsFromKubeconfig(r)
	}
	var wrapFilesForDir = func(sm *store.Storemanager, s string) ([]*FileWithPath, error) {
		filesForDirCalled++
		return filesForDir(sm, s)
	}
	var mockWriteConfig = func(*konf.Konfig) (string, error) { writeConfigCalledCount++; return "", nil }
	var mockDeleteOriginalConfig = func(*store.Storemanager, string) error { deleteOriginalConfigCalled++; return nil }

	type expCalls struct {
		DetermineConfigs     int
		WriteConfig          int
		DeleteOriginalConfig int
		FilesForDir          int
	}
	tt := map[string]struct {
		args      []string
		fsCreator func() afero.Fs
		expErr    error
		moveFlag  bool
		expCalls
	}{
		"single file, single context": {
			[]string{"./konf/store/dev-eu_dev-eu-1.yaml"},
			testhelper.FSWithFiles(fm.StoreDir, fm.SingleClusterSingleContextEU),
			nil,
			false,
			expCalls{DetermineConfigs: 1, WriteConfig: 1, FilesForDir: 1},
		},
		"single file, empty context": {
			[]string{"./konf/store/no-context.yaml"},
			testhelper.FSWithFiles(fm.StoreDir, fm.KonfWithoutContext),
			fmt.Errorf("no contexts found in the following file(s):\n\t- \"konf/store/no-context.yaml\"\n"),
			false,
			expCalls{DetermineConfigs: 1, WriteConfig: 0, FilesForDir: 1},
		},
		"single file, move flag provided": {
			[]string{"./konf/store/dev-eu_dev-eu-1.yaml"},
			testhelper.FSWithFiles(fm.StoreDir, fm.SingleClusterSingleContextEU),
			nil,
			true,
			expCalls{DetermineConfigs: 1, WriteConfig: 1, DeleteOriginalConfig: 1, FilesForDir: 1},
		},
		"directory with single file": {
			[]string{"./konf/store"},
			testhelper.FSWithFiles(fm.StoreDir, fm.SingleClusterSingleContextEU),
			nil,
			false,
			expCalls{DetermineConfigs: 1, WriteConfig: 1, DeleteOriginalConfig: 0, FilesForDir: 1},
		},
		"directory with multiple files": {
			[]string{"./konf/store"},
			testhelper.FSWithFiles(fm.StoreDir, fm.SingleClusterSingleContextEU, fm.SingleClusterSingleContextASIA),
			nil,
			false,
			expCalls{DetermineConfigs: 2, WriteConfig: 2, DeleteOriginalConfig: 0, FilesForDir: 1},
		},
		"directory with multiple files, move flag provided": {
			[]string{"./konf/store"},
			testhelper.FSWithFiles(fm.StoreDir, fm.SingleClusterSingleContextEU, fm.SingleClusterSingleContextASIA),
			nil,
			true,
			expCalls{DetermineConfigs: 2, WriteConfig: 2, DeleteOriginalConfig: 2, FilesForDir: 1},
		},
		"directory with multiple files, empty context": {
			[]string{"./konf/store"},
			testhelper.FSWithFiles(fm.StoreDir, fm.KonfWithoutContext, fm.KonfWithoutContext2),
			fmt.Errorf("no contexts found in the following file(s):\n\t- \"konf/store/no-context-2.yaml\"\n\t- \"konf/store/no-context.yaml\"\n"),
			false,
			expCalls{DetermineConfigs: 2, WriteConfig: 0, FilesForDir: 1},
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			determineConfigsCalled = 0
			writeConfigCalledCount = 0
			deleteOriginalConfigCalled = 0
			filesForDirCalled = 0
			sm := &store.Storemanager{Activedir: activeDir, Storedir: storeDir, Fs: tc.fsCreator()}

			icmd := newImportCmd()
			icmd.sm = sm
			icmd.determineConfigs = wrapDetermineConfig
			icmd.writeConfig = mockWriteConfig
			icmd.deleteOriginalConfig = mockDeleteOriginalConfig
			icmd.filesForDir = wrapFilesForDir
			icmd.move = tc.moveFlag
			cmd := icmd.cmd

			// TODO unfortunately I was not able to use ExecuteC here as this would run
			// the cobra.OnInitialize, which sets the filesystem to OS. It should be investigated
			// if there is another way
			err := cmd.RunE(cmd, tc.args)
			if !testhelper.EqualError(tc.expErr, err) {
				t.Errorf("Exp error %q, got %q", tc.expErr, err)
			}

			if tc.expCalls.DetermineConfigs != determineConfigsCalled {
				t.Errorf("Exp DetermineConfigsCalled to be %d, but got %d", tc.expCalls.DetermineConfigs, determineConfigsCalled)
			}

			if tc.expCalls.WriteConfig != writeConfigCalledCount {
				t.Errorf("Exp WriteConfigCalled to be %d, but got %d", tc.expCalls.WriteConfig, writeConfigCalledCount)
			}

			if tc.expCalls.DeleteOriginalConfig != deleteOriginalConfigCalled {
				t.Errorf("Exp DeleteOriginalConfigCalled to be %d, but got %d", tc.expCalls.DeleteOriginalConfig, deleteOriginalConfigCalled)
			}

			if tc.expCalls.FilesForDir != filesForDirCalled {
				t.Errorf("Exp FilesForDirCalled to be %d, but got %d", tc.expCalls.FilesForDir, filesForDirCalled)
			}

		})
	}
}

func TestDeleteOriginalConfig(t *testing.T) {
	fpath := "/dir/original-file.yaml"

	f := afero.NewMemMapFs()
	afero.WriteFile(f, fpath, nil, 0664)
	sm := &store.Storemanager{Fs: f} // for this simple case we do not need to set ActiveDir and StoreDir

	if err := deleteOriginalConfig(sm, fpath); err != nil {
		t.Fatalf("Could not delete original kubeconfig %q: '%v'", fpath, err)
	}

	if _, err := f.Stat(fpath); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Expected error to be FileNotFound, but got %v", err)
	}
}

// cases to handle:
// - dir with closing slash
// - dir without closing slash
// - a file
// - dir without any files
func TestFilesForDir(t *testing.T) {
	storeDir := "./konf/store"
	activeDir := "./konf/active"
	fm := testhelper.FilesystemManager{Storedir: storeDir, Activedir: activeDir}
	f := testhelper.FSWithFiles(fm.DSStore, fm.MultiClusterMultiContext, fm.SingleClusterSingleContextEU, fm.SingleClusterSingleContextASIA)()
	sm := &store.Storemanager{Activedir: activeDir, Storedir: storeDir, Fs: f}

	tt := map[string]struct {
		path   string
		expRes []string
	}{
		"dir with hidden files, slash path": {
			path: "./konf/store/",
			expRes: []string{
				"konf/store/multi_multi_konf.yaml",
				"konf/store/dev-eu_dev-eu-1.yaml",
				"konf/store/dev-asia_dev-asia-1.yaml",
			},
		},
		"dir with hidden files, no slash path": {
			path: "./konf/store",
			expRes: []string{
				"konf/store/multi_multi_konf.yaml",
				"konf/store/dev-eu_dev-eu-1.yaml",
				"konf/store/dev-asia_dev-asia-1.yaml",
			},
		},
		"single file": {
			path: "./konf/store/dev-eu_dev-eu-1.yaml",
			expRes: []string{
				"konf/store/dev-eu_dev-eu-1.yaml",
			},
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			files, err := filesForDir(sm, tc.path)
			if err != nil {
				t.Fatal(err)
			}

			res := make([]string, len(files))
			for i, file := range files {
				res[i] = file.FilePath
			}

			sort.Strings(res)
			sort.Strings(tc.expRes)

			if !cmp.Equal(res, tc.expRes) {
				t.Errorf("Exp and given filepaths differ:\n '%s'", cmp.Diff(res, tc.expRes))
			}

		})
	}

}
