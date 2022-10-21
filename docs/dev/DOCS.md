# Working on the Envoy Gateway Docs

The documentation for the Envoy Gateway lives in the `docs/` directory. Any
individual document can be written using either [reStructuredText] or [Markdown], 
you can choose the format that you're most comfortable with when working on the
documentation.

## Documentation Structure

The root of the site is in `docs/index.rst`. This is probably where to start
if you're trying to understand how things fit together. 

It's important to note that a given document _must_ have a reference in some
`.. toctree::` section for the document to be reachable. Not everything needs
to be in `docs/index.rst`'s `toctree` though.

## Documentation Workflow

To work with the docs, just edit reStructuredText or Markdown files in `docs`.
Before you compile the docs, you need to set the `ENVOY_VERSION` and 
`GATEWAYAPI_VERSION` environment variables. For example:

```bash
export ENVOY_VERSION=1.23
export GATEWAYAPI_VERSION=0.5.1
```

Then run

```bash
make docs
```

This will create `docs/html` with the built HTML pages. You can view the docs
either simply by pointing a web browser at the `file://` path to your
`docs/html`, or by firing up a static webserver from that directory, e.g.

```
cd docs/html ; python3 -m http.server
```

## Publishing Docs

Whenever docs are pushed to `main`, CI will publish the built docs to GitHub
Pages. For more details, see `.github/workflows/docs.yaml`.

[reStructuredText]: https://docutils.sourceforge.io/docs/ref/rst/restructuredtext.html
[Markdown]: https://daringfireball.net/projects/markdown/syntax

