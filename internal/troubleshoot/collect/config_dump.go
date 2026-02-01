// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package collect

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"time"

	adminv3 "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	troubleshootv1b2 "github.com/replicatedhq/troubleshoot/pkg/apis/troubleshoot/v1beta2"
	tbcollect "github.com/replicatedhq/troubleshoot/pkg/collect"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	kube "github.com/envoyproxy/gateway/internal/kubernetes"
)

const (
	SecretsConfigDumpTypeURL = "type.googleapis.com/envoy.admin.v3.SecretsConfigDump"
	// DefaultConfigDumpTimeout is the default timeout for config dump collection
	DefaultConfigDumpTimeout = 30 * time.Second
)

var _ tbcollect.Collector = &ConfigDump{}

// ConfigDump defines a collector that dumps the envoy configuration of the proxy pod.
type ConfigDump struct {
	BundlePath   string
	Namespace    string
	ClientConfig *rest.Config

	EnableSDS bool
	// Timeout is the timeout for collecting config dumps. If not set, defaults to DefaultConfigDumpTimeout.
	Timeout time.Duration
}

func (cd ConfigDump) Title() string {
	return "config-dump"
}

func (cd ConfigDump) IsExcluded() (bool, error) {
	return false, nil
}

func (cd ConfigDump) GetRBACErrors() []error {
	return nil
}

func (cd ConfigDump) HasRBACErrors() bool {
	return false
}

func (cd ConfigDump) CheckRBAC(_ context.Context, _ tbcollect.Collector, _ *troubleshootv1b2.Collect, _ *rest.Config, _ string) error {
	return nil
}

// getTimeout returns the configured timeout, or the default if not set
func (cd ConfigDump) getTimeout() time.Duration {
	if cd.Timeout > 0 {
		return cd.Timeout
	}
	return DefaultConfigDumpTimeout
}

func (cd ConfigDump) Collect(_ chan<- interface{}) (tbcollect.CollectorResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cd.getTimeout())
	defer cancel()

	client, err := kubernetes.NewForConfig(cd.ClientConfig)
	if err != nil {
		return nil, err
	}

	pods, err := listPods(ctx, client, cd.Namespace, labels.SelectorFromSet(map[string]string{
		"app.kubernetes.io/component":  "proxy",
		"app.kubernetes.io/managed-by": "envoy-gateway",
		"app.kubernetes.io/name":       "envoy",
	}))
	if err != nil {
		return nil, err
	}

	output := tbcollect.NewResult()

	cliClient, err := kube.NewForRestConfig(cd.ClientConfig)
	if err != nil {
		return output, err
	}

	logs := make([]string, 0, len(pods))
	for i := range pods {
		pod := &pods[i]
		nn := types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name}
		data, err := configDump(cliClient, nn, true)
		if err != nil {
			logs = append(logs, fmt.Sprintf("failed to get config dump for pod %s/%s: %v", pod.Namespace, pod.Name, err))
			continue
		}

		if !cd.EnableSDS {
			data, err = trimSDS(data)
			if err != nil {
				logs = append(logs, fmt.Sprintf("failed to trim SDS for pod %s/%s: %v", pod.Namespace, pod.Name, err))
				continue
			}
		}

		_ = output.SaveResult(cd.BundlePath, path.Join("config-dumps", pod.Namespace, fmt.Sprintf("%s.json", pod.Name)), bytes.NewBuffer(data))
	}
	if len(logs) > 0 {
		_ = output.SaveResult(cd.BundlePath, path.Join("config-dumps", "errors.log"), marshalErrors(logs))
	}

	return output, nil
}

func configDump(cli kube.CLIClient, nn types.NamespacedName, includeEds bool) ([]byte, error) {
	reqPath := "/config_dump"
	if includeEds {
		reqPath = fmt.Sprintf("%s?include_eds", reqPath)
	}
	return RequestWithPortForwarder(cli, nn, 19000, reqPath)
}

var marshaller = protojson.MarshalOptions{
	Indent: "    ", // pretty output
}

func trimSDS(input []byte) ([]byte, error) {
	dump := &adminv3.ConfigDump{}
	if err := protojson.Unmarshal(input, dump); err != nil {
		return input, err
	}

	cfgs := make([]*anypb.Any, 0, len(dump.Configs))
	for _, section := range dump.Configs {
		if section.TypeUrl == SecretsConfigDumpTypeURL {
			continue
		}

		cfgs = append(cfgs, section)
	}

	dump.Configs = cfgs
	res, err := marshaller.Marshal(dump)
	if err != nil {
		return nil, err
	}

	return res, nil
}
