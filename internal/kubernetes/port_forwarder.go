// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"

	netutil "github.com/envoyproxy/gateway/internal/utils/net"
)

type PortForwarder interface {
	// Start waits indefinitely for the port forward to become ready. Prefer
	// StartWithContext when the caller needs a bound on how long it can block.
	Start() error

	// StartWithContext is like Start, but returns ctx.Err() once ctx is done instead of
	// blocking forever. This can't interrupt a Kubernetes upgrade/dial already in flight -
	// client-go's dialer offers no cancellation hook for it - but it guarantees the caller
	// isn't blocked past ctx's deadline, and it stops the forwarder before returning so a
	// dial that does eventually complete doesn't outlive the caller.
	StartWithContext(ctx context.Context) error

	Stop()

	WaitForStop()

	// Address returns the address of the local forwarded address.
	Address() string
}

var _ PortForwarder = &localForwarder{}

type localForwarder struct {
	types.NamespacedName
	CLIClient

	localPort int
	podPort   int

	stopCh   chan struct{}
	stopOnce sync.Once
}

func NewLocalPortForwarder(client CLIClient, namespacedName types.NamespacedName, localPort, podPort int) (PortForwarder, error) {
	f := &localForwarder{
		stopCh:         make(chan struct{}),
		CLIClient:      client,
		NamespacedName: namespacedName,
		localPort:      localPort,
		podPort:        podPort,
	}
	if f.localPort == 0 {
		// get a random port
		p, err := netutil.LocalAvailablePort()
		if err != nil {
			return nil, fmt.Errorf("failed to get a local available port for Pod %q: %w", namespacedName, err)
		}
		f.localPort = p
	}

	return f, nil
}

func (f *localForwarder) Start() error {
	return f.StartWithContext(context.Background())
}

func (f *localForwarder) StartWithContext(ctx context.Context) error {
	errCh := make(chan error, 1)
	readyCh := make(chan struct{}, 1)
	go func() {
		for {
			select {
			case <-f.stopCh:
				return
			default:
			}

			fw, err := f.buildKubernetesPortForwarder(readyCh)
			if err != nil {
				errCh <- err
				return
			}

			if err := fw.ForwardPorts(); err != nil {
				errCh <- err
				return
			}

			readyCh = nil
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("failed to start port forwarder: %w", err)
	case <-readyCh:
		return nil
	case <-ctx.Done():
		// Give up rather than block the caller past ctx's deadline. Stop the forwarder
		// so that, if the wedged dial above does eventually return, it tears down
		// instead of lingering with a bound local port nobody is using anymore.
		f.Stop()
		return ctx.Err()
	}
}

func (f *localForwarder) buildKubernetesPortForwarder(readyCh chan struct{}) (*portforward.PortForwarder, error) {
	restClient, err := rest.RESTClientFor(f.RESTConfig())
	if err != nil {
		return nil, err
	}

	req := restClient.Post().Resource("pods").Namespace(f.Namespace).Name(f.Name).SubResource("portforward")
	serverURL := req.URL()

	roundTripper, upgrader, err := spdy.RoundTripperFor(f.RESTConfig())
	if err != nil {
		return nil, fmt.Errorf("failure creating roundtripper: %w", err)
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, serverURL)
	tunnelingDialer, err := portforward.NewSPDYOverWebsocketDialer(serverURL, f.RESTConfig())
	if err != nil {
		return nil, err
	}
	// Prefer Websocket dialer, fallback to SPDY dialer.
	dialer = portforward.NewFallbackDialer(tunnelingDialer, dialer, func(err error) bool {
		return httpstream.IsUpgradeFailure(err) || httpstream.IsHTTPSProxyError(err)
	})

	fw, err := portforward.NewOnAddresses(dialer,
		[]string{netutil.DefaultLocalAddress},
		[]string{fmt.Sprintf("%d:%d", f.localPort, f.podPort)},
		f.stopCh,
		readyCh,
		io.Discard,
		os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("failed establishing portforward: %w", err)
	}

	return fw, nil
}

func (f *localForwarder) Stop() {
	// Idempotent: StartWithContext may already have stopped the forwarder on a timeout
	// before the caller gets a chance to call Stop() itself (e.g. via a deferred call
	// guarded by a successful Start, or an explicit Stop() in an error path).
	f.stopOnce.Do(func() {
		close(f.stopCh)
	})
}

func (f *localForwarder) WaitForStop() {
	<-f.stopCh
}

func (f *localForwarder) Address() string {
	return net.JoinHostPort(netutil.DefaultLocalAddress, strconv.Itoa(f.localPort))
}
