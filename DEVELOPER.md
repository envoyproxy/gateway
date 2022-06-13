# Developer documentation

Envoy Gateway is built using a [make](https://www.gnu.org/software/make/)-based
build system. Our CI is based on [Github Actions](https://docs.github.com/en/actions) (see: [workflows](.github/workflows))

## Prerequisites

### go
* Version: 1.18.2
* Installation Guide: https://go.dev/doc/install to

### make
* Recommended Version: 4.3
* Installation Guide: https://www.gnu.org/software/make/).

### docker
* Optional when you want to build a Docker image or run make inside Docker.
* Recommened Version: 20.10.16
* Installation Guide: https://docs.docker.com/engine/install/

### linters
* [TODO](https://github.com/envoyproxy/gateway/issues/73)

* If you do not have these tools installed on your machine,
you can alternatively run `MAKE_IN_DOCKER=1 make <target>` to run `make` inside a Docker container which has all the
preinstalled tools needed to support all the `make` targets.

## Quick start

Run `make help` to see all the available targets to build, test and run `envoy-gateway`.
