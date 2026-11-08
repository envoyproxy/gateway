// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	routerv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/stretchr/testify/require"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

type authenticationCheckServer struct {
	authv3.UnimplementedAuthorizationServer
	calls  atomic.Int32
	failed atomic.Bool
}

func (server *authenticationCheckServer) Check(_ context.Context, check *authv3.CheckRequest) (*authv3.CheckResponse, error) {
	server.calls.Add(1)
	metadata := check.GetAttributes().GetMetadataContext().GetFilterMetadata()
	basic := metadata["envoy.filters.http.basic_auth/test"].GetFields()["username"].GetStringValue()
	payload := metadata["envoy.filters.http.jwt_authn"].GetFields()["issuer"].GetStructValue()
	if basic == "" && payload == nil {
		server.failed.Store(true)
	}
	if strings.Contains(check.GetAttributes().GetRequest().GetHttp().GetPath(), "external-mutation") {
		return &authv3.CheckResponse{Status: &statuspb.Status{}, HttpResponse: &authv3.CheckResponse_OkResponse{OkResponse: &authv3.OkHttpResponse{Headers: []*corev3.HeaderValueOption{{Header: &corev3.HeaderValue{Key: ":path", Value: "/public"}}}}}}, nil
	}
	if strings.Contains(check.GetAttributes().GetRequest().GetHttp().GetPath(), "external-deny") {
		return &authv3.CheckResponse{Status: &statuspb.Status{Code: 7}, HttpResponse: &authv3.CheckResponse_DeniedResponse{DeniedResponse: &authv3.DeniedHttpResponse{Status: &typev3.HttpStatus{Code: typev3.StatusCode_Forbidden}}}}, nil
	}
	return &authv3.CheckResponse{Status: &statuspb.Status{}, HttpResponse: &authv3.CheckResponse_OkResponse{OkResponse: &authv3.OkHttpResponse{}}}, nil
}

func TestAuthenticationProtocol(t *testing.T) {
	binary := os.Getenv("ENVOY_BINARY")
	if binary == "" {
		t.Skip("set ENVOY_BINARY to run HTTP and gRPC protocol tests against Envoy")
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	encode := base64.RawURLEncoding.EncodeToString
	jwks, _ := json.Marshal(map[string]any{"keys": []any{map[string]any{"kty": "RSA", "kid": "test", "alg": "RS256", "n": encode(key.N.Bytes()), "e": "AQAB"}}})
	sign := func(audience, scope string, expiry time.Time) string {
		header, _ := json.Marshal(map[string]any{"alg": "RS256", "kid": "test", "typ": "at+JWT"})
		payload, _ := json.Marshal(map[string]any{"iss": "https://issuer.example", "aud": audience, "exp": expiry.Unix(), "scope": scope})
		message := encode(header) + "." + encode(payload)
		digest := sha256.Sum256([]byte(message))
		signature, signErr := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
		require.NoError(t, signErr)
		return "Bearer " + message + "." + encode(signature)
	}
	backend := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) { response.WriteHeader(http.StatusOK) }))
	t.Cleanup(backend.Close)
	grpcListener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	server := grpc.NewServer()
	checker := &authenticationCheckServer{}
	authv3.RegisterAuthorizationServer(server, checker)
	go func() { _ = server.Serve(grpcListener) }()
	t.Cleanup(server.Stop)
	// #nosec G101 -- Fixed htpasswd fixture for the isolated protocol test.
	passwordHash := "eJy+BAeECxwgQcszRS/2Dxm/WMw="
	basic := &ir.BasicAuth{Name: "test", Users: ir.PrivateBytes("alice:{SHA}" + passwordHash + "\nbob:{SHA}" + passwordHash + "\n"), AllowMissing: true, EmitDynamicMetadata: true}
	jwtConfiguration := &ir.JWT{Name: "test", AllowMissing: true, Providers: []ir.JWTProvider{{Name: "issuer", Issuer: "https://issuer.example", Audiences: []string{"test-api"}, LocalJWKS: new(string(jwks))}}}
	basicRule := &ir.AuthorizationRule{Name: "basic", Action: egv1a1.AuthorizationActionAllow, Principal: ir.Principal{BasicAuth: &egv1a1.BasicAuthPrincipal{Users: []string{"alice"}}, BasicAuthFilter: basicAuthFilterName(basic)}}
	jwtRule := &ir.AuthorizationRule{Name: "jwt", Action: egv1a1.AuthorizationActionAllow, Principal: ir.Principal{JWT: &egv1a1.JWTPrincipal{Provider: "issuer", Scopes: []egv1a1.JWTScope{"read"}}}}
	listener := &ir.HTTPListener{}
	for _, path := range []string{"any", "basic", "jwt", "external", "external-deny", "external-mutation", "other", "public"} {
		security := &ir.SecurityFeatures{AuthenticationAny: true, BasicAuth: basic, JWT: jwtConfiguration}
		switch path {
		case "basic":
			security.Authorization = &ir.Authorization{DefaultAction: egv1a1.AuthorizationActionDeny, Rules: []*ir.AuthorizationRule{basicRule}}
		case "jwt":
			security.Authorization = &ir.Authorization{DefaultAction: egv1a1.AuthorizationActionDeny, Rules: []*ir.AuthorizationRule{jwtRule}}
		case "external", "external-deny", "external-mutation":
			security.Authorization = &ir.Authorization{DefaultAction: egv1a1.AuthorizationActionDeny, Rules: []*ir.AuthorizationRule{basicRule, jwtRule}}
			security.ExtAuth = &ir.ExtAuth{Name: "test", GRPC: &ir.GRPCExtAuthService{Destination: ir.RouteDestination{Name: "auth"}}, AuthenticationMetadataNamespaces: []string{"envoy.filters.http.jwt_authn", basicAuthFilterName(basic)}}
		case "other":
			security.BasicAuth = basic.DeepCopy()
			security.BasicAuth.Name = "other"
			security.BasicAuth.Users = ir.PrivateBytes("bob:{SHA}" + passwordHash + "\n")
		case "public":
			security = nil
		}
		listener.Routes = append(listener.Routes, &ir.HTTPRoute{Name: path, Security: security})
	}
	manager := &hcmv3.HttpConnectionManager{StatPrefix: "test", RouteSpecifier: &hcmv3.HttpConnectionManager_RouteConfig{RouteConfig: &routev3.RouteConfiguration{Name: "test"}}}
	filters := []httpFilter{&basicAuth{}, &jwt{}, &authenticationGate{}, &rbac{}, &extAuth{}}
	for _, filter := range filters {
		require.NoError(t, filter.patchHCM(manager, listener))
	}
	routerConfiguration, err := anypb.New(&routerv3.Router{})
	require.NoError(t, err)
	manager.HttpFilters = append(manager.HttpFilters, &hcmv3.HttpFilter{Name: egv1a1.EnvoyFilterRouter.String(), ConfigType: &hcmv3.HttpFilter_TypedConfig{TypedConfig: routerConfiguration}})
	manager.HttpFilters = orderAuthenticationGate(sortHTTPFilters(manager.HttpFilters, []egv1a1.FilterPosition{{Name: egv1a1.EnvoyFilterExtAuthz, After: new(egv1a1.EnvoyFilterJWTAuthn)}}))
	require.NoError(t, validateAuthenticationFilterOrder(manager, listener))
	virtual := &routev3.VirtualHost{Name: "test", Domains: []string{"*"}}
	for _, input := range listener.Routes {
		route := &routev3.Route{Name: input.Name, Match: &routev3.RouteMatch{PathSpecifier: &routev3.RouteMatch_Path{Path: "/" + input.Name}}, Action: &routev3.Route_Route{Route: &routev3.RouteAction{ClusterSpecifier: &routev3.RouteAction_Cluster{Cluster: "backend"}}}}
		for _, filter := range filters {
			require.NoError(t, filter.patchRoute(route, input, listener))
		}
		virtual.Routes = append(virtual.Routes, route)
	}
	manager.GetRouteConfig().VirtualHosts = []*routev3.VirtualHost{virtual}
	cluster := func(name, address string, http2 bool) map[string]any {
		host, port, splitErr := net.SplitHostPort(address)
		require.NoError(t, splitErr)
		number, parseErr := strconv.Atoi(port)
		require.NoError(t, parseErr)
		result := map[string]any{"name": name, "connect_timeout": "1s", "type": "STATIC", "load_assignment": map[string]any{"cluster_name": name, "endpoints": []any{map[string]any{"lb_endpoints": []any{map[string]any{"endpoint": map[string]any{"address": map[string]any{"socket_address": map[string]any{"address": host, "port_value": number}}}}}}}}}
		if http2 {
			result["http2_protocol_options"] = map[string]any{}
		}
		return result
	}
	portListener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := portListener.Addr().(*net.TCPAddr).Port
	require.NoError(t, portListener.Close())
	managerJSON, err := protojson.Marshal(manager)
	require.NoError(t, err)
	configuration := map[string]any{}
	require.NoError(t, json.Unmarshal(managerJSON, &configuration))
	configuration["@type"] = "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager"
	bootstrap := map[string]any{"static_resources": map[string]any{"listeners": []any{map[string]any{"name": "test", "address": map[string]any{"socket_address": map[string]any{"address": "127.0.0.1", "port_value": port}}, "filter_chains": []any{map[string]any{"filters": []any{map[string]any{"name": "envoy.filters.network.http_connection_manager", "typed_config": configuration}}}}}}, "clusters": []any{cluster("backend", strings.TrimPrefix(backend.URL, "http://"), false), cluster("auth", grpcListener.Addr().String(), true)}}}
	directory := t.TempDir()
	data, err := json.Marshal(bootstrap)
	require.NoError(t, err)
	configPath := filepath.Join(directory, "envoy.json")
	require.NoError(t, os.WriteFile(configPath, data, 0o600))
	logFile, err := os.Create(filepath.Join(directory, "envoy.log"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = logFile.Close() })
	// #nosec G702 -- The test runner selects the Envoy executable.
	process := exec.CommandContext(t.Context(), binary, "-c", configPath, "--concurrency", "1", "--log-level", "error", "--disable-hot-restart")
	process.Stdout = logFile
	process.Stderr = logFile
	require.NoError(t, process.Start())
	t.Cleanup(func() {
		_ = process.Process.Kill()
		_ = process.Wait()
		if t.Failed() {
			logs, _ := os.ReadFile(logFile.Name())
			t.Log(string(logs))
		}
	})
	address := fmt.Sprintf("http://127.0.0.1:%d", port)
	client := &http.Client{Timeout: time.Second}
	require.Eventually(t, func() bool {
		response, requestErr := client.Get(address + "/public")
		if requestErr != nil {
			return false
		}
		_ = response.Body.Close()
		return response.StatusCode == 200
	}, 10*time.Second, 20*time.Millisecond)
	basicHeader := func(user string) string {
		return "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":test-password"))
	}
	validJWT := sign("test-api", "read", time.Now().Add(time.Hour))
	for _, test := range []struct {
		name, path, credential string
		status                 int
	}{
		{"valid Basic with invalid JWT", "any?access_token=invalid", basicHeader("alice"), 401},
		{"valid Basic and JWT", "any?access_token=" + strings.TrimPrefix(validJWT, "Bearer "), basicHeader("alice"), 200},
		{"basic success", "any", basicHeader("alice"), 200},
		{"JWT without sub", "any", validJWT, 200},
		{"missing", "any", "", 403},
		{"invalid basic", "any", basicHeader("nobody"), 401},
		{"malformed basic", "any", "Basic !!!", 401},
		{"invalid JWT", "any", "Bearer invalid", 401},
		{"wrong audience", "any", sign("other", "read", time.Now().Add(time.Hour)), 403},
		{"expired", "any", sign("test-api", "read", time.Now().Add(-time.Hour)), 401},
		{"basic principal allows", "basic", basicHeader("alice"), 200},
		{"basic principal denies", "basic", basicHeader("bob"), 403},
		{"JWT cannot impersonate basic", "basic", validJWT, 403},
		{"JWT principal allows", "jwt", validJWT, 200},
		{"JWT scope denies", "jwt", sign("test-api", "other", time.Now().Add(time.Hour)), 403},
		{"basic cannot satisfy JWT", "jwt", basicHeader("alice"), 403},
		{"external mutation denied", "external-mutation", basicHeader("alice"), 500},
		{"external missing", "external", "", 403},
		{"external invalid", "external", "Bearer invalid", 401},
		{"external basic", "external", basicHeader("alice"), 200},
		{"external JWT", "external", validJWT, 200},
		{"local denies after external allow", "external", basicHeader("bob"), 403},
		{"external denies basic", "external-deny", basicHeader("alice"), 403},
		{"external denies JWT", "external-deny", validJWT, 403},
		{"route credentials isolated", "other", basicHeader("alice"), 401},
		{"public unaffected", "public", "", 200},
	} {
		t.Run(test.name, func(t *testing.T) {
			request, requestErr := http.NewRequestWithContext(t.Context(), "GET", address+"/"+test.path, nil)
			require.NoError(t, requestErr)
			if test.credential != "" {
				request.Header.Set("Authorization", test.credential)
			}
			request.Header.Set("x-authenticated-user", "alice")
			response, requestErr := client.Do(request)
			require.NoError(t, requestErr)
			defer response.Body.Close()
			body, readErr := io.ReadAll(response.Body)
			require.NoError(t, readErr)
			require.Equal(t, test.status, response.StatusCode, string(body))
		})
	}
	require.False(t, checker.failed.Load(), "external authorization received an unverified identity")
	require.GreaterOrEqual(t, checker.calls.Load(), int32(5))
	server.Stop()
	request, err := http.NewRequestWithContext(t.Context(), "GET", address+"/external", nil)
	require.NoError(t, err)
	request.Header.Set("Authorization", validJWT)
	response, err := client.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	require.NotEqual(t, 200, response.StatusCode)
}
