// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cert

import "runtime"

// CanonicalCertPath is the Debian/Ubuntu CA path used as a canonical value in golden files.
// The envoy-proxy image uses Debian, so this matches the runtime path.
const CanonicalCertPath = "/etc/ssl/certs/ca-certificates.crt"

// SystemCertPath is the default location of the system trust store, initialized at runtime once.
//
// This assumes the Envoy running in a very specific environment. For example, the default location of the system
// trust store on Debian derivatives like the envoy-proxy image being used by the infrastructure controller.
//
// TODO: this might be configurable by an env var or EnvoyGateway configuration.
var SystemCertPath = func() string {
	switch runtime.GOOS {
	case "darwin":
		// TODO: maybe automatically get the keychain cert? That might be macOS version dependent.
		// For now, we'll just use the root cert installed by Homebrew: brew install ca-certificates.
		//
		// See:
		// * https://apple.stackexchange.com/questions/226375/where-are-the-root-cas-stored-on-os-x
		// * https://superuser.com/questions/992167/where-are-digital-certificates-physically-stored-on-a-mac-os-x-machine
		return "/opt/homebrew/etc/ca-certificates/cert.pem"
	default:
		return CanonicalCertPath
	}
}()
