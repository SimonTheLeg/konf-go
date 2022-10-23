package cmd

import (
	"fmt"

	"github.com/simontheleg/konf-go/konf"
	log "github.com/simontheleg/konf-go/log"
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

	determineConfigs     func(afero.Fs, string) ([]*konfFile, error)
	writeConfig          func(afero.Fs, *konfFile) error
	deleteOriginalConfig func(afero.Fs, string) error

	move bool

	cmd *cobra.Command
}

func newImportCmd() *importCmd {
	fs := afero.NewOsFs()

	ic := &importCmd{
		fs:                   fs,
		determineConfigs:     determineConfigs,
		writeConfig:          writeConfig,
		deleteOriginalConfig: deleteOriginalConfig,
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

	ic.cmd.Flags().BoolVarP(&ic.move, "move", "m", false, "whether the original kubeconfig should be deleted after successful import (default is false)")

	return ic
}

// because import is a reserved word, we have to slightly rename this :)
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
		log.Info("Imported konf from %q successfully into %q\n", fpath, conf.FilePath)
	}

	if c.move {
		if err := c.deleteOriginalConfig(c.fs, fpath); err != nil {
			return err
		}
		log.Info("Successfully deleted original kubeconfig at %q", fpath)
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

		var k konfFile
		// TODO it might make sense to build in a duplicate detection here. This would ensure that the store is trustworthy, which in return makes it easy for
		// TODO the set command as it does not need any verification
		id := konf.IDFromClusterAndContext(cluster.Name, curCon.Name)
		k.FilePath = id.StorePath()
		k.Content.AuthInfos = append(k.Content.AuthInfos, user)
		k.Content.Clusters = append(k.Content.Clusters, cluster)
		k.Content.Contexts = append(k.Content.Contexts, curCon)

		k.Content.APIVersion = origConf.APIVersion
		k.Content.Kind = origConf.Kind
		k.Content.CurrentContext = curCon.Name

		konfs = append(konfs, &k)
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

func deleteOriginalConfig(f afero.Fs, path string) error {
	err := f.Remove(path)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(newImportCmd().cmd)
}
