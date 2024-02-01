// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"fmt"

	adminv3 "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	"google.golang.org/protobuf/reflect/protoreflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

type envoyConfigType string

var (
	BootstrapEnvoyConfigType envoyConfigType = "bootstrap"
	ClusterEnvoyConfigType   envoyConfigType = "cluster"
	EndpointEnvoyConfigType  envoyConfigType = "endpoint"
	ListenerEnvoyConfigType  envoyConfigType = "listener"
	RouteEnvoyConfigType     envoyConfigType = "route"
	AllEnvoyConfigType       envoyConfigType = "all"
)

func findXDSResourceFromConfigDump(resourceType envoyConfigType, globalConfigs *adminv3.ConfigDump) (protoreflect.ProtoMessage, error) {
	switch resourceType {
	case BootstrapEnvoyConfigType:
		for _, cfg := range globalConfigs.Configs {
			if cfg.GetTypeUrl() == "type.googleapis.com/envoy.admin.v3.BootstrapConfigDump" {
				return cfg, nil
			}
		}
	case EndpointEnvoyConfigType:
		for _, cfg := range globalConfigs.Configs {
			if cfg.GetTypeUrl() == "type.googleapis.com/envoy.admin.v3.EndpointsConfigDump" {
				return cfg, nil
			}
		}

	case ClusterEnvoyConfigType:
		for _, cfg := range globalConfigs.Configs {
			if cfg.GetTypeUrl() == "type.googleapis.com/envoy.admin.v3.ClustersConfigDump" {
				return cfg, nil
			}
		}
	case ListenerEnvoyConfigType:
		for _, cfg := range globalConfigs.Configs {
			if cfg.GetTypeUrl() == "type.googleapis.com/envoy.admin.v3.ListenersConfigDump" {
				return cfg, nil
			}
		}
	case RouteEnvoyConfigType:
		for _, cfg := range globalConfigs.Configs {
			if cfg.GetTypeUrl() == "type.googleapis.com/envoy.admin.v3.RoutesConfigDump" {
				return cfg, nil
			}
		}
	case AllEnvoyConfigType:
		return globalConfigs, nil
	default:
		return nil, fmt.Errorf("unknown resourceType %s", resourceType)
	}

	return nil, fmt.Errorf("unknown resourceType %s", resourceType)
}

func newK8sClient() (client.Client, error) {
	cli, err := client.New(config.GetConfigOrDie(), client.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Kubernetes client: %v", err)
	}

	gwv1.AddToScheme(cli.Scheme())
	gwv1b1.AddToScheme(cli.Scheme())
	gwv1a2.AddToScheme(cli.Scheme())
	egv1a1.AddToScheme(cli.Scheme())

	return cli, nil
}
