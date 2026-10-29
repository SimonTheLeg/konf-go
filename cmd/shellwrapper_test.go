package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/simontheleg/konf-go/testhelper"
)

func TestShellWrapperBehavior(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish"} {
		t.Run(shell, func(t *testing.T) {
			path, err := exec.LookPath(shell)
			if err != nil {
				t.Skipf("%s not installed", shell)
			}
			dir := t.TempDir()
			fake := `#!/bin/sh
printf '<%s>\n' "$@" > "$KONF_TEST_DIR/arguments"
printf '%s\n' "$KONF_SHELL_PID" > "$KONF_TEST_DIR/shell-pid"
case "${1-}" in
  fail) printf 'KUBECONFIGCHANGE:invalid\n'; exit 37 ;;
  set) printf 'KUBECONFIGCHANGE:/config with spaces.yaml\n' ;;
  completion|__complete|__completeNoDesc) printf '%s\n' '-n' 'two words' 'KUBECONFIGCHANGE:not-state' ;;
  cleanup) exit 0 ;;
  *) printf 'ordinary output\n' ;;
esac
`
			if err := os.WriteFile(filepath.Join(dir, "konf-go"), []byte(fake), 0700); err != nil {
				t.Fatal(err)
			}
			sc := newShellwrapperCmd()
			var wrapper bytes.Buffer
			sc.cmd.SetOut(&wrapper)
			if err := sc.shellwrapper(sc.cmd, []string{shell}); err != nil {
				t.Fatal(err)
			}
			preamble := "set -eu\n"
			if shell == "zsh" {
				preamble += "autoload -Uz add-zsh-hook\n"
			}
			checks := `
res=untouched
export KUBECONFIG=original
konf 'two words' '*' '' > "$KONF_TEST_DIR/output"
expected=$(printf '<%s>\n' 'two words' '*' '')
[ "$(cat "$KONF_TEST_DIR/arguments")" = "$expected" ] || exit 41
[ "$(cat "$KONF_TEST_DIR/output")" = 'ordinary output' ] || exit 42
[ "$res" = untouched ] || exit 43
[ "$(cat "$KONF_TEST_DIR/shell-pid")" = "$$" ] || exit 50
[ -z "${KONF_SHELL_PID-}" ] || exit 51
result=0
konf fail > "$KONF_TEST_DIR/output" || result=$?
[ "$result" = 37 ] || exit 44
[ "$KUBECONFIG" = original ] || exit 45
konf set 'two words'
[ "$KUBECONFIG" = '/config with spaces.yaml' ] || exit 46
for operation in completion __complete __completeNoDesc; do
  konf "$operation" zsh > "$KONF_TEST_DIR/output"
  expected=$(printf '%s\n' '-n' 'two words' 'KUBECONFIGCHANGE:not-state')
  [ "$(cat "$KONF_TEST_DIR/output")" = "$expected" ] || exit 47
  [ "$KUBECONFIG" = '/config with spaces.yaml' ] || exit 48
done
konf > "$KONF_TEST_DIR/output"
[ "$(cat "$KONF_TEST_DIR/output")" = 'ordinary output' ] || exit 49
`
			if shell == "fish" {
				preamble = ""
				checks = `
set -gx KUBECONFIG original
konf 'two words' '*' '' > "$KONF_TEST_DIR/output"; or exit 41
set expected (printf '<%s>\n' 'two words' '*' '' | string collect)
test (cat "$KONF_TEST_DIR/arguments" | string collect) = "$expected"; or exit 42
test (cat "$KONF_TEST_DIR/shell-pid") = "$fish_pid"; or exit 51
set -q KONF_SHELL_PID; and exit 52
konf fail > "$KONF_TEST_DIR/output"
test $status -eq 37; or exit 43
test "$KUBECONFIG" = original; or exit 44
konf set 'two words'; or exit 45
test "$KUBECONFIG" = '/config with spaces.yaml'; or exit 46
for operation in completion __complete __completeNoDesc
  konf "$operation" fish > "$KONF_TEST_DIR/output"; or exit 47
  set expected (printf '%s\n' '-n' 'two words' 'KUBECONFIGCHANGE:not-state' | string collect)
  test (cat "$KONF_TEST_DIR/output" | string collect) = "$expected"; or exit 48
  test "$KUBECONFIG" = '/config with spaces.yaml'; or exit 49
end
konf > "$KONF_TEST_DIR/output"; or exit 50
`
			}
			command := exec.Command(path, "-c", preamble+wrapper.String()+checks)
			command.Env = append(os.Environ(), "PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"), "KONF_TEST_DIR="+dir)
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("%s wrapper failed: %v\n%s", shell, err, output)
			}
		})
	}
}

func TestShellWrapperCmd(t *testing.T) {

	tt := map[string]struct {
		args   []string
		ExpErr error
	}{
		"zsh arg": {
			[]string{"zsh"},
			nil,
		},
		"bash arg": {
			[]string{"bash"},
			nil,
		},
		"fish arg": {
			[]string{"fish"},
			nil,
		},
		"invalid arg": {
			[]string{"invalid"},
			fmt.Errorf("konf currently does not support invalid"),
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			cs := newShellwrapperCmd()
			cmd := cs.cmd

			err := cmd.RunE(cmd, tc.args)

			if !testhelper.EqualError(err, tc.ExpErr) {
				t.Errorf("Want error '%s', got '%s'", tc.ExpErr, err)
			}
		})
	}
}
