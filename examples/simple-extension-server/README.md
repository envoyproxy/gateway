# Simple Extension Server

This example is a simple extension server for e2e testing purposes. For a detailed example of an extension server that leverages custom resources and extension hook context, see the `extension-server` directory.

The extension server modifies virtual hosts. If the vhost name contains the string "fail", the extension server returns an error. Otherwise, it adds another domain to the vhost domain list.

This server is used in the following tests:
- test/resilience/tests/extensionserver.go

