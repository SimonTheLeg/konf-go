package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/simontheleg/konf-go/config"
	"github.com/simontheleg/konf-go/konf"
	log "github.com/simontheleg/konf-go/log"
	"github.com/simontheleg/konf-go/prompt"
	"github.com/simontheleg/konf-go/store"
	"github.com/simontheleg/konf-go/utils"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type setCmd struct {
	sm *store.Storemanager

	cmd *cobra.Command
}

func newSetCommand() *setCmd {
	fs := afero.NewOsFs()
	sm := &store.Storemanager{Fs: fs, Activedir: config.ActiveDir(), Storedir: config.StoreDir(), LatestKonfPath: config.LatestKonfFilePath()}
	sc := &setCmd{
		sm: sm,
	}

	sc.cmd = &cobra.Command{
		Use:   `set`,
		Short: "Set kubeconfig to use in current shell",
		Args:  cobra.MaximumNArgs(1),
		Long: `Sets kubeconfig to use or start picker dialogue.
	
Examples:
-> 'set' run konf selection
-> 'set <konfig id>' set a specific konf
-> 'set -' set to last used konf
`,
		RunE:              sc.set,
		ValidArgsFunction: sc.completeSet,
	}

	return sc
}

func (c *setCmd) set(cmd *cobra.Command, args []string) error {
	// TODO if I stay with the mocking approach used in commands like
	// namespace. This part should be refactored to allow for mocking
	// the downstream funcs in order to test the if-else logic
	var id konf.KonfID
	var err error

	if len(args) == 0 {
		id, err = selectSingleKonf(c.sm, prompt.Terminal)
		if err != nil {
			return err
		}
	} else if args[0] == "-" {
		id, err = idOfLatestKonf(c.sm)
		if err != nil {
			return err
		}
	} else {
		id = konf.KonfID(args[0])
	}

	context, err := setContext(id, c.sm)
	if err != nil {
		return err
	}
	err = saveLatestKonf(c.sm, id)
	if err != nil {
		return fmt.Errorf("could not save latest konf. As a result 'konf set -' might not work: %q ", err)
	}

	log.Info("Setting context to %q\n", id)

	// By printing out to stdout, we pass the value to our zsh hook, which then sets $KUBECONFIG to it
	// Both operate on the convention to use "KUBECONFIGCHANGE:<new-path>". If you change this part in
	// here, do not forget to update shellwraper.go
	fmt.Println("KUBECONFIGCHANGE:" + context)

	return nil
}

func (c *setCmd) completeSet(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	konfs, err := c.sm.FetchAllKonfs()
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

// TODO make a decision where this code should be placed. Currently it does not
// make a lot of sense to bring it into its own package as it is at the nice
// intersection between utilizing two packages to fulfil business logic However
// it is also being used by two commands: "set" and "delete". But because
// they are in the same package, we also cannot easily duplicate the code for
// each
func selectSingleKonf(sm *store.Storemanager, pf prompt.RunFunc) (konf.KonfID, error) {
	k, err := sm.FetchAllKonfs()
	if err != nil {
		return "", err
	}
	p := createSetPrompt(k)
	selPos, err := pf(p)
	if err != nil {
		return "", err
	}

	if selPos >= len(k) {
		return "", fmt.Errorf("invalid selection %d", selPos)
	}
	sel := k[selPos]

	return konf.IDFromClusterAndContext(sel.Cluster, sel.Context), nil
}

func idOfLatestKonf(sm *store.Storemanager) (konf.KonfID, error) {
	b, err := afero.ReadFile(sm.Fs, sm.LatestKonfPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("could not select latest konf, because no konf was yet set")
		}
		return "", err
	}
	return konf.KonfID(b), nil
}

func setContext(id konf.KonfID, sm *store.Storemanager) (string, error) {
	k, err := afero.ReadFile(sm.Fs, sm.StorePathFromID(id))
	if err != nil {
		return "", err
	}

	ppid := os.Getppid()
	konfID := konf.IDFromProcessID(ppid)
	activeKonf := sm.ActivePathFromID(konfID)
	err = afero.WriteFile(sm.Fs, activeKonf, k, utils.KonfPerm)
	if err != nil {
		return "", err
	}

	return activeKonf, nil

}

func saveLatestKonf(sm *store.Storemanager, id konf.KonfID) error {
	return afero.WriteFile(sm.Fs, sm.LatestKonfPath, []byte(id), utils.KonfPerm)
}

func createSetPrompt(options []*store.Metadata) *promptui.Select {
	// TODO use ssh/terminal to get the terminalsize and set trunc accordingly https://stackoverflow.com/questions/16569433/get-terminal-size-in-go
	trunc := 25
	promptInactive, promptActive, label, fmap := prompt.NewTableOutputTemplates(trunc)

	// Wrapper is required as we need access to options, but the methodSignature from promptUI
	// requires you to only pass an index not the whole func
	// This wrapper allows us to unit-test the FuzzyFilterKonf func better
	var wrapFuzzyFilterKonf = func(input string, index int) bool {
		return prompt.FuzzyFilterKonf(input, options[index])
	}

	prompt := promptui.Select{
		Label: label,
		Items: options,
		Templates: &promptui.SelectTemplates{
			Active:   promptActive,
			Inactive: promptInactive,
			FuncMap:  fmap,
		},
		HideSelected: true,
		Stdout:       os.Stderr,
		Searcher:     wrapFuzzyFilterKonf,
		Size:         15,
	}
	return &prompt
}
