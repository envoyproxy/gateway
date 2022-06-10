# Make sure the tooling versions are same as defined in the 
# CI https://github.com/envoyproxy/gateway/blob/main/.github/workflows/build_and_test.yaml
# as well as in the Developer docs - https://github.com/envoyproxy/gateway/blob/main/DEVELOPER.md#prerequisites

# go
FROM golang:1.18.2 as builder
# docker CLI
RUN curl -fsSL https://get.docker.com | VERSION=20.10.16 sh 

WORKDIR /workspace
ENTRYPOINT ["make"]
