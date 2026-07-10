// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"errors"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	netutil "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/transport"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1a3 "sigs.k8s.io/gateway-api/apis/v1alpha3"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func NewClient(t *testing.T) (client.Client, *rest.Config) {
	cfg, err := config.GetConfig()
	require.NoError(t, err)

	// Install a transport wrapper that retries transient network errors
	// (e.g. EOF, connection reset, HTTP/2 GOAWAY, timeout) for idempotent
	// requests whose body can be replayed. This is necessary because the
	// gateway-api conformance suite rebuilds its client per test via
	// client.New(suite.RestConfig, ...) (see setClientsetForTest), so a
	// controller-runtime client wrapper returned from this function would be
	// bypassed. The transport wrapper is the only hook inherited by those
	// rebuilt clients, and it also covers the raw client returned here.
	//
	// Notably, client-go only retries GET on transient errors; DELETE and other
	// write verbs surface the error immediately (e.g. "Delete ... : EOF"),
	// which is the source of teardown flakes. Retrying replayable DELETEs here
	// closes that gap: if the first DELETE reached the apiserver but the
	// response was lost, the retry may return 404, which conformance cleanup
	// already treats as success.
	cfg.Wrap(transport.Wrappers(cfg.WrapTransport, newRetryTransportWrapper))

	c, err := client.New(cfg, client.Options{})
	require.NoError(t, err)

	// Install all the scheme to kubernetes client.
	CheckInstallScheme(t, c)

	return c, cfg
}

func CheckInstallScheme(t *testing.T, c client.Client) {
	require.NoError(t, gwapiv1a3.Install(c.Scheme()))
	require.NoError(t, gwapiv1a2.Install(c.Scheme()))
	require.NoError(t, gwapiv1b1.Install(c.Scheme()))
	require.NoError(t, gwapiv1.Install(c.Scheme()))
	require.NoError(t, egv1a1.AddToScheme(c.Scheme()))
	require.NoError(t, batchv1.AddToScheme(c.Scheme()))
}

// retryTransportMaxAttempts bounds the total attempts (1 initial + retries).
const retryTransportMaxAttempts = 4

// retryTransportBaseBackoff is the initial backoff between retries; it grows
// by retryTransportBackoffFactor on each attempt.
const (
	retryTransportBaseBackoff   = 100 * time.Millisecond
	retryTransportBackoffFactor = 2
)

// newRetryTransportWrapper returns a transport.WrapperFunc that layers a
// retrying RoundTripper over the base transport.
func newRetryTransportWrapper(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return &retryTransport{base: base}
}

// retryTransport retries transient, connection-level errors for idempotent
// requests whose body can be safely replayed.
type retryTransport struct {
	base http.RoundTripper
}

func (r *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Only retry idempotent verbs with a replayable body. POST/PATCH/PUT are
	// never retried; a request with a body and no GetBody cannot be replayed
	// safely after the body has been (partially) consumed.
	if !retryableMethod(req.Method) || !replayableBody(req) {
		return r.base.RoundTrip(req)
	}

	var (
		lastErr    error
		backoff    = retryTransportBaseBackoff
		attemptReq = req
	)

	for attempt := 0; attempt < retryTransportMaxAttempts; attempt++ {
		// On every attempt after the first, back off before retrying. The sleep
		// is done BEFORE creating a fresh body so that a context cancellation
		// observed here does not leak a body ReadCloser obtained from GetBody.
		if attempt > 0 {
			select {
			case <-req.Context().Done():
				return nil, req.Context().Err()
			case <-time.After(backoff):
			}
			backoff *= retryTransportBackoffFactor

			// Reset the body from GetBody so the base transport reads a fresh
			// copy; the original req.Body may have been consumed by the previous
			// attempt.
			if req.GetBody != nil {
				body, err := req.GetBody()
				if err != nil {
					return nil, err
				}
				attemptReq = req.Clone(req.Context())
				attemptReq.Body = body
			}
		}

		resp, err := r.base.RoundTrip(attemptReq)
		if !isTransientNetworkError(err) {
			// Either success (err == nil) or a non-retryable error: return as-is.
			return resp, err
		}

		// Transient error: drain and close any partial response before retrying
		// to avoid leaking the underlying connection.
		if resp != nil {
			_ = resp.Body.Close()
		}
		lastErr = err
	}

	return nil, lastErr
}

// retryableMethod reports whether the HTTP method is safe to retry without
// risking duplicate side effects.
func retryableMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodDelete:
		return true
	default:
		return false
	}
}

// replayableBody reports whether the request body can be re-sent on a retry.
// A request with no body, an explicit http.NoBody, or a body that can be
// obtained again via GetBody is replayable.
func replayableBody(req *http.Request) bool {
	return req.Body == nil || req.Body == http.NoBody || req.GetBody != nil
}

// isTransientNetworkError reports whether err is a connection-level error that
// may have occurred before or during the response, making a retry safe. It does
// not treat HTTP status codes (e.g. 5xx, 429) as retryable: those are returned
// to the caller, preserving the apiserver's status semantics (e.g. a 404 on a
// retried DELETE flows through to conformance cleanup's IsNotFound handling).
func isTransientNetworkError(err error) bool {
	if err == nil {
		return false
	}
	if netutil.IsProbableEOF(err) || netutil.IsConnectionReset(err) || netutil.IsHTTP2ConnectionLost(err) {
		return true
	}
	// net.Error covers request timeouts (DeadlineExceeded), which are safe to
	// retry for idempotent verbs.
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	return false
}

// compile-time assertion that retryTransport satisfies http.RoundTripper.
var _ http.RoundTripper = &retryTransport{}
