// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package common

import (
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
)

// ConvertResolvedMetricSinks converts IR metric sinks to bootstrap format.
func ConvertResolvedMetricSinks(irSinks []ir.ResolvedMetricSink) []bootstrap.MetricSink {
	result := make([]bootstrap.MetricSink, 0, len(irSinks))
	for i := range irSinks {
		sink := &irSinks[i]
		if len(sink.Backends) == 0 || len(sink.Backends[0].Settings) == 0 || len(sink.Backends[0].Settings[0].Endpoints) == 0 {
			continue
		}
		// Metrics are aggregated locally in Envoy and exported to one collector.
		ep := sink.Backends[0].Settings[0].Endpoints[0]
		ms := bootstrap.MetricSink{
			Address:                  ep.Host,
			Port:                     ep.Port,
			Authority:                sink.Authority,
			ReportCountersAsDeltas:   sink.ReportCountersAsDeltas,
			ReportHistogramsAsDeltas: sink.ReportHistogramsAsDeltas,
			Headers:                  sink.Headers,
			ResourceAttributes:       sink.ResourceAttributes,
		}
		if tls := sink.Backends[0].Settings[0].TLS; tls != nil {
			ms.TLS = &bootstrap.MetricSinkTLS{
				UseSystemTrustStore: tls.UseSystemTrustStore,
			}
			if tls.SNI != nil {
				ms.TLS.SNI = *tls.SNI
			}
			if tls.CACertificate != nil {
				ms.TLS.CACertificate = tls.CACertificate.Certificate
			}
		}
		result = append(result, ms)
	}
	return result
}
