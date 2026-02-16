// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// PathEscapedSlashAction determines the action for requests that contain %2F, %2f, %5C, or %5c
// sequences in the URI path.
// +kubebuilder:validation:Enum=KeepUnchanged;RejectRequest;UnescapeAndForward;UnescapeAndRedirect
type PathEscapedSlashAction string

const (
	// KeepUnchangedAction keeps escaped slashes as they arrive without changes
	KeepUnchangedAction PathEscapedSlashAction = "KeepUnchanged"
	// RejectRequestAction rejects client requests containing escaped slashes
	// with a 400 status. gRPC requests will be rejected with the INTERNAL (13)
	// error code.
	// The "httpN.downstream_rq_failed_path_normalization" counter is incremented
	// for each rejected request.
	RejectRequestAction PathEscapedSlashAction = "RejectRequest"
	// UnescapeAndRedirect unescapes %2F and %5C sequences and redirects to the new path
	// if these sequences were present.
	//
	// Redirect occurs after path normalization and merge slashes transformations if
	// they were configured. gRPC requests will be rejected with the INTERNAL (13)
	// error code.
	// This option minimizes possibility of path confusion exploits by forcing request
	// with unescaped slashes to traverse all parties: downstream client, intermediate
	// proxies, Envoy and upstream server.
	// The “httpN.downstream_rq_redirected_with_normalized_path” counter is incremented
	// for each redirected request.
	UnescapeAndRedirect PathEscapedSlashAction = "UnescapeAndRedirect"
	// UnescapeAndForward unescapes %2F and %5C sequences and forwards the request.
	// Note: this option should not be enabled if intermediaries perform path based access
	// control as it may lead to path confusion vulnerabilities.
	UnescapeAndForward PathEscapedSlashAction = "UnescapeAndForward"
)

// PathSettings provides settings that managing how the incoming path set by clients is handled.
type PathSettings struct {
	// EscapedSlashesAction determines how %2f, %2F, %5c, or %5C sequences in the path URI
	// should be handled.
	// The default is UnescapeAndRedirect.
	//
	// +optional
	EscapedSlashesAction *PathEscapedSlashAction `json:"escapedSlashesAction,omitempty"`
	// DisableMergeSlashes allows disabling the default configuration of merging adjacent
	// slashes in the path.
	// Note that slash merging is not part of the HTTP spec and is provided for convenience.
	//
	// +optional
	DisableMergeSlashes *bool `json:"disableMergeSlashes,omitempty"`
}
