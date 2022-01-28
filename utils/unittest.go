package utils

import (
	"github.com/simontheleg/konf-go/config"
	"github.com/spf13/afero"
)

const IntegrationtestDir = "/tmp/konfs"

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
	fs.MkdirAll(config.StoreDir(), KonfPerm)
}

func (*FilesystemManager) ActiveDir(fs afero.Fs) {
	fs.MkdirAll(config.ActiveDir(), KonfPerm)
}

func (*FilesystemManager) SingleClusterSingleContextEU(fs afero.Fs) {
	afero.WriteFile(fs, StorePathForID("dev-eu_dev-eu-1"), []byte(singleClusterSingleContextEU), KonfPerm)
	afero.WriteFile(fs, ActivePathForID("dev-eu_dev-eu-1"), []byte(singleClusterSingleContextEU), KonfPerm)
}

func (*FilesystemManager) SingleClusterSingleContextASIA(fs afero.Fs) {
	afero.WriteFile(fs, StorePathForID("dev-asia_dev-asia-1"), []byte(singleClusterSingleContextASIA), KonfPerm)
	afero.WriteFile(fs, ActivePathForID("dev-asia_dev-asia-1"), []byte(singleClusterSingleContextASIA), KonfPerm)
}

func (*FilesystemManager) InvalidYaml(fs afero.Fs) {
	afero.WriteFile(fs, ActivePathForID("no-konf"), []byte("I am no valid yaml"), KonfPerm)
	afero.WriteFile(fs, StorePathForID("no-konf"), []byte("I am no valid yaml"), KonfPerm)
}

func (*FilesystemManager) MultiClusterMultiContext(fs afero.Fs) {
	afero.WriteFile(fs, StorePathForID("multi_multi_konf"), []byte(multiClusterMultiContext), KonfPerm)
}

func (*FilesystemManager) MultiClusterSingleContext(fs afero.Fs) {
	afero.WriteFile(fs, StorePathForID("multi_konf"), []byte(multiClusterSingleContext), KonfPerm)
}

func (*FilesystemManager) SingleClusterMultiContext(fs afero.Fs) {
	afero.WriteFile(fs, StorePathForID("multi_konf"), []byte(singleClusterMultiContext), KonfPerm)
}

func (*FilesystemManager) LatestKonf(fs afero.Fs) {
	afero.WriteFile(fs, config.LatestKonfFile(), []byte("context_cluster"), KonfPerm)
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

	afero.WriteFile(fs, StorePathForID("no-context"), []byte(noContext), KonfPerm)
	afero.WriteFile(fs, ActivePathForID("no-context"), []byte(noContext), KonfPerm)
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
