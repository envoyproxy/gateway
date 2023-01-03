# Introduce egctl

## Motivation

EG should provide a command line tool with following capabilities:

- One-step installation/uninstallation experience
- Collect configuration from envoy proxy and gateway
- Update admin options such as changing log level
- Analyse system configuration to diagnose any issues in envoy gateway

This tool is named `egctl`.

## Syntax

Use the following syntax to run `egctl` commands from your terminal window:

```console
egctl [command] [entity] [name] [flags]
```

where `command`, `name`, and `flags` are:

* `command`: Specifies the operation that you want to perform on one or more resources,
  for example `install`, `uninstall`, `config`, `version`.

* `entity`: Specifies the entity the operation is being performed on such as `envoy-proxy` or `envoy-gateway`.

* `name`: Specifies the name of the specified instance. 

* `flags`: Specifies optional flags. For example, you can use the `-c` or `--config` flags to specify the values for installing.

If you need help, run `egctl help` from the terminal window.

## Operation

The following table includes short descriptions and the general syntax for all the `egctl` operations:

| Operation   | Syntax                                 | Description                                                                 |
| ----------- | -------------------------------------- | --------------------------------------------------------------------------- |
| `version`   | `egctl version`                        | Prints out build version information.                                       |
| `install`   | `egctl install ENTITY -c EGCONFIGFILE` | Install or reconfigure EG on a cluster.                                     |
| `uninstall` | `egctl uninstall`                      | Uninstall EG from a cluster                                                 |
| `get`       | `egctl get ENTITY`                     | Retrieve information about proxy configuration from envoy proxy and gateway |
| `log`       | `egctl log ENTITY --level trace`       | Change envoy proxy's log level                                              |
| `analyze`   | `egctl analyze`                        | Analyze EG configuration and print validation messages                      |

## Examples

Use the following set of examples to help you familiarize yourself with running the commonly used `egctl` operations:

```console
# Install EG using the definition in EG config file
egctl install envoy-gateway -c egconfig.yaml

# Install the managed Envoy Proxy fleet
egctl install envoy-proxy

# Uninstall EG
egctl uninstall

# Retrieve all information about proxy configuration from envoy
egctl get envoy-proxy all <instance_name>

# Retrieve listener information about proxy configuration from envoy 
egctl get envoy-proxy listener <instance_name>

# Retrieve information about envoy gateway
egctl get envoy-gateway

# Change log level of envoy proxy
egctl log envoy-proxy <instance_name> --level trace
```
