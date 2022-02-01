# This script tests the compatibility of the konf-go shellwrappe with different shells
# Therefore it has no shebang line and is intended to be executed directly by the shell to test

set -o errexit
set -o pipefail


if [[ -n ${ZSH_VERSION} ]]
then
  shell="zsh"
  autoload -U add-zsh-hook # this is required so the shellwrapper can be sourced
elif [[ -n ${BASH_VERSION} ]]
then
  shell="bash"
else
  echo "no valid shell detected"
  exit 1
fi

echo "${shell} detected. Running tests using ${shell} wrapper"

source <(konf-go shellwrapper ${shell})

KONFDIR=$(mktemp -d)
mkdir ${KONFDIR}/store
KONF=${KONFDIR}/store/test_test.yaml
touch ${KONF}
konf --konf-dir=${KONFDIR} set test_test

# if kubeconfig points to something in the active
if [[ $KUBECONFIG != "${KONFDIR}/active"* ]]; then
  echo "Expected KUBECONFIG to point to a file inside '${KONFDIR}/active', but got '${KUBECONFIG}'"
fi

echo "KUBECONFIG points to '${KUBECONFIG}', which looks fine"