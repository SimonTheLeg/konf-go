package testhelper

import (
	"github.com/simontheleg/konf-go/config"
	"github.com/simontheleg/konf-go/utils"
	"github.com/spf13/afero"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EqualError reports whether errors a and b are considered equal.
// They're equal if both are nil, or both are not nil and a.Error() == b.Error().
func EqualError(a, b error) bool {
	return a == nil && b == nil || a != nil && b != nil && a.Error() == b.Error()
}

type filefunc = func(afero.Fs)

// FSWithFiles is a testhelper that can be used to quickly setup a MemMapFs with required Files
func FSWithFiles(ff ...filefunc) afero.Fs {
	fs := afero.NewMemMapFs()

	for _, f := range ff {
		f(fs)
	}
	return fs
}

// FilesystemManager is used to manage filefuncs. It is feature identical to
// its string counterpart SampleKonfManager
type FilesystemManager struct{}

// StoreDir creates standard konf store
func (*FilesystemManager) StoreDir(fs afero.Fs) {
	fs.MkdirAll(config.StoreDir(), utils.KonfPerm)
}

// ActiveDir creates standard konf active
func (*FilesystemManager) ActiveDir(fs afero.Fs) {
	fs.MkdirAll(config.ActiveDir(), utils.KonfPerm)
}

// SingleClusterSingleContextEU creates a valid kubeconfig in store and active
func (*FilesystemManager) SingleClusterSingleContextEU(fs afero.Fs) {
	afero.WriteFile(fs, utils.StorePathForID("dev-eu_dev-eu-1"), []byte(singleClusterSingleContextEU), utils.KonfPerm)
	afero.WriteFile(fs, utils.ActivePathForID("dev-eu_dev-eu-1"), []byte(singleClusterSingleContextEU), utils.KonfPerm)
}

// SingleClusterSingleContextASIA creates a valid kubeconfig in store and active
func (*FilesystemManager) SingleClusterSingleContextASIA(fs afero.Fs) {
	afero.WriteFile(fs, utils.StorePathForID("dev-asia_dev-asia-1"), []byte(singleClusterSingleContextASIA), utils.KonfPerm)
	afero.WriteFile(fs, utils.ActivePathForID("dev-asia_dev-asia-1"), []byte(singleClusterSingleContextASIA), utils.KonfPerm)
}

// InvalidYaml creates an invalidYaml in store and active
func (*FilesystemManager) InvalidYaml(fs afero.Fs) {
	afero.WriteFile(fs, utils.ActivePathForID("no-konf"), []byte("I am no valid yaml"), utils.KonfPerm)
	afero.WriteFile(fs, utils.StorePathForID("no-konf"), []byte("I am no valid yaml"), utils.KonfPerm)
}

// MultiClusterMultiContext creates a kubeconfig with multiple clusters and contexts in store, resulting in an impure konfstore
func (*FilesystemManager) MultiClusterMultiContext(fs afero.Fs) {
	afero.WriteFile(fs, utils.StorePathForID("multi_multi_konf"), []byte(multiClusterMultiContext), utils.KonfPerm)
}

// MultiClusterSingleContext creates a kubeconfig with multiple clusters and one context in store, resulting in an impure konfstore
func (*FilesystemManager) MultiClusterSingleContext(fs afero.Fs) {
	afero.WriteFile(fs, utils.StorePathForID("multi_konf"), []byte(multiClusterSingleContext), utils.KonfPerm)
}

// SingleClusterMultiContext creates a kubeconfig with one cluster and multiple contexts in store, resulting in an impure konfstore
func (*FilesystemManager) SingleClusterMultiContext(fs afero.Fs) {
	afero.WriteFile(fs, utils.StorePathForID("multi_konf"), []byte(singleClusterMultiContext), utils.KonfPerm)
}

// LatestKonf creates a latestKonfFile pointing to an imaginary context and cluster
func (*FilesystemManager) LatestKonf(fs afero.Fs) {
	afero.WriteFile(fs, config.LatestKonfFile(), []byte("context_cluster"), utils.KonfPerm)
}

// KonfWithoutContext creates a kubeconfig which has no context, but still is valid
func (*FilesystemManager) KonfWithoutContext(fs afero.Fs) {
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

	afero.WriteFile(fs, utils.StorePathForID("no-context"), []byte(noContext), utils.KonfPerm)
	afero.WriteFile(fs, utils.ActivePathForID("no-context"), []byte(noContext), utils.KonfPerm)
}

// DSStore creates a .DS_Store file, that has caused quite some problems in the past
func (*FilesystemManager) DSStore(fs afero.Fs) {
	// in this case we cannot use StorePathForID, as this would append .yaml
	afero.WriteFile(fs, config.StoreDir()+"/.DS_Store", nil, utils.KonfPerm)
	afero.WriteFile(fs, config.ActiveDir()+"/.DS_Store", nil, utils.KonfPerm)
}

// SampleKonfManager is used to manage kubeconfig strings. It is feature identical to
// its file counterpart FilesystemManager
type SampleKonfManager struct{}

// SingleClusterSingleContextEU returns a valid kubeconfig
func (*SampleKonfManager) SingleClusterSingleContextEU() string {
	return singleClusterSingleContextEU
}

// SingleClusterSingleContextASIA returns a valid kubeconfig
func (*SampleKonfManager) SingleClusterSingleContextASIA() string {
	return singleClusterSingleContextASIA
}

// MultiClusterMultiContext returns a valid kubeconfig, that is unprocessed
func (*SampleKonfManager) MultiClusterMultiContext() string {
	return multiClusterMultiContext
}

// MultiClusterSingleContext returns a valid kubeconfig, that is unprocessed
func (*SampleKonfManager) MultiClusterSingleContext() string {
	return multiClusterSingleContext
}

var singleClusterSingleContextEU = `
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

var singleClusterSingleContextASIA = `
apiVersion: v1
clusters:
  - cluster:
      server: https://10.1.1.0
    name: dev-asia-1
contexts:
  - context:
      namespace: kube-public
      cluster: dev-asia-1
      user: dev-asia
    name: dev-asia
current-context: dev-asia
kind: Config
preferences: {}
users:
  - name: dev-asia
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

var singleClusterMultiContext = `
apiVersion: v1
clusters:
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

var multiClusterSingleContext = `
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
# Purposefully kept this wrong
current-context: dev-eu
kind: Config
preferences: {}
users:
  - name: dev-asia
    user: {}
`

// NamespaceFromName creates a simple namespace object for a name
func NamespaceFromName(name string) *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}
