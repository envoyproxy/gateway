// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1a3 "sigs.k8s.io/gateway-api/apis/v1alpha3"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

var (
	supportedXRouteTypes = []string{
		resource.KindHTTPRoute, resource.KindGRPCRoute, resource.KindTCPRoute,
		resource.KindUDPRoute, resource.KindTLSRoute,
	}

	supportedXPolicyTypes = []string{
		resource.KindBackendTLSPolicy, resource.KindBackendTrafficPolicy, resource.KindClientTrafficPolicy,
		resource.KindSecurityPolicy, resource.KindEnvoyPatchPolicy, resource.KindEnvoyExtensionPolicy,
	}

	supportedAllTypes = []string{
		resource.KindGatewayClass, resource.KindGateway,
	}
)

func init() {
	supportedAllTypes = append(supportedAllTypes, supportedXRouteTypes...)
	supportedAllTypes = append(supportedAllTypes, supportedXPolicyTypes...)
}

func newStatusCommand() *cobra.Command {
	var (
		quiet, verbose, allNamespaces bool
		resourceType, namespace       string
	)

	statusCommand := &cobra.Command{
		Use:   "status",
		Short: "Show the summary of the status of resources in Envoy Gateway",
		Example: `  # Show the status of gatewayclass resources.
  egctl x status gatewayclass

  # Show the status of gateway resources with less information under default namespace.
  egctl x status gateway -q

  # Show the status of gateway resources with details under default namespace.
  egctl x status gateway -v

  # Show the status of httproute resources with details under a specific namespace.
  egctl x status httproute -v -n foobar

  # Show the status of all route resources under all namespaces.
  egctl x status xroute -A

  # Show the status of all resources under all namespaces.
  egctl x status all -A
	`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
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

			switch strings.ToLower(resourceType) {
			case "all":
				for _, rt := range supportedAllTypes {
					if err = runStatus(ctx, cmd.OutOrStdout(), k8sClient, rt, namespace, quiet, verbose, allNamespaces, true, true); err != nil {
						return err
					}
				}
				return nil
			case "xroute":
				for _, rt := range supportedXRouteTypes {
					if err = runStatus(ctx, cmd.OutOrStdout(), k8sClient, rt, namespace, quiet, verbose, allNamespaces, true, true); err != nil {
						return err
					}
				}
				return nil
			case "xpolicy":
				for _, rt := range supportedXPolicyTypes {
					if err = runStatus(ctx, cmd.OutOrStdout(), k8sClient, rt, namespace, quiet, verbose, allNamespaces, true, true); err != nil {
						return err
					}
				}
				return nil
			default:
				return runStatus(ctx, cmd.OutOrStdout(), k8sClient, resourceType, namespace, quiet, verbose, allNamespaces, false, false)
			}
		},
	}

	statusCommand.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Show the first status of resources only")
	statusCommand.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show the status of resources with details")
	statusCommand.PersistentFlags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "Get the status of resources from all namespaces")
	statusCommand.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "Specific a namespace to get the status of resources")

	return statusCommand
}

func newStatusTableWriter(out io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(out, 10, 0, 3, ' ', 0)
}

func writeStatusTable(table *tabwriter.Writer, header []string, body [][]string) {
	fmt.Fprintln(table, strings.Join(header, "\t"))
	for _, b := range body {
		fmt.Fprintln(table, strings.Join(b, "\t"))
	}
}

// runStatus find and write the summary table of status for a specific resource type.
func runStatus(ctx context.Context, logOut io.Writer, cli client.Client, inputResourceType, namespace string, quiet, verbose, allNamespaces, ignoreEmpty, typedName bool) error {
	var (
		resourcesList client.ObjectList
		resourceKind  string
		table         = newStatusTableWriter(logOut)
	)

	if allNamespaces {
		namespace = ""
	}

	switch strings.ToLower(inputResourceType) {
	case "gc", "gatewayclass":
		gc := gwapiv1.GatewayClassList{}
		if err := cli.List(ctx, &gc, client.InNamespace(namespace)); err != nil {
			return err
		}
		resourcesList = &gc
		resourceKind = resource.KindGatewayClass

	case "gtw", "gateway":
		gtw := gwapiv1.GatewayList{}
		if err := cli.List(ctx, &gtw, client.InNamespace(namespace)); err != nil {
			return err
		}
		resourcesList = &gtw
		resourceKind = resource.KindGateway

	case "httproute":
		httproute := gwapiv1.HTTPRouteList{}
		if err := cli.List(ctx, &httproute, client.InNamespace(namespace)); err != nil {
			return err
		}
		resourcesList = &httproute
		resourceKind = resource.KindHTTPRoute

	case "grpcroute":
		grpcroute := gwapiv1.GRPCRouteList{}
		if err := cli.List(ctx, &grpcroute, client.InNamespace(namespace)); err != nil {
			return err
		}
		resourcesList = &grpcroute
		resourceKind = resource.KindGRPCRoute

	case "tcproute":
		tcproute := gwapiv1a2.TCPRouteList{}
		if err := cli.List(ctx, &tcproute, client.InNamespace(namespace)); err != nil {
			return err
		}
		resourcesList = &tcproute
		resourceKind = resource.KindTCPRoute

	case "udproute":
		udproute := gwapiv1a2.UDPRouteList{}
		if err := cli.List(ctx, &udproute, client.InNamespace(namespace)); err != nil {
			return err
		}
		resourcesList = &udproute
		resourceKind = resource.KindUDPRoute

	case "tlsroute":
		tlsroute := gwapiv1a2.TLSRouteList{}
		if err := cli.List(ctx, &tlsroute, client.InNamespace(namespace)); err != nil {
			return err
		}
		resourcesList = &tlsroute
		resourceKind = resource.KindTLSRoute

	case "btlspolicy", "backendtlspolicy":
		btlspolicy := gwapiv1a3.BackendTLSPolicyList{}
		if err := cli.List(ctx, &btlspolicy, client.InNamespace(namespace)); err != nil {
			return err
		}
		resourcesList = &btlspolicy
		resourceKind = resource.KindBackendTLSPolicy

	case "btp", "backendtrafficpolicy":
		btp := egv1a1.BackendTrafficPolicyList{}
		if err := cli.List(ctx, &btp, client.InNamespace(namespace)); err != nil {
			return err
		}
		resourcesList = &btp
		resourceKind = resource.KindBackendTrafficPolicy

	case "ctp", "clienttrafficpolicy":
		ctp := egv1a1.ClientTrafficPolicyList{}
		if err := cli.List(ctx, &ctp, client.InNamespace(namespace)); err != nil {
			return err
		}
		resourcesList = &ctp
		resourceKind = resource.KindClientTrafficPolicy

	case "epp", "envoypatchpolicy":
		epp := egv1a1.EnvoyPatchPolicyList{}
		if err := cli.List(ctx, &epp, client.InNamespace(namespace)); err != nil {
			return err
		}
		resourcesList = &epp
		resourceKind = resource.KindEnvoyPatchPolicy

	case "eep", "envoyextensionpolicy":
		eep := egv1a1.EnvoyExtensionPolicyList{}
		if err := cli.List(ctx, &eep, client.InNamespace(namespace)); err != nil {
			return err
		}
		resourcesList = &eep
		resourceKind = resource.KindEnvoyExtensionPolicy

	case "sp", "securitypolicy":
		sp := egv1a1.SecurityPolicyList{}
		if err := cli.List(ctx, &sp, client.InNamespace(namespace)); err != nil {
			return err
		}
		resourcesList = &sp
		resourceKind = resource.KindSecurityPolicy

	default:
		return fmt.Errorf("unknown input resource type: %s, supported input types are: %s",
			inputResourceType, strings.Join(supportedAllTypes, ", "))
	}

	namespaced, err := cli.IsObjectNamespaced(resourcesList)
	if err != nil {
		return err
	}

	needNamespaceHeader := allNamespaces && namespaced
	header := fetchStatusHeader(resourceKind, verbose, needNamespaceHeader)
	body := fetchStatusBody(resourcesList, resourceKind, quiet, verbose, needNamespaceHeader, typedName)

	if ignoreEmpty && len(body) == 0 {
		return nil
	}

	writeStatusTable(table, header, body)
	if err = table.Flush(); err != nil {
		return err
	}

	// Separate tables by newline if there are multiple tables.
	if ignoreEmpty && typedName {
		fmt.Print("\n")
	}

	return nil
}

// extendStatusHeader extends header in the way of:
//   - Insert `NAMESPACE` at first if needed
//   - Append various details if verbose is on
func extendStatusHeader(header []string, verbose, needNamespace bool) []string {
	if needNamespace {
		header = append([]string{"NAMESPACE"}, header...)
	}
	if verbose {
		header = append(header, []string{"MESSAGE", "OBSERVED GENERATION", "LAST TRANSITION TIME"}...)
	}

	return header
}

// extendStatusBodyWithNamespaceAndName extends current body with namespace and name at head.
func extendStatusBodyWithNamespaceAndName(body [][]string, namespace, name string, needNamespace bool) [][]string {
	for i := 0; i < len(body); i++ {
		if needNamespace {
			body[i] = append([]string{namespace, name}, body[i]...)
		} else {
			body[i] = append([]string{name}, body[i]...)
		}
		// Only display once for the first row.
		namespace, name = "", ""
	}
	return body
}

func kindName(kind, name string) string {
	return strings.ToLower(kind) + "/" + name
}

func fetchStatusHeader(resourceKind string, verbose, needNamespace bool) (header []string) {
	defaultHeader := []string{"NAME", "TYPE", "STATUS", "REASON"}
	xRouteHeader := []string{"NAME", "PARENT", "TYPE", "STATUS", "REASON"}
	xPolicyHeader := []string{"NAME", "ANCESTOR REFERENCE", "TYPE", "STATUS", "REASON"}

	switch {
	case strings.HasSuffix(resourceKind, "Route"):
		return extendStatusHeader(xRouteHeader, verbose, needNamespace)
	case strings.HasSuffix(resourceKind, "Policy"):
		return extendStatusHeader(xPolicyHeader, verbose, needNamespace)
	default:
		return extendStatusHeader(defaultHeader, verbose, needNamespace)
	}
}

func fetchStatusBody(resourcesList client.ObjectList, resourceKind string, quiet, verbose, needNamespace, typedName bool) (body [][]string) {
	v := reflect.ValueOf(resourcesList).Elem()
	itemsField := v.FieldByName("Items")

	for i := 0; i < itemsField.Len(); i++ {
		var (
			name, namespace string
			rows            [][]string
			item            = itemsField.Index(i)
			nameField       = item.FieldByName("Name")
			statusField     = item.FieldByName("Status")
		)

		if typedName {
			name = kindName(resourceKind, nameField.String())
		} else {
			name = nameField.String()
		}

		if needNamespace {
			namespaceField := item.FieldByName("Namespace")
			namespace = namespaceField.String()
		}

		switch {
		// For xRoute, the conditions are storing in `Resource.Status.Parents[i].Conditions`.
		case strings.HasSuffix(resourceKind, "Route"):
			parentsField := statusField.FieldByName("Parents")
			for j := 0; j < parentsField.Len(); j++ {
				parentItem := parentsField.Index(j)
				conditions := fetchConditions(parentItem, quiet, verbose)

				// Extend conditions with parent.
				parentRef := parentItem.FieldByName("ParentRef")
				parentName := kindName(
					parentRef.FieldByName("Kind").Elem().String(),
					parentRef.FieldByName("Name").String(),
				)
				for k := 0; k < len(conditions); k++ {
					conditions[k] = append([]string{parentName}, conditions[k]...)
					parentName = ""
				}

				rows = append(rows, conditions...)
			}

		// For xPolicy, the conditions are storing in `Resource.Status.Ancestors[i].Conditions`.
		case strings.HasSuffix(resourceKind, "Policy"):
			ancestorsField := statusField.FieldByName("Ancestors")
			for j := 0; j < ancestorsField.Len(); j++ {
				policyAncestorStatus := ancestorsField.Index(j)
				conditions := fetchConditions(policyAncestorStatus, quiet, verbose)

				// Extend conditions with ancestor.
				ancestorRef := policyAncestorStatus.FieldByName("AncestorRef")
				ancestorName := kindName(
					ancestorRef.FieldByName("Kind").Elem().String(),
					ancestorRef.FieldByName("Name").String(),
				)
				for k := 0; k < len(conditions); k++ {
					conditions[k] = append([]string{ancestorName}, conditions[k]...)
					ancestorName = ""
				}

				rows = append(rows, conditions...)
			}

		// For others, the conditions are storing in `Resource.Status.Conditions`.
		default:
			conditions := fetchConditions(statusField, quiet, verbose)
			rows = append(rows, conditions...)
		}

		rows = extendStatusBodyWithNamespaceAndName(rows, namespace, name, needNamespace)
		body = append(body, rows...)
	}

	return body
}

// fetchConditions fetches conditions from the `Conditions` field of parent
// by calling fetchCondition for each condition.
func fetchConditions(parent reflect.Value, quiet, verbose bool) [][]string {
	var rows [][]string

	conditionsField := parent.FieldByName("Conditions")
	conditions := conditionsField.Interface().([]metav1.Condition)

	// All conditions are sorted in descending order by time.
	for i := len(conditions) - 1; i >= 0; i-- {
		row := fetchCondition(conditions[i], verbose)
		rows = append(rows, row)

		if quiet {
			break
		}
	}

	return rows
}

// fetchCondition fetches the Type, Status, Reason of one condition, and more if verbose.
func fetchCondition(condition metav1.Condition, verbose bool) []string {
	row := []string{condition.Type, string(condition.Status), condition.Reason}

	// Write more details about this condition if verbose is on.
	if verbose {
		row = append(row, []string{
			condition.Message,
			strconv.FormatInt(condition.ObservedGeneration, 10),
			condition.LastTransitionTime.String(),
		}...)
	}

	return row
}
