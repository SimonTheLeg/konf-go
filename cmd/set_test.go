package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/manifoldco/promptui"
	"github.com/simontheleg/konf-go/config"
	"github.com/simontheleg/konf-go/konf"
	"github.com/simontheleg/konf-go/prompt"
	"github.com/simontheleg/konf-go/store"
	"github.com/simontheleg/konf-go/testhelper"
	"github.com/simontheleg/konf-go/utils"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func TestSelectLastKonf(t *testing.T) {
	storeDir := "./konf/store"
	activeDir := "./konf/active"
	latestKonfPath := "./konf/latestkonf" // it is fine to use an imaginary file location here
	fm := testhelper.FilesystemManager{Storedir: storeDir, Activedir: activeDir, LatestKonfPath: latestKonfPath}

	tt := map[string]struct {
		FSCreator func() afero.Fs
		ExpID     konf.KonfID
		ExpError  error
	}{
		"latestKonf set": {
			FSCreator: testhelper.FSWithFiles(fm.LatestKonf),
			ExpID:     "context_cluster",
			ExpError:  nil,
		},
		"no latestKonf": {
			FSCreator: testhelper.FSWithFiles(),
			ExpID:     "",
			ExpError:  fmt.Errorf("could not select latest konf, because no konf was yet set"),
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			id, err := idOfLatestKonf(tc.FSCreator())

			if !testhelper.EqualError(tc.ExpError, err) {
				t.Errorf("Want error %q, got %q", tc.ExpError, err)
			}

			if tc.ExpID != id {
				t.Errorf("Want ID %q, got %q", tc.ExpID, id)
			}
		})
	}
}

func TestCompleteSet(t *testing.T) {
	// since cobra takes care of the majority of the complexity (like parsing out results that don't match completion start),
	// we only need to test regular cases
	storeDir := "./konf/store"
	activeDir := "./konf/active"
	fm := testhelper.FilesystemManager{Storedir: storeDir, Activedir: activeDir}

	tt := map[string]struct {
		fsCreator    func() afero.Fs
		expComp      []string
		expCompDirec cobra.ShellCompDirective
	}{
		"normal results": {
			testhelper.FSWithFiles(fm.StoreDir, fm.SingleClusterSingleContextASIA, fm.SingleClusterSingleContextEU),
			[]string{"dev-asia_dev-asia-1", "dev-eu_dev-eu-1"},
			cobra.ShellCompDirectiveNoFileComp,
		},
		"no results": {
			testhelper.FSWithFiles(fm.StoreDir),
			[]string{},
			cobra.ShellCompDirectiveNoFileComp,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			fs := tc.fsCreator()
			sm := &store.Storemanager{Activedir: activeDir, Storedir: storeDir, Fs: fs}

			scmd := newSetCommand()
			scmd.sm = sm

			res, compdirec := scmd.completeSet(scmd.cmd, []string{}, "")

			if !cmp.Equal(res, tc.expComp) {
				t.Errorf("Exp and given comps differ: \n '%s'", cmp.Diff(tc.expComp, res))
			}

			if compdirec != tc.expCompDirec {
				t.Errorf("Exp compdirec %q, got %q", tc.expCompDirec, compdirec)
			}
		})
	}
}

func TestSaveLatestKonf(t *testing.T) {
	expFile := "./konf/latestkonf"
	expID := konf.KonfID("context_cluster")

	f := afero.NewMemMapFs()
	err := saveLatestKonf(f, expID)
	if err != nil {
		t.Errorf("Could not save last konf: %q", err)
	}
	finf, err := f.Stat(expFile)
	if err != nil {
		t.Errorf("Could not stat file: %q", err)
	}
	if finf == nil {
		t.Errorf("Exp file %q to be present, but it isnt", expFile)
	}
	id, _ := afero.ReadFile(f, expFile)
	if konf.KonfID(id) != expID {
		t.Errorf("Exp id to be %q but is %q", expID, id)
	}
}

func TestSetContext(t *testing.T) {
	storeDir := config.StoreDir()
	ppid := os.Getppid()
	sm := testhelper.SampleKonfManager{}

	tt := map[string]struct {
		InID        konf.KonfID
		StoreExists bool
		ExpErr      error
		ExpKonfPath string
	}{
		"normal write": {
			"dev-eu_dev-eu",
			true,
			nil,
			konf.IDFromProcessID(ppid).ActivePath(),
		},
		"invalid id": {
			"i-am-invalid",
			false,
			fs.ErrNotExist,
			"",
		},
	}

	for name, tc := range tt {

		t.Run(name, func(t *testing.T) {
			f := afero.NewMemMapFs()

			if tc.StoreExists {
				afero.WriteFile(f, storeDir+"/"+string(tc.InID)+".yaml", []byte(sm.SingleClusterSingleContextEU()), utils.KonfPerm)
			}

			resKonfPath, resError := setContext(tc.InID, f)

			if !errors.Is(resError, tc.ExpErr) {
				t.Errorf("Want error '%s', got '%s'", tc.ExpErr, resError)
			}

			if resKonfPath != tc.ExpKonfPath {
				t.Errorf("Want konfPath '%s', got '%s'", tc.ExpKonfPath, resKonfPath)
			}

			if tc.ExpKonfPath != "" {
				_, err := f.Stat(tc.ExpKonfPath)
				if err != nil {
					if errors.Is(err, fs.ErrNotExist) {
						t.Errorf("Exp file %q to be present, but it is not", tc.ExpKonfPath)
					} else {
						t.Fatalf("Unexpected error occurred: '%s'", err)
					}
				}
				res, err := afero.ReadFile(f, tc.ExpKonfPath)
				if err != nil {
					t.Errorf("Wanted to read file %q, but failed: %q", tc.ExpKonfPath, err)
				}
				if string(res) != sm.SingleClusterSingleContextEU() {
					t.Errorf("Exp content %q, got %q", res, sm.SingleClusterSingleContextEU())
				}
			}
		})

	}
}

func TestSelectContext(t *testing.T) {
	storeDir := "./konf/store"
	activeDir := "./konf/active"
	fm := testhelper.FilesystemManager{Storedir: storeDir, Activedir: activeDir}
	f := testhelper.FSWithFiles(fm.StoreDir, fm.SingleClusterSingleContextEU, fm.SingleClusterSingleContextASIA)()
	sm := &store.Storemanager{Fs: f, Activedir: config.ActiveDir(), Storedir: config.StoreDir()}

	// cases
	// - invalid selection
	// - prompt failure
	tt := map[string]struct {
		pf     prompt.RunFunc
		expID  konf.KonfID
		expErr error
	}{
		"select asia": {
			func(s *promptui.Select) (int, error) { return 0, nil },
			"dev-asia_dev-asia-1",
			nil,
		},
		"select eu": {
			func(s *promptui.Select) (int, error) { return 1, nil },
			"dev-eu_dev-eu-1",
			nil,
		},
		"prompt failure": {
			func(s *promptui.Select) (int, error) { return 1, fmt.Errorf("err") },
			"",
			fmt.Errorf("err"),
		},
		"invalid selection": {
			func(s *promptui.Select) (int, error) { return 2, nil },
			"",
			fmt.Errorf("invalid selection 2"),
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			res, err := selectSingleKonf(sm, tc.pf)

			if !testhelper.EqualError(err, tc.expErr) {
				t.Errorf("Exp err %q, got %q", tc.expErr, err)
			}

			if res != tc.expID {
				t.Errorf("Exp id %q, got %q", tc.expID, res)
			}
		})
	}
}
