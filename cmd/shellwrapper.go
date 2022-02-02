package cmd

import (
	"fmt"

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

	return &sc
}

func (c *shellwrapperCmd) shellwrapper(cmd *cobra.Command, args []string) error {
	var wrapper string
	var zsh = `
konf() {
  res=$(konf-go $@)
  # only change $KUBECONFIG if instructed by konf-go
  if [[ $res == "KUBECONFIGCHANGE:"* ]]
  then
    # this basically takes the line and cuts out the KUBECONFIGCHANGE Part
    export KUBECONFIG="${res#*KUBECONFIGCHANGE:}"
  else
    # this makes --help work
    echo "${res}"
  fi
}
konf_cleanup() {
  konf-go cleanup
}
add-zsh-hook zshexit konf_cleanup
`

	var bash = `
konf() {
  res=$(konf-go $@)
  # only change $KUBECONFIG if instructed by konf-go
  if [[ $res == "KUBECONFIGCHANGE:"* ]]
  then
    # this basically takes the line and cuts out the KUBECONFIGCHANGE Part
    export KUBECONFIG="${res#*KUBECONFIGCHANGE:}"
  else
    # this makes --help work
    echo "${res}"
  fi
}
konf_cleanup() {
  konf-go cleanup
}

trap konf_cleanup EXIT
`

	switch args[0] {
	case "zsh":
		wrapper = zsh
	case "bash":
		wrapper = bash
	default:
		return fmt.Errorf("konf currently does not support %s", args[0])
	}

	fmt.Println(wrapper)

	return nil
}

func init() {
	rootCmd.AddCommand(newShellwrapperCmd().cmd)
}
