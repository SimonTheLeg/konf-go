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

func FSWithFiles(ff ...filefunc) afero.Fs {
	fs := afero.NewMemMapFs()

	for _, f := range ff {
		f(fs)
	}
	return fs
}

type FilesystemManager struct{}

func (*FilesystemManager) StoreDir(fs afero.Fs) {
	fs.MkdirAll(config.StoreDir(), utils.KonfPerm)
}

func (*FilesystemManager) ActiveDir(fs afero.Fs) {
	fs.MkdirAll(config.ActiveDir(), utils.KonfPerm)
}

func (*FilesystemManager) SingleClusterSingleContextEU(fs afero.Fs) {
	afero.WriteFile(fs, utils.StorePathForID("dev-eu_dev-eu-1"), []byte(singleClusterSingleContextEU), utils.KonfPerm)
	afero.WriteFile(fs, utils.ActivePathForID("dev-eu_dev-eu-1"), []byte(singleClusterSingleContextEU), utils.KonfPerm)
}

func (*FilesystemManager) SingleClusterSingleContextASIA(fs afero.Fs) {
	afero.WriteFile(fs, utils.StorePathForID("dev-asia_dev-asia-1"), []byte(singleClusterSingleContextASIA), utils.KonfPerm)
	afero.WriteFile(fs, utils.ActivePathForID("dev-asia_dev-asia-1"), []byte(singleClusterSingleContextASIA), utils.KonfPerm)
}

func (*FilesystemManager) InvalidYaml(fs afero.Fs) {
	afero.WriteFile(fs, utils.ActivePathForID("no-konf"), []byte("I am no valid yaml"), utils.KonfPerm)
	afero.WriteFile(fs, utils.StorePathForID("no-konf"), []byte("I am no valid yaml"), utils.KonfPerm)
}

func (*FilesystemManager) MultiClusterMultiContext(fs afero.Fs) {
	afero.WriteFile(fs, utils.StorePathForID("multi_multi_konf"), []byte(multiClusterMultiContext), utils.KonfPerm)
}

func (*FilesystemManager) MultiClusterSingleContext(fs afero.Fs) {
	afero.WriteFile(fs, utils.StorePathForID("multi_konf"), []byte(multiClusterSingleContext), utils.KonfPerm)
}

func (*FilesystemManager) SingleClusterMultiContext(fs afero.Fs) {
	afero.WriteFile(fs, utils.StorePathForID("multi_konf"), []byte(singleClusterMultiContext), utils.KonfPerm)
}

func (*FilesystemManager) LatestKonf(fs afero.Fs) {
	afero.WriteFile(fs, config.LatestKonfFile(), []byte("context_cluster"), utils.KonfPerm)
}

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

type SampleKonfManager struct{}

func (*SampleKonfManager) SingleClusterSingleContextEU() string {
	return singleClusterSingleContextEU
}

func (*SampleKonfManager) SingleClusterSingleContextASIA() string {
	return singleClusterSingleContextASIA
}

func (*SampleKonfManager) MultiClusterMultiContext() string {
	return multiClusterMultiContext
}

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
