package cmd

import (
	"github.com/simontheleg/konf-go/konf"
	"github.com/simontheleg/konf-go/log"
	"github.com/simontheleg/konf-go/prompt"
	"github.com/simontheleg/konf-go/store"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type deleteCmd struct {
	fs afero.Fs

	fetchconfs       func(afero.Fs) ([]*store.Metadata, error)
	selectSingleKonf func(afero.Fs, prompt.RunFunc) (konf.KonfID, error)
	deleteKonfWithID func(afero.Fs, konf.KonfID) error
	idsForGlobs      func(afero.Fs, []string) ([]konf.KonfID, error)
	prompt           prompt.RunFunc

	cmd *cobra.Command
}

func newDeleteCommand() *deleteCmd {
	dc := &deleteCmd{
		fs:               afero.NewOsFs(),
		fetchconfs:       store.FetchAllKonfs,
		selectSingleKonf: selectSingleKonf,
		deleteKonfWithID: deleteKonfWithID,
		idsForGlobs:      idsForGlobs,
		prompt:           prompt.Terminal,
	}

	dc.cmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete kubeconfig",
		Long: `Delete one or multiple kubeconfigs

Examples:
-> 'delete' run selection prompt for deletion
-> 'delete <konfig id> [<konfig id 2>]' delete specific konf(s)
-> 'delete "my-konf*"' delete konf matching fileglob
`,
		RunE:              dc.delete,
		ValidArgsFunction: dc.completeDelete,
	}

	return dc
}

func (c *deleteCmd) delete(cmd *cobra.Command, args []string) error {
	var ids []konf.KonfID
	var err error

	if len(args) == 0 {
		var id konf.KonfID
		id, err = c.selectSingleKonf(c.fs, c.prompt)
		if err != nil {
			return err
		}
		ids = append(ids, id)
	} else {
		ids, err = c.idsForGlobs(c.fs, args)
		if err != nil {
			return err
		}
	}

	for _, id := range ids {
		if err := c.deleteKonfWithID(c.fs, id); err != nil {
			return err
		}
	}

	log.Info("Deletion successful. If for security reasons you want to remove any currently active konfs, close the shell sessions they are used in.")
	return nil
}

func deleteKonfWithID(fs afero.Fs, id konf.KonfID) error {
	if err := fs.Remove(id.StorePath()); err != nil {
		return err
	}
	log.Info("Successfully deleted konf %q at %q", id, id.StorePath())
	return nil
}

// idsForGlobs takes in a slice of patterns and returns corresponding IDs from
// the konfStore
func idsForGlobs(f afero.Fs, patterns []string) ([]konf.KonfID, error) {
	var ids []konf.KonfID
	for _, pattern := range patterns {
		metadata, err := store.FetchKonfsForGlob(f, pattern) // resolve any globs among the arguments
		if err != nil {
			return nil, err
		}
		for _, f := range metadata {
			id := konf.IDFromClusterAndContext(f.Cluster, f.Context)
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func (c *deleteCmd) completeDelete(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	konfs, err := store.FetchAllKonfs(c.fs)
	if err != nil {
		// if the store is just empty, return no suggestions, instead of throwing an error
		if _, ok := err.(*store.EmptyStore); ok {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}

		cobra.CompDebugln(err.Error(), true)
		return nil, cobra.ShellCompDirectiveError
	}

	sug := []string{}
	for _, k := range konfs {
		// with the current design of 'set', we need to return the ID here in the autocomplete as the first part of the completion
		// as it is directly passed to set
		sug = append(sug, string(konf.IDFromClusterAndContext(k.Cluster, k.Context)))
	}

	return sug, cobra.ShellCompDirectiveNoFileComp
}

func init() {
	rootCmd.AddCommand(newDeleteCommand().cmd)
}
