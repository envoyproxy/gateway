package version

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/envoyproxy/gateway/internal/ir"
)

var (
	EnvoyGatewayVersion string
	GatewayAPIVersion   string
	EnvoyVersion        = strings.Split(ir.DefaultProxyImage, ":")[1]
	GitCommitID         string
)

func init() {
	bi, ok := debug.ReadBuildInfo()
	if ok {
		for _, dep := range bi.Deps {
			if dep.Path == "sigs.k8s.io/gateway-api" {
				GatewayAPIVersion = dep.Version
			}
		}
	}
}

// Print shows the versions of the Envoy Gateway.
func Print() error {
	fmt.Printf("ENVOY_GATEWAY_VERSION: %s\n", EnvoyGatewayVersion)
	fmt.Printf("ENVOY_VERSION: %s\n", EnvoyVersion)
	fmt.Printf("GATEWAYAPI_VERSION: %s\n", GatewayAPIVersion)
	fmt.Printf("GIT_COMMIT_ID: %s\n", GitCommitID)

	return nil
}
