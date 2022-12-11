# Introduce egctl

## Motivation

EG should provide a command line tool with following capabilities:

- One-stop installation/uninstallation experience
- Collect configuration form data plane
- Change envoy logging configuration on the fly
- Analyse system configuration

This tool is named `egctl`.

## Syntax

Use the following syntax to run `egctl` commands from your terminal window:

```console
egctl [command] [subcommand] [name] [flags]
```

where `command`, `name`, and `flags` are:

* `command`: Specifies the operation that you want to perform on one or more resources,
  for example `install`, `uninstall`, `proxy-config`, `version`.

* `name`: Specifies the name of the specified pod. 

* `flags`: Specifies optional flags. For example, you can use the `-f` or `--file` flags to specify the values for installing.

If you need help, run `kubectl help` from the terminal window.

## Operation

The following table includes short descriptions and the general syntax for all the `egctl` operations:

| Operation      | Syntax                          | Description                                               |
| -------------- |---------------------------------|-----------------------------------------------------------|
| `version`      | `egctl version`                 | Prints out build version information.                     |
| `install`      | `egctl install -f FILENAME`     | Install or reconfigure EG on a cluster.                   |
| `uninstall`    | `egctl uninstall`               | Uninstall EG from a cluster                               |
| `proxy-config` | `egctl proxy-config SUBCOMMAND` | Retrieve information about proxy configuration from envoy |
| `analyze`      | `egctl analyze`                 | Analyze configuration and print validation messages       |

## Examples

Use the following set of examples to help you familiarize yourself with running the commonly used `egctl` operations:

```console
# Install EG using the definition in custom-values.yaml
egctl install -f custom-values.yaml

# Uninstall EG
egctl uninstall

# Retrieve all information about proxy configuration from envoy
egctl proxy-config all <pod_name>

# Retrieve listener information about proxy configuration from envoy
egctl proxy-config listener <pod_name>

# Change log level of envoy
egctl proxy-config log <pod_name> --level trace
```
