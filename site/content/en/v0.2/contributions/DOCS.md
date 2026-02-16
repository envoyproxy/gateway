---
title: "Working on Envoy Gateway Docs"
description: "This section tells the development of 
 Envoy Gateway Documents."
---

The documentation for the Envoy Gateway lives in the `docs/` directory. Any
individual document can be written using either [reStructuredText] or [Markdown],
you can choose the format that you're most comfortable with when working on the
documentation.

## Documentation Structure

We supported the versioned Docs now, the directory name under docs represents
the version of docs. The root of the latest site is in `docs/latest/index.rst`.
This is probably where to start if you're trying to understand how things fit together.

Note that the new contents should be added to `docs/latest` and will be cut off at
the next release. The contents under `docs/v0.2.0` are auto-generated,
and usually do not need to make changes to them, unless if you find the current release pages have
some incorrect contents. If so, you should send a PR to update contents both of `docs/latest`
and `docs/v0.2.0`.

It's important to note that a given document _must_ have a reference in some
`.. toctree::` section for the document to be reachable. Not everything needs
to be in `docs/index.rst`'s `toctree` though.

You can access the website which represents the current release in default,
and you can access the website which contains the latest version changes in
[Here][latest-website] or at the footer of the pages.

## Documentation Workflow

To work with the docs, just edit reStructuredText or Markdown files in `docs`,
then run

```bash
make docs
```

This will create `docs/html` with the built HTML pages. You can view the docs
either simply by pointing a web browser at the `file://` path to your
`docs/html`, or by firing up a static webserver from that directory, e.g.

``` shell
make docs-serve
```

If you want to generate a new release version of the docs, like `v0.3.0`, then run

```bash
make docs-release TAG=v0.3.0
```

This will update the VERSION file at the project root, which records current release version,
and it will be used in the pages version context and binary version output. Also, this will generate
new dir `docs/v0.3.0`, which contains docs at v0.3.0 and updates artifact links to `v0.3.0`
in all files under `docs/v0.3.0/user`, like `quickstart.md`, `http-routing.md` and etc.

## Publishing Docs

Whenever docs are pushed to `main`, CI will publish the built docs to GitHub
Pages. For more details, see `.github/workflows/docs.yaml`.

[reStructuredText]: https://docutils.sourceforge.io/docs/ref/rst/restructuredtext.html
[Markdown]: https://daringfireball.net/projects/markdown/syntax
[latest-website]: https://gateway.envoyproxy.io/latest
