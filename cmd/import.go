package cmd

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/simontheleg/konf-go/config"
	"github.com/simontheleg/konf-go/konf"
	log "github.com/simontheleg/konf-go/log"
	"github.com/simontheleg/konf-go/store"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type importCmd struct {
	sm *store.Storemanager

	filesForDir          func(*store.Storemanager, string) ([]*FileWithPath, error)
	determineConfigs     func(io.Reader) ([]*konf.Konfig, error)
	writeConfig          func(*konf.Konfig) (string, error)
	deleteOriginalConfig func(*store.Storemanager, string) error

	move bool

	cmd *cobra.Command
}

func newImportCmd() *importCmd {
	fs := afero.NewOsFs()
	sm := &store.Storemanager{Activedir: config.ActiveDir(), Storedir: config.StoreDir(), Fs: fs}

	ic := &importCmd{
		sm: sm,

		filesForDir:          filesForDir,
		determineConfigs:     konf.KonfsFromKubeconfig,
		writeConfig:          sm.WriteKonfToStore,
		deleteOriginalConfig: deleteOriginalConfig,
	}

	ic.cmd = &cobra.Command{
		Use:   "import",
		Short: "Import kubeconfigs into konf store",
		Long: `Import one or multiple kubeconfigs

Examples:
-> 'konf import /mydir/myfile.yaml' will import a single kubeconfig
-> 'konf import /mydir' will import all files in that directory

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
	searchpath := args[0] // safe, as we specify cobra.ExactArgs(1)

	files, err := c.filesForDir(c.sm, searchpath)
	if err != nil {
		return err
	}

	// we need to wrap this here, as we require the original importpath
	type importKonf struct {
		Konf       *konf.Konfig
		ImportPath string
	}
	konfs := []*importKonf{}
	for _, file := range files {
		ks, err := c.determineConfigs(file.File)
		if err != nil {
			return err
		}
		for _, k := range ks {
			konfs = append(konfs, &importKonf{Konf: k, ImportPath: file.FilePath})
		}
	}

	if len(konfs) == 0 {
		errMsg := "no contexts found in the following file(s):\n"
		for _, file := range files {
			errMsg += fmt.Sprintf("\t- %q\n", file.FilePath)
		}
		return fmt.Errorf(errMsg)
	}

	for _, k := range konfs {
		_, err = c.writeConfig(k.Konf)
		if err != nil {
			return err
		}
		log.Info("Imported konf from %q successfully into %q\n", k.ImportPath, k.Konf.StorePath)
	}

	if c.move {
		for _, f := range files {
			if err := c.deleteOriginalConfig(c.sm, f.FilePath); err != nil {
				return err
			}
			log.Info("Successfully deleted original kubeconfig file at %q", f.FilePath)
		}
	}

	return nil
}

func deleteOriginalConfig(sm *store.Storemanager, path string) error {
	// TODO refactor: This action should be provided by a convenience func inside the store package
	err := sm.Fs.Remove(path)
	if err != nil {
		return err
	}
	return nil
}

// wrapper struct, so we can return the original path as well
type FileWithPath struct {
	FilePath string
	File     afero.File
}

// filesForDir extracts all relevant files from a dir.
//
// Relevant is defined as in no subdirectories and no hidden files. If a file
// instead of a dir is supplied, the file will be returned
func filesForDir(sm *store.Storemanager, path string) ([]*FileWithPath, error) {
	fileinfo, err := sm.Fs.Stat(path)
	if err != nil {
		return nil, err
	}

	files := []*FileWithPath{}

	if fileinfo.IsDir() {
		fileinfos, err := afero.ReadDir(sm.Fs, path)
		if err != nil {
			return nil, err
		}
		for _, p := range fileinfos {
			if p.IsDir() || strings.HasPrefix(p.Name(), ".") {
				continue // skip any directories or hidden files
			}
			fpath := filepath.Join(path, p.Name())
			file, err := sm.Fs.Open(fpath)
			if err != nil {
				return nil, err
			}
			files = append(files, &FileWithPath{FilePath: fpath, File: file})
		}
	} else {
		file, err := sm.Fs.Open(path)
		if err != nil {
			return nil, err
		}
		// by calling file.Name(), we resolve any path ambiguities (e.g. "./dir" and
		// "dir")
		files = append(files, &FileWithPath{FilePath: file.Name(), File: file})
	}

	return files, nil
}
