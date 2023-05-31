// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Import all Envoy filter types so they are registered and deserialization does not fail
// when using them in the "typed_config" attributes.
package extensions

// nolint: lll
//
//go:generate sh -c "echo '// Copyright Envoy Gateway Authors' > extensions.gen.go"
//go:generate sh -c "echo '// SPDX-License-Identifier: Apache-2.0' >> extensions.gen.go"
//go:generate sh -c "echo '// The full text of the Apache license is available in the LICENSE file at' >> extensions.gen.go"
//go:generate sh -c "echo '// the root of the repo.\n' >> extensions.gen.go"
//go:generate sh -c "echo '// GENERATED FILE -- DO NOT EDIT\n' >> extensions.gen.go"
//go:generate sh -c "echo 'package extensions\n\nimport (' >> extensions.gen.go"
//go:generate sh -c "go list github.com/envoyproxy/go-control-plane/... | grep 'v[3-9]' | grep -v /pkg/ | xargs -I{} echo '\t_ \"{}\"' >> extensions.gen.go"
//go:generate sh -c "echo ')' >> extensions.gen.go"
