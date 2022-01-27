package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/simontheleg/konf-go/utils"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	k8s "k8s.io/client-go/tools/clientcmd/api/v1"
	"sigs.k8s.io/yaml"
)

type konfFile struct {
	FilePath string
	Content  k8s.Config
}

type importCmd struct {
	fs afero.Fs

	determineConfigs func(afero.Fs, string) ([]*konfFile, error)
	writeConfig      func(afero.Fs, *konfFile) error

	cmd *cobra.Command
}

func newImportCmd() *importCmd {
	fs := afero.NewOsFs()

	ic := &importCmd{
		fs:               fs,
		determineConfigs: determineConfigs,
		writeConfig:      writeConfig,
	}

	ic.cmd = &cobra.Command{
		Use:   "import",
		Short: "Import kubeconfigs into konf store",
		Long: `Import kubeconfigs into konf store

It is important that you import all configs first, as konf requires each config to only
contain a single context. Import will take care of splitting if necessary.`,
		Args: cobra.ExactArgs(1),
		RunE: ic.importf,
	}

	ic.cmd.SetOut(os.Stderr)

	return ic
}

// because import is a reserved word, we have to slighly rename this :)
func (c *importCmd) importf(cmd *cobra.Command, args []string) error {
	fpath := args[0] // safe, as we specify cobra.ExactArgs(1)

	confs, err := c.determineConfigs(c.fs, fpath)
	if err != nil {
		return err
	}

	if len(confs) == 0 {
		return fmt.Errorf("no contexts found in file %q", fpath)
	}

	for _, conf := range confs {
		err = c.writeConfig(c.fs, conf)
		if err != nil {
			return err
		}
		log.Printf("Imported konf from %q successfully into %q\n", fpath, conf.FilePath)
	}

	return nil
}

// determineConfigs returns the individual configs from a konfigfile
// This is required as konfig requires each kubeconfig in its store to
// only contain a single context
// If more than one cluster is in a kubeconfig, determineConfig will split it up
// into multiple konfigFile and returns them as a slice
func determineConfigs(f afero.Fs, fpath string) ([]*konfFile, error) {

	b, err := afero.ReadFile(f, fpath)
	if err != nil {
		return nil, err
	}

	var origConf k8s.Config
	err = yaml.Unmarshal(b, &origConf)
	if err != nil {
		return nil, err
	}

	// basically should be as simple as
	// 1. Loop through all the contexts
	// 2. Find the corresponding cluster for each context
	// 3. Find the corresponding user for each context
	// 4. Create a new konfigFile for each context mapped to its cluster

	var konfs = []*konfFile{}
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

		var konf konfFile
		// TODO it might make sense to build in a duplicate detection here. This would ensure that the store is trustworthy, which in return makes it easy for
		// TODO the set command as it does not need any verification
		id := utils.IDFromClusterAndContext(cluster.Name, curCon.Name)
		konf.FilePath = utils.StorePathForID(id)
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

func writeConfig(f afero.Fs, kf *konfFile) error {
	b, err := yaml.Marshal(kf.Content)
	if err != nil {
		return err
	}

	err = afero.WriteFile(f, kf.FilePath, b, utils.KonfPerm)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(newImportCmd().cmd)
}
