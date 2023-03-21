// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"fmt"
	"io"
	"net/http"

	adminv3 "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/gateway/internal/cmd/options"
	kube "github.com/envoyproxy/gateway/internal/kubernetes"
)

var (
	output       string
	podName      string
	podNamespace string
)

const (
	adminPort     = 19000   // TODO: make this configurable until EG support
	containerName = "envoy" // TODO: make this configurable until EG support
)

func retrieveConfigDump(args []string, includeEds bool) (*adminv3.ConfigDump, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("pod name is required")
	}

	podName = args[0]

	if podName == "" {
		return nil, fmt.Errorf("pod name is required")
	}

	if podNamespace == "" {
		return nil, fmt.Errorf("pod namespace is required")
	}

	fw, err := portForwarder(types.NamespacedName{
		Namespace: podNamespace,
		Name:      podName,
	})
	if err != nil {
		return nil, err
	}
	if err := fw.Start(); err != nil {
		return nil, err
	}
	defer fw.Stop()

	configDump, err := extractConfigDump(fw, includeEds)
	if err != nil {
		return nil, err
	}

	return configDump, nil
}

func portForwarder(nn types.NamespacedName) (kube.PortForwarder, error) {
	c, err := kube.NewCLIClient(options.DefaultConfigFlags.ToRawKubeConfigLoader())
	if err != nil {
		return nil, fmt.Errorf("build CLI client fail: %w", err)
	}

	pod, err := c.Pod(nn)
	if err != nil {
		return nil, fmt.Errorf("get pod %s fail: %w", nn, err)
	}
	if pod.Status.Phase != "Running" {
		return nil, fmt.Errorf("pod %s is not running", nn)
	}

	fw, err := kube.NewLocalPortForwarder(c, nn, 0, adminPort)
	if err != nil {
		return nil, err
	}

	return fw, nil
}

func marshalEnvoyProxyConfig(configDump protoreflect.ProtoMessage, output string) ([]byte, error) {
	out, err := protojson.MarshalOptions{
		Multiline: true,
	}.Marshal(configDump)
	if err != nil {
		return nil, err
	}

	if output == "yaml" {
		out, err = yaml.JSONToYAML(out)
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

func extractConfigDump(fw kube.PortForwarder, includeEds bool) (*adminv3.ConfigDump, error) {
	out, err := configDumpRequest(fw.Address(), includeEds)
	if err != nil {
		return nil, err
	}

	configDump := &adminv3.ConfigDump{}
	if err := protojson.Unmarshal(out, configDump); err != nil {
		return nil, err
	}

	return configDump, nil
}

func configDumpRequest(address string, includeEds bool) ([]byte, error) {
	url := fmt.Sprintf("http://%s/config_dump", address)
	if includeEds {
		url = fmt.Sprintf("%s?include_eds", url)
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	return io.ReadAll(resp.Body)
}
