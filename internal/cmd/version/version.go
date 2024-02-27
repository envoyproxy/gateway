// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package version

import (
	"encoding/json"
	"fmt"
	"io"
	"runtime/debug"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/gateway/api/v1alpha1"
)

type Info struct {
	EnvoyGatewayVersion    string `json:"envoyGatewayVersion"`
	GatewayAPIVersion      string `json:"gatewayAPIVersion"`
	EnvoyProxyVersion      string `json:"envoyProxyVersion"`
	ShutdownManagerVersion string `json:"shutdownManagerVersion"`
	GitCommitID            string `json:"gitCommitID"`
}

func Get() Info {
	return Info{
		EnvoyGatewayVersion:    envoyGatewayVersion,
		GatewayAPIVersion:      gatewayAPIVersion,
		EnvoyProxyVersion:      envoyProxyVersion,
		ShutdownManagerVersion: shutdownManagerVersion,
		GitCommitID:            gitCommitID,
	}
}

var (
	envoyGatewayVersion    string
	gatewayAPIVersion      string
	envoyProxyVersion      = strings.Split(*v1alpha1.DefaultKubernetesContainerImage(v1alpha1.DefaultEnvoyProxyImage), ":")[1]
	shutdownManagerVersion string
	gitCommitID            string
)

func init() {
	bi, ok := debug.ReadBuildInfo()
	if ok {
		for _, dep := range bi.Deps {
			if dep.Path == "sigs.k8s.io/gateway-api" {
				gatewayAPIVersion = dep.Version
			}
		}
	}
}

// Print shows the versions of the Envoy Gateway.
func Print(w io.Writer, format string) error {
	v := Get()
	switch format {
	case "json":
		if marshalled, err := json.MarshalIndent(v, "", "  "); err == nil {
			_, _ = fmt.Fprintln(w, string(marshalled))
		}
	case "yaml":
		if marshalled, err := yaml.Marshal(v); err == nil {
			_, _ = fmt.Fprintln(w, string(marshalled))
		}
	default:
		_, _ = fmt.Fprintf(w, "ENVOY_GATEWAY_VERSION: %s\n", v.EnvoyGatewayVersion)
		_, _ = fmt.Fprintf(w, "ENVOY_PROXY_VERSION: %s\n", v.EnvoyProxyVersion)
		_, _ = fmt.Fprintf(w, "GATEWAYAPI_VERSION: %s\n", v.GatewayAPIVersion)
		_, _ = fmt.Fprintf(w, "SHUTDOWN_MANAGER_VERSION: %s\n", v.ShutdownManagerVersion)
		_, _ = fmt.Fprintf(w, "GIT_COMMIT_ID: %s\n", v.GitCommitID)
	}

	return nil
}
