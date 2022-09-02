# Developer documentation

Envoy Gateway is built using a [make][make]-based build system. Our CI is based on [Github Actions][gha]
(see: [workflows](.github/workflows)).

## Prerequisites

### go
* Version: 1.18.2
* Installation Guide: https://go.dev/doc/install

### make
* Recommended Version: 4.0 or later
* Installation Guide: https://www.gnu.org/software/make

### docker
* Optional when you want to build a Docker image or run `make` inside Docker.
* Recommended Version: 20.10.16
* Installation Guide: https://docs.docker.com/engine/install

### python3
* Need a `python3` program
* Must have a functioning `venv` module; this is part of the standard
  library, but some distributions (such as Debian and Ubuntu) replace
  it with a stub and require you to install a `python3-venv` package
  separately.

## Quick start
* Run `make help` to see all the available targets to build, test and run `envoy-gateway`.

### Building the `envoy-gateway` binary
* Run `make build` to build the binary that gets generated in the `bin/` directory

### Running tests
* Run `make test` to run the golang tests.

### Running code linters
* Run `make lint` to make sure your code passes all the linter checks.

### Building and Pushing the Image
* Run `IMAGE=docker.io/you/gateway-dev make image` to build the docker image.
* Run `IMAGE=docker.io/you/gateway-dev make push-multiarch` to build and push the multi-arch docker image.

**_NOTE:_**  Replace `IMAGE` with your registry's image name.

### Creating a Kind Cluster to deploy Envoy Gateway
* Run `make create-cluster` to create a Kind cluster and `make kube-install-image` to build a image and load
it into the Kind cluster.

### Deploying Envoy Gateway in Kubernetes
* Run `make kube-deploy` to deploy Envoy Gateway resources as well as the Gateway API
CRDs into a Kubernetes cluster (linked to the current kube context).
* Run `make kube-undeploy` to delete the resources from the cluster created using `kube-deploy`.

**_NOTE:_**  Above command deploys the `envoyproxy/gateway-dev:latest` image into your cluster.
Once https://github.com/envoyproxy/gateway/issues/323 is resolved, you should be able to deploy
your custom image into the cluster.

### Run Gateway API Conformance Tests
* Run `make conformance` to run Gateway API Conformance tests using `envoy-gateway` in a
local Kind cluster. Go [here](https://gateway-api.sigs.k8s.io/concepts/conformance/) to learn
more about the tests.

**_NOTE:_**  This command is currently Work in Progress. :construction::construction::construction::construction:

[make]: https://www.gnu.org/software/make/
[gha]: https://docs.github.com/en/actions
