# Make sure the tooling versions are same as defined in the 
# CI https://github.com/envoyproxy/gateway/blob/main/.github/workflows/build_and_test.yaml
# as well as in the Developer docs - https://github.com/envoyproxy/gateway/blob/main/DEVELOPER.md#prerequisites

# go
FROM golang:1.18.2 as builder

ENV YAMLLINT_VERSION=1.24.2
ENV GOLINT_VERSION=v1.46.2
ENV CODESPELL_VERSION=v2.1.0

# docker CLI
RUN curl -fsSL https://get.docker.com | VERSION=20.10.16 sh

# python
RUN apt-get update && apt-get install -y --no-install-recommends python3-pip

# golangci Lint
Run curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s ${GOLINT_VERSION}

# pip install
RUN python3 -m pip install --no-cache-dir yamllint==${YAMLLINT_VERSION}
RUN python3 -m pip install --no-cache-dir codespell==${CODESPELL_VERSION}

WORKDIR /workspace
ENTRYPOINT ["make"]