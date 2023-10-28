package cmd

import (
	"errors"
	"io/fs"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/simontheleg/konf-go/konf"
	"github.com/simontheleg/konf-go/prompt"
	"github.com/simontheleg/konf-go/store"
	"github.com/simontheleg/konf-go/testhelper"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"k8s.io/utils/strings/slices"
)

func TestDeleteKonfWithID(t *testing.T) {
	storeDir := "./konf/store"
	activeDir := "./konf/active"
	fm := testhelper.FilesystemManager{Storedir: storeDir, Activedir: activeDir}

	tt := map[string]struct {
		fsCreator   func() afero.Fs
		idToDelete  konf.KonfID
		expError    error
		expFiles    []string
		notExpFiles []string
	}{
		"file was found": {
			fsCreator:   testhelper.FSWithFiles(fm.SingleClusterSingleContextEU, fm.SingleClusterSingleContextASIA),
			idToDelete:  "dev-eu_dev-eu-1",
			expError:    nil,
			expFiles:    []string{storeDir + "/dev-asia_dev-asia-1.yaml"},
			notExpFiles: []string{storeDir + "/dev-eu_dev-eu-1.yaml"},
		},
		"file was not found": {
			fsCreator:   testhelper.FSWithFiles(fm.SingleClusterSingleContextASIA),
			idToDelete:  "dev-eu_dev-eu-1",
			expError:    fs.ErrNotExist,
			expFiles:    []string{storeDir + "/dev-asia_dev-asia-1.yaml"},
			notExpFiles: []string{},
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {

			fsm := tc.fsCreator()
			sm := &store.Storemanager{Fs: fsm, Activedir: activeDir, Storedir: storeDir}

			err := deleteKonfWithID(sm, tc.idToDelete)

			if !errors.Is(err, tc.expError) {
				t.Errorf("Exp err to be %q, got %q", tc.expError, err)
			}

			for _, f := range tc.expFiles {
				if _, err := fsm.Stat(f); err != nil {
					t.Errorf("Exp file %q to exist, but it does not", f)
				}
			}

			for _, s := range tc.notExpFiles {
				_, err := fsm.Stat(s)

				if err == nil {
					t.Errorf("Exp file '%s' to be deleted, but it still exists", s)
				}

				if err != nil && !errors.Is(err, fs.ErrNotExist) {
					t.Fatalf("An unexpected error has occurred: %q", err)
				}
			}

		})
	}
}

func TestIDsForGlobs(t *testing.T) {
	storeDir := "./konf/store"
	activeDir := "./konf/active"
	fm := testhelper.FilesystemManager{Storedir: storeDir, Activedir: activeDir}

	tt := map[string]struct {
		fsCreator func() afero.Fs
		patterns  []string
		expIDs    []string
		expError  error
	}{
		"single argument no glob": {
			fsCreator: testhelper.FSWithFiles(fm.SingleClusterSingleContextEU, fm.SingleClusterSingleContextASIA, fm.SingleClusterSingleContextEU2, fm.SingleClusterSingleContextASIA2),
			patterns:  []string{"dev-eu_dev-eu-1"},
			expIDs:    []string{"dev-eu_dev-eu-1"},
			expError:  nil,
		},
		"single argument glob": {
			fsCreator: testhelper.FSWithFiles(fm.SingleClusterSingleContextEU, fm.SingleClusterSingleContextASIA, fm.SingleClusterSingleContextEU2, fm.SingleClusterSingleContextASIA2),
			patterns:  []string{"dev-eu_dev-eu*"},
			expIDs:    []string{"dev-eu_dev-eu-1", "dev-eu_dev-eu-2"},
			expError:  nil,
		},
		"two arguments no glob": {
			fsCreator: testhelper.FSWithFiles(fm.SingleClusterSingleContextEU, fm.SingleClusterSingleContextASIA, fm.SingleClusterSingleContextEU2, fm.SingleClusterSingleContextASIA2),
			patterns:  []string{"dev-eu_dev-eu-1", "dev-asia_dev-asia-1"},
			expIDs:    []string{"dev-eu_dev-eu-1", "dev-asia_dev-asia-1"},
			expError:  nil,
		},
		"two arguments one glob": {
			fsCreator: testhelper.FSWithFiles(fm.SingleClusterSingleContextEU, fm.SingleClusterSingleContextASIA, fm.SingleClusterSingleContextEU2, fm.SingleClusterSingleContextASIA2),
			patterns:  []string{"dev-eu_dev-eu*", "dev-asia_dev-asia-1"},
			expIDs:    []string{"dev-eu_dev-eu-1", "dev-eu_dev-eu-2", "dev-asia_dev-asia-1"},
			expError:  nil,
		},
		"two arguments two globs": {
			fsCreator: testhelper.FSWithFiles(fm.SingleClusterSingleContextEU, fm.SingleClusterSingleContextASIA, fm.SingleClusterSingleContextEU2, fm.SingleClusterSingleContextASIA2),
			patterns:  []string{"dev-eu_dev-eu*", "dev-asia_dev-asia*"},
			expIDs:    []string{"dev-eu_dev-eu-1", "dev-eu_dev-eu-2", "dev-asia_dev-asia-1", "dev-asia_dev-asia-2"},
			expError:  nil,
		},
		"no match": {
			fsCreator: testhelper.FSWithFiles(fm.SingleClusterSingleContextEU, fm.SingleClusterSingleContextASIA, fm.SingleClusterSingleContextEU2, fm.SingleClusterSingleContextASIA2),
			patterns:  []string{"no-match"},
			expIDs:    []string{},
			expError:  &store.NoMatch{Pattern: "no-match"},
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {

			fs := tc.fsCreator()
			sm := &store.Storemanager{Activedir: activeDir, Storedir: storeDir, Fs: fs}

			res, err := idsForGlobs(sm, tc.patterns)

			if !testhelper.EqualError(tc.expError, err) {
				t.Errorf("Exp error %q, got %q", tc.expError, err)
			}

			var ids []string
			for _, r := range res {
				ids = append(ids, string(r))
			}

			sort.Strings(tc.expIDs)
			sort.Strings(ids)

			if !slices.Equal(tc.expIDs, ids) {
				t.Errorf("Exp ids to be %v, got %v", tc.expIDs, ids)
			}

		})
	}
}

func TestDelete(t *testing.T) {

	selectSingleKonfCalled := 0
	idsForGlobsCalled := 0
	deleteKonfWithIDCalled := 0

	var mockSelectSingleKonf = func(*store.Storemanager, prompt.RunFunc) (konf.KonfID, error) {
		selectSingleKonfCalled++
		return "id1", nil
	}

	var mockIDsForGlobs = func(*store.Storemanager, []string) ([]konf.KonfID, error) {
		idsForGlobsCalled++
		return []konf.KonfID{"id1", "id2", "id3"}, nil
	}

	var mockDeleteKonfWithID = func(*store.Storemanager, konf.KonfID) error {
		deleteKonfWithIDCalled++
		return nil
	}

	cmd := &deleteCmd{
		selectSingleKonf: mockSelectSingleKonf,
		idsForGlobs:      mockIDsForGlobs,
		deleteKonfWithID: mockDeleteKonfWithID,
	}

	tt := map[string]struct {
		args                      []string
		expSelectSingleKonfCalled int
		expIdsForGlobsCalled      int
		expDeleteKonfWithIDCalled int
	}{
		"select one": {
			args:                      []string{},
			expSelectSingleKonfCalled: 1,
			expIdsForGlobsCalled:      0,
			expDeleteKonfWithIDCalled: 1,
		},
		"multiple arguments supplied": {
			args:                      []string{"id1", "id2", "id3"},
			expSelectSingleKonfCalled: 0,
			expIdsForGlobsCalled:      1,
			expDeleteKonfWithIDCalled: 3,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {

			selectSingleKonfCalled = 0
			idsForGlobsCalled = 0
			deleteKonfWithIDCalled = 0

			err := cmd.delete(cmd.cmd, tc.args)

			if err != nil {
				t.Fatalf("An unexpected error occured: %v", err)
			}

			if tc.expSelectSingleKonfCalled != selectSingleKonfCalled {
				t.Errorf("Exp SelectSingleKonf to be called %d times, was called %d times", tc.expSelectSingleKonfCalled, selectSingleKonfCalled)
			}

			if tc.expIdsForGlobsCalled != idsForGlobsCalled {
				t.Errorf("Exp IDsForGlobsCalled to be called %d times, was called %d times", tc.expIdsForGlobsCalled, idsForGlobsCalled)
			}

			if tc.expDeleteKonfWithIDCalled != deleteKonfWithIDCalled {
				t.Errorf("Exp DeleteKonfWithID to be called %d times, was called %d times", tc.expDeleteKonfWithIDCalled, deleteKonfWithIDCalled)
			}

		})

	}

}

func TestCompleteDelete(t *testing.T) {
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

			dcmd := newDeleteCommand()
			dcmd.sm = sm
			dcmd.fetchconfs = sm.FetchAllKonfs

			res, compdirec := dcmd.completeDelete(dcmd.cmd, []string{}, "")

			if !cmp.Equal(res, tc.expComp) {
				t.Errorf("Exp and given comps differ: \n '%s'", cmp.Diff(tc.expComp, res))
			}

			if compdirec != tc.expCompDirec {
				t.Errorf("Exp compdirec %q, got %q", tc.expCompDirec, compdirec)
			}
		})
	}
}
