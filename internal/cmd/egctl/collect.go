// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"github.com/envoyproxy/gateway/internal/cmd/options"
	tb "github.com/envoyproxy/gateway/internal/troubleshoot"
	"github.com/envoyproxy/gateway/internal/utils/path"
)

type collectOptions struct {
	outPath               string
	envoyGatewayNamespace string
	proxyNamespaces       []string
	enableSDS             bool
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
			cmdutil.CheckErr(runCollect(c.OutOrStderr(), *collectOpts))
		},
	}

	flags := collectCommand.Flags()
	options.AddKubeConfigFlags(flags)

	collectCommand.PersistentFlags().StringVarP(&collectOpts.outPath, "output", "o", "",
		"Specify the output file path for collected data. If not specified, a timestamped file will be created in the current directory.")
	collectCommand.PersistentFlags().StringVarP(&collectOpts.envoyGatewayNamespace, "envoy-system-namespace", "", "envoy-gateway-system",
		"Specify the namespace where the Envoy Gateway controller is installed.")
	collectCommand.PersistentFlags().StringArrayVarP(&collectOpts.proxyNamespaces, "envoy-proxy-namespaces", "", []string{},
		"Specify the namespaces where Envoy proxies are running.")
	collectCommand.PersistentFlags().BoolVarP(&collectOpts.enableSDS, "sds", "", false,
		"Specify if the SDS configuration will retrieved or not, Default to false")

	return collectCommand
}

func runCollect(errWriter io.Writer, collectOpts collectOptions) error {
	cc := options.DefaultConfigFlags.ToRawKubeConfigLoader()
	restConfig, err := cc.ClientConfig()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()
	go waitForSignal(ctx, cancel)

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
		overridePath, err := path.ValidateOutputPath(collectOpts.outPath)
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

	collectedNamespaces := []string{collectOpts.envoyGatewayNamespace}
	proxyNamespaces := sets.New(collectOpts.proxyNamespaces...)
	if len(proxyNamespaces) > 0 {
		collectedNamespaces = append(collectedNamespaces, proxyNamespaces.UnsortedList()...)
	}

	result, err := tb.CollectResult(ctx, restConfig,
		tb.WithBundlePath(bundlePath),
		tb.WithCollectedNamespaces(collectedNamespaces),
		tb.WithSDS(collectOpts.enableSDS),
	)
	if err != nil {
		_, _ = fmt.Fprintln(errWriter, "warning: failed to collect all data:", err)
	}
	return result.ArchiveBundle(bundlePath, fmt.Sprintf("%s.tar.gz", basename))
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
