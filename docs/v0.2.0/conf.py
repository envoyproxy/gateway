import os
import re

# Add any Sphinx extension module names here, as strings. They can be
# extensions coming with Sphinx (named 'sphinx.ext.*') or your custom
# ones.
extensions = [
    'sphinx.ext.duration',
    'sphinx.ext.autosectionlabel',
    'myst_parser',
]

autosectionlabel_prefix_document = True
myst_heading_anchors = 3

html_theme = 'alabaster'

# The master toctree document.
master_doc = 'index'

# General information about the project.
version = os.environ["BUILD_VERSION"]
envoyVersion = os.environ["ENVOY_PROXY_VERSION"]
gatewayAPIVersion = os.environ["GATEWAYAPI_VERSION"]

project = 'Envoy Gateway'
author = 'Envoy Gateway Project Authors'

copyright = 'Envoy Gateway Project Authors | <a href="https://github.com/envoyproxy/gateway">GitHub</a> | <a href="/latest">Latest Docs</a>'

source_suffix = {
    '.rst': 'restructuredtext',
    '.md': 'markdown',
}

variables_to_export = [
    "envoyVersion",
    "gatewayAPIVersion",
]

frozen_locals = dict(locals())
rst_epilog = '\n'.join(map(lambda x: f".. |{x}| replace:: {frozen_locals[x]}", variables_to_export))
del frozen_locals
