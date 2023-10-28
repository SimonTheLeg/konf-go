package testhelper

import (
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
func FSWithFiles(ff ...filefunc) func() afero.Fs {
	return func() afero.Fs {
		fs := afero.NewMemMapFs()

		for _, f := range ff {
			f(fs)
		}
		return fs
	}
}

// FilesystemManager is used to manage filefuncs. It is feature identical to
// its string counterpart SampleKonfManager
type FilesystemManager struct {
	Storedir       string
	Activedir      string
	LatestKonfPath string
}

// it makes sense to reimplement the path generation for store and active, so this package does
// not need to depend on the konf or id package and we can freely use it.
// TODO remove this once id and store pkg are properly separated
func (f *FilesystemManager) storePathForID(id string) string {
	return f.Storedir + "/" + id + ".yaml"
}

func (f *FilesystemManager) activePathForID(id string) string {
	return f.Activedir + "/" + id + ".yaml"
}

// StoreDir creates standard konf store
func (f *FilesystemManager) StoreDir(fs afero.Fs) {
	fs.MkdirAll(f.Storedir, utils.KonfPerm)
}

// ActiveDir creates standard konf active
func (f *FilesystemManager) ActiveDir(fs afero.Fs) {
	fs.MkdirAll(f.Activedir, utils.KonfPerm)
}

// SingleClusterSingleContextEU creates a valid kubeconfig in store and active
func (f *FilesystemManager) SingleClusterSingleContextEU(fs afero.Fs) {
	afero.WriteFile(fs, f.storePathForID("dev-eu_dev-eu-1"), []byte(singleClusterSingleContextEU), utils.KonfPerm)
	afero.WriteFile(fs, f.activePathForID("dev-eu_dev-eu-1"), []byte(singleClusterSingleContextEU), utils.KonfPerm)
}

// SingleClusterSingleContextASIA creates a valid kubeconfig in store and active
func (f *FilesystemManager) SingleClusterSingleContextASIA(fs afero.Fs) {
	afero.WriteFile(fs, f.storePathForID("dev-asia_dev-asia-1"), []byte(singleClusterSingleContextASIA), utils.KonfPerm)
	afero.WriteFile(fs, f.activePathForID("dev-asia_dev-asia-1"), []byte(singleClusterSingleContextASIA), utils.KonfPerm)
}

// SingleClusterSingleContextEU2 creates a second valid kubeconfig in store and active. It is mainly used for glob testing
func (f *FilesystemManager) SingleClusterSingleContextEU2(fs afero.Fs) {
	afero.WriteFile(fs, f.storePathForID("dev-eu_dev-eu-2"), []byte(singleClusterSingleContextEU2), utils.KonfPerm)
	afero.WriteFile(fs, f.activePathForID("dev-eu_dev-eu-2"), []byte(singleClusterSingleContextEU2), utils.KonfPerm)
}

// SingleClusterSingleContextASIA2 creates a second valid kubeconfig in store and active. It is mainly used for glob testing
func (f *FilesystemManager) SingleClusterSingleContextASIA2(fs afero.Fs) {
	afero.WriteFile(fs, f.storePathForID("dev-asia_dev-asia-2"), []byte(singleClusterSingleContextASIA2), utils.KonfPerm)
	afero.WriteFile(fs, f.activePathForID("dev-asia_dev-asia-2"), []byte(singleClusterSingleContextASIA2), utils.KonfPerm)
}

// InvalidYaml creates an invalidYaml in store and active
func (f *FilesystemManager) InvalidYaml(fs afero.Fs) {
	afero.WriteFile(fs, f.storePathForID("no-konf"), []byte("I am no valid yaml"), utils.KonfPerm)
	afero.WriteFile(fs, f.activePathForID("no-konf"), []byte("I am no valid yaml"), utils.KonfPerm)
}

// MultiClusterMultiContext creates a kubeconfig with multiple clusters and contexts in store, resulting in an impure konfstore
func (f *FilesystemManager) MultiClusterMultiContext(fs afero.Fs) {
	afero.WriteFile(fs, f.storePathForID("multi_multi_konf"), []byte(multiClusterMultiContext), utils.KonfPerm)
}

// MultiClusterSingleContext creates a kubeconfig with multiple clusters and one context in store, resulting in an impure konfstore
func (f *FilesystemManager) MultiClusterSingleContext(fs afero.Fs) {
	afero.WriteFile(fs, f.storePathForID("multi_konf"), []byte(multiClusterSingleContext), utils.KonfPerm)
}

// SingleClusterMultiContext creates a kubeconfig with one cluster and multiple contexts in store, resulting in an impure konfstore
func (f *FilesystemManager) SingleClusterMultiContext(fs afero.Fs) {
	afero.WriteFile(fs, f.storePathForID("multi_konf"), []byte(singleClusterMultiContext), utils.KonfPerm)
}

// LatestKonf creates a latestKonfFile pointing to an imaginary context and cluster
func (f *FilesystemManager) LatestKonf(fs afero.Fs) {
	afero.WriteFile(fs, f.LatestKonfPath, []byte("context_cluster"), utils.KonfPerm)
}

// KonfWithoutContext creates a kubeconfig which has no context, but still is valid
func (f *FilesystemManager) KonfWithoutContext(fs afero.Fs) {
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

	afero.WriteFile(fs, f.storePathForID("no-context"), []byte(noContext), utils.KonfPerm)
	afero.WriteFile(fs, f.activePathForID("no-context"), []byte(noContext), utils.KonfPerm)
}

// KonfWithoutContext2 creates a kubeconfig which has no context, but still is valid
func (f *FilesystemManager) KonfWithoutContext2(fs afero.Fs) {
	var noContext = `
apiVersion: v1
clusters:
  - cluster:
      server: https://10.1.1.0
    name: dev-eu-2
kind: Config
preferences: {}
users:
  - name: dev-eu
    user: {}
`

	afero.WriteFile(fs, f.storePathForID("no-context-2"), []byte(noContext), utils.KonfPerm)
	afero.WriteFile(fs, f.activePathForID("no-context-2"), []byte(noContext), utils.KonfPerm)
}

// DSStore creates a .DS_Store file, that has caused quite some problems in the past
func (f *FilesystemManager) DSStore(fs afero.Fs) {
	// in this case we cannot use StorePathForID, as this would append .yaml
	afero.WriteFile(fs, f.Storedir+"/.DS_Store", nil, utils.KonfPerm)
	afero.WriteFile(fs, f.Activedir+"/.DS_Store", nil, utils.KonfPerm)
}

// EmptyDir creates an EmptyDir in StoreDir and ActiveDir
func (f *FilesystemManager) EmptyDir(fs afero.Fs) {
	// in this case we cannot use StorePathForID, as this would append .yaml
	fs.Mkdir(f.Storedir+"empty-dir", utils.KonfDirPerm)
	fs.Mkdir(f.Activedir+"empty-dir", utils.KonfDirPerm)
}

// EUDir creates an dir called "eu" in StoreDir and ActiveDir. It is mainly used to test globing
func (f *FilesystemManager) EUDir(fs afero.Fs) {
	// in this case we cannot use StorePathForID, as this would append .yaml
	fs.Mkdir(f.Storedir+"eu", utils.KonfDirPerm)
	fs.Mkdir(f.Activedir+"eu", utils.KonfDirPerm)
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
var singleClusterSingleContextEU2 = `
apiVersion: v1
clusters:
  - cluster:
      server: https://10.1.1.0
    name: dev-eu-2
contexts:
  - context:
      namespace: kube-public
      cluster: dev-eu-2
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
var singleClusterSingleContextASIA2 = `
apiVersion: v1
clusters:
  - cluster:
      server: https://10.1.1.0
    name: dev-asia-2
contexts:
  - context:
      namespace: kube-public
      cluster: dev-asia-2
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
