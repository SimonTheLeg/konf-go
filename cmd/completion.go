package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// Main modifications from the auto-generated completion.go by cobra are to accommodate the konf wrapper.
// A detailed explanation can be found on a per shell basis down below
// The preset for this file is taken from https://github.com/spf13/cobra/blob/master/shell_completions.md

type completionCmd struct {
	cmd *cobra.Command
}

func newCompletionCmd() *completionCmd {
	cc := completionCmd{}

	cc.cmd = &cobra.Command{
		Use:   "completion [bash|zsh]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:

  # To load completions for each session, add this to your zshrc:

  source <(konf completion bash)

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it. Simply add the following to your .zshrc:

  autoload -U compinit && compinit

  # To load completions for each session, add this to your zshrc:
	source <(konf completion zsh)

`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh"},
		Args:                  cobra.ExactValidArgs(1),
		RunE:                  cc.completion,
	}

	return &cc
}

func (c *completionCmd) completion(cmd *cobra.Command, args []string) error {
	switch args[0] {

	case "zsh":
		// This allows for also using 'source <(konf completion zsh)' with zsh, similar to bash.
		// Basically it just adds the compdef command so it can be run. Taken from kubectl, who
		// do a similar thing
		zshHeader := "#compdef _konf konf\ncompdef _konf konf\n"

		// So per default cobra makes use of the words[] array that zsh provides to you in completion funcs.
		// Words is an array that contains all words that have been typed by the user before hitting tab
		// Now cobra takes words[1] which is equal to the name of the comand and uses this to call completion on it
		// However in our case this does not work as words[1] points to 'konf' which is the wrapper and not the binary
		// In order to solve this we have to ensure that words[1] equates to konf-go, which is the binary.
		// Currently I have found, the fastest way to do this is by inserting a line to overwrite words[1]. This is
		// because the words[1] reference is used throughout the script and I would not want to replace all of it
		var b bytes.Buffer
		err := rootCmd.GenZshCompletion(&b)
		if err != nil {
			return err
		}
		anchor := "local -a completions" // this is basically a line early in the original script that we are going to cling onto
		genZsh := strings.Replace(b.String(), anchor, anchor+"\n    words[1]=\"konf-go\"", 1)

		os.Stdout.WriteString(zshHeader + genZsh)

	case "bash":
		var b bytes.Buffer
		err := rootCmd.GenBashCompletionV2(&b, true)
		if err != nil {
			return err
		}
		anchor := "local requestComp lastParam lastChar args"
		genBash := strings.Replace(b.String(), anchor, anchor+"\n    words[0]=\"konf-go\"", 1) // basically the same as for zsh, but this words[] is zero-indexed

		os.Stdout.WriteString(genBash)

	default:
		return fmt.Errorf("konf currently does not support autocompletions for %s", args[0])
	}

	return nil
}

func init() {
	rootCmd.AddCommand(newCompletionCmd().cmd)
}
