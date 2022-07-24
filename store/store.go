package store

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/simontheleg/konf-go/config"
	"github.com/simontheleg/konf-go/log"
	"github.com/simontheleg/konf-go/utils"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
	k8s "k8s.io/client-go/tools/clientcmd/api/v1"
)

// TableOutput describes a formatting of kubekonf information, that is being used to present the user a nice table selection
type TableOutput struct {
	// Since we have no other use for structured information, we can safely leave this in set.go for now
	Context string
	Cluster string
	File    string
}

// KubeConfigOverload describes a state in which a kubeconfig has multiple Contexts or Clusters
// This can be undesirable for konf when such a kubeconfig is in its store
type KubeConfigOverload struct {
	path string
}

func (k *KubeConfigOverload) Error() string {
	return fmt.Sprintf("Impure Store: The kubeconfig %q contains multiple contexts and/or clusters. Please only use 'konf import' for populating the store\n", k.path)
}

// EmptyStore describes a state in which no kubeconfig is inside the store
// It makes sense to have this in a separate case as it does not matter for some operations (e.g. importing) but detrimental for others (e.g. running the selection prompt)
type EmptyStore struct{}

func (k *EmptyStore) Error() string {
	return fmt.Sprintf("The konf store at %q is empty. Please run 'konf import' to populate it", config.StoreDir())
}

func FetchKonfs(f afero.Fs) ([]*TableOutput, error) {
	var konfs []fs.FileInfo

	err := afero.Walk(f, config.StoreDir(), func(path string, info fs.FileInfo, err error) error {
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

		konfs = append(konfs, info)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// cut out the root element, which gets added in the previous step
	// this is safe as the element is guaranteed to be at the first position
	konfs = konfs[1:]

	// similar to fs.ReadDir, sort the entries for easier viewing for the user and to
	// be consistent with what shells return during auto-completion
	sort.Slice(konfs, func(i, j int) bool { return konfs[i].Name() < konfs[j].Name() })

	if len(konfs) == 0 {
		return nil, &EmptyStore{}
	}

	out := []*TableOutput{}
	// TODO the logic of this loop should be extracted into the walkFn above to avoid looping twice
	// TODO (possibly the walkfunction should also be extracted into its own function)
	for _, konf := range konfs {

		id := utils.IDFromFileInfo(konf)
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

		t := TableOutput{}
		t.Context = kubeconf.Contexts[0].Name
		t.Cluster = kubeconf.Clusters[0].Name
		t.File = path
		out = append(out, &t)
	}
	return out, nil
}
