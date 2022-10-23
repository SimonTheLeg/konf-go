package konf

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	k8s "k8s.io/client-go/tools/clientcmd/api/v1"
)

var singleClusterSingleContext = `
apiVersion: v1
clusters:
- cluster:
    server: https://10.1.1.0
  name: dev-eu-1
contexts:
- context:
    namespace: kube-public
    cluster: dev-eu-1
    user: dev-eu
  name: dev-eu
current-context: dev-eu
kind: Config
preferences: {}
users:
- name: dev-eu
  user: {}
`

var multiClusterMultiContext = `
apiVersion: v1
clusters:
  - cluster:
      server: https://192.168.0.1
    name: dev-asia-1
  - cluster:
      server: https://10.1.1.0
    name: dev-eu-1
contexts:
  - context:
      namespace: kube-system
      cluster: dev-asia-1
      user: dev-asia
    name: dev-asia
  - context:
      namespace: kube-public
      cluster: dev-eu-1
      user: dev-eu
    name: dev-eu
current-context: dev-eu
kind: Config
preferences: {}
users:
  - name: dev-asia
    user: {}
  - name: dev-eu
    user: {}
`

var noContext = `
apiVersion: v1
clusters:
  - cluster:
      server: https://10.1.1.0
    name: dev-eu-1
kind: Config
preferences: {}
users:
  - name: dev-eu
    user: {}
`

func TestKonfsFromKubeConfig(t *testing.T) {
	tt := map[string]struct {
		kubeconfig string
		expKonfs   []*Konfig
	}{
		// TODO multi-context, no context
		"single context": {
			kubeconfig: singleClusterSingleContext,
			expKonfs: []*Konfig{
				{
					Id:        "dev-eu_dev-eu-1",
					StorePath: "./konf/store/dev-eu_dev-eu-1.yaml",
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
				},
			},
		},
		"multi context": {
			kubeconfig: multiClusterMultiContext,
			expKonfs: []*Konfig{
				{
					Id:        "dev-asia_dev-asia-1",
					StorePath: "./konf/store/dev-asia_dev-asia-1.yaml",
					Kubeconfig: k8s.Config{
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
								Name: "dev-asia",
							},
						},
					},
				},
				{
					Id:        "dev-eu_dev-eu-1",
					StorePath: "./konf/store/dev-eu_dev-eu-1.yaml",
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
				},
			},
		},
		"no context": {
			kubeconfig: noContext,
			expKonfs:   []*Konfig{},
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			res, err := KonfsFromKubeconfig(strings.NewReader(tc.kubeconfig))
			if err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(res, tc.expKonfs) {
				t.Errorf("Exp and given konfs differ: \n '%s'", cmp.Diff(tc.expKonfs, res))
			}

		})
	}
}
