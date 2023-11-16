// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// BasicAuth defines the configuration for the HTTP Basic Authentication.
type BasicAuth struct {
	// Username-hashed password pairs used to verify user credentials in the
	// "Authorization" header.
	//
	// The value needs to be the htpasswd format, for example: "user1:{SHA}hashed_user1_password".
	// Right now, only SHA hash algorithm is supported.
	// Reference to https://httpd.apache.org/docs/2.4/programs/htpasswd.html
	//
	// +kubebuilder:validation:MinItems=1
	Users []string `json:"users"`
}
