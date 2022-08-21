package store

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/simontheleg/konf-go/config"
	"github.com/simontheleg/konf-go/testhelper"
	"github.com/spf13/afero"
)

func TestFetchAllKonfs(t *testing.T) {
	fm := testhelper.FilesystemManager{}

	tt := map[string]struct {
		FSCreator   func() afero.Fs
		CheckError  func(*testing.T, error) // currently this convoluted mess is needed so we can accurately check for types. errors.As does not work in our case
		ExpTableOut []*Metadata
	}{
		"empty store": {
			FSCreator:   testhelper.FSWithFiles(fm.StoreDir),
			CheckError:  expEmptyStore,
			ExpTableOut: nil,
		},
		"valid konfs and a wrong konf": {
			FSCreator:  testhelper.FSWithFiles(fm.StoreDir, fm.SingleClusterSingleContextEU, fm.SingleClusterSingleContextASIA, fm.InvalidYaml),
			CheckError: expNil,
			ExpTableOut: []*Metadata{
				{
					Context: "dev-asia",
					Cluster: "dev-asia-1",
					File:    "./konf/store/dev-asia_dev-asia-1.yaml",
				},
				{
					Context: "dev-eu",
					Cluster: "dev-eu-1",
					File:    "./konf/store/dev-eu_dev-eu-1.yaml",
				},
			},
		},
		"overloaded konf (cluster)": {
			FSCreator:   testhelper.FSWithFiles(fm.StoreDir, fm.MultiClusterSingleContext),
			CheckError:  expKubeConfigOverload,
			ExpTableOut: nil,
		},
		"overloaded konf (context)": {
			FSCreator:   testhelper.FSWithFiles(fm.StoreDir, fm.SingleClusterMultiContext),
			CheckError:  expKubeConfigOverload,
			ExpTableOut: nil,
		},
		"the nice MacOS .DS_Store file": {
			FSCreator:  testhelper.FSWithFiles(fm.StoreDir, fm.DSStore, fm.SingleClusterSingleContextEU),
			CheckError: expNil,
			ExpTableOut: []*Metadata{
				{
					Context: "dev-eu",
					Cluster: "dev-eu-1",
					File:    "./konf/store/dev-eu_dev-eu-1.yaml",
				},
			},
		},
		"ignore directories": {
			FSCreator:  testhelper.FSWithFiles(fm.StoreDir, fm.SingleClusterSingleContextEU, fm.EmptyDir),
			CheckError: expNil,
			ExpTableOut: []*Metadata{
				{
					Context: "dev-eu",
					Cluster: "dev-eu-1",
					File:    "./konf/store/dev-eu_dev-eu-1.yaml",
				},
			},
		},
		"only directories in store": {
			FSIn:        testhelper.FSWithFiles(fm.StoreDir, fm.EmptyDir),
			CheckError:  expEmptyStore,
			ExpTableOut: nil,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			out, err := FetchKonfs(tc.FSIn)

			tc.CheckError(t, err)

			if !cmp.Equal(tc.ExpTableOut, out) {
				t.Errorf("Exp and given Tableoutputs differ:\n'%s'", cmp.Diff(tc.ExpTableOut, out))
			}
		})
	}
}

func expEmptyStore(t *testing.T, err error) {
	if _, ok := err.(*EmptyStore); !ok {
		t.Errorf("Expected err to be of type EmptyStore")
	}
}

func expKubeConfigOverload(t *testing.T, err error) {
	if _, ok := err.(*KubeConfigOverload); !ok {
		t.Errorf("Expected err to be of type KubeConfigOverload")
	}
}

func expNil(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Expected err to be nil, but got %q", err)
	}
}
