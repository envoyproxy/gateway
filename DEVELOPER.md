# Developer documentation

Envoy Gateway is built using a [make](https://www.gnu.org/software/make/)-based
build system. Our CI is based on [Github Actions](https://docs.github.com/en/actions) (see: [workflows](.github/workflows))

## Prerequisites

* Go, currently we use 1.18.x. You can refer to https://go.dev/doc/install to
download and install Go on your system.
* [make](https://www.gnu.org/software/make/).
* [Docker](https://docs.docker.com/engine/install/), this is optional when you want to build the image.
* The project also uses the below tools for linting. If you do not have these tools installed on your machine,
you can alternatively run `MAKE_IN_DOCKER=1 make <target>` to run `make` inside a Docker container which has all the
preinstalled tools needed to support all the `make` targets.
  * [TODO](https://github.com/envoyproxy/gateway/issues/73)


## Quick start

[TODO](https://github.com/envoyproxy/gateway/issues/101) Run `make help` to see all the available targets to build, test
and run `envoy-gateway`.
