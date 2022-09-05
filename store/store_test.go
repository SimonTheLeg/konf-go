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
	config.InitWithOverrides(
		&config.Config{
			KonfDir: "./konf",
		},
	)

	tt := map[string]struct {
		fsCreator   func() afero.Fs
		checkError  func(*testing.T, error) // currently this convoluted mess is needed so we can accurately check for types. errors.As does not work in our case
		expTableOut []*Metadata
	}{
		"empty store": {
			fsCreator:   testhelper.FSWithFiles(fm.StoreDir),
			checkError:  expEmptyStore,
			expTableOut: nil,
		},
		"valid konfs and a wrong konf": {
			fsCreator:  testhelper.FSWithFiles(fm.StoreDir, fm.SingleClusterSingleContextEU, fm.SingleClusterSingleContextASIA, fm.InvalidYaml),
			checkError: expNil,
			expTableOut: []*Metadata{
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
			fsCreator:   testhelper.FSWithFiles(fm.StoreDir, fm.MultiClusterSingleContext),
			checkError:  expKubeConfigOverload,
			expTableOut: nil,
		},
		"overloaded konf (context)": {
			fsCreator:   testhelper.FSWithFiles(fm.StoreDir, fm.SingleClusterMultiContext),
			checkError:  expKubeConfigOverload,
			expTableOut: nil,
		},
		"the nice MacOS .DS_Store file": {
			fsCreator:  testhelper.FSWithFiles(fm.StoreDir, fm.DSStore, fm.SingleClusterSingleContextEU),
			checkError: expNil,
			expTableOut: []*Metadata{
				{
					Context: "dev-eu",
					Cluster: "dev-eu-1",
					File:    "./konf/store/dev-eu_dev-eu-1.yaml",
				},
			},
		},
		"ignore directories": {
			fsCreator:  testhelper.FSWithFiles(fm.StoreDir, fm.SingleClusterSingleContextEU, fm.EmptyDir),
			checkError: expNil,
			expTableOut: []*Metadata{
				{
					Context: "dev-eu",
					Cluster: "dev-eu-1",
					File:    "./konf/store/dev-eu_dev-eu-1.yaml",
				},
			},
		},
		"only directories in store": {
			fsCreator:   testhelper.FSWithFiles(fm.StoreDir, fm.EmptyDir),
			checkError:  expEmptyStore,
			expTableOut: nil,
		},
	}

	for name, tc := range tt {

		t.Run(name, func(t *testing.T) {
			out, err := FetchAllKonfs(tc.fsCreator())

			tc.checkError(t, err)

			if !cmp.Equal(tc.expTableOut, out) {
				t.Errorf("Exp and given Tableoutputs differ:\n'%s'", cmp.Diff(tc.expTableOut, out))
			}
		})
	}
}

// This test is mainly required due to an interesting design by afero. More specifically afero
// parses out any leading "./" characters for all files inside the folder, but not for the root
// element itself. But because konfDir can be configured to anything we want, we need to test
// for these cases as well
func TestFetchAllKonfsCustomKonfDir(t *testing.T) {
	fm := testhelper.FilesystemManager{}
	conf := &config.Config{
		KonfDir: "konf", // override with custom dir here
	}
	config.InitWithOverrides(conf)

	tt := map[string]struct {
		fsCreator      func() afero.Fs
		checkError     func(*testing.T, error) // currently this convoluted mess is needed so we can accurately check for types. errors.As does not work in our case
		expTableOutput []*Metadata
	}{
		"empty store": {
			fsCreator:      testhelper.FSWithFiles(fm.StoreDir),
			checkError:     expEmptyStore,
			expTableOutput: nil,
		},
		"valid konfs and a wrong konf": {
			fsCreator:  testhelper.FSWithFiles(fm.StoreDir, fm.SingleClusterSingleContextEU, fm.SingleClusterSingleContextASIA, fm.InvalidYaml),
			checkError: expNil,
			expTableOutput: []*Metadata{
				{
					Context: "dev-asia",
					Cluster: "dev-asia-1",
					File:    "konf/store/dev-asia_dev-asia-1.yaml",
				},
				{
					Context: "dev-eu",
					Cluster: "dev-eu-1",
					File:    "konf/store/dev-eu_dev-eu-1.yaml",
				},
			},
		},
		"overloaded konf (cluster)": {
			fsCreator:      testhelper.FSWithFiles(fm.StoreDir, fm.MultiClusterSingleContext),
			checkError:     expKubeConfigOverload,
			expTableOutput: nil,
		},
		"overloaded konf (context)": {
			fsCreator:      testhelper.FSWithFiles(fm.StoreDir, fm.SingleClusterMultiContext),
			checkError:     expKubeConfigOverload,
			expTableOutput: nil,
		},
		"the nice MacOS .DS_Store file": {
			fsCreator:  testhelper.FSWithFiles(fm.StoreDir, fm.DSStore, fm.SingleClusterSingleContextEU),
			checkError: expNil,
			expTableOutput: []*Metadata{
				{
					Context: "dev-eu",
					Cluster: "dev-eu-1",
					File:    "konf/store/dev-eu_dev-eu-1.yaml",
				},
			},
		},
		"ignore directories": {
			fsCreator:  testhelper.FSWithFiles(fm.StoreDir, fm.SingleClusterSingleContextEU, fm.EmptyDir),
			checkError: expNil,
			expTableOutput: []*Metadata{
				{
					Context: "dev-eu",
					Cluster: "dev-eu-1",
					File:    "konf/store/dev-eu_dev-eu-1.yaml",
				},
			},
		},
		"only directories in store": {
			fsCreator:      testhelper.FSWithFiles(fm.StoreDir, fm.EmptyDir),
			checkError:     expEmptyStore,
			expTableOutput: nil,
		},
	}

	for name, tc := range tt {

		t.Run(name, func(t *testing.T) {
			out, err := FetchAllKonfs(tc.fsCreator())

			tc.checkError(t, err)

			if !cmp.Equal(tc.expTableOutput, out) {
				t.Errorf("Exp and given Tableoutputs differ:\n'%s'", cmp.Diff(tc.expTableOutput, out))
			}
		})
	}
}

func TestFetchKonfsForGlob(t *testing.T) {
	fm := testhelper.FilesystemManager{}
	// TODO figure out if there is a better way to handle configuration, so we don't have to set the konfdir every time
	config.InitWithOverrides(
		&config.Config{
			KonfDir: "./konf",
		},
	)

	tt := map[string]struct {
		fsCreator   func() afero.Fs
		checkError  func(*testing.T, error) // currently this convoluted mess is needed so we can accurately check for types. errors.As does not work in our case
		glob        string
		expTableOut []*Metadata
	}{
		"match eu konf": {
			fsCreator:  testhelper.FSWithFiles(fm.StoreDir, fm.SingleClusterSingleContextEU, fm.SingleClusterSingleContextASIA, fm.InvalidYaml),
			checkError: expNil,
			glob:       "dev-eu*",
			expTableOut: []*Metadata{
				{
					Context: "dev-eu",
					Cluster: "dev-eu-1",
					File:    "./konf/store/dev-eu_dev-eu-1.yaml",
				},
			},
		},
		"match eu konf no expansion": {
			fsCreator:  testhelper.FSWithFiles(fm.StoreDir, fm.SingleClusterSingleContextEU, fm.SingleClusterSingleContextASIA, fm.InvalidYaml),
			checkError: expNil,
			glob:       "dev-eu_dev-eu-1",
			expTableOut: []*Metadata{
				{
					Context: "dev-eu",
					Cluster: "dev-eu-1",
					File:    "./konf/store/dev-eu_dev-eu-1.yaml",
				},
			},
		},
		"match asia konf": {
			fsCreator:  testhelper.FSWithFiles(fm.StoreDir, fm.SingleClusterSingleContextEU, fm.SingleClusterSingleContextASIA, fm.InvalidYaml),
			checkError: expNil,
			glob:       "dev-asia*",
			expTableOut: []*Metadata{
				{
					Context: "dev-asia",
					Cluster: "dev-asia-1",
					File:    "./konf/store/dev-asia_dev-asia-1.yaml",
				},
			},
		},
		"directory ignore takes precedence over glob": {
			fsCreator:  testhelper.FSWithFiles(fm.StoreDir, fm.EUDir, fm.SingleClusterSingleContextEU, fm.SingleClusterSingleContextASIA),
			checkError: expNil,
			glob:       "dev-eu*",
			expTableOut: []*Metadata{
				{
					Context: "dev-eu",
					Cluster: "dev-eu-1",
					File:    "./konf/store/dev-eu_dev-eu-1.yaml",
				},
			},
		},
		"no match, but valid konfs exist": {
			fsCreator:   testhelper.FSWithFiles(fm.SingleClusterSingleContextEU),
			checkError:  expNoMatch,
			glob:        "no-match",
			expTableOut: nil,
		},
	}

	for name, tc := range tt {

		t.Run(name, func(t *testing.T) {
			out, err := FetchKonfsForGlob(tc.fsCreator(), tc.glob)

			tc.checkError(t, err)

			if !cmp.Equal(tc.expTableOut, out) {
				t.Errorf("Exp and given Tableoutputs differ:\n'%s'", cmp.Diff(tc.expTableOut, out))
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

func expNoMatch(t *testing.T, err error) {
	if _, ok := err.(*NoMatch); !ok {
		t.Errorf("Expected err to be of type NoMatch")
	}
}
