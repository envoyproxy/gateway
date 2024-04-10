// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

const BasicAuthUsersSecretKey = ".htpasswd"

// BasicAuth defines the configuration for 	the HTTP Basic Authentication.
type BasicAuth struct {
	// The Kubernetes secret which contains the username-password pairs in
	// htpasswd format, used to verify user credentials in the "Authorization"
	// header.
	//
	// This is an Opaque secret. The username-password pairs should be stored in
	// the key ".htpasswd". As the key name indicates, the value needs to be the
	// htpasswd format, for example: "user1:{SHA}hashed_user1_password".
	// Right now, only SHA hash algorithm is supported.
	// Reference to https://httpd.apache.org/docs/2.4/programs/htpasswd.html
	// for more details.
	//
	// Note: The secret must be in the same namespace as the SecurityPolicy.
	Users gwapiv1b1.SecretObjectReference `json:"users"`
}
