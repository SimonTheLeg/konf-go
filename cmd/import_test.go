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
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	k8s "k8s.io/client-go/tools/clientcmd/api/v1"
)

func TestImport(t *testing.T) {
	fm := testhelper.FilesystemManager{}
	var determineConfigsCalled bool
	var writeConfigCalledCount int
	var deleteOriginalConfigCalled bool
	// using just a wrapper here instead of a full mock, makes testing it slightly easier
	var wrapDetermineConfig = func(r io.Reader) ([]*konf.Config, error) {
		determineConfigsCalled = true
		return konf.KonfsFromKubeconfig(r)
	}
	var mockWriteConfig = func(afero.Fs, *konf.Config) error { writeConfigCalledCount++; return nil }
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

var devEUControlGroup = &konf.Config{
	StorePath: konf.IDFromClusterAndContext("dev-eu-1", "dev-eu").StorePath(),
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

func TestWriteConfig(t *testing.T) {
	fm := testhelper.FilesystemManager{}
	f := testhelper.FSWithFiles(fm.ActiveDir, fm.StoreDir)()

	exp := `apiVersion: v1
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

	err := writeConfig(f, devEUControlGroup)
	if err != nil {
		t.Errorf("Exp err to be nil but got %q", err)
	}

	b, err := afero.ReadFile(f, devEUControlGroup.StorePath)
	if err != nil {
		t.Errorf("Exp read in file without any issues, but got %q", err)
	}

	res := string(b)
	if res != exp {
		t.Errorf("\nExp:\n%s\ngot\n%s\n", exp, res)
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
