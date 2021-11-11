package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	k8s "k8s.io/client-go/tools/clientcmd/api/v1"
	"sigs.k8s.io/yaml"
)

type konfigFile struct {
	FileName string
	Content  k8s.Config
}

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

		err = os.MkdirAll(filepath.Dir(viper.GetString("konfStore")), 0770)
		if err != nil {
			return err
		}

		for _, conf := range confs {
			fmt.Printf("Attempting to create file %s\n", conf.FileName)
			f, err = os.Create(conf.FileName)
			if err != nil {
				return err
			}
			err = writeConfig(f, &conf.Content)
			if err != nil {
				return err
			}
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
		// I have chosen this combination as it is fairly unique among multiple configs. I decided against using just context.name as a lot of times the context is just called "default", which results in lots of naming collisions
		konf.FileName = viper.GetString("konfStore") + curCon.Name + "_" + cluster.Name + ".yaml"
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// importCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// importCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
