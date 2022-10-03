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
* Run `IMAGE=envoyproxy/gateway-dev TAG=latest make kube-deploy` to deploy Envoy Gateway resources, including the Gateway API CRDs,
with the `envoyproxy/gateway-dev:latest` Envoy Gateway image into a Kubernetes cluster (linked to the current kube context).
* Run `make kube-undeploy` to delete the resources from the cluster created using `kube-deploy`.

**_NOTE:_**  Replace `IMAGE` with your registry's image name.

### Run Gateway API Conformance Tests
* Run `make conformance` to run Gateway API Conformance tests using `envoy-gateway` in a
local Kind cluster. Go [here](https://gateway-api.sigs.k8s.io/concepts/conformance/) to learn
more about the tests.

**_NOTE:_** Conformance tests against a kind cluster is currently unsupported on Mac computers.
As a workaround, you could run this against your own Kubernetes cluster (such as Kubernetes on Docker Desktop) using this command -
`IMAGE=docker.io/you/gateway-dev make push-multiarch && IMAGE=docker.io/you/gateway-dev make kube-deploy && make run-conformance`
which builds and pushes the Envoy-Gateway image to your hub, deploys Envoy Gateway resources into your cluster
and runs the Gateway API conformance tests.

### Debugging the Envoy Config
An easy way to view the envoy config that Envoy Gateway is using is to port-forward to the admin interface port (currently `19000`)
on the Envoy Gateway deployment so that it can be accessed locally.

`kubectl port-forward deploy/envoy-default-eg -n envoy-gateway-system 19000:19000`

Now you are able to view the running Envoy configuration by navigating to `127.0.0.1:19000/config_dump`.

There are many other endpoints on the [Envoy admin interface](https://www.envoyproxy.io/docs/envoy/v1.23.0/operations/admin#operations-admin-interface) that may be helpful when debugging.

[make]: https://www.gnu.org/software/make/
[gha]: https://docs.github.com/en/actions
