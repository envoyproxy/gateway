package translator

import (
	"errors"
	"fmt"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/tetratelabs/multierror"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

// TranslateXdsIR translates the XDS IR into xDS resources
func TranslateXdsIR(ir *ir.Xds) (*types.ResourceVersionTable, error) {
	if ir == nil {
		return nil, errors.New("ir is nil")
	}

	tCtx := new(types.ResourceVersionTable)

	for _, httpListener := range ir.HTTP {
		// 1:1 between IR HTTPListener and xDS Listener
		xdsListener, err := buildXdsListener(httpListener)
		if err != nil {
			return nil, multierror.Append(err, errors.New("error building xds listener"))
		}

		// 1:1 between IR TLSListenerConfig and xDS Secret
		if httpListener.TLS != nil {
			// Build downstream TLS details.
			tSocket, err := buildXdsDownstreamTLSSocket(httpListener.Name, httpListener.TLS)
			if err != nil {
				return nil, multierror.Append(err, errors.New("error building xds listener tls socket"))
			}
			xdsListener.FilterChains[0].TransportSocket = tSocket

			secret, err := buildXdsDownstreamTLSSecret(httpListener.Name, httpListener.TLS)
			if err != nil {
				return nil, multierror.Append(err, errors.New("error building xds listener tls secret"))
			}
			tCtx.AddXdsResource(resource.SecretType, secret)
		}

		// Allocate virtual host for this httpListener.
		// 1:1 between IR HTTPListener and xDS VirtualHost
		routeName := getXdsRouteName(httpListener.Name)
		vHost := &route.VirtualHost{
			Name:    routeName,
			Domains: httpListener.Hostnames,
		}

		for _, httpRoute := range httpListener.Routes {
			// 1:1 between IR HTTPRoute and xDS config.route.v3.Route
			xdsRoute, err := buildXdsRoute(httpRoute)
			if err != nil {
				return nil, multierror.Append(err, errors.New("error building xds route"))
			}
			vHost.Routes = append(vHost.Routes, xdsRoute)

			// 1:1 between IR HTTPRoute and xDS Cluster
			xdsCluster, err := buildXdsCluster(httpRoute)
			if err != nil {
				return nil, multierror.Append(err, errors.New("error building xds cluster"))
			}
			tCtx.AddXdsResource(resource.ClusterType, xdsCluster)

		}

		xdsRouteCfg := &route.RouteConfiguration{
			Name: routeName,
		}
		xdsRouteCfg.VirtualHosts = append(xdsRouteCfg.VirtualHosts, vHost)

		tCtx.AddXdsResource(resource.ListenerType, xdsListener)
		tCtx.AddXdsResource(resource.RouteType, xdsRouteCfg)
	}

	return tCtx, nil
}

func getXdsRouteName(listenerName string) string {
	return fmt.Sprintf("route_%s", listenerName)
}

func getXdsListenerName(listenerName string, listenerPort uint32) string {
	return fmt.Sprintf("listener_%s_%d", listenerName, listenerPort)
}

func getXdsSecretName(listenerName string) string {
	return fmt.Sprintf("secret_%s", listenerName)
}

func getXdsClusterName(routeName string) string {
	return fmt.Sprintf("cluster_%s", routeName)
}

// Point to xds cluster.
func makeConfigSource() *core.ConfigSource {
	source := &core.ConfigSource{}
	source.ResourceApiVersion = resource.DefaultAPIVersion
	source.ConfigSourceSpecifier = &core.ConfigSource_ApiConfigSource{
		ApiConfigSource: &core.ApiConfigSource{
			TransportApiVersion:       resource.DefaultAPIVersion,
			ApiType:                   core.ApiConfigSource_GRPC,
			SetNodeOnFirstMessageOnly: true,
			GrpcServices: []*core.GrpcService{{
				TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &core.GrpcService_EnvoyGrpc{ClusterName: "xds_cluster"},
				},
			}},
		},
	}
	return source
}
