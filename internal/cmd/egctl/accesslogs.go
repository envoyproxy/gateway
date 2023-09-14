// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	alv3cfg "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	alv3ext "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/grpc/v3"
	alv3svc "github.com/envoyproxy/go-control-plane/envoy/service/accesslog/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/jsonpb"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	apisv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

const (
	// accessLogServerAddress is the listening address of the access-log server.
	accessLogServerAddress = "0.0.0.0"
	// accessLogServerPort is the default listening port of the access-log server.
	accessLogServerPort = 9001
	// accessLogServerClusterNamePrefix is the prefix of access-log cluster name.
	accessLogServerClusterNamePrefix = "egctl-access-logs"
)

func NewAccessLogsCommand() *cobra.Command {
	var (
		port      int
		listener  string
		namespace string
	)

	accessLogsCommand := &cobra.Command{
		Use:     "access-logs",
		Short:   "Streaming access logs from Envoy Proxy.",
		Example: ``, // TODO(sh2): add example
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("missing the name of gateway")
			}

			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			wg := &sync.WaitGroup{}
			gateway := args[0]
			if err := accessLogs(ctx, wg, port, namespace, gateway, listener); err != nil {
				return err
			}

			wg.Wait()

			// TODO(sh2): recycle the created resources

			return nil
		},
	}

	accessLogsCommand.PersistentFlags().IntVarP(&port, "port", "p", accessLogServerPort, "Listening port of access-log server.")
	accessLogsCommand.PersistentFlags().StringVarP(&listener, "listener", "l", "", "Listener name of the gateway.")
	accessLogsCommand.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "Namespace of the gateway.")

	return accessLogsCommand
}

type accessLogServer struct {
	marshaler jsonpb.Marshaler
}

var _ alv3svc.AccessLogServiceServer = &accessLogServer{}

func newAccessLogServer() alv3svc.AccessLogServiceServer {
	return &accessLogServer{}
}

func (a *accessLogServer) StreamAccessLogs(server alv3svc.AccessLogService_StreamAccessLogsServer) error {
	log.Println("Start streaming access logs from server")
	for {
		recv, err := server.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		str, _ := a.marshaler.MarshalToString(recv)
		log.Println(str) // TODO(sh2): prettify the json output
	}
}

func serveAccessLogServer(ctx context.Context, wg *sync.WaitGroup, server *grpc.Server, addr string) {
	defer wg.Done()

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("failed to listen on address %s: %v\n", addr, err)
		return
	}

	go func() {
		<-ctx.Done()
		log.Println("Shutting down access-log server")
		server.Stop()
	}()

	log.Printf("Serving access-log server on %s\n", addr)
	if err = server.Serve(listener); err != nil {
		log.Printf("failed to start access-log server: %v\n", err)
	}
}

func isValidPort(port int) bool {
	if port < 1024 || port > 65535 {
		return false
	}
	return true
}

func validateAccessLogsInputs(port int) error {
	if !isValidPort(port) {
		return fmt.Errorf("%d is not a valid port", port)
	}

	return nil
}

func accessLogs(ctx context.Context, wg *sync.WaitGroup, port int, namespace, gateway, listener string) error {
	if err := validateAccessLogsInputs(port); err != nil {
		return err
	}

	// TODO(sh2): check whether the envoy patch policy is enabled, return err if not

	// TODO(sh2): check whether the gateway is exist, and its listener

	grpcServer := grpc.NewServer()
	alv3svc.RegisterAccessLogServiceServer(grpcServer, newAccessLogServer())

	addr := net.JoinHostPort(accessLogServerAddress, strconv.Itoa(port))
	wg.Add(1)
	go serveAccessLogServer(ctx, wg, grpcServer, addr)

	return nil
}

func expectedAccessLogsClusterName() string {
	// The cluster name is of the form <prefix>-<hash>, where hash is generated based on unix time.
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%d", time.Now().Unix())))
	hash := fmt.Sprintf("%x", h.Sum(nil))[:8]
	return fmt.Sprintf("%s-%s", accessLogServerClusterNamePrefix, hash)
}

func expectedJsonPatchListenerName(namespace, gateway, listener string) string {
	// The listener name is of the form <GatewayNamespace>/<GatewayName>/<GatewayListenerName>.
	return fmt.Sprintf("%s/%s/%s", namespace, gateway, listener)
}

func expectedJsonPatchAccessLogHttpGrpcConfig(clusterName string) (string, error) {
	var buf bytes.Buffer

	httpGrpcAccessLogConfig := &alv3ext.HttpGrpcAccessLogConfig{
		CommonConfig: &alv3ext.CommonGrpcAccessLogConfig{
			LogName:             clusterName,
			TransportApiVersion: corev3.ApiVersion_V3,
			GrpcService: &corev3.GrpcService{
				TargetSpecifier: &corev3.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &corev3.GrpcService_EnvoyGrpc{
						ClusterName: clusterName,
					},
				},
			},
		},
	}

	httpGrpcAccessLogConfigAny, err := anypb.New(httpGrpcAccessLogConfig)
	if err != nil {
		return "", err
	}

	accessLog := &alv3cfg.AccessLog{
		Name: wellknown.HTTPGRPCAccessLog,
		ConfigType: &alv3cfg.AccessLog_TypedConfig{
			TypedConfig: httpGrpcAccessLogConfigAny,
		},
	}

	m := jsonpb.Marshaler{OrigName: true}
	w := io.Writer(&buf)
	if err = m.Marshal(w, accessLog); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func expectedJsonPatchAccessLogClusterConfig(address, clusterName string, port uint32) (string, error) {
	var buf bytes.Buffer

	endpoint := &endpointv3.LbEndpoint_Endpoint{
		Endpoint: &endpointv3.Endpoint{
			Address: &corev3.Address{
				Address: &corev3.Address_SocketAddress{
					SocketAddress: &corev3.SocketAddress{
						Address:       address,
						PortSpecifier: &corev3.SocketAddress_PortValue{PortValue: port},
					},
				},
			},
		},
	}

	cluster := &clusterv3.Cluster{
		Name:           clusterName,
		ConnectTimeout: durationpb.New(10 * time.Second),
		LoadAssignment: &endpointv3.ClusterLoadAssignment{
			ClusterName: clusterName,
			Endpoints: []*endpointv3.LocalityLbEndpoints{
				{
					LbEndpoints: []*endpointv3.LbEndpoint{
						{HostIdentifier: endpoint},
					},
				},
			},
		},
	}

	m := jsonpb.Marshaler{OrigName: true}
	w := io.Writer(&buf)
	if err := m.Marshal(w, cluster); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func expectedEnvoyPatchPolicy(namespace, gateway, listener, address, clusterName string, port uint32) (*egv1a1.EnvoyPatchPolicy, error) {
	ns := gwv1a2.Namespace(namespace)

	accessLogHttpGrpcConfig, err := expectedJsonPatchAccessLogHttpGrpcConfig(clusterName)
	if err != nil {
		return nil, err
	}

	accessLogClusterConfig, err := expectedJsonPatchAccessLogClusterConfig(address, clusterName, port)
	if err != nil {
		return nil, err
	}

	policy := &egv1a1.EnvoyPatchPolicy{
		TypeMeta: metav1.TypeMeta{
			Kind:       egv1a1.KindEnvoyPatchPolicy,
			APIVersion: egv1a1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      clusterName,
		},
		Spec: egv1a1.EnvoyPatchPolicySpec{
			Type: egv1a1.JSONPatchEnvoyPatchType,
			TargetRef: gwv1a2.PolicyTargetReference{
				Kind:      "Gateway",
				Group:     gwv1b1.GroupName,
				Name:      gwv1a2.ObjectName(gateway),
				Namespace: &ns,
			},
			JSONPatches: []egv1a1.EnvoyJSONPatchConfig{
				{
					Type: egv1a1.ListenerEnvoyResourceType,
					Name: expectedJsonPatchListenerName(namespace, gateway, listener),
					Operation: egv1a1.JSONPatchOperation{
						Op:   "add",
						Path: "/default_filter_chain/filters/0/typed_config/access_log/0",
						Value: apisv1.JSON{
							Raw: []byte(accessLogHttpGrpcConfig),
						},
					},
				},
				{
					Type: egv1a1.ClusterEnvoyResourceType,
					Name: clusterName,
					Operation: egv1a1.JSONPatchOperation{
						Op:   "add",
						Path: "",
						Value: apisv1.JSON{
							Raw: []byte(accessLogClusterConfig),
						},
					},
				},
			},
		},
	}

	return policy, nil
}
