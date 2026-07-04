Fixed `HTTPRoute` URL rewrites with `PathPrefix "/"` so `replacePrefixMatch` can prepend a new prefix without dropping the separator between the new prefix and the original request path.
