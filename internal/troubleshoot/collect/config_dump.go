// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package collect

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path"

	kube "github.com/envoyproxy/gateway/internal/kubernetes"
	troubleshootv1b2 "github.com/replicatedhq/troubleshoot/pkg/apis/troubleshoot/v1beta2"
	tbcollect "github.com/replicatedhq/troubleshoot/pkg/collect"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var _ tbcollect.Collector = &ConfigDump{}

// ConfigDump defines a collector that dumps the envoy configuration of the proxy pod.
type ConfigDump struct {
	BundlePath   string
	Namespace    string
	ClientConfig *rest.Config
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

func (cd ConfigDump) Collect(_ chan<- interface{}) (tbcollect.CollectorResult, error) {
	client, err := kubernetes.NewForConfig(cd.ClientConfig)
	if err != nil {
		return nil, err
	}

	pods, err := listPods(context.TODO(), client, cd.Namespace, labels.SelectorFromSet(map[string]string{
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

	for _, pod := range pods {
		nn := types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name}
		data, err := configDump(cliClient, nn, true)
		if err != nil {
			continue
		}

		k := fmt.Sprintf("%s-%s.json", pod.Namespace, pod.Name)
		_ = output.SaveResult(cd.BundlePath, path.Join("config-dumps", k), bytes.NewBuffer(data))
	}

	return output, nil
}

func configDump(cli kube.CLIClient, nn types.NamespacedName, includeEds bool) ([]byte, error) {
	fw, err := kube.NewLocalPortForwarder(cli, nn, 0, 19000)
	if err != nil {
		return nil, err
	}

	if err := fw.Start(); err != nil {
		return nil, err
	}
	defer fw.Stop()

	url := fmt.Sprintf("http://%s/config_dump", fw.Address())
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
