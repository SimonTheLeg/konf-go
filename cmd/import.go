package cmd

import (
	"bytes"
	"io"
	"log"
	"os"

	"github.com/simontheleg/konf-go/utils"
	"github.com/spf13/cobra"
	k8s "k8s.io/client-go/tools/clientcmd/api/v1"
	"sigs.k8s.io/yaml"
)

type konfigFile struct {
	FileName string
	Content  k8s.Config
}

// TODO rewrite this package to make use of afero

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import kubeconfigs into konf store",
	Long: `Import kubeconfigs into konf store

It is important that you import all configs first, as konf requires each config to only
contain a single context. Import will take care of splitting if necessary.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		fpath := args[0] // safe, as we specify cobra.ExactArgs(1)
		f, err := os.Open(fpath)
		if err != nil {
			return err
		}

		confs, err := determineConfigs(f)
		if err != nil {
			return err
		}

		for _, conf := range confs {
			f, err = os.Create(conf.FileName)
			if err != nil {
				return err
			}
			err = writeConfig(f, &conf.Content)
			if err != nil {
				return err
			}
			log.Printf("Imported konf from %s successfully into %s\n", fpath, conf.FileName)
		}

		return nil
	},
}

// determineConfigs returns the individual configs from a konfigfile
// This is required as konfig requires each kubeconfig in its store to
// only contain a single context
// If more than one cluster is in a kubeconfig, determineConfig will split it up
// into multiple konfigFile and returns them as a slice
func determineConfigs(conf io.Reader) ([]*konfigFile, error) {

	// basically should be as simple as
	// 1. Loop through all the contexts
	// 2. Find the corresponding cluster for each context
	// 3. Find the corresponding user for each context
	// 4. Create a new konfigFile for each context mapped to its cluster

	var origConf k8s.Config
	b := new(bytes.Buffer)
	_, err := b.ReadFrom(conf)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(b.Bytes(), &origConf)
	if err != nil {
		return nil, err
	}

	var konfs = []*konfigFile{}
	for _, curCon := range origConf.Contexts {

		cluster := k8s.NamedCluster{}
		for _, curCl := range origConf.Clusters {
			if curCl.Name == curCon.Context.Cluster {
				cluster = curCl
				break
			}
		}
		user := k8s.NamedAuthInfo{}
		for _, curU := range origConf.AuthInfos {
			if curU.Name == curCon.Context.AuthInfo {
				user = curU
				break
			}
		}

		var konf konfigFile
		// TODO it might make sense to build in a duplicate detection here. This would ensure that the store is trustworthy, which in return makes it easy for
		// TODO the set command as it does not need any verification
		id := utils.IDFromClusterAndContext(cluster.Name, curCon.Name)
		konf.FileName = utils.StorePathForID(id)
		konf.Content.AuthInfos = append(konf.Content.AuthInfos, user)
		konf.Content.Clusters = append(konf.Content.Clusters, cluster)
		konf.Content.Contexts = append(konf.Content.Contexts, curCon)

		konf.Content.APIVersion = origConf.APIVersion
		konf.Content.Kind = origConf.Kind
		konf.Content.CurrentContext = curCon.Name

		konfs = append(konfs, &konf)
	}

	return konfs, nil
}

func writeConfig(w io.Writer, conf *k8s.Config) error {
	c, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}

	_, err = w.Write(c)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(importCmd)
}
