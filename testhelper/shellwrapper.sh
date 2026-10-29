# This script tests the compatibility of the konf-go shellwrapper with different shells
# Therefore it has no shebang line and is intended to be executed directly by the shell to test

set -eu
set -o pipefail

if [[ -n ${ZSH_VERSION:-} ]]
then
  shell="zsh"
  autoload -U add-zsh-hook # this is required so the shellwrapper can be sourced
elif [[ -n ${BASH_VERSION:-} ]]
then
  shell="bash"
else
  echo "no valid shell detected"
  exit 1
fi

echo "${shell} detected. Running tests using ${shell} wrapper"

root=$(mktemp -d "${TMPDIR:-/tmp}/konf-shellwrapper.XXXXXX")
trap 'rm -rf "$root"' EXIT
export HOME="$root/home"
KONFDIR="$root/konf dir"
mkdir -p "$HOME" "$KONFDIR/store"
konf-go shellwrapper "$shell" > "$root/wrapper"
source "$root/wrapper"

konf_cleanup() {
  local result=$?
  KONF_SHELL_PID=$$ command konf-go --konf-dir="$KONFDIR" --silent cleanup || result=$?
  rm -rf "$root"
  return "$result"
}

cat > "$KONFDIR/store/test_test.yaml" <<'EOF'
apiVersion: v1
kind: Config
clusters:
- name: test
  cluster:
    server: https://example.invalid
contexts:
- name: test
  context:
    cluster: test
    user: test
    namespace: original
current-context: test
users:
- name: test
  user: {}
EOF

konf --konf-dir="$KONFDIR" --silent set test_test
expected="$KONFDIR/active/$$.yaml"
if [[ "$KUBECONFIG" != "$expected" ]] || [[ ! -f "$KUBECONFIG" ]]; then
  echo "Expected KUBECONFIG '$expected', got '$KUBECONFIG'" >&2
  exit 1
fi

konf --konf-dir="$KONFDIR" --silent ns testing
grep -q 'namespace: testing' "$expected"
grep -q 'namespace: original' "$KONFDIR/store/test_test.yaml"

konf --konf-dir="$KONFDIR" --silent set -
[[ "$KUBECONFIG" = "$expected" ]]
grep -q 'namespace: original' "$expected"

if konf --konf-dir="$KONFDIR" --silent set missing > "$root/output" 2> "$root/error"; then
  echo "Expected missing selection to fail" >&2
  exit 1
fi
[[ "$KUBECONFIG" = "$expected" ]]
konf completion "$shell" > "$root/completion"
[[ -s "$root/completion" ]]

echo "PASS: $shell wrapper arguments, shell PID, restore, namespace isolation, failures and completion"