package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type shellwrapperCmd struct {
	cmd *cobra.Command
}

func newShellwrapperCmd() *shellwrapperCmd {
	sc := shellwrapperCmd{}

	sc.cmd = &cobra.Command{
		Use:   "shellwrapper",
		Short: "Shell wrapper and hooks for konf command",
		Long: `Shell wrapper and hooks for konf command

The output of this command should be sourced in your .rc file.

See https://github.com/SimonTheLeg/konf-go#installation on how to do so
`,
		RunE: sc.shellwrapper,
		Args: cobra.ExactArgs(1),
	}

	sc.cmd.SetOut(os.Stderr)

	return &sc
}

func (c *shellwrapperCmd) shellwrapper(cmd *cobra.Command, args []string) error {
	var wrapper string

	zsh := `
konf() {
  res=$(konf-go $@)
  # protect against an empty command
  # Note we cannot do something like if "$1 == set" and only run the export on set commands as cmd flags can be at any position in our cli
  if [[ $res != "" ]] then
    export KUBECONFIG=$res
  fi
}
konf_cleanup() {
  konf-go cleanup
}
add-zsh-hook zshexit konf_cleanup
`

	if args[0] == "zsh" { // safe as we specify cobra.ExactArgs(1)
		wrapper = zsh
	} else {
		return fmt.Errorf("konf currently does not support %s", args[0])
	}

	fmt.Println(wrapper)

	return nil
}

func init() {
	rootCmd.AddCommand(newShellwrapperCmd().cmd)
}
