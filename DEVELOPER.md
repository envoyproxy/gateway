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
* Run `make create-cluster` to create a [Kind][kind] cluster.
* Run `make kube-install-image` to build an image and load it into the Kind cluster.

**_NOTE:_** Envoy Gateway is tested against Kubernetes v1.24.0.

### Deploying Envoy Gateway in Kubernetes
* Run `IMAGE=envoyproxy/gateway-dev TAG=latest make kube-deploy` to deploy Envoy Gateway resources, including the Gateway API CRDs,
with the `envoyproxy/gateway-dev:latest` Envoy Gateway image into a Kubernetes cluster (linked to the current kube context).
* Run `make kube-undeploy` to delete the resources from the cluster created using `kube-deploy`.

**_NOTE:_**  Replace `IMAGE` with your registry's image name.

### Configure a demo setup
* Run `make kube-demo` to deploy a demo backend service, gatewayclass, gateway and httproute resource
(similar to steps outlined in the [Quickstart](https://github.com/envoyproxy/gateway/blob/main/docs/user/QUICKSTART.md) docs) and test the configuration.
* Run `make kube-demo-undeploy` to delete the resources created by the `make kube-demo` command.

### Run Gateway API Conformance Tests
The commands below build and push the Envoy Gateway image to your hub, deploy Envoy Gateway to the Kubernetes cluster,
and run the Gateway API conformance tests. Refer to the Gateway API [conformance homepage][conform] to learn more about
the tests.

#### On a Linux Host
* Run `make conformance` to run Gateway API conformance tests against Envoy Gateway in a local Kind cluster.

#### On a Mac Host
Since Mac doesn't support [directly exposing][kind_lb] the Docker network to the Mac host, use one of the following
workarounds to run conformance tests:

- Run Envoy Gateway in your own Kubernetes cluster and then run `IMAGE=docker.io/you/gateway-dev make push-multiarch &&
IMAGE=docker.io/you/gateway-dev make kube-deploy && make run-conformance`

- Run Docker Desktop with [Kubernetes support][docker_kube] and then run `IMAGE=docker.io/you/gateway-dev make
push-multiarch && IMAGE=docker.io/you/gateway-dev make kube-install-image && IMAGE=docker.io/you/gateway-dev make
kube-deploy && make run-conformance`
- Install and run [Docker Mac Net Connect][mac_connect] and then run `IMAGE=docker.io/you/gateway-dev make
push-multiarch && IMAGE=docker.io/you/gateway-dev make kube-install-image && IMAGE=docker.io/you/gateway-dev make
kube-deploy && make run-conformance`.

**_NOTE:_**  Replace `IMAGE` with your registry's image name.

### Debugging the Envoy Config
An easy way to view the envoy config that Envoy Gateway is using is to port-forward to the admin interface port (currently `19000`)
on the Envoy deployment that corresponds to a Gateway so that it can be accessed locally.

`kubectl port-forward deploy/envoy-${GATEWAY_NAMESPACE}-${GATEWAY_NAME} -n envoy-gateway-system 19000:19000`

Now you are able to view the running Envoy configuration by navigating to `127.0.0.1:19000/config_dump`.

There are many other endpoints on the [Envoy admin interface](https://www.envoyproxy.io/docs/envoy/v1.23.0/operations/admin#operations-admin-interface) that may be helpful when debugging.

[make]: https://www.gnu.org/software/make/
[gha]: https://docs.github.com/en/actions
[kind]: https://kind.sigs.k8s.io/
[conform]: https://gateway-api.sigs.k8s.io/concepts/conformance/
[kind_lb]: https://kind.sigs.k8s.io/docs/user/loadbalancer/
[docker_kube]: https://docs.docker.com/desktop/kubernetes/
[mac_connect]: https://github.com/chipmk/docker-mac-net-connect
