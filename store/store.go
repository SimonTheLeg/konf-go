package store

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/simontheleg/konf-go/config"
	"github.com/simontheleg/konf-go/konf"
	"github.com/simontheleg/konf-go/log"
	"github.com/simontheleg/konf-go/utils"
	"github.com/spf13/afero"
	k8s "k8s.io/client-go/tools/clientcmd/api/v1"
	"sigs.k8s.io/yaml"
)

// Metadata describes a formatting of kubekonf information.
// It is mainly being used to present the user a nice table selection
type Metadata struct {
	Context string
	Cluster string
	File    string
}

// FetchAllKonfs retrieves metadata for all konfs currently in the store
func FetchAllKonfs(f afero.Fs) ([]*Metadata, error) {
	return FetchKonfsForGlob(f, "*")
}

// FetchKonfsForGlob returns all konfs whose name matches the supplied pattern.
// Pattern matching is done using [filepath.Match]. The pattern should only
// include the name of the file itself not its full path. Also it should not
// include the extension of the file. All relation to the konfs StoreDir will be
// handled automatically.
//
// [filepath.Match]: https://pkg.go.dev/path/filepath#Match
func FetchKonfsForGlob(f afero.Fs, pattern string) ([]*Metadata, error) {
	var konfs []fs.FileInfo
	var filesChecked int

	err := afero.Walk(f, config.StoreDir(), func(path string, info fs.FileInfo, errPath error) error {
		// do not add directories. This is important as later we check the number of items in konf to determine whether store is empty or not
		// without this check we would display an empty prompt if the user has only directories in their storeDir
		if info.IsDir() && path != config.StoreDir() {
			return filepath.SkipDir
		}

		// skip any hidden files
		if strings.HasPrefix(info.Name(), ".") {
			// I have decided to not print any log line on this, which differs from the logic
			// for malformed kubeconfigs. I think this makes sense as konf import will never produce
			// a hidden file and the purpose of this check is rather to protect against
			// automatically created files like the .DS_Store on MacOs. On the other side however
			// it is quite easy to create a malformed kubeconfig without noticing
			return nil
		}

		// only increment filesChecked after we have sorted out directories and hidden files
		filesChecked++

		// skip any files that do not match our glob
		patternPath := config.StoreDir() + "/" + pattern + ".yaml"
		patternPath = strings.TrimPrefix(patternPath, "./") // we need this as afero.Walk trims out any leading "./"
		match, err := filepath.Match(patternPath, path)
		if err != nil {
			return fmt.Errorf("Could not apply glob %q: %v", pattern, err)
		}
		if !match {
			return nil
		}

		konfs = append(konfs, info)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// at this point it is worth mentioning, that we do not need to remove the
	// root element from the list of konfs anymore. This is because filepath.Match
	// never matches for the root element, and therefore the root iself is not
	// part of the list anymore

	// if the walkfunc only ran once, it means that the storedir does not contain any file which could be a kubeconfig
	// It will always run at least once because we do not skip the rootDir
	if filesChecked == 1 {
		return nil, &EmptyStore{}
	}

	// similar to fs.ReadDir, sort the entries for easier viewing for the user and to
	// be consistent with what shells return during auto-completion
	sort.Slice(konfs, func(i, j int) bool { return konfs[i].Name() < konfs[j].Name() })

	if len(konfs) == 0 {
		return nil, &NoMatch{Pattern: pattern}
	}

	out := []*Metadata{}
	// TODO the logic of this loop should be extracted into the walkFn above to avoid looping twice
	// TODO (possibly the walkfunction should also be extracted into its own function)
	for _, k := range konfs {

		id := konf.IDFromFileInfo(k)
		path := id.StorePath()
		file, err := f.Open(path)
		if err != nil {
			return nil, err
		}
		val, err := afero.ReadAll(file)
		if err != nil {
			return nil, err
		}
		kubeconf := &k8s.Config{}
		err = yaml.Unmarshal(val, kubeconf)
		if err != nil {
			log.Warn("file %q does not contain a valid kubeconfig. Skipping for evaluation", path)
			continue
		}

		if len(kubeconf.Contexts) > 1 || len(kubeconf.Clusters) > 1 {
			// This directly returns, as an impure store is a danger for other usage down the road
			return nil, &KubeConfigOverload{path}
		}

		t := Metadata{}
		t.Context = kubeconf.Contexts[0].Name
		t.Cluster = kubeconf.Clusters[0].Name
		t.File = path
		out = append(out, &t)
	}
	return out, nil
}

func WriteKonfToStore(f afero.Fs, konf *konf.Konfig) (storepath string, err error) {
	b, err := yaml.Marshal(konf.Kubeconfig)
	if err != nil {
		return "", err
	}

	storepath = konf.Id.StorePath()

	err = afero.WriteFile(f, storepath, b, utils.KonfPerm)
	if err != nil {
		return "", err
	}

	return storepath, nil
}
