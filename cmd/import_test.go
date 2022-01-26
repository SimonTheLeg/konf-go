package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/simontheleg/konf-go/utils"
	k8s "k8s.io/client-go/tools/clientcmd/api/v1"
)

var devEU konfigFile
var devAsia konfigFile

func TestDetermineConfigs(t *testing.T) {
	setup(t)
	sm := utils.SampleKonfManager{}

	tt := map[string]struct {
		InConfig           string
		ExpError           error
		ExpNumOfKonfigFile int
		ExpKonfigFiles     []*konfigFile
	}{
		"SingleClusterSingleContext": {
			InConfig:           sm.SingleClusterSingleContextEU(),
			ExpError:           nil,
			ExpNumOfKonfigFile: 1,
			ExpKonfigFiles: []*konfigFile{
				&devEU,
			},
		},
		"multiClusterMultiContext": {
			InConfig:           sm.MultiClusterMultiContext(),
			ExpError:           nil,
			ExpNumOfKonfigFile: 2,
			ExpKonfigFiles: []*konfigFile{
				&devAsia,
				&devEU,
			},
		},
		"multiClusterSingleContext": {
			InConfig:           sm.MultiClusterSingleContext(),
			ExpError:           nil,
			ExpNumOfKonfigFile: 1,
			ExpKonfigFiles: []*konfigFile{
				&devAsia,
			},
		},
		"emptyConfig": {
			InConfig:           "",
			ExpError:           nil,
			ExpNumOfKonfigFile: 0,
			ExpKonfigFiles:     []*konfigFile{},
		},
		// All for the coverage ;)
		"invalidConfig": {
			InConfig:           "I am not valid yaml",
			ExpError:           fmt.Errorf("error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type v1.Config"),
			ExpNumOfKonfigFile: 0,
			ExpKonfigFiles:     nil,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			res, err := determineConfigs(strings.NewReader(tc.InConfig))

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
	setup(t)

	var b bytes.Buffer

	writeConfig(&b, &devEU.Content)

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

	res := b.String()
	if res != exp {
		t.Errorf("\nExp:\n%s\ngot\n%s\n", exp, res)
	}

	// TODO it would be really nice to check if the returned kubeconfig yaml is valid in sense of it being complete
	// Unfortunately I was not able to find a good way to perform this check using the client-go package
}

// All certificate & token values are mocked
func setup(t *testing.T) {
	devEU = konfigFile{
		FileName: utils.StorePathForID(utils.IDFromClusterAndContext("dev-eu-1", "dev-eu")),
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

	devAsia = konfigFile{
		FileName: utils.StorePathForID(utils.IDFromClusterAndContext("dev-asia-1", "dev-asia")),
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

