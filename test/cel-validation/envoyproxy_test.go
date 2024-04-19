// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build celvalidation
// +build celvalidation

package celvalidation

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestEnvoyProxyProvider(t *testing.T) {
	ctx := context.Background()
	baseEnvoyProxy := egv1a1.EnvoyProxy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "proxy",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: egv1a1.EnvoyProxySpec{},
	}

	cases := []struct {
		desc         string
		mutate       func(envoy *egv1a1.EnvoyProxy)
		mutateStatus func(envoy *egv1a1.EnvoyProxy)
		wantErrors   []string
	}{
		{
			desc:       "nil provider",
			mutate:     func(envoy *egv1a1.EnvoyProxy) {},
			wantErrors: []string{},
		},
		{
			desc: "unsupported provider",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: "foo",
					},
				}
			},
			wantErrors: []string{"Unsupported value: \"foo\": supported values: \"Kubernetes\""},
		},
		{
			desc: "invalid service type",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Type: ptr.To(egv1a1.ServiceType("foo")),
							},
						},
					},
				}
			},
			wantErrors: []string{"Unsupported value: \"foo\": supported values: \"ClusterIP\", \"LoadBalancer\", \"NodePort\""},
		},
		{
			desc: "allocateLoadBalancerNodePorts-pass-case1",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Type:                          ptr.To(egv1a1.ServiceTypeLoadBalancer),
								AllocateLoadBalancerNodePorts: ptr.To(true),
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "allocateLoadBalancerNodePorts-pass-case2",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Type: ptr.To(egv1a1.ServiceTypeClusterIP),
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "allocateLoadBalancerNodePorts-fail",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Type:                          ptr.To(egv1a1.ServiceTypeClusterIP),
								AllocateLoadBalancerNodePorts: ptr.To(true),
							},
						},
					},
				}
			},
			wantErrors: []string{"allocateLoadBalancerNodePorts can only be set for LoadBalancer type"},
		},
		{
			desc: "loadBalancerSourceRanges-pass-case1",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Type:                     ptr.To(egv1a1.ServiceTypeLoadBalancer),
								LoadBalancerSourceRanges: []string{"1.1.1.1"},
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "loadBalancerSourceRanges-pass-case2",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Type: ptr.To(egv1a1.ServiceTypeClusterIP),
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "loadBalancerSourceRanges-fail",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Type:                     ptr.To(egv1a1.ServiceTypeClusterIP),
								LoadBalancerSourceRanges: []string{"1.1.1.1"},
							},
						},
					},
				}
			},
			wantErrors: []string{"loadBalancerSourceRanges can only be set for LoadBalancer type"},
		},
		{
			desc: "ServiceTypeLoadBalancer-with-valid-IP",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Type:           ptr.To(egv1a1.ServiceTypeLoadBalancer),
								LoadBalancerIP: ptr.To("20.205.243.166"), // github ip for test only
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "ServiceTypeLoadBalancer-with-empty-IP",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Type: ptr.To(egv1a1.ServiceTypeLoadBalancer),
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "ServiceTypeLoadBalancer-with-invalid-IP",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Type:           ptr.To(egv1a1.ServiceTypeLoadBalancer),
								LoadBalancerIP: ptr.To("1.2.3.4."),
							},
						},
					},
				}
			},
			wantErrors: []string{"loadBalancerIP must be a valid IPv4 address"},
		},
		{
			desc: "ServiceTypeLoadBalancer-with-invalid-IP",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Type:           ptr.To(egv1a1.ServiceTypeLoadBalancer),
								LoadBalancerIP: ptr.To("a.b.c.d"),
							},
						},
					},
				}
			},
			wantErrors: []string{"loadBalancerIP must be a valid IPv4 address"},
		},
		{
			desc: "ServiceTypeClusterIP-with-empty-IP",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Type: ptr.To(egv1a1.ServiceTypeClusterIP),
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "ServiceTypeClusterIP-with-LoadBalancerIP",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyService: &egv1a1.KubernetesServiceSpec{
								Type:           ptr.To(egv1a1.ServiceTypeClusterIP),
								LoadBalancerIP: ptr.To("20.205.243.166"), // github ip for test only
							},
						},
					},
				}
			},
			wantErrors: []string{"loadBalancerIP can only be set for LoadBalancer type"},
		},
		{
			desc: "invalid-ProxyAccessLogFormat",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						AccessLog: &egv1a1.ProxyAccessLog{
							Settings: []egv1a1.ProxyAccessLogSetting{
								{
									Format: egv1a1.ProxyAccessLogFormat{
										Type: "foo",
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{"Unsupported value: \"foo\": supported values: \"Text\", \"JSON\""},
		},
		{
			desc: "ProxyAccessLogFormat-with-TypeText-but-no-text",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						AccessLog: &egv1a1.ProxyAccessLog{
							Settings: []egv1a1.ProxyAccessLogSetting{
								{
									Format: egv1a1.ProxyAccessLogFormat{
										Type: egv1a1.ProxyAccessLogFormatTypeText,
									},
									Sinks: []egv1a1.ProxyAccessLogSink{
										{
											Type: egv1a1.ProxyAccessLogSinkTypeFile,
											File: &egv1a1.FileEnvoyProxyAccessLog{
												Path: "foo/bar",
											},
										},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{"If AccessLogFormat type is Text, text field needs to be set"},
		},
		{
			desc: "ProxyAccessLogFormat-with-TypeJSON-but-no-json",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						AccessLog: &egv1a1.ProxyAccessLog{
							Settings: []egv1a1.ProxyAccessLogSetting{
								{
									Format: egv1a1.ProxyAccessLogFormat{
										Type: egv1a1.ProxyAccessLogFormatTypeJSON,
									},
									Sinks: []egv1a1.ProxyAccessLogSink{
										{
											Type: egv1a1.ProxyAccessLogSinkTypeFile,
											File: &egv1a1.FileEnvoyProxyAccessLog{
												Path: "foo/bar",
											},
										},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{"If AccessLogFormat type is JSON, json field needs to be set"},
		},
		{
			desc: "ProxyAccessLogFormat-with-TypeJSON-but-got-text",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						AccessLog: &egv1a1.ProxyAccessLog{
							Settings: []egv1a1.ProxyAccessLogSetting{
								{
									Format: egv1a1.ProxyAccessLogFormat{
										Type: egv1a1.ProxyAccessLogFormatTypeJSON,
										Text: ptr.To("[%START_TIME%]"),
									},
									Sinks: []egv1a1.ProxyAccessLogSink{
										{
											Type: egv1a1.ProxyAccessLogSinkTypeFile,
											File: &egv1a1.FileEnvoyProxyAccessLog{
												Path: "foo/bar",
											},
										},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{"If AccessLogFormat type is JSON, json field needs to be set"},
		},
		{
			desc: "ProxyAccessLogSink-with-TypeALS-but-no-als",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						AccessLog: &egv1a1.ProxyAccessLog{
							Settings: []egv1a1.ProxyAccessLogSetting{
								{
									Format: egv1a1.ProxyAccessLogFormat{
										Type: egv1a1.ProxyAccessLogFormatTypeJSON,
										JSON: map[string]string{
											"foo": "bar",
										},
									},
									Sinks: []egv1a1.ProxyAccessLogSink{
										{
											Type: egv1a1.ProxyAccessLogSinkTypeALS,
										},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{"If AccessLogSink type is ALS, als field needs to be set"},
		},
		{
			desc: "ProxyAccessLogSink-with-TypeFile-but-no-file",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						AccessLog: &egv1a1.ProxyAccessLog{
							Settings: []egv1a1.ProxyAccessLogSetting{
								{
									Format: egv1a1.ProxyAccessLogFormat{
										Type: egv1a1.ProxyAccessLogFormatTypeText,
										Text: ptr.To("[%START_TIME%]"),
									},
									Sinks: []egv1a1.ProxyAccessLogSink{
										{
											Type: egv1a1.ProxyAccessLogSinkTypeFile,
										},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{"If AccessLogSink type is File, file field needs to be set"},
		},
		{
			desc: "ProxyAccessLogSink-with-TypeOpenTelemetry-but-no-openTelemetry",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						AccessLog: &egv1a1.ProxyAccessLog{
							Settings: []egv1a1.ProxyAccessLogSetting{
								{
									Format: egv1a1.ProxyAccessLogFormat{
										Type: egv1a1.ProxyAccessLogFormatTypeText,
										Text: ptr.To("[%START_TIME%]"),
									},
									Sinks: []egv1a1.ProxyAccessLogSink{
										{
											Type: egv1a1.ProxyAccessLogSinkTypeOpenTelemetry,
										},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{"If AccessLogSink type is OpenTelemetry, openTelemetry field needs to be set"},
		},
		{
			desc: "ProxyAccessLog-settings-pass",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						AccessLog: &egv1a1.ProxyAccessLog{
							Settings: []egv1a1.ProxyAccessLogSetting{
								{
									Format: egv1a1.ProxyAccessLogFormat{
										Type: egv1a1.ProxyAccessLogFormatTypeText,
										Text: ptr.To("[%START_TIME%]"),
									},
									Sinks: []egv1a1.ProxyAccessLogSink{
										{
											Type: egv1a1.ProxyAccessLogSinkTypeFile,
											File: &egv1a1.FileEnvoyProxyAccessLog{
												Path: "foo/bar",
											},
										},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "accesslog-ALS",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						AccessLog: &egv1a1.ProxyAccessLog{
							Settings: []egv1a1.ProxyAccessLogSetting{
								{
									Format: egv1a1.ProxyAccessLogFormat{
										Type: egv1a1.ProxyAccessLogFormatTypeJSON,
										JSON: map[string]string{
											"attr1": "value1",
											"attr2": "value2",
										},
									},
									Sinks: []egv1a1.ProxyAccessLogSink{
										{
											Type: egv1a1.ProxyAccessLogSinkTypeALS,
											ALS: &egv1a1.ALSEnvoyProxyAccessLog{
												BackendRefs: []egv1a1.BackendRef{
													{
														BackendObjectReference: gwapiv1.BackendObjectReference{
															Name: "fake-service",
															Port: ptr.To(gwapiv1.PortNumber(9000)),
														},
													},
												},
												Type: egv1a1.ALSEnvoyProxyAccessLogTypeHTTP,
											},
										},
									},
								},
							},
						},
					},
				}
			},
		},
		{
			desc: "invalid-accesslog-ALS-type",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						AccessLog: &egv1a1.ProxyAccessLog{
							Settings: []egv1a1.ProxyAccessLogSetting{
								{
									Format: egv1a1.ProxyAccessLogFormat{
										Type: "Text",
										Text: ptr.To("[%START_TIME%]"),
									},
									Sinks: []egv1a1.ProxyAccessLogSink{
										{
											Type: egv1a1.ProxyAccessLogSinkTypeALS,
											ALS: &egv1a1.ALSEnvoyProxyAccessLog{
												BackendRefs: []egv1a1.BackendRef{
													{
														BackendObjectReference: gwapiv1.BackendObjectReference{
															Name: "fake-service",
															Port: ptr.To(gwapiv1.PortNumber(9000)),
														},
													},
												},
												Type: egv1a1.ALSEnvoyProxyAccessLogTypeTCP,
												HTTP: &egv1a1.ALSEnvoyProxyHTTPAccessLogConfig{},
											},
										},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{"The http field may only be set when type is HTTP."},
		},
		{
			desc: "invalid-accesslog-ALS-backendrefs",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						AccessLog: &egv1a1.ProxyAccessLog{
							Settings: []egv1a1.ProxyAccessLogSetting{
								{
									Format: egv1a1.ProxyAccessLogFormat{
										Type: "Text",
										Text: ptr.To("[%START_TIME%]"),
									},
									Sinks: []egv1a1.ProxyAccessLogSink{
										{
											Type: egv1a1.ProxyAccessLogSinkTypeALS,
											ALS: &egv1a1.ALSEnvoyProxyAccessLog{
												BackendRefs: []egv1a1.BackendRef{
													{
														BackendObjectReference: gwapiv1.BackendObjectReference{
															Name: "fake-service",
															Kind: ptr.To(gwapiv1.Kind("foo")),
														},
													},
												},
												Type: egv1a1.ALSEnvoyProxyAccessLogTypeHTTP,
											},
										},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{"BackendRefs only supports Service Kind."},
		},
		{
			desc: "invalid-accesslog-ALS-no-backendrefs",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						AccessLog: &egv1a1.ProxyAccessLog{
							Settings: []egv1a1.ProxyAccessLogSetting{
								{
									Format: egv1a1.ProxyAccessLogFormat{
										Type: "Text",
										Text: ptr.To("[%START_TIME%]"),
									},
									Sinks: []egv1a1.ProxyAccessLogSink{
										{
											Type: egv1a1.ProxyAccessLogSinkTypeALS,
											ALS: &egv1a1.ALSEnvoyProxyAccessLog{
												Type: egv1a1.ALSEnvoyProxyAccessLogTypeHTTP,
											},
										},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{"Invalid value: \"null\""},
		},
		{
			desc: "invalid-accesslog-ALS-empty-backendrefs",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						AccessLog: &egv1a1.ProxyAccessLog{
							Settings: []egv1a1.ProxyAccessLogSetting{
								{
									Format: egv1a1.ProxyAccessLogFormat{
										Type: "Text",
										Text: ptr.To("[%START_TIME%]"),
									},
									Sinks: []egv1a1.ProxyAccessLogSink{
										{
											Type: egv1a1.ProxyAccessLogSinkTypeALS,
											ALS: &egv1a1.ALSEnvoyProxyAccessLog{
												BackendRefs: []egv1a1.BackendRef{},
												Type:        egv1a1.ALSEnvoyProxyAccessLogTypeHTTP,
											},
										},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{"should have at least 1 items"},
		},
		{
			desc: "invalid-accesslog-ALS-multi-backendrefs",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						AccessLog: &egv1a1.ProxyAccessLog{
							Settings: []egv1a1.ProxyAccessLogSetting{
								{
									Format: egv1a1.ProxyAccessLogFormat{
										Type: "Text",
										Text: ptr.To("[%START_TIME%]"),
									},
									Sinks: []egv1a1.ProxyAccessLogSink{
										{
											Type: egv1a1.ProxyAccessLogSinkTypeALS,
											ALS: &egv1a1.ALSEnvoyProxyAccessLog{
												BackendRefs: []egv1a1.BackendRef{
													{
														BackendObjectReference: gwapiv1.BackendObjectReference{
															Name: "fake-service",
															Port: ptr.To(gwapiv1.PortNumber(8080)),
														},
													},
													{
														BackendObjectReference: gwapiv1.BackendObjectReference{
															Name: "fake-service",
															Port: ptr.To(gwapiv1.PortNumber(8080)),
														},
													},
												},
												Type: egv1a1.ALSEnvoyProxyAccessLogTypeHTTP,
											},
										},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{"must have at most 1 items"},
		},
		{
			desc: "accesslog-OpenTelemetry",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						AccessLog: &egv1a1.ProxyAccessLog{
							Settings: []egv1a1.ProxyAccessLogSetting{
								{
									Format: egv1a1.ProxyAccessLogFormat{
										Type: "Text",
										Text: ptr.To("[%START_TIME%]"),
									},
									Sinks: []egv1a1.ProxyAccessLogSink{
										{
											Type: egv1a1.ProxyAccessLogSinkTypeOpenTelemetry,
											OpenTelemetry: &egv1a1.OpenTelemetryEnvoyProxyAccessLog{
												Host: ptr.To("0.0.0.0"),
												Port: 8080,
											},
										},
									},
								},
							},
						},
					},
				}
			},
		},
		{
			desc: "invalid-accesslog-backendref",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						AccessLog: &egv1a1.ProxyAccessLog{
							Settings: []egv1a1.ProxyAccessLogSetting{
								{
									Format: egv1a1.ProxyAccessLogFormat{
										Type: "Text",
										Text: ptr.To("[%START_TIME%]"),
									},
									Sinks: []egv1a1.ProxyAccessLogSink{
										{
											Type: egv1a1.ProxyAccessLogSinkTypeOpenTelemetry,
											OpenTelemetry: &egv1a1.OpenTelemetryEnvoyProxyAccessLog{
												BackendRefs: []egv1a1.BackendRef{
													{
														BackendObjectReference: gwapiv1.BackendObjectReference{
															Name: "fake-service",
															Kind: ptr.To(gwapiv1.Kind("foo")),
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{"only support Service Kind."},
		},
		{
			desc: "accesslog-backendref",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						AccessLog: &egv1a1.ProxyAccessLog{
							Settings: []egv1a1.ProxyAccessLogSetting{
								{
									Format: egv1a1.ProxyAccessLogFormat{
										Type: "Text",
										Text: ptr.To("[%START_TIME%]"),
									},
									Sinks: []egv1a1.ProxyAccessLogSink{
										{
											Type: egv1a1.ProxyAccessLogSinkTypeOpenTelemetry,
											OpenTelemetry: &egv1a1.OpenTelemetryEnvoyProxyAccessLog{
												BackendRefs: []egv1a1.BackendRef{
													{
														BackendObjectReference: gwapiv1.BackendObjectReference{
															Name: "fake-service",
															Kind: ptr.To(gwapiv1.Kind("Service")),
															Port: ptr.To(gwapiv1.PortNumber(8080)),
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}
			},
		},
		{
			desc: "multi-accesslog-backendref",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						AccessLog: &egv1a1.ProxyAccessLog{
							Settings: []egv1a1.ProxyAccessLogSetting{
								{
									Format: egv1a1.ProxyAccessLogFormat{
										Type: "Text",
										Text: ptr.To("[%START_TIME%]"),
									},
									Sinks: []egv1a1.ProxyAccessLogSink{
										{
											Type: egv1a1.ProxyAccessLogSinkTypeOpenTelemetry,
											OpenTelemetry: &egv1a1.OpenTelemetryEnvoyProxyAccessLog{
												BackendRefs: []egv1a1.BackendRef{
													{
														BackendObjectReference: gwapiv1.BackendObjectReference{
															Name: "fake-service",
															Kind: ptr.To(gwapiv1.Kind("Service")),
															Port: ptr.To(gwapiv1.PortNumber(8080)),
														},
													},
													{
														BackendObjectReference: gwapiv1.BackendObjectReference{
															Name: "fake-service",
															Kind: ptr.To(gwapiv1.Kind("Service")),
															Port: ptr.To(gwapiv1.PortNumber(8080)),
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{"must have at most 1 items"},
		},
		{
			desc: "accesslog-backendref-empty-kind",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						AccessLog: &egv1a1.ProxyAccessLog{
							Settings: []egv1a1.ProxyAccessLogSetting{
								{
									Format: egv1a1.ProxyAccessLogFormat{
										Type: "Text",
										Text: ptr.To("[%START_TIME%]"),
									},
									Sinks: []egv1a1.ProxyAccessLogSink{
										{
											Type: egv1a1.ProxyAccessLogSinkTypeOpenTelemetry,
											OpenTelemetry: &egv1a1.OpenTelemetryEnvoyProxyAccessLog{
												BackendRefs: []egv1a1.BackendRef{
													{
														BackendObjectReference: gwapiv1.BackendObjectReference{
															Name: "fake-service",
															Port: ptr.To(gwapiv1.PortNumber(8080)),
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}
			},
		},
		{
			desc: "accesslog-backend-empty",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						AccessLog: &egv1a1.ProxyAccessLog{
							Settings: []egv1a1.ProxyAccessLogSetting{
								{
									Format: egv1a1.ProxyAccessLogFormat{
										Type: "Text",
										Text: ptr.To("[%START_TIME%]"),
									},
									Sinks: []egv1a1.ProxyAccessLogSink{
										{
											Type:          egv1a1.ProxyAccessLogSinkTypeOpenTelemetry,
											OpenTelemetry: &egv1a1.OpenTelemetryEnvoyProxyAccessLog{},
										},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{"host or backendRefs needs to be set"},
		},
		{
			desc: "ProxyMetricSink-with-TypeOpenTelemetry-but-no-openTelemetry",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						Metrics: &egv1a1.ProxyMetrics{
							Sinks: []egv1a1.ProxyMetricSink{
								{
									Type: egv1a1.MetricSinkTypeOpenTelemetry,
								},
							},
						},
					},
				}
			},
			wantErrors: []string{"If MetricSink type is OpenTelemetry, openTelemetry field needs to be set"},
		},
		{
			desc: "ProxyMetrics-sinks-pass",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						Metrics: &egv1a1.ProxyMetrics{
							Sinks: []egv1a1.ProxyMetricSink{
								{
									Type: egv1a1.MetricSinkTypeOpenTelemetry,
									OpenTelemetry: &egv1a1.ProxyOpenTelemetrySink{
										Host: ptr.To("0.0.0.0"),
										Port: 3217,
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "ProxyMetrics-sinks-backend-empty",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						Metrics: &egv1a1.ProxyMetrics{
							Sinks: []egv1a1.ProxyMetricSink{
								{
									Type:          egv1a1.MetricSinkTypeOpenTelemetry,
									OpenTelemetry: &egv1a1.ProxyOpenTelemetrySink{},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{"host or backendRefs needs to be set"},
		},
		{
			desc: "ProxyMetrics-sinks-backendref",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						Metrics: &egv1a1.ProxyMetrics{
							Sinks: []egv1a1.ProxyMetricSink{
								{
									Type: egv1a1.MetricSinkTypeOpenTelemetry,
									OpenTelemetry: &egv1a1.ProxyOpenTelemetrySink{
										BackendRefs: []egv1a1.BackendRef{
											{
												BackendObjectReference: gwapiv1.BackendObjectReference{
													Name: "fake-service",
													Port: ptr.To(gwapiv1.PortNumber(8080)),
												},
											},
										},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "ProxyMetrics-sinks-multi-backendref",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						Metrics: &egv1a1.ProxyMetrics{
							Sinks: []egv1a1.ProxyMetricSink{
								{
									Type: egv1a1.MetricSinkTypeOpenTelemetry,
									OpenTelemetry: &egv1a1.ProxyOpenTelemetrySink{
										BackendRefs: []egv1a1.BackendRef{
											{
												BackendObjectReference: gwapiv1.BackendObjectReference{
													Name: "fake-service",
													Port: ptr.To(gwapiv1.PortNumber(8080)),
												},
											},
											{
												BackendObjectReference: gwapiv1.BackendObjectReference{
													Name: "fake-service",
													Port: ptr.To(gwapiv1.PortNumber(8080)),
												},
											},
										},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{"must have at most 1 items"},
		},
		{
			desc: "ProxyMetrics-sinks-backendref-empty-kind",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						Metrics: &egv1a1.ProxyMetrics{
							Sinks: []egv1a1.ProxyMetricSink{
								{
									Type: egv1a1.MetricSinkTypeOpenTelemetry,
									OpenTelemetry: &egv1a1.ProxyOpenTelemetrySink{
										BackendRefs: []egv1a1.BackendRef{
											{
												BackendObjectReference: gwapiv1.BackendObjectReference{
													Name: "fake-service",
													Port: ptr.To(gwapiv1.PortNumber(8080)),
												},
											},
										},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "ProxyMetrics-sinks-invalid-backendref",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						Metrics: &egv1a1.ProxyMetrics{
							Sinks: []egv1a1.ProxyMetricSink{
								{
									Type: egv1a1.MetricSinkTypeOpenTelemetry,
									OpenTelemetry: &egv1a1.ProxyOpenTelemetrySink{
										BackendRefs: []egv1a1.BackendRef{
											{
												BackendObjectReference: gwapiv1.BackendObjectReference{
													Name: "fake-service",
													Kind: ptr.To(gwapiv1.Kind("foo")),
													Port: ptr.To(gwapiv1.PortNumber(8080)),
												},
											},
										},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{"only support Service Kind."},
		},
		{
			desc: "invalid-tracing-backendref-invalid-kind",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						Tracing: &egv1a1.ProxyTracing{
							Provider: egv1a1.TracingProvider{
								Type: egv1a1.TracingProviderTypeOpenTelemetry,
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "fake-service",
											Kind: ptr.To(gwapiv1.Kind("foo")),
										},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{"only support Service Kind."},
		},
		{
			desc: "tracing-backendref-empty-kind",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						Tracing: &egv1a1.ProxyTracing{
							Provider: egv1a1.TracingProvider{
								Type: egv1a1.TracingProviderTypeOpenTelemetry,
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "fake-service",
											Port: ptr.To(gwapiv1.PortNumber(8080)),
										},
									},
								},
							},
						},
					},
				}
			},
		},
		{
			desc: "tracing-backendref",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						Tracing: &egv1a1.ProxyTracing{
							Provider: egv1a1.TracingProvider{
								Type: egv1a1.TracingProviderTypeOpenTelemetry,
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "fake-service",
											Kind: ptr.To(gwapiv1.Kind("Service")),
											Port: ptr.To(gwapiv1.PortNumber(8080)),
										},
									},
								},
							},
						},
					},
				}
			},
		},
		{
			desc: "tracing-multi-backendref",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						Tracing: &egv1a1.ProxyTracing{
							Provider: egv1a1.TracingProvider{
								Type: egv1a1.TracingProviderTypeOpenTelemetry,
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "fake-service",
											Kind: ptr.To(gwapiv1.Kind("Service")),
											Port: ptr.To(gwapiv1.PortNumber(8080)),
										},
									},
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "fake-service",
											Kind: ptr.To(gwapiv1.Kind("Service")),
											Port: ptr.To(gwapiv1.PortNumber(8080)),
										},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{"must have at most 1 items"},
		},
		{
			desc: "tracing-empty-backend",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Telemetry: &egv1a1.ProxyTelemetry{
						Tracing: &egv1a1.ProxyTracing{
							Provider: egv1a1.TracingProvider{
								Type: egv1a1.TracingProviderTypeOpenTelemetry,
							},
						},
					},
				}
			},
			wantErrors: []string{"host or backendRefs needs to be set"},
		},
		{
			desc: "ProxyHpa-maxReplicas-is-required",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyHpa: &egv1a1.KubernetesHorizontalPodAutoscalerSpec{},
						},
					},
				}
			},
			wantErrors: []string{"spec.provider.kubernetes.envoyHpa.maxReplicas: Required value"},
		},
		{
			desc: "ProxyHpa-minReplicas-less-than-0",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyHpa: &egv1a1.KubernetesHorizontalPodAutoscalerSpec{
								MinReplicas: ptr.To[int32](-1),
								MaxReplicas: ptr.To[int32](2),
							},
						},
					},
				}
			},
			wantErrors: []string{"minReplicas must be greater than 0"},
		},
		{
			desc: "ProxyHpa-maxReplicas-less-than-0",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyHpa: &egv1a1.KubernetesHorizontalPodAutoscalerSpec{
								MaxReplicas: ptr.To[int32](-1),
							},
						},
					},
				}
			},
			wantErrors: []string{"maxReplicas must be greater than 0"},
		},
		{
			desc: "ProxyHpa-maxReplicas-less-than-minReplicas",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyHpa: &egv1a1.KubernetesHorizontalPodAutoscalerSpec{
								MinReplicas: ptr.To[int32](5),
								MaxReplicas: ptr.To[int32](2),
							},
						},
					},
				}
			},
			wantErrors: []string{"maxReplicas cannot be less than minReplicas"},
		},
		{
			desc: "ProxyHpa-maxReplicas-equals-to-minReplicas",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyHpa: &egv1a1.KubernetesHorizontalPodAutoscalerSpec{
								MinReplicas: ptr.To[int32](2),
								MaxReplicas: ptr.To[int32](2),
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "ProxyHpa-valid",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					Provider: &egv1a1.EnvoyProxyProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
							EnvoyHpa: &egv1a1.KubernetesHorizontalPodAutoscalerSpec{
								MinReplicas: ptr.To[int32](5),
								MaxReplicas: ptr.To[int32](10),
							},
						},
					},
				}
			},
		},
		{
			desc: "ProxyFilterOrder-with-before-and-after",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					FilterOrder: []egv1a1.FilterPosition{
						{
							Name:   egv1a1.EnvoyFilterRateLimit,
							Before: ptr.To(egv1a1.EnvoyFilterCORS),
							After:  ptr.To(egv1a1.EnvoyFilterBasicAuthn),
						},
					},
				}
			},
			wantErrors: []string{"only one of before or after can be specified"},
		},
		{
			desc: "ProxyFilterOrder-without-before-or-after",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					FilterOrder: []egv1a1.FilterPosition{
						{
							Name: egv1a1.EnvoyFilterRateLimit,
						},
					},
				}
			},
			wantErrors: []string{"one of before or after must be specified"},
		},
		{
			desc: "ProxyFilterOrder-with-before",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					FilterOrder: []egv1a1.FilterPosition{
						{
							Name:   egv1a1.EnvoyFilterRateLimit,
							Before: ptr.To(egv1a1.EnvoyFilterCORS),
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "ProxyFilterOrder-with-after",
			mutate: func(envoy *egv1a1.EnvoyProxy) {
				envoy.Spec = egv1a1.EnvoyProxySpec{
					FilterOrder: []egv1a1.FilterPosition{
						{
							Name:  egv1a1.EnvoyFilterRateLimit,
							After: ptr.To(egv1a1.EnvoyFilterBasicAuthn),
						},
					},
				}
			},
			wantErrors: []string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			proxy := baseEnvoyProxy.DeepCopy()
			proxy.Name = fmt.Sprintf("proxy-%v", time.Now().UnixNano())

			if tc.mutate != nil {
				tc.mutate(proxy)
			}
			err := c.Create(ctx, proxy)

			if tc.mutateStatus != nil {
				tc.mutateStatus(proxy)
				err = c.Status().Update(ctx, proxy)
			}

			if (len(tc.wantErrors) != 0) != (err != nil) {
				t.Fatalf("Unexpected response while creating EnvoyProxy; got err=\n%v\n;want error=%v", err, tc.wantErrors)
			}

			var missingErrorStrings []string
			for _, wantError := range tc.wantErrors {
				if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(wantError)) {
					missingErrorStrings = append(missingErrorStrings, wantError)
				}
			}
			if len(missingErrorStrings) != 0 {
				t.Errorf("Unexpected response while creating EnvoyProxy; got err=\n%v\n;missing strings within error=%q", err, missingErrorStrings)
			}
		})
	}
}
