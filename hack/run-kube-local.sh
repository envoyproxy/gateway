set -euo pipefail

GOARCH=$(go env GOARCH)
GOOS=$(go env GOOS)
KUBECONFIG="${KUBECONFIG:-$HOME/.kube/config}"

# Check that kubectl is installed.
if ! [ "$(which kubectl)" ] ; then
    echo "kubectl not installed"
    exit 1
fi

# Check the kubectl config file.
if ! [ "$(stat ${KUBECONFIG})" ] ; then
    echo "kubeconfig not set"
    exit 1
fi

# Run the envoy gateway binary
./bin/${GOOS}/${GOARCH}/envoy-gateway server "$@"
