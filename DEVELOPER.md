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
* Recommended Version: 20.10.16
* Installation Guide: https://docs.docker.com/engine/install/

### linters
* If you already have tools: `golangci-lint, yamllint and codespell` installed on your machine, you can run `make <target>`
directly on your machine.
* If you do not have these tools installed on your machine,
you can alternatively run `MAKE_IN_DOCKER=1 make <target>` to run `make` inside a Docker container which has all the
preinstalled tools needed to support all the `make` targets.
* Installation Guide: [golangci-lint](https://github.com/golangci/golangci-lint#install), [yamllint](https://github.com/adrienverge/yamllint#installation), 
[codespell](https://github.com/codespell-project/codespell#installation)

## Quick start

Run `make help` to see all the available targets to build, test and run `envoy-gateway`.
