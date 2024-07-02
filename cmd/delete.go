package cmd

import (
	"github.com/simontheleg/konf-go/config"
	"github.com/simontheleg/konf-go/konf"
	"github.com/simontheleg/konf-go/log"
	"github.com/simontheleg/konf-go/prompt"
	"github.com/simontheleg/konf-go/store"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type deleteCmd struct {
	sm               *store.Storemanager
	fetchconfs       func() ([]*store.Metadata, error)
	selectSingleKonf func(*store.Storemanager, prompt.RunFunc) (konf.KonfID, error)
	deleteKonfWithID func(*store.Storemanager, konf.KonfID) error
	idsForGlobs      func(*store.Storemanager, []string) ([]konf.KonfID, error)
	prompt           prompt.RunFunc

	cmd *cobra.Command
}

func newDeleteCommand() *deleteCmd {
	fs := afero.NewOsFs()
	sm := &store.Storemanager{Fs: fs, Activedir: config.ActiveDir(), Storedir: config.StoreDir()}
	dc := &deleteCmd{
		sm:               sm,
		fetchconfs:       sm.FetchAllKonfs,
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
		id, err = c.selectSingleKonf(c.sm, c.prompt)
		if err != nil {
			return err
		}
		ids = append(ids, id)
	} else {
		ids, err = c.idsForGlobs(c.sm, args)
		if err != nil {
			return err
		}
	}

	for _, id := range ids {
		if err := c.deleteKonfWithID(c.sm, id); err != nil {
			return err
		}
	}

	log.Info("Deletion successful. If for security reasons you want to remove any currently active konfs, close the shell sessions they are used in.")
	return nil
}

func deleteKonfWithID(sm *store.Storemanager, id konf.KonfID) error {
	path := sm.StorePathFromID(id)
	if err := sm.Fs.Remove(path); err != nil {
		return err
	}
	log.Info("Successfully deleted konf %q at %q", id, path)
	return nil
}

// idsForGlobs takes in a slice of patterns and returns corresponding IDs from
// the konfStore
func idsForGlobs(sm *store.Storemanager, patterns []string) ([]konf.KonfID, error) {
	var ids []konf.KonfID
	for _, pattern := range patterns {
		metadata, err := sm.FetchKonfsForGlob(pattern) // resolve any globs among the arguments
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
	konfs, err := c.fetchconfs()
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
