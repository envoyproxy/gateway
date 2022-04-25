// Copyright Project Contour Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package envoygatewayconfig

import (
	"fmt"
	"time"

	"github.com/imdario/mergo"
	envoygateway_api_v1alpha1 "github.com/projectcontour/contour/apis/envoygateway/v1alpha1"
	"github.com/projectcontour/contour/internal/timeout"
	"github.com/projectcontour/contour/pkg/config"
	"k8s.io/utils/pointer"
)

// UIntPtr returns a pointer to a uint value.
func UIntPtr(val uint) *uint {
	return &val
}

// UInt32Ptr returns a pointer to a uint32 value.
func UInt32Ptr(val uint32) *uint32 {
	return &val
}

// OverlayOnDefaults overlays the settings in the provided spec onto the
// default settings, and returns the results.
func OverlayOnDefaults(spec envoygateway_api_v1alpha1.EnvoyGatewayConfigurationSpec) (envoygateway_api_v1alpha1.EnvoyGatewayConfigurationSpec, error) {
	res := Defaults()

	if err := mergo.Merge(&res, spec, mergo.WithOverride); err != nil {
		return envoygateway_api_v1alpha1.EnvoyGatewayConfigurationSpec{}, err
	}

	return res, nil
}

// Defaults returns the default settings Contour uses if no user-specified
// configuration is provided.
func Defaults() envoygateway_api_v1alpha1.EnvoyGatewayConfigurationSpec {
	return envoygateway_api_v1alpha1.EnvoyGatewayConfigurationSpec{
		XDSServer: &envoygateway_api_v1alpha1.XDSServerConfig{
			Type:    envoygateway_api_v1alpha1.ContourServerType,
			Address: "0.0.0.0",
			Port:    8001,
			TLS: &envoygateway_api_v1alpha1.TLS{
				CAFile:   "/certs/ca.crt",
				CertFile: "/certs/tls.crt",
				KeyFile:  "/certs/tls.key",
				Insecure: pointer.Bool(false),
			},
		},
		Ingress: &envoygateway_api_v1alpha1.IngressConfig{
			ClassNames:    nil,
			StatusAddress: "",
		},
		Debug: &envoygateway_api_v1alpha1.DebugConfig{
			Address:                 "127.0.0.1",
			Port:                    6060,
			DebugLogLevel:           envoygateway_api_v1alpha1.InfoLog,
			KubernetesDebugLogLevel: UIntPtr(0),
		},
		Health: &envoygateway_api_v1alpha1.HealthConfig{
			Address: "0.0.0.0",
			Port:    8000,
		},
		Envoy: &envoygateway_api_v1alpha1.EnvoyConfig{
			Listener: &envoygateway_api_v1alpha1.EnvoyListenerConfig{
				UseProxyProto:             pointer.Bool(false),
				DisableAllowChunkedLength: pointer.Bool(false),
				DisableMergeSlashes:       pointer.Bool(false),
				ConnectionBalancer:        "",
				TLS: &envoygateway_api_v1alpha1.EnvoyTLS{
					MinimumProtocolVersion: "1.2",
					CipherSuites:           defaultCipherSuites(),
				},
			},
			Service: &envoygateway_api_v1alpha1.NamespacedName{
				Namespace: "projectcontour",
				Name:      "envoy",
			},
			HTTPListener: &envoygateway_api_v1alpha1.EnvoyListener{
				Address:   "0.0.0.0",
				Port:      8080,
				AccessLog: "/dev/stdout",
			},
			HTTPSListener: &envoygateway_api_v1alpha1.EnvoyListener{
				Address:   "0.0.0.0",
				Port:      8443,
				AccessLog: "/dev/stdout",
			},
			Health: &envoygateway_api_v1alpha1.HealthConfig{
				Address: "0.0.0.0",
				Port:    8002,
			},
			Metrics: &envoygateway_api_v1alpha1.MetricsConfig{
				Address: "0.0.0.0",
				Port:    8002,
				TLS:     nil,
			},
			ClientCertificate: nil,
			Logging: &envoygateway_api_v1alpha1.EnvoyLogging{
				AccessLogFormat:       envoygateway_api_v1alpha1.EnvoyAccessLog,
				AccessLogFormatString: "",
				AccessLogFields:       nil,
				AccessLogLevel:        envoygateway_api_v1alpha1.LogLevelInfo,
			},
			DefaultHTTPVersions: []envoygateway_api_v1alpha1.HTTPVersionType{
				"HTTP/1.1",
				"HTTP/2",
			},
			Timeouts: &envoygateway_api_v1alpha1.TimeoutParameters{
				RequestTimeout:                nil,
				ConnectionIdleTimeout:         nil,
				StreamIdleTimeout:             nil,
				MaxConnectionDuration:         nil,
				DelayedCloseTimeout:           nil,
				ConnectionShutdownGracePeriod: nil,
				ConnectTimeout:                nil,
			},
			Cluster: &envoygateway_api_v1alpha1.ClusterParameters{
				DNSLookupFamily: envoygateway_api_v1alpha1.AutoClusterDNSFamily,
			},
			Network: &envoygateway_api_v1alpha1.NetworkParameters{
				XffNumTrustedHops: UInt32Ptr(0),
				EnvoyAdminPort:    pointer.Int(9001),
			},
		},
		Gateway: nil,
		HTTPProxy: &envoygateway_api_v1alpha1.HTTPProxyConfig{
			DisablePermitInsecure: pointer.Bool(false),
			RootNamespaces:        nil,
			FallbackCertificate:   nil,
		},
		EnableExternalNameService: pointer.Bool(false),
		RateLimitService:          nil,
		Policy: &envoygateway_api_v1alpha1.PolicyConfig{
			RequestHeadersPolicy:  &envoygateway_api_v1alpha1.HeadersPolicy{},
			ResponseHeadersPolicy: &envoygateway_api_v1alpha1.HeadersPolicy{},
			ApplyToIngress:        pointer.Bool(false),
		},
		Metrics: &envoygateway_api_v1alpha1.MetricsConfig{
			Address: "0.0.0.0",
			Port:    8000,
			TLS:     nil,
		},
	}
}

func defaultCipherSuites() []envoygateway_api_v1alpha1.TLSCipherType {
	var res []envoygateway_api_v1alpha1.TLSCipherType

	for _, cipherSuite := range config.DefaultTLSCiphers {
		res = append(res, envoygateway_api_v1alpha1.TLSCipherType(cipherSuite))
	}

	return res
}

type Timeouts struct {
	Request                       timeout.Setting
	ConnectionIdle                timeout.Setting
	StreamIdle                    timeout.Setting
	MaxConnectionDuration         timeout.Setting
	DelayedClose                  timeout.Setting
	ConnectionShutdownGracePeriod timeout.Setting
	ConnectTimeout                time.Duration // Since "infinite" is not valid ConnectTimeout value, use time.Duration instead of timeout.Setting.
}

func ParseTimeoutPolicy(timeoutParameters *envoygateway_api_v1alpha1.TimeoutParameters) (Timeouts, error) {
	var (
		err      error
		timeouts Timeouts
	)

	if timeoutParameters == nil {
		return timeouts, nil
	}

	if timeoutParameters.RequestTimeout != nil {
		timeouts.Request, err = timeout.Parse(*timeoutParameters.RequestTimeout)
		if err != nil {
			return Timeouts{}, fmt.Errorf("failed to parse request timeout: %s", err)
		}
	}
	if timeoutParameters.ConnectionIdleTimeout != nil {
		timeouts.ConnectionIdle, err = timeout.Parse(*timeoutParameters.ConnectionIdleTimeout)
		if err != nil {
			return Timeouts{}, fmt.Errorf("failed to parse connection idle timeout: %s", err)
		}
	}
	if timeoutParameters.StreamIdleTimeout != nil {
		timeouts.StreamIdle, err = timeout.Parse(*timeoutParameters.StreamIdleTimeout)
		if err != nil {
			return Timeouts{}, fmt.Errorf("failed to parse stream idle timeout: %s", err)
		}
	}
	if timeoutParameters.MaxConnectionDuration != nil {
		timeouts.MaxConnectionDuration, err = timeout.Parse(*timeoutParameters.MaxConnectionDuration)
		if err != nil {
			return Timeouts{}, fmt.Errorf("failed to parse max connection duration: %s", err)
		}
	}
	if timeoutParameters.DelayedCloseTimeout != nil {
		timeouts.DelayedClose, err = timeout.Parse(*timeoutParameters.DelayedCloseTimeout)
		if err != nil {
			return Timeouts{}, fmt.Errorf("failed to parse delayed close timeout: %s", err)
		}
	}
	if timeoutParameters.ConnectionShutdownGracePeriod != nil {
		timeouts.ConnectionShutdownGracePeriod, err = timeout.Parse(*timeoutParameters.ConnectionShutdownGracePeriod)
		if err != nil {
			return Timeouts{}, fmt.Errorf("failed to parse connection shutdown grace period: %s", err)
		}
	}
	if timeoutParameters.ConnectTimeout != nil {
		timeouts.ConnectTimeout, err = time.ParseDuration(*timeoutParameters.ConnectTimeout)
		if err != nil {
			return Timeouts{}, fmt.Errorf("failed to parse connect timeout: %s", err)
		}
	}

	return timeouts, nil
}
