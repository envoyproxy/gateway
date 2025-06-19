# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Envoy Gateway is an open source project for managing Envoy Proxy as a standalone or Kubernetes-based application gateway. It provides an expressive, extensible API (based on Gateway API) that makes Envoy accessible to application developers without requiring deep knowledge of Envoy's complex xDS APIs.

## Essential Development Commands

### Build and Development
- `make build` - Build envoy-gateway for host platform
- `make build-multiarch` - Build envoy-gateway for multiple platforms
- `make test` - Run all Go unit tests
- `make generate` - Generate go code from templates and tags (includes kube-generate, docs-api, helm-generate)
- `make clean` - Remove all files created during builds

### Code Quality and Linting
- `make lint` - Run all linters (golint, yamllint, whitenoise lint, codespell, shellcheck)
- `make fix-golint` - Run golangci-lint with auto-fix to resolve code issues
- `make format` - Update and check dependencies with go mod tidy
- `make gen-check` - Check if generated files are up to date

### Testing
- `make test` - Run Go unit tests
- `make e2e` - Run end-to-end tests (creates cluster, deploys EG, runs tests, cleans up)
- `make conformance` - Run Gateway API conformance tests
- `make benchmark` - Run benchmark tests
- `go test -v ./internal/gatewayapi/` - Run tests for specific package
- `go test -run TestSpecificFunction ./path/to/package` - Run specific test function
- `make testdata` - Update test golden files after making changes

### Kubernetes Development
- `make kube-deploy` - Install Envoy Gateway into Kubernetes cluster
- `make kube-undeploy` - Uninstall Envoy Gateway from Kubernetes cluster
- `make create-cluster` - Create a kind cluster for testing
- `make delete-cluster` - Delete kind cluster
- `make manifests` - Generate Kubernetes manifests and CRDs

### Docker Images
- `make image` - Build docker images for host platform
- `make push` - Push docker images to registry

## High-Level Architecture

### Core Components

1. **API Layer** (`/api/v1alpha1/`)
   - Defines CRDs and types for Envoy Gateway
   - Key resources: EnvoyGateway, EnvoyProxy, BackendTrafficPolicy, ClientTrafficPolicy, SecurityPolicy
   - Policy resources for authentication (OIDC, JWT, BasicAuth), traffic management, extensions

2. **Gateway API Translation** (`/internal/gatewayapi/`)
   - Translates Gateway API resources (Gateway, HTTPRoute, etc.) into internal IR
   - Handles policy attachment and merging logic
   - Manages resource validation and status reporting

3. **XDS Translation** (`/internal/xds/`)
   - Converts internal IR into Envoy xDS configuration
   - Components: Bootstrap, Translator, Cache, Server
   - Serves xDS configuration to Envoy proxies

4. **Infrastructure Management** (`/internal/infrastructure/`)
   - Manages deployment and lifecycle of Envoy proxy infrastructure
   - Kubernetes provider: deploys as Deployments/DaemonSets with Services, ConfigMaps

5. **Intermediate Representation** (`/internal/ir/`)
   - XDS IR: Configuration for data plane (listeners, routes, clusters)
   - Infra IR: Infrastructure requirements (proxy deployment specs)

### Key Binaries
- `cmd/envoy-gateway/` - Main control plane binary
- `cmd/egctl/` - CLI tool for debugging and interaction

### Flow
1. Gateway API resources and policies created in Kubernetes
2. Gateway API translator processes resources → Internal IR
3. XDS translator converts IR → Envoy xDS configuration
4. Infrastructure manager ensures proxy deployment
5. XDS server delivers configuration to Envoy proxies

## Testing Architecture

### Test Organization
- `/test/` - Main test directory with subdirectories for different test types
- Tests co-located with source code in `*_test.go` files
- Table-driven tests with input/output YAML files in `testdata/` directories

### Test Types and Commands
- **Unit tests**: `make test` or `go test ./...`
- **Integration tests**: `make kube-test` 
- **E2E tests**: `make e2e` (full setup) or `make run-e2e` (existing cluster)
- **Conformance tests**: `make conformance` or `make run-conformance`
- **CEL validation**: `make go.test.cel`
- **Benchmark tests**: `make benchmark`
- **Resilience tests**: `make resilience`

### Test-Specific Commands
- `make testdata` - Update golden files after changes
- `E2E_RUN_TEST="TestName" make run-e2e` - Run specific e2e test
- `CONFORMANCE_RUN_TEST="TestName" make run-conformance` - Run specific conformance test

## Development Patterns

### Code Generation
- Run `make generate` after modifying API types or adding new CRDs
- This regenerates deepcopy methods, manifests, and documentation
- Always run `make gen-check` to verify generated files are up to date

### Testing Workflow
1. Write unit tests alongside code changes
2. Run `make test` for quick feedback
3. Run `make testdata` if test outputs change
4. Run `make e2e` for comprehensive validation before PR

### Debugging and Development
- Use `egctl` CLI for debugging deployed configurations
- Access logs and metrics through Kubernetes standard tools
- Use `make kube-demo` to deploy test scenarios

## Important File Locations

- `/Makefile` - Main build targets (wrapper around tools/make/*.mk)
- `/tools/make/` - All make target implementations
- `/api/v1alpha1/` - API definitions and CRDs
- `/internal/gatewayapi/testdata/` - Test input/output files for Gateway API translation
- `/internal/xds/testdata/` - Test input/output files for XDS translation
- `/examples/` - Example configurations and sample applications

## Project-Specific Notes

- The codebase uses a multi-layer translation approach: Gateway API → IR → xDS
- Policy attachment follows Gateway API standards with strategic merge semantics
- Extension points include WASM, Lua, and external processing filters
- The project supports both Kubernetes and standalone deployments
- All APIs are currently v1alpha1 and subject to change