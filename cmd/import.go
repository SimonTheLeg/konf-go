package cmd

import (
	"fmt"
	"io"

	"github.com/simontheleg/konf-go/konf"
	log "github.com/simontheleg/konf-go/log"
	"github.com/simontheleg/konf-go/store"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	k8s "k8s.io/client-go/tools/clientcmd/api/v1"
)

type konfFile struct {
	FilePath string
	Content  k8s.Config
}

type importCmd struct {
	fs afero.Fs

	determineConfigs     func(io.Reader) ([]*konf.Konfig, error)
	writeConfig          func(afero.Fs, *konf.Konfig) (string, error)
	deleteOriginalConfig func(afero.Fs, string) error

	move bool

	cmd *cobra.Command
}

func newImportCmd() *importCmd {
	fs := afero.NewOsFs()

	ic := &importCmd{
		fs:                   fs,
		determineConfigs:     konf.KonfsFromKubeconfig,
		writeConfig:          store.WriteKonfToStore,
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

	f, err := c.fs.Open(fpath)
	if err != nil {
		return err
	}

	konfs, err := c.determineConfigs(f)
	if err != nil {
		return err
	}

	if len(konfs) == 0 {
		return fmt.Errorf("no contexts found in file %q", fpath)
	}

	for _, k := range konfs {
		_, err = c.writeConfig(c.fs, k)
		if err != nil {
			return err
		}
		log.Info("Imported konf from %q successfully into %q\n", fpath, k.StorePath)
	}

	if c.move {
		if err := c.deleteOriginalConfig(c.fs, fpath); err != nil {
			return err
		}
		log.Info("Successfully deleted original kubeconfig at %q", fpath)
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
