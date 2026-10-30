# Wasm test fixtures

`main.go` implements a Proxy Wasm extension that reads a backend alias
from route metadata and makes an HTTP callout. The extension backend E2E test
uses it to verify that shared filters resolve different bindings on different routes.

The static file server image build compiles the module and serves it at
`/wasm/backend_callout.wasm`.

To build locally from the repository root:

```shell
make -C examples/wasm-module-test backend-callout
```
