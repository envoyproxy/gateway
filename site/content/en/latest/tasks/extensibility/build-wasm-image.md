---
title: "Build a Wasm image"
---

Envoy Gateway supports two types of Wasm extensions within the [EnvoyExtensionPolicy][] API: HTTP Wasm Extensions and Image Wasm Extensions. 
Packaging a Wasm extension as an OCI image is beneficial because it simplifies versioning and distribution for users. 
Additionally, users can leverage existing image toolchain to build and manage Wasm images.

This document describes how to build OCI images which are consumable by Envoy Gateway.  

## Wasm Image Formats

There are two types of images that are supported by Envoy Gateway. One is in the Docker format, and another is the standard 
OCI specification compliant format. Please note that both of them are supported by any OCI registries. You can choose 
either format depending on your preference, and both types of images are consumable by Envoy Gateway [EnvoyExtensionPolicy][] API.

## Build Wasm Docker image

We assume that you have a valid Wasm binary named `plugin.wasm`. Then you can build a Wasm Docker image with the Docker CLI.

1. First, we prepare the following Dockerfile:

```
$ cat Dockerfile
FROM scratch

COPY plugin.wasm ./
```

**Note: you must have exactly one `COPY` instruction in the Dockerfile in order to end up having only one layer in produced images.**

2. Then, build your image via `docker build` command

```
$ docker build . -t my-registry/mywasm:0.1.0
```

3. Finally, push the image to your registry via `docker push` command

```
$ docker push my-registry/mywasm:0.1.0
```

## Build Wasm OCI image

We assume that you have a valid Wasm binary named `plugin.wasm`, and you have [buildah](https://buildah.io/) installed on your machine. 
Then you can build a Wasm OCI image with the `buildah` CLI.

1. First, we create a working container from `scratch` base image with `buildah from` command.

```
$ buildah --name mywasm from scratch
mywasm
```

2. Then copy the Wasm binary into that base image by `buildah copy` command to create the layer.

```
$ buildah copy mywasm plugin.wasm ./
af82a227630327c24026d7c6d3057c3d5478b14426b74c547df011ca5f23d271
```

**Note: you must execute `buildah copy` exactly once in order to end up having only one layer in produced images**

4. Now, you can build an OCI image and push it to your registry via `buildah commit` command

```
$ buildah commit mywasm docker://my-remote-registry/mywasm:0.1.0
```

[EnvoyExtensionPolicy]: ../../../api/extension_types#envoyextensionpolicy
