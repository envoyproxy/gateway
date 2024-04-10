---
title: "Install egctl"
weight: -80
---

{{% alert title="What is egctl?" color="primary" %}}

`egctl` is a command line tool to provide additional functionality for Envoy Gateway users.

{{% /alert %}}


This guide shows how to install the egctl CLI. egctl can be installed either from source, or from pre-built binary releases.

### From The Envoy Gateway Project

The Envoy Gateway project provides two ways to fetch and install egctl. These are the official methods to get egctl releases. Installation through those methods can be found below the official methods.

### From the Binary Releases

Every [release](https://github.com/envoyproxy/gateway/releases) of egctl provides binary releases for a variety of OSes. These binary versions can be manually downloaded and installed.

1. Download your [desired version](https://github.com/envoyproxy/gateway/releases)
2. Unpack it (tar -zxvf egctl_latest_linux_amd64.tar.gz)
3. Find the egctl binary in the unpacked directory, and move it to its desired destination (mv bin/linux/amd64/egctl /usr/local/bin/egctl)

From there, you should be able to run: `egctl help`.

### From Script

`egctl` now has an installer script that will automatically grab the latest release version of egctl and install it locally.

You can fetch that script, and then execute it locally. It's well documented so that you can read through it and understand what it is doing before you run it.

```shell
curl -fsSL -o get-egctl.sh https://gateway.envoyproxy.io/get-egctl.sh

chmod +x get-egctl.sh

# get help info of the 
bash get-egctl.sh --help

# install the latest development version of egctl
bash VERSION=latest get-egctl.sh
```

Yes, you can just use the below command if you want to live on the edge.

```shell
curl https://gateway.envoyproxy.io/get-egctl.sh | VERSION=latest bash 
```

{{% alert title="Next Steps" color="warning" %}}

You can refer to [User Guides](../../user/egctl) to more details about egctl.

{{% /alert %}}
