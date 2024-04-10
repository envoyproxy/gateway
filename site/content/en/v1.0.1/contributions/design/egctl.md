---
title: "egctl Design"
---

## Motivation

EG should provide a command line tool with following capabilities:

- Collect configuration from envoy proxy and gateway
- Analyse system configuration to diagnose any issues in envoy gateway

This tool is named `egctl`.

## Syntax

Use the following syntax to run `egctl` commands from your terminal window:

```console
egctl [command] [entity] [name] [flags]
```

where `command`, `name`, and `flags` are:

* `command`: Specifies the operation that you want to perform on one or more resources,
  for example `config`, `version`.

* `entity`: Specifies the entity the operation is being performed on such as `envoy-proxy` or `envoy-gateway`.

* `name`: Specifies the name of the specified instance. 

* `flags`: Specifies optional flags. For example, you can use the `-c` or `--config` flags to specify the values for installing.

If you need help, run `egctl help` from the terminal window.

## Operation

The following table includes short descriptions and the general syntax for all the `egctl` operations:

| Operation     | Syntax                           | Description                                                                          |
| --------------| -------------------------------- | -------------------------------------------------------------------------------------|
| `version`     | `egctl version`                  | Prints out build version information.                                                |
| `config`      | `egctl config ENTITY`            | Retrieve information about proxy configuration from envoy proxy and gateway          |
| `analyze`     | `egctl analyze`                  | Analyze EG configuration and print validation messages                               |
| `experimental`| `egctl experimental`             | Subcommand for experimental features. These do not guarantee backwards compatibility |

## Examples

Use the following set of examples to help you familiarize yourself with running the commonly used `egctl` operations:

```console
# Retrieve all information about proxy configuration from envoy
egctl config envoy-proxy all <instance_name>

# Retrieve listener information about proxy configuration from envoy 
egctl config envoy-proxy listener <instance_name>

# Retrieve the relevant rate limit configuration from the Rate Limit instance
egctl config envoy-ratelimit
```
