// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	pb "github.com/envoyproxy/gateway/proto/extension"
)

func logOnce(t *testing.T, req *clusterv3.Cluster, resources ...string) map[string]any {
	t.Helper()
	r := clusterRequest(resources...)
	r.Cluster = req
	return logLine(t, true, "PostClusterModify", r)
}

func logLine(t *testing.T, logRequests bool, hook string, req proto.Message) map[string]any {
	t.Helper()
	var buf bytes.Buffer
	s := New(slog.New(slog.NewJSONHandler(&buf, nil)), nil, "extserver", false)
	s.logRequests = logRequests
	s.logRequest(hook, req)

	var line map[string]any
	if err := json.Unmarshal(buf.Bytes(), &line); err != nil {
		t.Fatalf("log line is not JSON: %v\n%s", err, buf.String())
	}
	return line
}

func TestLogRequestDecodesEnvoyExtensionTypes(t *testing.T) {
	// DnsCluster is one of the types that failed to resolve before extensions.gen.go was added.
	c := &clusterv3.Cluster{Name: "c", ClusterDiscoveryType: &clusterv3.Cluster_ClusterType{
		ClusterType: &clusterv3.Cluster_CustomClusterType{
			Name:        "envoy.cluster.dns",
			TypedConfig: &anypb.Any{TypeUrl: "type.googleapis.com/envoy.extensions.clusters.dns.v3.DnsCluster"},
		},
	}}
	line := logOnce(t, c, testPool)

	if _, ok := line["requestJSONError"]; ok {
		t.Fatalf("unexpected JSON error: %v", line["requestJSONError"])
	}
	if _, ok := line["request"].(map[string]any); !ok {
		t.Fatalf("expected request as nested JSON, got %T", line["request"])
	}
	res, ok := line["extensionResources"].([]any)
	if !ok || len(res) != 1 {
		t.Fatalf("expected one decoded extension resource, got %v", line["extensionResources"])
	}
	if kind := res[0].(map[string]any)["kind"]; kind != "InferencePool" {
		t.Fatalf("expected InferencePool, got %v", kind)
	}
}

func TestLogRequestFallsBackForUnknownTypes(t *testing.T) {
	c := &clusterv3.Cluster{Name: "c", ClusterDiscoveryType: &clusterv3.Cluster_ClusterType{
		ClusterType: &clusterv3.Cluster_CustomClusterType{
			Name:        "unknown",
			TypedConfig: &anypb.Any{TypeUrl: "type.googleapis.com/example.Unknown"},
		},
	}}
	line := logOnce(t, c)

	if _, ok := line["requestJSONError"]; !ok {
		t.Fatal("expected requestJSONError for an unknown type")
	}
	if req, ok := line["request"].(string); !ok || req == "" {
		t.Fatalf("expected the request in text format, got %v", line["request"])
	}
}

func TestLogRequestOmitsRequestByDefault(t *testing.T) {
	r := clusterRequest(testPool)
	line := logLine(t, false, "PostClusterModify", r)

	if _, ok := line["request"]; ok {
		t.Fatalf("expected no request without --log-requests, got %v", line["request"])
	}
	if name := line["name"]; name != r.Cluster.GetName() {
		t.Fatalf("expected cluster name %q, got %v", r.Cluster.GetName(), name)
	}
	res, ok := line["extensionResources"].([]any)
	if !ok || len(res) != 1 {
		t.Fatalf("expected one extension resource, got %v", line["extensionResources"])
	}
	if want := "inference.networking.k8s.io/v1/InferencePool/default/pool-a"; res[0] != want {
		t.Fatalf("expected only the resource identity %q, got %v", want, res[0])
	}
}

func TestLogRequestRedactsSecrets(t *testing.T) {
	const privateKey = "-----BEGIN PRIVATE KEY-----not-a-real-key"
	req := &pb.PostTranslateModifyRequest{Secrets: []*tlsv3.Secret{{
		Name: "default/tls",
		Type: &tlsv3.Secret_TlsCertificate{TlsCertificate: &tlsv3.TlsCertificate{
			PrivateKey: &corev3.DataSource{Specifier: &corev3.DataSource_InlineString{InlineString: privateKey}},
		}},
	}}}

	for _, logRequests := range []bool{false, true} {
		line := logLine(t, logRequests, "PostTranslateModify", req)
		b, err := json.Marshal(line)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(b, []byte("not-a-real-key")) {
			t.Fatalf("private key logged with logRequests=%v: %s", logRequests, b)
		}
		if logRequests && !bytes.Contains(b, []byte("default/tls")) {
			t.Fatalf("expected the secret name to be logged: %s", b)
		}
	}
	if req.Secrets[0].GetTlsCertificate() == nil {
		t.Fatal("logging modified the request")
	}
}
