package cmd

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/simontheleg/konf-go/utils"
	"github.com/spf13/afero"
	k8s "k8s.io/client-go/tools/clientcmd/api/v1"
)

func TestImport(t *testing.T) {
	utils.InitTestViper()
	fm := utils.FilesystemManager{}
	var determineConfigsCalled bool
	var writeConfigCalledCount int
	// using just a wrapper here instead of a full mock, makes testing it slightly easier
	var wrapDetermineConfig = func(f afero.Fs, fpath string) ([]*konfFile, error) {
		determineConfigsCalled = true
		return determineConfigs(f, fpath)
	}
	var mockWriteConfig = func(afero.Fs, *konfFile) error { writeConfigCalledCount += 1; return nil }

	type ExpCalls struct {
		DetermineConfigs bool
		WriteConfig      int
	}
	tt := map[string]struct {
		Args   []string
		Fs     afero.Fs
		ExpErr error
		ExpCalls
	}{
		"single context": {
			[]string{"./konf/store/dev-eu_dev-eu-1.yaml"},
			utils.FSWithFiles(fm.StoreDir, fm.SingleClusterSingleContextEU),
			nil,
			ExpCalls{DetermineConfigs: true, WriteConfig: 1},
		},
		"empty context": {
			[]string{"./konf/store/no-context.yaml"},
			utils.FSWithFiles(fm.StoreDir, fm.KonfWithoutContext),
			fmt.Errorf("no contexts found in file \"./konf/store/no-context.yaml\""),
			ExpCalls{DetermineConfigs: true, WriteConfig: 0},
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			determineConfigsCalled = false
			writeConfigCalledCount = 0

			icmd := newImportCmd()
			icmd.fs = tc.Fs
			icmd.determineConfigs = wrapDetermineConfig
			icmd.writeConfig = mockWriteConfig
			cmd := icmd.cmd
			cmd.SetArgs(tc.Args)

			_, err := cmd.ExecuteC()
			if !utils.EqualError(tc.ExpErr, err) {
				t.Errorf("Exp error %q, got %q", tc.ExpErr, err)
			}

			if tc.ExpCalls.DetermineConfigs != determineConfigsCalled {
				t.Errorf("Exp DetermineConfigsCalled to be %t, but got %t", tc.ExpCalls.DetermineConfigs, determineConfigsCalled)
			}

			if tc.ExpCalls.WriteConfig != writeConfigCalledCount {
				t.Errorf("Exp WriteConfigCalled to be %d, but got %d", tc.ExpCalls.WriteConfig, writeConfigCalledCount)
			}

		})
	}
}

func devEUControlGroup() *konfFile {
	utils.InitTestViper()
	return &konfFile{
		FilePath: utils.StorePathForID(utils.IDFromClusterAndContext("dev-eu-1", "dev-eu")),
		Content: k8s.Config{
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
}

func devASIAControlGroup() *konfFile {
	utils.InitTestViper()
	return &konfFile{
		FilePath: utils.StorePathForID(utils.IDFromClusterAndContext("dev-asia-1", "dev-asia")),
		Content: k8s.Config{
			APIVersion:     "v1",
			Kind:           "Config",
			CurrentContext: "dev-asia",
			Clusters: []k8s.NamedCluster{
				{
					Name: "dev-asia-1",
					Cluster: k8s.Cluster{
						Server: "https://192.168.0.1",
					},
				},
			},
			Contexts: []k8s.NamedContext{
				{
					Name: "dev-asia",
					Context: k8s.Context{
						Cluster:   "dev-asia-1",
						Namespace: "kube-system",
						AuthInfo:  "dev-asia",
					},
				},
			},
			AuthInfos: []k8s.NamedAuthInfo{
				{
					Name:     "dev-asia",
					AuthInfo: k8s.AuthInfo{},
				},
			},
		},
	}
}

func TestDetermineConfigs(t *testing.T) {
	fm := utils.FilesystemManager{}
	utils.InitTestViper()

	devEU := devEUControlGroup()
	devAsia := devASIAControlGroup()

	tt := map[string]struct {
		Fs                 afero.Fs
		konfpath           string
		ExpError           error
		ExpNumOfKonfigFile int
		ExpKonfigFiles     []*konfFile
	}{
		"SingleClusterSingleContext": {
			Fs:                 utils.FSWithFiles(fm.StoreDir, fm.SingleClusterSingleContextEU),
			konfpath:           "./konf/store/dev-eu_dev-eu-1.yaml",
			ExpError:           nil,
			ExpNumOfKonfigFile: 1,
			ExpKonfigFiles: []*konfFile{
				devEU,
			},
		},
		"multiClusterMultiContext": {
			Fs:                 utils.FSWithFiles(fm.StoreDir, fm.MultiClusterMultiContext),
			konfpath:           "./konf/store/multi_multi_konf.yaml",
			ExpError:           nil,
			ExpNumOfKonfigFile: 2,
			ExpKonfigFiles: []*konfFile{
				devAsia,
				devEU,
			},
		},
		"multiClusterSingleContext": {
			Fs:                 utils.FSWithFiles(fm.StoreDir, fm.MultiClusterSingleContext),
			konfpath:           "./konf/store/multi_konf.yaml",
			ExpError:           nil,
			ExpNumOfKonfigFile: 1,
			ExpKonfigFiles: []*konfFile{
				devAsia,
			},
		},
		"emptyConfig": {
			Fs:                 utils.FSWithFiles(),
			konfpath:           "i-dont-exist.yaml",
			ExpError:           fmt.Errorf("open i-dont-exist.yaml: file does not exist"),
			ExpNumOfKonfigFile: 0,
			ExpKonfigFiles:     nil,
		},
		// All for the coverage ;)
		"invalidConfig": {
			Fs:                 utils.FSWithFiles(fm.StoreDir, fm.InvalidYaml),
			konfpath:           "./konf/store/no-konf.yaml",
			ExpError:           fmt.Errorf("error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type v1.Config"),
			ExpNumOfKonfigFile: 0,
			ExpKonfigFiles:     nil,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			res, err := determineConfigs(tc.Fs, tc.konfpath)

			if !utils.EqualError(err, tc.ExpError) {
				t.Errorf("Want error '%s', got '%s'", tc.ExpError, err)
			}

			if len(tc.ExpKonfigFiles) != tc.ExpNumOfKonfigFile {
				t.Errorf("Want %d, got %d kubeconfigs", tc.ExpNumOfKonfigFile, len(tc.ExpKonfigFiles))
			}

			if !cmp.Equal(tc.ExpKonfigFiles, res) {
				t.Errorf("Exp and given KonfigFiles differ:\n'%s'", cmp.Diff(tc.ExpKonfigFiles, res))
			}

		})
	}

}

func TestWriteConfig(t *testing.T) {
	utils.InitTestViper()
	fm := utils.FilesystemManager{}
	f := utils.FSWithFiles(fm.ActiveDir, fm.StoreDir)
	kf := devEUControlGroup()

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

	err := writeConfig(f, kf)
	if err != nil {
		t.Errorf("Exp err to be nil but got %q", err)
	}

	b, err := afero.ReadFile(f, kf.FilePath)
	if err != nil {
		t.Errorf("Exp read in file without any issues, but got %q", err)
	}

	res := string(b)
	if res != exp {
		t.Errorf("\nExp:\n%s\ngot\n%s\n", exp, res)
	}

	// TODO it would be really nice to check if the returned kubeconfig yaml is valid in sense of it being complete
	// Unfortunately I was not able to find a good way to perform this check using the client-go package
}
