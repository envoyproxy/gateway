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
egctl [command] [entry] [name] [flags]
```

where `command`, `name`, and `flags` are:

* `command`: Specifies the operation that you want to perform on one or more resources,
  for example `install`, `uninstall`, `config`, `version`.

* `entry`: Specifies the entity the operation is being performed on such as `envoy-proxy` or `envoy-gateway`.

* `name`: Specifies the name of the specified instance. 

* `flags`: Specifies optional flags. For example, you can use the `-c` or `--config` flags to specify the values for installing.

If you need help, run `kubectl help` from the terminal window.

## Operation

The following table includes short descriptions and the general syntax for all the `egctl` operations:

| Operation   | Syntax                           | Description                                                                 |
| ----------- | -------------------------------- | --------------------------------------------------------------------------- |
| `version`   | `egctl version`                  | Prints out build version information.                                       |
| `install`   | `egctl install -c CUSTOMPROFILE` | Install or reconfigure EG on a cluster.                                     |
| `uninstall` | `egctl uninstall`                | Uninstall EG from a cluster                                                 |
| `config`    | `egctl config ENTRY`             | Retrieve information about proxy configuration from envoy proxy and gateway |
| `log`       | `egctl log ENTRY --level trace`  | Change envoy proxy's log level                                              |
| `analyze`   | `egctl analyze`                  | Analyze configuration and print validation messages                         |

## Examples

Use the following set of examples to help you familiarize yourself with running the commonly used `egctl` operations:

```console
# Install EG using the definition in custom profile file
egctl install envoy-gateway -c custom-profile.yaml

# Install the managed Envoy Proxy fleet
egctl install envoy-proxy

# Uninstall EG
egctl uninstall

# Retrieve all information about proxy configuration from envoy
egctl config envoy-proxy all <instance_name>

# Retrieve listener information about proxy configuration from envoy 
egctl config envoy-proxy listener <instance_name>

# Retrieve information about envoy gateway
egctl config envoy-gateway

# Change log level of envoy proxy
egctl log envoy-proxy <instance_name> --level trace
```
