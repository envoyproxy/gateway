// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	tbcollect "github.com/replicatedhq/troubleshoot/pkg/collect"
	"github.com/replicatedhq/troubleshoot/pkg/convert"
	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"github.com/envoyproxy/gateway/internal/cmd/options"
	"github.com/envoyproxy/gateway/internal/troubleshoot/collect"
)

type collectOptions struct {
	outPath               string
	envoyGatewayNamespace string
}

func newCollectCommand() *cobra.Command {
	collectOpts := &collectOptions{}
	collectCommand := &cobra.Command{
		Use:   "collect",
		Short: "Collect configurations from the cluster to help diagnose any issues offline",
		Example: `  # Collect configurations from current context.
  egctl experimental collect
	`,
		Run: func(c *cobra.Command, args []string) {
			cmdutil.CheckErr(runCollect(*collectOpts))
		},
	}

	flags := collectCommand.Flags()
	options.AddKubeConfigFlags(flags)

	collectCommand.PersistentFlags().StringVarP(&collectOpts.outPath, "output", "o", "",
		"Specify the output file path for collected data. If not specified, a timestamped file will be created in the current directory.")
	collectCommand.PersistentFlags().StringVarP(&collectOpts.envoyGatewayNamespace, "envoy-system-namespace", "", "envoy-gateway-system",
		"Specify the namespace where the Envoy Gateway controller is installed.")

	return collectCommand
}

func runCollect(collectOpts collectOptions) error {
	cc := options.DefaultConfigFlags.ToRawKubeConfigLoader()
	restConfig, err := cc.ClientConfig()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()
	go waitForSignal(ctx, cancel)

	progressChan := make(chan interface{})
	go func() {
		select {
		case <-ctx.Done():
			close(progressChan)
		case msg := <-progressChan:
			fmt.Printf("Collecting support bundle: %v\n", msg)
		}
	}()

	tmpDir, err := os.MkdirTemp("", "envoy-gateway-support-bundle")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tmpDir)

	basename := ""
	if collectOpts.outPath != "" {
		// use override output path
		overridePath, err := convert.ValidateOutputPath(collectOpts.outPath)
		if err != nil {
			return fmt.Errorf("override output file path: %w", err)
		}
		basename = strings.TrimSuffix(overridePath, ".tar.gz")
	} else {
		// use default output path
		basename = fmt.Sprintf("envoy-gateway-%s", time.Now().Format("2006-01-02T15_04_05"))
	}
	bundlePath := filepath.Join(tmpDir, strings.TrimSuffix(basename, ".tar.gz"))
	if err := os.MkdirAll(bundlePath, 0o777); err != nil {
		return fmt.Errorf("create bundle dir: %w", err)
	}

	var result tbcollect.CollectorResult
	collectors := []tbcollect.Collector{
		// Collect the custom resources from Gateway API and EG
		collect.CustomResource{
			ClientConfig: restConfig,
			BundlePath:   bundlePath,
			IncludeGroups: []string{
				"gateway.envoyproxy.io",
				"gateway.networking.k8s.io",
			},
		},
		// Collect resources from EnvoyGateway system namespace
		collect.EnvoyGatewayResource{
			ClientConfig: restConfig,
			BundlePath:   bundlePath,
			Namespace:    collectOpts.envoyGatewayNamespace,
		},
	}
	total := len(collectors)
	allCollectedData := make(map[string][]byte)
	for i, collector := range collectors {
		res, err := collector.Collect(progressChan)
		if err != nil {
			progressChan <- fmt.Errorf("failed to run collector: %s: %w", collector.Title(), err)
			progressChan <- tbcollect.CollectProgress{
				CurrentName:    collector.Title(),
				CurrentStatus:  "failed",
				CompletedCount: i + 1,
				TotalCount:     total,
			}
			continue
		}
		for k, v := range res {
			allCollectedData[k] = v
		}
	}
	result = allCollectedData

	filename := fmt.Sprintf("%s.tar.gz", basename)
	return result.ArchiveSupportBundle(bundlePath, filename)
}

func waitForSignal(c context.Context, cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-c.Done():
	case <-sigCh:
		cancel()
	}
}
