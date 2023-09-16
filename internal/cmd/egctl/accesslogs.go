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
	"os"
	"os/signal"
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
	kcorev1 "k8s.io/api/core/v1"
	kapisv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
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
		port      uint32
		listener  string
		namespace string
	)

	accessLogsCommand := &cobra.Command{
		Use:   "access-logs",
		Short: "Streaming access logs from Envoy Proxy.",
		Example: `  # Start streaming access logs from the Envoy Proxy of one gateway.
  egctl experimental access-logs <gateway-name>

  # Start streaming access logs from the Envoy Proxy of one gateway in specific namespace.
  egctl experimental access-logs <gateway-name> -n <namespace>

  # Start streaming access logs from the Envoy Proxy of one gateway with specific listener.
  egctl experimental access-logs <gateway-name> -l <listener-name>

  # Start access-log server with specific port.
  egctl experimental access-logs <gateway-name> -p <port>
    `,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("missing the name of gateway")
			}
			return accessLogs(namespace, args[0], listener, port)
		},
	}

	accessLogsCommand.PersistentFlags().Uint32VarP(&port, "port", "p", accessLogServerPort, "Listening port of access-log server.")
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

func newScheme() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := gwv1b1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := gwv1a2.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := egv1a1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	return scheme, nil
}

func newClient() (client.Client, error) {
	scheme, err := newScheme()
	if err != nil {
		return nil, err
	}

	cli, err := client.New(config.GetConfigOrDie(), client.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}

	return cli, nil
}

func (a *accessLogServer) StreamAccessLogs(server alv3svc.AccessLogService_StreamAccessLogsServer) error {
	for {
		recv, err := server.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		nodeID := recv.GetIdentifier().GetNode().GetId()
		entry := recv.GetHttpLogs().GetLogEntry()
		n := len(entry)
		log.Printf("Streaming %d access log entry from %s", n, nodeID)
		for i := 0; i < n; i++ {
			str, _ := a.marshaler.MarshalToString(entry[i])
			log.Println(str)
		}
	}
}

func serveAccessLogServer(ctx context.Context, wg *sync.WaitGroup, server *grpc.Server, addr string) {
	defer wg.Done()

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("failed to listen on address %s: %v", addr, err)
		return
	}

	go func() {
		<-ctx.Done()
		log.Println("Shutting down access-log server")
		server.Stop()
	}()

	log.Printf("Serving access-log server on %s", addr)
	if err = server.Serve(listener); err != nil {
		log.Printf("failed to start access-log server: %v", err)
	}
}

// isValidGateway check the status of given gateway and return its first listener name.
func isValidGateway(ctx context.Context, cli client.Client, namespace, gateway string) (string, error) {
	gw := &gwv1b1.Gateway{}
	nn := types.NamespacedName{
		Namespace: namespace,
		Name:      gateway,
	}
	if err := cli.Get(ctx, nn, gw); err != nil {
		return "", fmt.Errorf("gateway '%s' not found", nn.String())
	}
	if len(gw.Spec.Listeners) == 0 {
		return "", fmt.Errorf("gateway '%s' has no listeners", nn.String())
	}

	return string(gw.Spec.Listeners[0].Name), nil
}

func isValidPort(port uint32) bool {
	if port < 1024 || port > 65535 {
		return false
	}
	return true
}

func validateAccessLogsInputs(port uint32) error {
	if !isValidPort(port) {
		return fmt.Errorf("%d is not a valid port", port)
	}

	return nil
}

func accessLogs(namespace, gateway, listener string, port uint32) error {
	if err := validateAccessLogsInputs(port); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cli, err := newClient()
	if err != nil {
		return fmt.Errorf("failed to create kube client: %v", err)
	}

	// TODO(sh2): Checking the status of EnvoyPatchPolicy.
	log.Println("NOTE: PLEASE MAKE SURE THE ENVOY PATCH POLICY IS ENABLED, " +
		"YOU CAN REFER TO https://gateway.envoyproxy.io/latest/user/envoy-patch-policy.html FOR MORE INFO.")

	// Checking the status of gateway and its listeners.
	defaultListener, err := isValidGateway(ctx, cli, namespace, gateway)
	if err != nil {
		return err
	}
	if len(listener) == 0 {
		listener = defaultListener
		log.Printf("Using '%s' listener in gateway '%s' as default", listener, gateway)
	}

	// Preparing EnvoyPatchPolicy.
	clusterName := expectedAccessLogsClusterName()
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	address, err := getNodeIP(ctx, cli, hostname)
	if err != nil {
		return err
	}

	policy, err := expectedEnvoyPatchPolicy(namespace, gateway, listener, address, clusterName, port)
	if err != nil {
		return err
	}

	log.Printf("Creating %s/envoy-patch-policy: %s", namespace, clusterName)
	if err = createOrUpdateEnvoyPatchPolicy(ctx, cli, policy); err != nil {
		return err
	}

	wg := &sync.WaitGroup{}
	if err = runAccessLogServer(ctx, wg, port); err != nil {
		return err
	}

	wg.Wait()
	log.Printf("Deleting %s/envoy-patch-policy: %s", namespace, clusterName)
	// Using new context instead of old context (may get canceled).
	if err = deleteEnvoyPatchPolicy(context.Background(), cli, policy); err != nil {
		return err
	}

	return nil
}

func runAccessLogServer(ctx context.Context, wg *sync.WaitGroup, port uint32) error {
	grpcServer := grpc.NewServer()
	alv3svc.RegisterAccessLogServiceServer(grpcServer, newAccessLogServer())

	addr := net.JoinHostPort(accessLogServerAddress, fmt.Sprintf("%d", port))
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

func expectedJSONPatchListenerName(namespace, gateway, listener string) string {
	// The listener name is of the form <GatewayNamespace>/<GatewayName>/<GatewayListenerName>.
	return fmt.Sprintf("%s/%s/%s", namespace, gateway, listener)
}

func expectedJSONPatchAccessLogHTTPGrpcConfig(clusterName string) (string, error) {
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

func expectedJSONPatchAccessLogClusterConfig(address, clusterName string, port uint32) (string, error) {
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
		Name:                 clusterName,
		ConnectTimeout:       durationpb.New(10 * time.Second),
		Http2ProtocolOptions: &corev3.Http2ProtocolOptions{},
		ClusterDiscoveryType: &clusterv3.Cluster_Type{
			Type: clusterv3.Cluster_STRICT_DNS,
		},
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

	accessLogHTTPGrpcConfig, err := expectedJSONPatchAccessLogHTTPGrpcConfig(clusterName)
	if err != nil {
		return nil, err
	}

	accessLogClusterConfig, err := expectedJSONPatchAccessLogClusterConfig(address, clusterName, port)
	if err != nil {
		return nil, err
	}

	policy := &egv1a1.EnvoyPatchPolicy{
		TypeMeta: kmetav1.TypeMeta{
			Kind:       egv1a1.KindEnvoyPatchPolicy,
			APIVersion: egv1a1.GroupVersion.String(),
		},
		ObjectMeta: kmetav1.ObjectMeta{
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
					Name: expectedJSONPatchListenerName(namespace, gateway, listener),
					Operation: egv1a1.JSONPatchOperation{
						Op:   "add",
						Path: "/default_filter_chain/filters/0/typed_config/access_log/0",
						Value: kapisv1.JSON{
							Raw: []byte(accessLogHTTPGrpcConfig),
						},
					},
				},
				{
					Type: egv1a1.ClusterEnvoyResourceType,
					Name: clusterName,
					Operation: egv1a1.JSONPatchOperation{
						Op:   "add",
						Path: "",
						Value: kapisv1.JSON{
							Raw: []byte(accessLogClusterConfig),
						},
					},
				},
			},
		},
	}

	return policy, nil
}

func createOrUpdateEnvoyPatchPolicy(ctx context.Context, cli client.Client, policy *egv1a1.EnvoyPatchPolicy) error {
	key := types.NamespacedName{
		Namespace: policy.Namespace,
		Name:      policy.Name,
	}
	cur := &egv1a1.EnvoyPatchPolicy{}
	if err := cli.Get(ctx, key, cur); err != nil {
		if kerrors.IsNotFound(err) {
			if err = cli.Create(ctx, policy); err != nil {
				return err
			}
		}
	} else {
		if err = cli.Update(ctx, policy); err != nil {
			return err
		}
	}

	return nil
}

func deleteEnvoyPatchPolicy(ctx context.Context, cli client.Client, policy *egv1a1.EnvoyPatchPolicy) error {
	if err := cli.Delete(ctx, policy); err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return nil
}

func getNodeIP(ctx context.Context, cli client.Client, name string) (string, error) {
	key := types.NamespacedName{Name: name}
	node := &kcorev1.Node{}
	if err := cli.Get(ctx, key, node); err != nil {
		return "", err
	}

	var ip string
	for _, address := range node.Status.Addresses {
		if address.Type == kcorev1.NodeInternalIP {
			ip = address.Address
		}
	}

	return ip, nil
}
