// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"
)

// fakeTransport is a programmable http.RoundTripper for testing retryTransport.
// It records every request it sees (including the body bytes read on each
// attempt) and returns the configured response/error sequence.
type fakeTransport struct {
	// calls is the ordered list of (resp, err) to return, one per RoundTrip.
	// If there are fewer calls than attempts, the last entry is reused.
	calls []call
	// seenBodies records the raw body bytes read on each attempt, in order.
	seenBodies [][]byte
	// reqs records each request (post body-reset) seen by the transport.
	reqs []*http.Request
	idx  int
}

type call struct {
	// status is the HTTP status code to return on a success attempt. Ignored
	// when err != nil.
	status int
	err    error
}

func (f *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Read and snapshot the body the transport actually received, so we can
	// assert that retries replay the full body (not a consumed/truncated one).
	var bodyBytes []byte
	if req.Body != nil && req.Body != http.NoBody {
		bodyBytes, _ = io.ReadAll(req.Body)
		req.Body.Close()
	}
	f.seenBodies = append(f.seenBodies, bodyBytes)
	f.reqs = append(f.reqs, req)

	i := f.idx
	if i >= len(f.calls) {
		i = len(f.calls) - 1
	}
	f.idx++

	c := f.calls[i]
	if c.err != nil {
		return nil, c.err
	}
	// Construct the response here (not in the call struct literal) so the
	// bodyclose linter can track its ownership from construction through return
	// to the caller, which closes it via closeBody. http.NoBody needs no close.
	return okResponse(c.status), nil
}

// newGetRequest builds a GET request with no body.
func newGetRequest(t *testing.T) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.com/pods", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	return req
}

// newDeleteRequest builds a DELETE request carrying a JSON body, the way
// client-go serializes DeleteOptions via bytes.NewReader(bodyBytes). The
// stdlib auto-populates req.GetBody for *bytes.Reader, making it replayable.
func newDeleteRequest(t *testing.T, body string) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete,
		"https://example.com/apis/gateway.networking.k8s.io/v1/gatewayclasses/upgrade",
		bytes.NewReader([]byte(body)))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	return req
}

// newPostRequest builds a POST request carrying a non-replayable body
// (a plain io.Reader with no GetBody), so the retry loop must NOT retry it.
func newPostRequest(t *testing.T) *http.Request {
	t.Helper()
	// Use an io.Reader type that stdlib does NOT special-case, so GetBody
	// stays nil and the body is not replayable.
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost,
		"https://example.com/pods", &eofReader{})
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	return req
}

// eofReader is an io.Reader that yields its bytes once; it is not one of the
// stdlib-special-cased types, so http.NewRequestWithContext leaves GetBody nil.
type eofReader struct{ n int }

func (r *eofReader) Read(p []byte) (int, error) {
	if r.n > 0 {
		r.n--
		p[0] = 'x'
		return 1, nil
	}
	return 0, io.EOF
}

// okResponse builds a minimal response. It uses http.NoBody (a sentinel that
// needs no closing) so the bodyclose linter is satisfied and the fake transport
// can return the same response across retries without body-aliasing concerns;
// tests only inspect StatusCode.
func okResponse(status int) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       http.NoBody,
		Header:     make(http.Header),
	}
}

// closeBody drains and closes resp.Body; a no-op for nil resp or http.NoBody.
func closeBody(t *testing.T, resp *http.Response) {
	t.Helper()
	if resp == nil || resp.Body == nil || resp.Body == http.NoBody {
		return
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
}

func TestRetryTransport_RetriesEOFOnGetThenSucceeds(t *testing.T) {
	ft := &fakeTransport{calls: []call{
		{err: io.EOF},
		{err: io.ErrUnexpectedEOF},
		{status: http.StatusOK},
	}}
	rt := &retryTransport{base: ft}

	resp, err := rt.RoundTrip(newGetRequest(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	closeBody(t, resp)
	if got := len(ft.seenBodies); got != 3 {
		t.Fatalf("expected 3 attempts, got %d", got)
	}
}

func TestRetryTransport_RetriesEOFOnDeleteAndReplaysBody(t *testing.T) {
	// Simulate the teardown flake: first DELETE reaches the apiserver but the
	// response is lost (EOF). The retry replays the same body and the second
	// attempt returns 404 (which conformance cleanup treats as success).
	payload := `{"kind":"DeleteOptions","apiVersion":"v1","propagationPolicy":"Background"}`
	ft := &fakeTransport{calls: []call{
		{err: io.EOF},
		{status: http.StatusNotFound},
	}}
	rt := &retryTransport{base: ft}

	resp, err := rt.RoundTrip(newDeleteRequest(t, payload))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 to be passed through, got %d", resp.StatusCode)
	}
	closeBody(t, resp)
	if got := len(ft.seenBodies); got != 2 {
		t.Fatalf("expected 2 attempts, got %d", got)
	}
	// Both attempts must have received the FULL payload: the retry must not
	// send a consumed/truncated body.
	for i, b := range ft.seenBodies {
		if string(b) != payload {
			t.Fatalf("attempt %d body mismatch:\n got: %q\nwant: %q", i, b, payload)
		}
	}
}

func TestRetryTransport_DoesNotRetryPost(t *testing.T) {
	ft := &fakeTransport{calls: []call{
		{err: io.EOF},
		{status: http.StatusOK},
	}}
	rt := &retryTransport{base: ft}

	// POST with a non-replayable body must surface the first error without
	// retrying, so the success response is never reached; resp stays nil.
	resp, err := rt.RoundTrip(newPostRequest(t))
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected EOF to be returned without retry, got %v", err)
	}
	closeBody(t, resp)
	if got := len(ft.seenBodies); got != 1 {
		t.Fatalf("POST must not be retried; expected 1 attempt, got %d", got)
	}
}

func TestRetryTransport_DoesNotRetryNonReplayableBody(t *testing.T) {
	// A GET with a body but no GetBody is not replayable: must not retry even
	// though the method is idempotent.
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet,
		"https://example.com/pods", &eofReader{})
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	ft := &fakeTransport{calls: []call{
		{err: io.EOF},
		{status: http.StatusOK},
	}}
	rt := &retryTransport{base: ft}

	resp, err := rt.RoundTrip(req)
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected EOF returned without retry, got %v", err)
	}
	closeBody(t, resp)
	if got := len(ft.seenBodies); got != 1 {
		t.Fatalf("non-replayable body must not be retried; expected 1 attempt, got %d", got)
	}
}

func TestRetryTransport_DoesNotRetrySuccessStatus(t *testing.T) {
	// A 500/429 is a successful transport response (err==nil); it must be
	// returned to the caller, NOT retried, so apiserver status semantics are
	// preserved (e.g. a 404 on a retried DELETE flows to IsNotFound handling).
	for _, status := range []int{http.StatusInternalServerError, http.StatusTooManyRequests, http.StatusNotFound} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			ft := &fakeTransport{calls: []call{{status: status}}}
			rt := &retryTransport{base: ft}

			resp, err := rt.RoundTrip(newDeleteRequest(t, "{}"))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.StatusCode != status {
				t.Fatalf("expected status %d, got %d", status, resp.StatusCode)
			}
			closeBody(t, resp)
			if got := len(ft.seenBodies); got != 1 {
				t.Fatalf("HTTP status must not be retried; expected 1 attempt, got %d", got)
			}
		})
	}
}

func TestRetryTransport_ExhaustsAttemptsAndReturnsLastError(t *testing.T) {
	ft := &fakeTransport{calls: []call{{err: io.EOF}}}
	rt := &retryTransport{base: ft}

	resp, err := rt.RoundTrip(newGetRequest(t))
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected last error EOF, got %v", err)
	}
	closeBody(t, resp)
	if got := len(ft.seenBodies); got != retryTransportMaxAttempts {
		t.Fatalf("expected %d attempts, got %d", retryTransportMaxAttempts, got)
	}
}

func TestRetryTransport_HonorsContextCancellation(t *testing.T) {
	// A context cancelled before the first retry sleep must short-circuit the
	// between-attempt backoff rather than waiting for the timer.
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancelled up front: the between-attempt select must fire Done.

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.com/pods", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	ft := &fakeTransport{calls: []call{{err: io.EOF}}}
	rt := &retryTransport{base: ft}

	start := time.Now()
	resp, err := rt.RoundTrip(req)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	closeBody(t, resp)
	// The first attempt always runs (no sleep before attempt 0); the cancel is
	// observed on the first between-attempt select, so exactly 1 attempt runs
	// and we should NOT have waited the full backoff.
	if got := len(ft.seenBodies); got != 1 {
		t.Fatalf("expected 1 attempt before context cancel observed, got %d", got)
	}
	if elapsed := time.Since(start); elapsed >= retryTransportBaseBackoff {
		t.Fatalf("context cancel should short-circuit backoff; waited %v", elapsed)
	}
}

func TestRetryTransport_ContextCancelDuringBackoffLeaksNoBody(t *testing.T) {
	// Regression guard for the body-creation ordering: when a retried request
	// has a replayable body (GetBody != nil) and the context is cancelled during
	// the between-attempt backoff, GetBody must NOT have been invoked — so no
	// body ReadCloser is created and left unclosed. (An earlier version derived
	// the body before the sleep, leaking it on cancel.)
	var getBodyCalls int
	payload := `{"kind":"DeleteOptions"}`
	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete,
		"https://example.com/apis/gateway.networking.k8s.io/v1/gatewayclasses/upgrade",
		bytes.NewReader([]byte(payload)))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	// Wrap GetBody to count invocations; stdlib already set it for *bytes.Reader.
	origGetBody := req.GetBody
	req.GetBody = func() (io.ReadCloser, error) {
		getBodyCalls++
		return origGetBody()
	}

	// Cancel during the first backoff sleep: schedule cancel after a tiny delay
	// so attempt 0 (which sees io.EOF) has run, then the retry's select fires.
	ctx, cancel := context.WithCancel(req.Context())
	req = req.WithContext(ctx)
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()

	ft := &fakeTransport{calls: []call{{err: io.EOF}}}
	rt := &retryTransport{base: ft}

	resp, err := rt.RoundTrip(req)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	closeBody(t, resp)
	// Attempt 0 ran (no backoff before it); the retry's backoff select observed
	// the cancellation BEFORE GetBody was consulted, so GetBody stays at 0.
	if getBodyCalls != 0 {
		t.Fatalf("GetBody should not be called on cancel-during-backoff; called %d time(s)", getBodyCalls)
	}
	if got := len(ft.seenBodies); got != 1 {
		t.Fatalf("expected 1 attempt, got %d", got)
	}
}

func TestRetryTransport_ConnectionResetAndTimeoutAreRetried(t *testing.T) {
	// Use a timeout error (net.Error) and verify it is treated as transient.
	timeoutErr := &timeoutError{}
	ft := &fakeTransport{calls: []call{
		{err: timeoutErr},
		{status: http.StatusOK},
	}}
	rt := &retryTransport{base: ft}

	resp, err := rt.RoundTrip(newGetRequest(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	closeBody(t, resp)
	if got := len(ft.seenBodies); got != 2 {
		t.Fatalf("expected 2 attempts, got %d", got)
	}
}

// timeoutError implements net.Error for testing the Timeout() branch.
type timeoutError struct{}

func (timeoutError) Error() string   { return "i/o timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

func TestIsTransientNetworkError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"io.EOF", io.EOF, true},
		{"io.ErrUnexpectedEOF", io.ErrUnexpectedEOF, true},
		{"timeout", &timeoutError{}, true},
		{"generic", errors.New("boom"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTransientNetworkError(tt.err); got != tt.want {
				t.Fatalf("isTransientNetworkError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestReplayableBody(t *testing.T) {
	// No body.
	getReq := newGetRequest(t)
	if !replayableBody(getReq) {
		t.Fatal("GET with nil body should be replayable")
	}
	// bytes.Reader-backed body => GetBody auto-set by stdlib => replayable.
	delReq := newDeleteRequest(t, `{"a":1}`)
	if !replayableBody(delReq) {
		t.Fatal("DELETE with bytes.Reader body should be replayable via GetBody")
	}
	if delReq.GetBody == nil {
		t.Fatal("expected stdlib to populate GetBody for *bytes.Reader body")
	}
	// Non-replayable body (no GetBody).
	postReq := newPostRequest(t)
	if replayableBody(postReq) {
		t.Fatal("POST with non-replayable body should not be replayable")
	}
}
