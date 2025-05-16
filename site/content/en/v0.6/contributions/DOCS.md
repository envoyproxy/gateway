---
title: Developer Guide
weight: 10
---

## Prerequisites

### Core Tools
- **Go** (v1.24 or later)  
  📌 [Installation Guide](https://go.dev/doc/install)  
- **Make** (v4.0 or later)  
  📌 [Installation Guide](https://www.gnu.org/software/make)  
- **Python 3** (with `venv` module)  
  ```bash
  # For Debian/Ubuntu users:
  sudo apt-get install python3-venv

Quick Start
make help  # View all available commands

Building Envoy Gateway
# Build all binaries:
make build

# Build specific components:
make build BINS="envoy-gateway"  # Control plane
make build BINS="egctl"         # CLI tool


 Output Location: bin/<OS>/<ARCH>/ (e.g., bin/linux/amd64/)


Testing

# Run unit tests:
make test

# End-to-end tests:
make e2e

# Generate test data:
make testdata

Kubernetes Development
Local Cluster (Kind)

# Create a test cluster:
make create-cluster

# Deploy with latest image:
TAG=latest make kube-deploy

# Deploy custom image:
make kube-install-image
IMAGE_PULL_POLICY=IfNotPresent make kube-deploy




Demo Setup

# Deploy demo resources:
make kube-demo

# Clean up demo:
make kube-demo-undeploy


Platform-Specific Notes
MacOS Users
For conformance tests:

Preferred: Use Docker Desktop with Kubernetes, or

Alternative:

brew install chipmk/tap/docker-mac-net-connect
TAG=latest make conformance



