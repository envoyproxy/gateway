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

__Note:__ If you do not have these tools installed on your machine, you can alternatively run
`MAKE_IN_DOCKER=1 make <target>` to run `make` inside a Docker container which has all the preinstalled tools needed to
support all the `make` targets.

## Quick start

Run `make help` to see all the available targets to build, test and run `envoy-gateway`.

[make]: https://www.gnu.org/software/make/
[gha]: https://docs.github.com/en/actions
