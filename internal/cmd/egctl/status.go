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
	"reflect"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func newStatusCommand() *cobra.Command {
	var (
		quiet, verbose, allNamespaces bool
		resourceType, namespace       string
	)

	statusCommand := &cobra.Command{
		Use:   "status",
		Short: "Show the summary of the status of resources in Envoy Gateway",
		Example: `  # Show the status of gatewayclass resources under default namespace.
  egctl x status gatewayclass

  # Show the status of gateway resources with less information under default namespace.
  egctl x status gateway -q

  # Show the status of gateway resources with details under default namespace.
  egctl x status gateway -v

  # Show the status of httproute resources with details under a specific namespace.
  egctl x status httproute -v -n foobar

  # Show the status of httproute resources under all namespaces.
  egctl x status httproute -A
	`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			table := newStatusTableWriter(os.Stdout)
			k8sClient, err := newK8sClient()
			if err != nil {
				return err
			}

			switch {
			case len(args) == 1:
				resourceType = args[0]
			case len(args) > 1:
				return fmt.Errorf("unknown args: %s", strings.Join(args[1:], ","))
			default:
				return fmt.Errorf("invalid args: must specific a resources type")
			}

			return runStatus(ctx, k8sClient, table, resourceType, namespace, quiet, verbose, allNamespaces)
		},
	}

	statusCommand.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Show the status of resources only")
	statusCommand.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show the status of resources with details")
	statusCommand.PersistentFlags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "Get resources from all namespaces")
	statusCommand.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "Specific a namespace to get resources")

	return statusCommand
}

func newStatusTableWriter(out io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(out, 10, 0, 3, ' ', 0)
}

func runStatus(ctx context.Context, cli client.Client, table *tabwriter.Writer, resourceType, namespace string, quiet, verbose, allNamespaces bool) error {
	var resourcesList client.ObjectList

	if allNamespaces {
		namespace = ""
	}

	switch strings.ToLower(resourceType) {
	case "gc", "gatewayclass":
		gc := gwv1.GatewayClassList{}
		if err := cli.List(ctx, &gc, client.InNamespace(namespace)); err != nil {
			return err
		}
		resourcesList = &gc

	case "gtw", "gateway":
		gtw := gwv1.GatewayList{}
		if err := cli.List(ctx, &gtw, client.InNamespace(namespace)); err != nil {
			return err
		}
		resourcesList = &gtw

	case "httproute":
		httproute := gwv1.HTTPRouteList{}
		if err := cli.List(ctx, &httproute, client.InNamespace(namespace)); err != nil {
			return err
		}
		resourcesList = &httproute

	case "grpcroute":
		grpcroute := gwv1a2.GRPCRouteList{}
		if err := cli.List(ctx, &grpcroute, client.InNamespace(namespace)); err != nil {
			return err
		}
		resourcesList = &grpcroute

	case "tcproute":
		tcproute := gwv1a2.TCPRouteList{}
		if err := cli.List(ctx, &tcproute, client.InNamespace(namespace)); err != nil {
			return err
		}
		resourcesList = &tcproute

	case "udproute":
		udproute := gwv1a2.UDPRouteList{}
		if err := cli.List(ctx, &udproute, client.InNamespace(namespace)); err != nil {
			return err
		}
		resourcesList = &udproute

	case "tlsroute":
		tlsroute := gwv1a2.TLSRouteList{}
		if err := cli.List(ctx, &tlsroute, client.InNamespace(namespace)); err != nil {
			return err
		}
		resourcesList = &tlsroute

	case "btlspolicy", "backendtlspolicy":
		btlspolicy := gwv1a2.BackendTLSPolicyList{}
		if err := cli.List(ctx, &btlspolicy, client.InNamespace(namespace)); err != nil {
			return err
		}
		resourcesList = &btlspolicy

	case "btp", "backendtrafficpolicy":
		btp := egv1a1.BackendTrafficPolicyList{}
		if err := cli.List(ctx, &btp, client.InNamespace(namespace)); err != nil {
			return err
		}
		resourcesList = &btp

	case "ctp", "clienttrafficpolicy":
		ctp := egv1a1.ClientTrafficPolicyList{}
		if err := cli.List(ctx, &ctp, client.InNamespace(namespace)); err != nil {
			return err
		}
		resourcesList = &ctp

	case "epp", "enovypatchpolicy":
		epp := egv1a1.EnvoyPatchPolicyList{}
		if err := cli.List(ctx, &epp, client.InNamespace(namespace)); err != nil {
			return err
		}
		resourcesList = &epp

	case "sp", "securitypolicy":
		sp := egv1a1.SecurityPolicyList{}
		if err := cli.List(ctx, &sp, client.InNamespace(namespace)); err != nil {
			return err
		}
		resourcesList = &sp

	default:
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}

	namespaced, err := cli.IsObjectNamespaced(resourcesList)
	if err != nil {
		return err
	}

	needNamespaceHeader := allNamespaces && namespaced
	writeStatusHeaders(table, verbose, needNamespaceHeader)

	if err = writeStatusBodies(table, resourcesList, resourceType, quiet, verbose, needNamespaceHeader); err != nil {
		return err
	}

	return table.Flush()
}

func writeStatusHeaders(table *tabwriter.Writer, verbose, needNamespace bool) {
	headers := []string{"NAME", "TYPE", "STATUS", "REASON"}

	if needNamespace {
		headers = append([]string{"NAMESPACE"}, headers...)
	}
	if verbose {
		headers = append(headers, []string{"MESSAGE", "OBSERVED GENERATION", "LAST TRANSITION TIME"}...)
	}

	fmt.Fprintln(table, strings.Join(headers, "\t"))
}

func writeStatusBodies(table *tabwriter.Writer, resourcesList client.ObjectList, resourceType string, quiet, verbose, needNamespace bool) error {
	v := reflect.ValueOf(resourcesList).Elem()

	itemsField := v.FieldByName("Items")
	if !itemsField.IsValid() {
		return fmt.Errorf("failed to load `.Items` field from %s", resourceType)
	}

	for i := 0; i < itemsField.Len(); i++ {
		item := itemsField.Index(i)

		var name, namespace string
		nameField := item.FieldByName("Name")
		if !nameField.IsValid() {
			return fmt.Errorf("failed to find `.Items[i].Name` field from %s", resourceType)
		}
		name = nameField.String()

		if needNamespace {
			namespaceField := item.FieldByName("Namespace")
			if !namespaceField.IsValid() {
				return fmt.Errorf("failed to find `.Items[i].Namespace` field from %s", resourceType)
			}
			namespace = namespaceField.String()
		}

		statusField := item.FieldByName("Status")
		if !statusField.IsValid() {
			return fmt.Errorf("failed to find `.Items[i].Status` field from %s", resourceType)
		}

		// Different resources store the conditions at different position.
		switch {
		case strings.Contains(resourceType, "route"):
			// Scrape conditions from `Resource.Status.Parents[i].Conditions` field
			parentsField := statusField.FieldByName("Parents")
			if !parentsField.IsValid() {
				return fmt.Errorf("failed to find `.Items[i].Status.Parents` field from %s", resourceType)
			}

			for j := 0; j < parentsField.Len(); j++ {
				parentItem := parentsField.Index(j)
				if err := findAndWriteConditions(table, parentItem, resourceType, name, namespace, quiet, verbose, needNamespace); err != nil {
					return err
				}
			}

		case resourceType == "btlspolicy" || resourceType == "backendtlspolicy":
			// Scrape conditions from `Resource.Status.Ancestors[i].Conditions` field
			ancestorsField := statusField.FieldByName("Ancestors")
			if !ancestorsField.IsValid() {
				return fmt.Errorf("failed to find `.Items[i].Status.Ancestors` field from %s", resourceType)
			}

			for j := 0; j < ancestorsField.Len(); j++ {
				ancestorItem := ancestorsField.Index(j)
				if err := findAndWriteConditions(table, ancestorItem, resourceType, name, namespace, quiet, verbose, needNamespace); err != nil {
					return err
				}
			}

		default:
			// Scrape conditions from `Resource.Status.Conditions` field
			if err := findAndWriteConditions(table, statusField, resourceType, name, namespace, quiet, verbose, needNamespace); err != nil {
				return err
			}
		}
	}

	return nil
}

func findAndWriteConditions(table *tabwriter.Writer, parent reflect.Value, resourceType, name, namespace string, quiet, verbose, needNamespace bool) error {
	conditionsField := parent.FieldByName("Conditions")
	if !conditionsField.IsValid() {
		return fmt.Errorf("failed to find `Conditions` field for %s", resourceType)
	}

	conditions := conditionsField.Interface().([]metav1.Condition)
	writeConditions(table, conditions, name, namespace, quiet, verbose, needNamespace)

	return nil
}

func writeConditions(table *tabwriter.Writer, conditions []metav1.Condition, name, namespace string, quiet, verbose, needNamespace bool) {
	// Sort in descending order by time of each condition.
	for i := len(conditions) - 1; i >= 0; i-- {
		if i < len(conditions)-1 {
			name, namespace = "", ""
		}

		writeCondition(table, conditions[i], name, namespace, verbose, needNamespace)

		if quiet {
			break
		}
	}
}

func writeCondition(table *tabwriter.Writer, condition metav1.Condition, name, namespace string, verbose, needNamespace bool) {
	row := []string{name, condition.Type, string(condition.Status), condition.Reason}

	// Write conditions corresponding to its headers.
	if needNamespace {
		row = append([]string{namespace}, row...)
	}
	if verbose {
		row = append(row, []string{
			condition.Message,
			strconv.FormatInt(condition.ObservedGeneration, 10),
			condition.LastTransitionTime.String(),
		}...)
	}

	fmt.Fprintln(table, strings.Join(row, "\t"))
}
