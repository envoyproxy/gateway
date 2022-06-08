# Make sure versions are same as 
# https://github.com/envoyproxy/gateway/blob/main/.github/workflows/build_and_test.yaml
FROM golang:1.18.2 as builder

# Install Docker CLI
RUN curl -fsSL https://get.docker.com | VERSION=20.10.16 sh 

WORKDIR /workspace
ENTRYPOINT ["make"]
