// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/envoyproxy/gateway/internal/cmd/egctl/migrate"
	"github.com/envoyproxy/gateway/internal/cmd/egctl/migrate/deprecated"
)

// migrateFlags holds the flags shared by every migration path.
type migrateFlags struct {
	files         []string
	fromCluster   bool
	namespace     string
	allNamespaces bool
	mode          string
}

func (f *migrateFlags) options() migrate.Options {
	ns := f.namespace
	if f.allNamespaces {
		ns = ""
	}
	return migrate.Options{
		Files:       f.files,
		FromCluster: f.fromCluster,
		Namespace:   ns,
		Mode:        migrate.Mode(f.mode),
	}
}

// newK8sClientFunc creates the client for --from-cluster. Tests replace it with a fake client.
var newK8sClientFunc = newK8sClient

func newMigrateCommand() *cobra.Command {
	flags := &migrateFlags{}
	modes := make([]string, 0, len(migrate.Modes))
	for _, m := range migrate.Modes {
		modes = append(modes, string(m))
	}

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate Envoy Gateway resources",
		Long: `Migrate Envoy Gateway resources, read from files or from the cluster.

The migrated YAML is printed to stdout (--mode print), written back to the files (--mode in-place),
or not written at all (--mode check-only). A report of every migrated field is written to stderr.

Exit codes:
  0  nothing to migrate, or everything was migrated
  1  --mode check-only found something to migrate
  2  at least one field can't be migrated automatically, see the report`,
	}

	pf := cmd.PersistentFlags()
	pf.StringSliceVarP(&flags.files, "file", "f", nil, "Files or directories (searched recursively) to migrate, or - for stdin.")
	pf.BoolVar(&flags.fromCluster, "from-cluster", false, "Read the resources from the cluster instead of files.")
	pf.StringVarP(&flags.namespace, "namespace", "n", "default", "Namespace to read resources from with --from-cluster.")
	pf.BoolVarP(&flags.allNamespaces, "all-namespaces", "A", false, "Read resources from all namespaces with --from-cluster.")
	pf.StringVar(&flags.mode, "mode", string(migrate.ModePrint), "One of "+strings.Join(modes, ", ")+".")

	cmd.AddCommand(newMigrateDeprecatedCommand(flags))
	return cmd
}

func newMigrateDeprecatedCommand(flags *migrateFlags) *cobra.Command {
	path := &deprecated.Path{}

	cmd := &cobra.Command{
		Use:   "deprecated",
		Short: "Migrate deprecated API fields to their replacements",
		Long: `Migrate the deprecated Envoy Gateway API fields, which are removed in v1.11, to their replacements.

ClientTrafficPolicy, BackendTrafficPolicy, SecurityPolicy, EnvoyExtensionPolicy and EnvoyProxy
resources are rewritten. Comments and field order are kept.

These are reported as errors and never rewritten:
  - Telemetry host/port that isn't the DNS name of a Service: set backendRefs to a Service or a Backend.
  - The EnvoyGateway configuration, in a ConfigMap, an EnvoyGateway file or the Helm values:
    update config.envoyGateway in the Helm values.
  - With --from-cluster, resources managed by Helm, Argo CD or Flux: fix the source instead.`,
		Example: `  # Print the migrated manifests
  egctl x migrate deprecated -f manifests/

  # Rewrite the manifests in place
  egctl x migrate deprecated -f manifests/ --mode in-place

  # Report what would be migrated, exit 1 if anything would be
  egctl x migrate deprecated -f manifests/ --mode check-only

  # Print the migrated resources from all namespaces of the cluster
  egctl x migrate deprecated --from-cluster -A`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts := flags.options()
			if err := opts.Validate(); err != nil {
				return err
			}
			var cli client.Client
			if opts.FromCluster {
				var err error
				if cli, err = newK8sClientFunc(); err != nil {
					return err
				}
			}
			err := migrate.Run(cmd.Context(), path, opts, cli, cmd.OutOrStdout(), cmd.ErrOrStderr())
			if err != nil {
				return fmt.Errorf("migrate deprecated: %w", err)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&path.EnvoyGatewayNamespace, "eg-namespace", "envoy-gateway-system",
		"Namespace of the envoy-gateway-config ConfigMap, used with --from-cluster.")
	return cmd
}
