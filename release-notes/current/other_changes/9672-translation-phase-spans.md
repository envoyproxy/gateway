Added per-phase tracing spans to the Gateway API and xDS translators, each recording
the size of the input it processed, so that a slow translation can be attributed to a
specific phase — listener processing, HTTP and gRPC route processing, the main policy
types, EnvoyPatchPolicy JSON patches, extension server hooks, and xDS resource
validation — instead of showing up as one opaque multi-second span.
