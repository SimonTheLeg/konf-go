package store

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/simontheleg/konf-go/konf"
	"github.com/simontheleg/konf-go/testhelper"
	"github.com/spf13/afero"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	k8s "k8s.io/client-go/tools/clientcmd/api/v1"
)

func TestFetchAllKonfs(t *testing.T) {
	storeDir := "./konf/store"
	activeDir := "./konf/active"
	fm := testhelper.FilesystemManager{Storedir: storeDir, Activedir: activeDir}

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
			sm := Storemanager{Activedir: activeDir, Storedir: storeDir, Fs: tc.fsCreator()}
			out, err := sm.FetchAllKonfs()

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
	storeDir := "konf/store"
	activeDir := "konf/active"
	fm := testhelper.FilesystemManager{Storedir: storeDir, Activedir: activeDir}

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
			sm := Storemanager{Activedir: activeDir, Storedir: storeDir, Fs: tc.fsCreator()}
			out, err := sm.FetchAllKonfs()

			tc.checkError(t, err)

			if !cmp.Equal(tc.expTableOutput, out) {
				t.Errorf("Exp and given Tableoutputs differ:\n'%s'", cmp.Diff(tc.expTableOutput, out))
			}
		})
	}
}

func TestFetchKonfsForGlob(t *testing.T) {
	storeDir := "./konf/store"
	activeDir := "./konf/active"
	fm := testhelper.FilesystemManager{Storedir: storeDir, Activedir: activeDir}

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
			fs := tc.fsCreator()
			sm := &Storemanager{Activedir: activeDir, Storedir: storeDir, Fs: fs}

			out, err := sm.FetchKonfsForGlob(tc.glob)

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

func TestWriteKonfToStore(t *testing.T) {
	storeDir := "./konf/store"
	activeDir := "./konf/active"
	fm := testhelper.FilesystemManager{Storedir: storeDir, Activedir: activeDir}
	f := testhelper.FSWithFiles(fm.ActiveDir, fm.StoreDir)()
	sm := &Storemanager{Activedir: activeDir, Storedir: storeDir, Fs: f}

	expContent := `apiVersion: v1
clusters:
- cluster:
    server: https://10.1.1.0
  name: dev-eu-1
contexts:
- context:
    cluster: dev-eu-1
    namespace: kube-public
    user: dev-eu
  name: dev-eu
current-context: dev-eu
kind: Config
preferences: {}
users:
- name: dev-eu
  user: {}
`

	expPath := "./konf/store/dev-eu_dev-eu-1.yaml"

	var devEUControlGroup = &konf.Konfig{
		Id: konf.IDFromClusterAndContext("dev-eu-1", "dev-eu"),
		Kubeconfig: k8s.Config{
			APIVersion:     "v1",
			Kind:           "Config",
			CurrentContext: "dev-eu",
			Clusters: []k8s.NamedCluster{
				{
					Name: "dev-eu-1",
					Cluster: k8s.Cluster{
						Server: "https://10.1.1.0",
					},
				},
			},
			Contexts: []k8s.NamedContext{
				{
					Name: "dev-eu",
					Context: k8s.Context{
						Cluster:   "dev-eu-1",
						Namespace: "kube-public",
						AuthInfo:  "dev-eu",
					},
				},
			},
			AuthInfos: []k8s.NamedAuthInfo{
				{
					Name: "dev-eu",
				},
			},
		},
	}

	p, err := sm.WriteKonfToStore(devEUControlGroup)
	if err != nil {
		t.Errorf("Exp err to be nil but got %q", err)
	}

	if p != expPath {
		t.Errorf("Exp path to be %q, but got %q", expPath, p)
	}

	b, err := afero.ReadFile(f, expPath)
	if err != nil {
		t.Errorf("Exp read in file without any issues, but got %q", err)
	}

	res := string(b)
	if res != expContent {
		t.Errorf("\nExp:\n%s\ngot\n%s\n", expContent, res)
	}

	// check if the konf is also valid for creating a clientset
	conf, err := clientcmd.NewClientConfigFromBytes(b)
	if err != nil {
		t.Errorf("Exp to create clientconfig, but got %q", err)
	}
	cc, err := conf.ClientConfig()
	if err != nil {
		t.Errorf("Exp to extract rest.config, but got %q", err)
	}
	_, err = kubernetes.NewForConfig(cc)
	if err != nil {
		t.Errorf("Exp to create clientset, but got %q", err)
	}

}

func TestActivePathFromID(t *testing.T) {
	sm := Storemanager{Activedir: "something/active", Storedir: "something/store"}
	konfID := konf.IDFromClusterAndContext("mycluster", "mycontext")
	res := sm.ActivePathFromID(konfID)
	expRes := "something/active/mycontext_mycluster.yaml"
	if res != expRes {
		t.Errorf("wanted id %q, got %q", expRes, res)
	}
}

func TestStorePathFromID(t *testing.T) {
	sm := Storemanager{Activedir: "something/active", Storedir: "something/store"}
	konfID := konf.IDFromClusterAndContext("mycluster", "mycontext")
	res := sm.StorePathFromID(konfID)
	expRes := "something/store/mycontext_mycluster.yaml"
	if res != expRes {
		t.Errorf("wanted id %q, got %q", expRes, res)
	}
}
