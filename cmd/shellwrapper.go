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
  local res result
  local KONF_SHELL_PID=$$
  export KONF_SHELL_PID
  # Completion output does not change shell state and can be streamed directly.
  case "${1-}" in
    completion|__complete|__completeNoDesc) command konf-go "$@"; return $? ;;
  esac
  if res=$(command konf-go "$@"); then
    result=0
  else
    result=$?
  fi
  # only change $KUBECONFIG if instructed by konf-go
  if [[ $result == 0 && $res == "KUBECONFIGCHANGE:"* ]]
  then
    # this basically takes the line and cuts out the KUBECONFIGCHANGE Part
    export KUBECONFIG="${res#*KUBECONFIGCHANGE:}"
  else
    # this makes --help work
    printf '%s\n' "$res"
  fi
  return "$result"
}
konf_cleanup() {
  KONF_SHELL_PID=$$ command konf-go cleanup
}
add-zsh-hook zshexit konf_cleanup
`

	var bash = `
konf() {
  local res result
  local KONF_SHELL_PID=$$
  export KONF_SHELL_PID
  # Completion output does not change shell state and can be streamed directly.
  case "${1-}" in
    completion|__complete|__completeNoDesc) command konf-go "$@"; return $? ;;
  esac
  if res=$(command konf-go "$@"); then
    result=0
  else
    result=$?
  fi
  # only change $KUBECONFIG if instructed by konf-go
  if [[ $result == 0 && $res == "KUBECONFIGCHANGE:"* ]]
  then
    # this basically takes the line and cuts out the KUBECONFIGCHANGE Part
    export KUBECONFIG="${res#*KUBECONFIGCHANGE:}"
  else
    # this makes --help work
    printf '%s\n' "$res"
  fi
  return "$result"
}
konf_cleanup() {
  KONF_SHELL_PID=$$ command konf-go cleanup
}

trap konf_cleanup EXIT
`

	var fish = `
function konf -w konf-go
    set -lx KONF_SHELL_PID $fish_pid
    if test (count $argv) -gt 0; and contains -- $argv[1] completion __complete __completeNoDesc
        command konf-go $argv
        return $status
    end
    set -f res (konf-go $argv)
    set -f result $status
    # only change $KUBECONFIG if instructed by konf-go
    if test $result -eq 0; and string match -q 'KUBECONFIGCHANGE:*' $res
        # this basically takes the line and cuts out the KUBECONFIGCHANGE Part
        set -gx KUBECONFIG (string replace -r '^KUBECONFIGCHANGE:' '' $res)
    else
        # this makes --help work
        # because fish does not support bracketed vars, we use printf instead
        printf "%s\n" $res
    end
    return $result
end

function konf_cleanup
    set -lx KONF_SHELL_PID $fish_pid
    command konf-go cleanup
end

trap konf_cleanup EXIT
`

	switch args[0] {
	case "zsh":
		wrapper = zsh
	case "bash":
		wrapper = bash
	case "fish":
		wrapper = fish
	default:
		return fmt.Errorf("konf currently does not support %s", args[0])
	}

	_, err := fmt.Fprintln(cmd.OutOrStdout(), wrapper)
	return err
}
