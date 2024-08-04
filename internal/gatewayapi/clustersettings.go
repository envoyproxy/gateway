// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	perr "github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func translateTrafficFeatures(policy *egv1a1.ClusterSettings) (*ir.TrafficFeatures, error) {
	if policy == nil {
		return nil, nil
	}
	ret := &ir.TrafficFeatures{}

	if timeout, err := buildTimeout(*policy, nil); err != nil {
		return nil, err
	} else {
		ret.Timeout = timeout
	}

	if bc, err := buildBackendConnection(*policy); err != nil {
		return nil, err
	} else {
		ret.BackendConnection = bc
	}

	if ka, err := buildTCPKeepAlive(*policy); err != nil {
		return nil, err
	} else {
		ret.TCPKeepalive = ka
	}

	if cb, err := buildCircuitBreaker(*policy); err != nil {
		return nil, err
	} else {
		ret.CircuitBreaker = cb
	}

	if lb, err := buildLoadBalancer(*policy); err != nil {
		return nil, err
	} else {
		ret.LoadBalancer = lb
	}

	ret.ProxyProtocol = buildProxyProtocol(*policy)

	ret.HealthCheck = buildHealthCheck(*policy)

	ret.DNS = translateDNS(*policy)

	// If nothing was set in any of the above calls, return nil instead of an empty
	// container
	var empty ir.TrafficFeatures
	if empty == *ret {
		ret = nil
	}

	return ret, nil
}

func buildTimeout(policy egv1a1.ClusterSettings, r *ir.HTTPRoute) (*ir.Timeout, error) {
	if policy.Timeout == nil {
		return nil, nil
	}
	var (
		tto  *ir.TCPTimeout
		hto  *ir.HTTPTimeout
		terr bool
		errs error
	)

	pto := policy.Timeout

	if pto.TCP != nil && pto.TCP.ConnectTimeout != nil {
		d, err := time.ParseDuration(string(*pto.TCP.ConnectTimeout))
		if err != nil {
			terr = true
			errs = errors.Join(errs, fmt.Errorf("invalid ConnectTimeout value %s", *pto.TCP.ConnectTimeout))
		} else {
			tto = &ir.TCPTimeout{
				ConnectTimeout: ptr.To(metav1.Duration{Duration: d}),
			}
		}
	}

	if pto.HTTP != nil {
		var cit *metav1.Duration
		var mcd *metav1.Duration

		if pto.HTTP.ConnectionIdleTimeout != nil {
			d, err := time.ParseDuration(string(*pto.HTTP.ConnectionIdleTimeout))
			if err != nil {
				terr = true
				errs = errors.Join(errs, fmt.Errorf("invalid ConnectionIdleTimeout value %s", *pto.HTTP.ConnectionIdleTimeout))
			} else {
				cit = ptr.To(metav1.Duration{Duration: d})
			}
		}

		if pto.HTTP.MaxConnectionDuration != nil {
			d, err := time.ParseDuration(string(*pto.HTTP.MaxConnectionDuration))
			if err != nil {
				terr = true
				errs = errors.Join(errs, fmt.Errorf("invalid MaxConnectionDuration value %s", *pto.HTTP.MaxConnectionDuration))
			} else {
				mcd = ptr.To(metav1.Duration{Duration: d})
			}
		}

		hto = &ir.HTTPTimeout{
			ConnectionIdleTimeout: cit,
			MaxConnectionDuration: mcd,
		}
	}

	// http request timeout is translated during the gateway-api route resource translation
	// merge route timeout setting with backendtrafficpolicy timeout settings
	if terr {
		if r != nil && r.Traffic != nil && r.Traffic.Timeout != nil {
			return r.Traffic.Timeout.DeepCopy(), errs
		}
	} else {
		// http request timeout is translated during the gateway-api route resource translation
		// merge route timeout setting with backendtrafficpolicy timeout settings
		if r != nil &&
			r.Traffic != nil &&
			r.Traffic.Timeout != nil &&
			r.Traffic.Timeout.HTTP != nil &&
			r.Traffic.Timeout.HTTP.RequestTimeout != nil {
			if hto == nil {
				hto = &ir.HTTPTimeout{
					RequestTimeout: r.Traffic.Timeout.HTTP.RequestTimeout,
				}
			} else {
				hto.RequestTimeout = r.Traffic.Timeout.HTTP.RequestTimeout
			}
		}

		if hto != nil || tto != nil {
			return &ir.Timeout{
				TCP:  tto,
				HTTP: hto,
			}, nil
		}
	}

	return nil, errs
}

func buildBackendConnection(policy egv1a1.ClusterSettings) (*ir.BackendConnection, error) {
	if policy.Connection == nil {
		return nil, nil
	}
	var (
		bcIR = &ir.BackendConnection{}
		bc   = &egv1a1.BackendConnection{}
	)

	if policy.Connection != nil {
		bc = policy.Connection

		if bc.BufferLimit != nil {
			bf, ok := bc.BufferLimit.AsInt64()
			if !ok {
				return nil, fmt.Errorf("invalid BufferLimit value %s", bc.BufferLimit.String())
			}
			if bf < 0 || bf > math.MaxUint32 {
				return nil, fmt.Errorf("BufferLimit value %s is out of range", bc.BufferLimit.String())
			}

			bcIR.BufferLimitBytes = ptr.To(uint32(bf))
		}
	}

	return bcIR, nil
}

func buildTCPKeepAlive(policy egv1a1.ClusterSettings) (*ir.TCPKeepalive, error) {
	if policy.TCPKeepalive == nil {
		return nil, nil
	}

	pka := policy.TCPKeepalive
	ka := &ir.TCPKeepalive{}

	if pka.Probes != nil {
		ka.Probes = pka.Probes
	}

	if pka.IdleTime != nil {
		d, err := time.ParseDuration(string(*pka.IdleTime))
		if err != nil {
			return nil, fmt.Errorf("invalid IdleTime value %s", *pka.IdleTime)
		}
		ka.IdleTime = ptr.To(uint32(d.Seconds()))
	}

	if pka.Interval != nil {
		d, err := time.ParseDuration(string(*pka.Interval))
		if err != nil {
			return nil, fmt.Errorf("invalid Interval value %s", *pka.Interval)
		}
		ka.Interval = ptr.To(uint32(d.Seconds()))
	}
	return ka, nil
}

func buildCircuitBreaker(policy egv1a1.ClusterSettings) (*ir.CircuitBreaker, error) {
	if policy.CircuitBreaker == nil {
		return nil, nil
	}

	var cb *ir.CircuitBreaker
	pcb := policy.CircuitBreaker

	if pcb != nil {
		cb = &ir.CircuitBreaker{}

		if pcb.MaxConnections != nil {
			if ui32, ok := int64ToUint32(*pcb.MaxConnections); ok {
				cb.MaxConnections = &ui32
			} else {
				return nil, fmt.Errorf("invalid MaxConnections value %d", *pcb.MaxConnections)
			}
		}

		if pcb.MaxParallelRequests != nil {
			if ui32, ok := int64ToUint32(*pcb.MaxParallelRequests); ok {
				cb.MaxParallelRequests = &ui32
			} else {
				return nil, fmt.Errorf("invalid MaxParallelRequests value %d", *pcb.MaxParallelRequests)
			}
		}

		if pcb.MaxPendingRequests != nil {
			if ui32, ok := int64ToUint32(*pcb.MaxPendingRequests); ok {
				cb.MaxPendingRequests = &ui32
			} else {
				return nil, fmt.Errorf("invalid MaxPendingRequests value %d", *pcb.MaxPendingRequests)
			}
		}

		if pcb.MaxParallelRetries != nil {
			if ui32, ok := int64ToUint32(*pcb.MaxParallelRetries); ok {
				cb.MaxParallelRetries = &ui32
			} else {
				return nil, fmt.Errorf("invalid MaxParallelRetries value %d", *pcb.MaxParallelRetries)
			}
		}

		if pcb.MaxRequestsPerConnection != nil {
			if ui32, ok := int64ToUint32(*pcb.MaxRequestsPerConnection); ok {
				cb.MaxRequestsPerConnection = &ui32
			} else {
				return nil, fmt.Errorf("invalid MaxRequestsPerConnection value %d", *pcb.MaxRequestsPerConnection)
			}
		}

	}

	return cb, nil
}

func buildLoadBalancer(policy egv1a1.ClusterSettings) (*ir.LoadBalancer, error) {
	if policy.LoadBalancer == nil {
		return nil, nil
	}
	var lb *ir.LoadBalancer
	switch policy.LoadBalancer.Type {
	case egv1a1.ConsistentHashLoadBalancerType:
		consistentHash, err := buildConsistentHashLoadBalancer(*policy.LoadBalancer)
		if err != nil {
			return nil, perr.WithMessage(err, "ConsistentHash")
		}

		lb = &ir.LoadBalancer{
			ConsistentHash: consistentHash,
		}
	case egv1a1.LeastRequestLoadBalancerType:
		lb = &ir.LoadBalancer{}
		if policy.LoadBalancer.SlowStart != nil {
			if policy.LoadBalancer.SlowStart.Window != nil {
				lb.LeastRequest = &ir.LeastRequest{
					SlowStart: &ir.SlowStart{
						Window: policy.LoadBalancer.SlowStart.Window,
					},
				}
			}
		}
	case egv1a1.RandomLoadBalancerType:
		lb = &ir.LoadBalancer{
			Random: &ir.Random{},
		}
	case egv1a1.RoundRobinLoadBalancerType:
		lb = &ir.LoadBalancer{
			RoundRobin: &ir.RoundRobin{
				SlowStart: &ir.SlowStart{},
			},
		}
		if policy.LoadBalancer.SlowStart != nil {
			if policy.LoadBalancer.SlowStart.Window != nil {
				lb.RoundRobin = &ir.RoundRobin{
					SlowStart: &ir.SlowStart{
						Window: policy.LoadBalancer.SlowStart.Window,
					},
				}
			}
		}
	}

	return lb, nil
}

func buildConsistentHashLoadBalancer(policy egv1a1.LoadBalancer) (*ir.ConsistentHash, error) {
	consistentHash := &ir.ConsistentHash{}

	if policy.ConsistentHash.TableSize != nil {
		tableSize := policy.ConsistentHash.TableSize

		if *tableSize > MaxConsistentHashTableSize || !big.NewInt(int64(*tableSize)).ProbablyPrime(0) {
			return nil, fmt.Errorf("invalid TableSize value %d", *tableSize)
		}

		consistentHash.TableSize = tableSize
	}

	switch policy.ConsistentHash.Type {
	case egv1a1.SourceIPConsistentHashType:
		consistentHash.SourceIP = ptr.To(true)
	case egv1a1.HeaderConsistentHashType:
		consistentHash.Header = &ir.Header{
			Name: policy.ConsistentHash.Header.Name,
		}
	case egv1a1.CookieConsistentHashType:
		consistentHash.Cookie = policy.ConsistentHash.Cookie
	}

	return consistentHash, nil
}

func buildProxyProtocol(policy egv1a1.ClusterSettings) *ir.ProxyProtocol {
	if policy.ProxyProtocol == nil {
		return nil
	}
	var pp *ir.ProxyProtocol
	switch policy.ProxyProtocol.Version {
	case egv1a1.ProxyProtocolVersionV1:
		pp = &ir.ProxyProtocol{
			Version: ir.ProxyProtocolVersionV1,
		}
	case egv1a1.ProxyProtocolVersionV2:
		pp = &ir.ProxyProtocol{
			Version: ir.ProxyProtocolVersionV2,
		}
	}

	return pp
}

func buildHealthCheck(policy egv1a1.ClusterSettings) *ir.HealthCheck {
	if policy.HealthCheck == nil {
		return nil
	}

	irhc := &ir.HealthCheck{}
	irhc.Passive = buildPassiveHealthCheck(*policy.HealthCheck)
	irhc.Active = buildActiveHealthCheck(*policy.HealthCheck)

	return irhc
}

func buildPassiveHealthCheck(policy egv1a1.HealthCheck) *ir.OutlierDetection {
	if policy.Passive == nil {
		return nil
	}

	hc := policy.Passive
	irOD := &ir.OutlierDetection{
		Interval:                       hc.Interval,
		SplitExternalLocalOriginErrors: hc.SplitExternalLocalOriginErrors,
		ConsecutiveLocalOriginFailures: hc.ConsecutiveLocalOriginFailures,
		ConsecutiveGatewayErrors:       hc.ConsecutiveGatewayErrors,
		Consecutive5xxErrors:           hc.Consecutive5xxErrors,
		BaseEjectionTime:               hc.BaseEjectionTime,
		MaxEjectionPercent:             hc.MaxEjectionPercent,
	}
	return irOD
}

func buildActiveHealthCheck(policy egv1a1.HealthCheck) *ir.ActiveHealthCheck {
	if policy.Active == nil {
		return nil
	}

	hc := policy.Active
	irHC := &ir.ActiveHealthCheck{
		Timeout:            hc.Timeout,
		Interval:           hc.Interval,
		UnhealthyThreshold: hc.UnhealthyThreshold,
		HealthyThreshold:   hc.HealthyThreshold,
	}
	switch hc.Type {
	case egv1a1.ActiveHealthCheckerTypeHTTP:
		irHC.HTTP = buildHTTPActiveHealthChecker(hc.HTTP)
	case egv1a1.ActiveHealthCheckerTypeTCP:
		irHC.TCP = buildTCPActiveHealthChecker(hc.TCP)
	}

	return irHC
}

func buildHTTPActiveHealthChecker(h *egv1a1.HTTPActiveHealthChecker) *ir.HTTPHealthChecker {
	if h == nil {
		return nil
	}

	irHTTP := &ir.HTTPHealthChecker{
		Path:   h.Path,
		Method: h.Method,
	}
	if irHTTP.Method != nil {
		*irHTTP.Method = strings.ToUpper(*irHTTP.Method)
	}

	// deduplicate http statuses
	statusSet := sets.NewInt()
	for _, r := range h.ExpectedStatuses {
		statusSet.Insert(int(r))
	}
	irStatuses := make([]ir.HTTPStatus, 0, statusSet.Len())

	for _, r := range statusSet.List() {
		irStatuses = append(irStatuses, ir.HTTPStatus(r))
	}
	irHTTP.ExpectedStatuses = irStatuses

	irHTTP.ExpectedResponse = translateActiveHealthCheckPayload(h.ExpectedResponse)
	return irHTTP
}

func buildTCPActiveHealthChecker(h *egv1a1.TCPActiveHealthChecker) *ir.TCPHealthChecker {
	if h == nil {
		return nil
	}

	irTCP := &ir.TCPHealthChecker{
		Send:    translateActiveHealthCheckPayload(h.Send),
		Receive: translateActiveHealthCheckPayload(h.Receive),
	}
	return irTCP
}

func translateActiveHealthCheckPayload(p *egv1a1.ActiveHealthCheckPayload) *ir.HealthCheckPayload {
	if p == nil {
		return nil
	}

	irPayload := &ir.HealthCheckPayload{}
	switch p.Type {
	case egv1a1.ActiveHealthCheckPayloadTypeText:
		irPayload.Text = p.Text
	case egv1a1.ActiveHealthCheckPayloadTypeBinary:
		irPayload.Binary = make([]byte, len(p.Binary))
		copy(irPayload.Binary, p.Binary)
	}

	return irPayload
}

func translateDNS(policy egv1a1.ClusterSettings) *ir.DNS {
	if policy.DNS == nil {
		return nil
	}
	return &ir.DNS{
		RespectDNSTTL:  policy.DNS.RespectDNSTTL,
		DNSRefreshRate: policy.DNS.DNSRefreshRate,
	}
}
