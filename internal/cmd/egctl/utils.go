// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"fmt"

	adminv3 "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	"google.golang.org/protobuf/reflect/protoreflect"
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
		for _, config := range globalConfigs.Configs {
			if config.GetTypeUrl() == "type.googleapis.com/envoy.admin.v3.BootstrapConfigDump" {
				return config, nil
			}
		}
	case EndpointEnvoyConfigType:
		for _, config := range globalConfigs.Configs {
			if config.GetTypeUrl() == "type.googleapis.com/envoy.admin.v3.EndpointsConfigDump" {
				return config, nil
			}
		}

	case ClusterEnvoyConfigType:
		for _, config := range globalConfigs.Configs {
			if config.GetTypeUrl() == "type.googleapis.com/envoy.admin.v3.ClustersConfigDump" {
				return config, nil
			}
		}
	case ListenerEnvoyConfigType:
		for _, config := range globalConfigs.Configs {
			if config.GetTypeUrl() == "type.googleapis.com/envoy.admin.v3.ListenersConfigDump" {
				return config, nil
			}
		}
	case RouteEnvoyConfigType:
		for _, config := range globalConfigs.Configs {
			if config.GetTypeUrl() == "type.googleapis.com/envoy.admin.v3.RoutesConfigDump" {
				return config, nil
			}
		}
	case AllEnvoyConfigType:
		return globalConfigs, nil
	default:
		return nil, fmt.Errorf("unknown resourceType %s", resourceType)
	}

	return nil, fmt.Errorf("unknown resourceType %s", resourceType)
}
