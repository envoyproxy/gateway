// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	perr "github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
	"github.com/envoyproxy/gateway/internal/wasm"
)

// oci URL prefix
const ociURLPrefix = "oci://"

func (t *Translator) ProcessEnvoyExtensionPolicies(envoyExtensionPolicies []*egv1a1.EnvoyExtensionPolicy,
	gateways []*GatewayContext,
	routes []RouteContext,
	resources *Resources,
	xdsIR XdsIRMap,
) []*egv1a1.EnvoyExtensionPolicy {
	var res []*egv1a1.EnvoyExtensionPolicy

	// Sort based on timestamp
	sort.Slice(envoyExtensionPolicies, func(i, j int) bool {
		return envoyExtensionPolicies[i].CreationTimestamp.Before(&(envoyExtensionPolicies[j].CreationTimestamp))
	})

	// First build a map out of the routes and gateways for faster lookup since users might have thousands of routes or more.
	routeMap := map[policyTargetRouteKey]*policyRouteTargetContext{}
	for _, route := range routes {
		key := policyTargetRouteKey{
			Kind:      string(GetRouteType(route)),
			Name:      route.GetName(),
			Namespace: route.GetNamespace(),
		}
		routeMap[key] = &policyRouteTargetContext{RouteContext: route}
	}

	gatewayMap := map[types.NamespacedName]*policyGatewayTargetContext{}
	for _, gw := range gateways {
		key := utils.NamespacedName(gw)
		gatewayMap[key] = &policyGatewayTargetContext{GatewayContext: gw}
	}

	// Map of Gateway to the routes attached to it
	gatewayRouteMap := make(map[string]sets.Set[string])

	// Translate
	// 1. First translate Policies targeting xRoutes
	// 2. Finally, the policies targeting Gateways

	// Process the policies targeting xRoutes
	for _, policy := range envoyExtensionPolicies {
		if policy.Spec.TargetRef.Kind != KindGateway {
			policy := policy.DeepCopy()
			res = append(res, policy)

			// Negative statuses have already been assigned so its safe to skip
			route, resolveErr := resolveEEPolicyRouteTargetRef(policy, routeMap)
			if route == nil {
				continue
			}

			// Find the Gateway that the route belongs to and add it to the
			// gatewayRouteMap and ancestor list, which will be used to check
			// policy overrides and populate its ancestor status.
			parentRefs := GetParentReferences(route)
			ancestorRefs := make([]gwapiv1a2.ParentReference, 0, len(parentRefs))
			for _, p := range parentRefs {
				if p.Kind == nil || *p.Kind == KindGateway {
					namespace := route.GetNamespace()
					if p.Namespace != nil {
						namespace = string(*p.Namespace)
					}
					gwNN := types.NamespacedName{
						Namespace: namespace,
						Name:      string(p.Name),
					}

					key := gwNN.String()
					if _, ok := gatewayRouteMap[key]; !ok {
						gatewayRouteMap[key] = make(sets.Set[string])
					}
					gatewayRouteMap[key].Insert(utils.NamespacedName(route).String())

					// Do need a section name since the policy is targeting to a route
					ancestorRefs = append(ancestorRefs, getAncestorRefForPolicy(gwNN, p.SectionName))
				}
			}

			// Set conditions for resolve error, then skip current xroute
			if resolveErr != nil {
				status.SetResolveErrorForPolicyAncestors(&policy.Status,
					ancestorRefs,
					t.GatewayControllerName,
					policy.Generation,
					resolveErr,
				)

				continue
			}

			// Set conditions for translation error if it got any
			if err := t.translateEnvoyExtensionPolicyForRoute(policy, route, xdsIR, resources); err != nil {
				status.SetTranslationErrorForPolicyAncestors(&policy.Status,
					ancestorRefs,
					t.GatewayControllerName,
					policy.Generation,
					status.Error2ConditionMsg(err),
				)
			}

			// Set Accepted condition if it is unset
			status.SetAcceptedForPolicyAncestors(&policy.Status, ancestorRefs, t.GatewayControllerName)
		}
	}

	// Process the policies targeting Gateways
	for _, policy := range envoyExtensionPolicies {
		if policy.Spec.TargetRef.Kind == KindGateway {
			policy := policy.DeepCopy()
			res = append(res, policy)

			// Negative statuses have already been assigned so its safe to skip
			gateway, resolveErr := resolveEEPolicyGatewayTargetRef(policy, gatewayMap)
			if gateway == nil {
				continue
			}

			// Find its ancestor reference by resolved gateway, even with resolve error
			gatewayNN := utils.NamespacedName(gateway)
			ancestorRefs := []gwapiv1a2.ParentReference{
				// Don't need a section name since the policy is targeting to a gateway
				getAncestorRefForPolicy(gatewayNN, nil),
			}

			// Set conditions for resolve error, then skip current gateway
			if resolveErr != nil {
				status.SetResolveErrorForPolicyAncestors(&policy.Status,
					ancestorRefs,
					t.GatewayControllerName,
					policy.Generation,
					resolveErr,
				)

				continue
			}

			// Set conditions for translation error if it got any
			if err := t.translateEnvoyExtensionPolicyForGateway(policy, gateway, xdsIR, resources); err != nil {
				status.SetTranslationErrorForPolicyAncestors(&policy.Status,
					ancestorRefs,
					t.GatewayControllerName,
					policy.Generation,
					status.Error2ConditionMsg(err),
				)
			}

			// Set Accepted condition if it is unset
			status.SetAcceptedForPolicyAncestors(&policy.Status, ancestorRefs, t.GatewayControllerName)

			// Check if this policy is overridden by other policies targeting at
			// route level
			if r, ok := gatewayRouteMap[gatewayNN.String()]; ok {
				// Maintain order here to ensure status/string does not change with the same data
				routes := r.UnsortedList()
				sort.Strings(routes)
				message := fmt.Sprintf("This policy is being overridden by other envoyExtensionPolicies for these routes: %v", routes)

				status.SetConditionForPolicyAncestors(&policy.Status,
					ancestorRefs,
					t.GatewayControllerName,
					egv1a1.PolicyConditionOverridden,
					metav1.ConditionTrue,
					egv1a1.PolicyReasonOverridden,
					message,
					policy.Generation,
				)
			}
		}
	}

	return res
}

func resolveEEPolicyGatewayTargetRef(policy *egv1a1.EnvoyExtensionPolicy, gateways map[types.NamespacedName]*policyGatewayTargetContext) (*GatewayContext, *status.PolicyResolveError) {
	targetNs := policy.Namespace

	// Check if the gateway exists
	key := types.NamespacedName{
		Name:      string(policy.Spec.TargetRef.Name),
		Namespace: targetNs,
	}
	gateway, ok := gateways[key]

	// Gateway not found
	if !ok {
		return nil, nil
	}

	// Ensure Policy and target are in the same namespace
	if policy.Namespace != targetNs {
		message := fmt.Sprintf("Namespace:%s TargetRef.Namespace:%s, EnvoyExtensionPolicy can only target a resource in the same namespace.",
			policy.Namespace, targetNs)

		return gateway.GatewayContext, &status.PolicyResolveError{
			Reason:  gwapiv1a2.PolicyReasonInvalid,
			Message: message,
		}
	}

	// Check if another policy targeting the same Gateway exists
	if gateway.attached {
		message := "Unable to target Gateway, another EnvoyExtensionPolicy has already attached to it"

		return gateway.GatewayContext, &status.PolicyResolveError{
			Reason:  gwapiv1a2.PolicyReasonConflicted,
			Message: message,
		}
	}

	// Set context and save
	gateway.attached = true
	gateways[key] = gateway

	return gateway.GatewayContext, nil
}

func resolveEEPolicyRouteTargetRef(policy *egv1a1.EnvoyExtensionPolicy, routes map[policyTargetRouteKey]*policyRouteTargetContext) (RouteContext, *status.PolicyResolveError) {
	targetNs := policy.Namespace

	// Check if the route exists
	key := policyTargetRouteKey{
		Kind:      string(policy.Spec.TargetRef.Kind),
		Name:      string(policy.Spec.TargetRef.Name),
		Namespace: targetNs,
	}

	route, ok := routes[key]
	// Route not found
	if !ok {
		return nil, nil
	}

	// Ensure Policy and target are in the same namespace
	if policy.Namespace != targetNs {
		message := fmt.Sprintf("Namespace:%s TargetRef.Namespace:%s, EnvoyExtensionPolicy can only target a resource in the same namespace.",
			policy.Namespace, targetNs)

		return route.RouteContext, &status.PolicyResolveError{
			Reason:  gwapiv1a2.PolicyReasonInvalid,
			Message: message,
		}
	}

	// Check if another policy targeting the same xRoute exists
	if route.attached {
		message := fmt.Sprintf("Unable to target %s, another EnvoyExtensionPolicy has already attached to it",
			string(policy.Spec.TargetRef.Kind))

		return route.RouteContext, &status.PolicyResolveError{
			Reason:  gwapiv1a2.PolicyReasonConflicted,
			Message: message,
		}
	}

	// Set context and save
	route.attached = true
	routes[key] = route

	return route.RouteContext, nil
}

func (t *Translator) translateEnvoyExtensionPolicyForRoute(
	policy *egv1a1.EnvoyExtensionPolicy,
	route RouteContext,
	xdsIR XdsIRMap,
	resources *Resources,
) error {
	var (
		extProcs  []ir.ExtProc
		wasms     []ir.Wasm
		err, errs error
	)

	if extProcs, err = t.buildExtProcs(policy, resources); err != nil {
		err = perr.WithMessage(err, "ExtProc")
		errs = errors.Join(errs, err)
	}

	if wasms, err = t.buildWasms(policy, resources); err != nil {
		err = perr.WithMessage(err, "WASM")
		errs = errors.Join(errs, err)
	}

	// Early return if got any errors
	if errs != nil {
		return errs
	}

	// Apply IR to all relevant routes
	prefix := irRoutePrefix(route)
	for _, x := range xdsIR {
		for _, http := range x.HTTP {
			for _, r := range http.Routes {
				// Apply if there is a match
				if strings.HasPrefix(r.Name, prefix) {
					r.ExtProcs = extProcs
					r.Wasms = wasms
				}
			}
		}
	}

	return nil
}

func (t *Translator) translateEnvoyExtensionPolicyForGateway(policy *egv1a1.EnvoyExtensionPolicy,
	gateway *GatewayContext, xdsIR XdsIRMap, resources *Resources,
) error {
	var (
		extProcs  []ir.ExtProc
		wasms     []ir.Wasm
		err, errs error
	)

	if extProcs, err = t.buildExtProcs(policy, resources); err != nil {
		err = perr.WithMessage(err, "ExtProc")
		errs = errors.Join(errs, err)
	}

	if wasms, err = t.buildWasms(policy, resources); err != nil {
		err = perr.WithMessage(err, "WASM")
		errs = errors.Join(errs, err)
	}

	// Early return if got any errors
	if errs != nil {
		return errs
	}

	irKey := t.getIRKey(gateway.Gateway)
	// Should exist since we've validated this
	x := xdsIR[irKey]

	policyTarget := irStringKey(policy.Namespace, string(policy.Spec.TargetRef.Name))

	for _, http := range x.HTTP {
		gatewayName := http.Name[0:strings.LastIndex(http.Name, "/")]
		if t.MergeGateways && gatewayName != policyTarget {
			continue
		}

		// A Policy targeting the most specific scope(xRoute) wins over a policy
		// targeting a lesser specific scope(Gateway).
		for _, r := range http.Routes {
			// if already set - there's a route level policy, so skip
			if r.ExtProcs != nil ||
				r.Wasms != nil {
				continue
			}

			if r.ExtProcs == nil {
				r.ExtProcs = extProcs
			}
			if r.Wasms == nil {
				r.Wasms = wasms
			}
		}
	}

	return nil
}

func (t *Translator) buildExtProcs(policy *egv1a1.EnvoyExtensionPolicy, resources *Resources) ([]ir.ExtProc, error) {
	var extProcIRList []ir.ExtProc

	if policy == nil {
		return nil, nil
	}

	for idx, ep := range policy.Spec.ExtProc {
		name := irConfigNameForEEP(policy, idx)
		extProcIR, err := t.buildExtProc(name, utils.NamespacedName(policy), ep, idx, resources)
		if err != nil {
			return nil, err
		}
		extProcIRList = append(extProcIRList, *extProcIR)
	}
	return extProcIRList, nil
}

func (t *Translator) buildExtProc(
	name string,
	policyNamespacedName types.NamespacedName,
	extProc egv1a1.ExtProc,
	extProcIdx int,
	resources *Resources,
) (*ir.ExtProc, error) {
	var (
		ds        *ir.DestinationSetting
		authority string
		err       error
	)

	var dsl []*ir.DestinationSetting
	for i := range extProc.BackendRefs {
		if err = t.validateExtServiceBackendReference(
			&extProc.BackendRefs[i].BackendObjectReference,
			policyNamespacedName.Namespace,
			egv1a1.KindEnvoyExtensionPolicy,
			resources); err != nil {
			return nil, err
		}

		ds, err = t.processExtServiceDestination(
			&extProc.BackendRefs[i].BackendObjectReference,
			policyNamespacedName,
			egv1a1.KindEnvoyExtensionPolicy,
			ir.GRPC,
			resources)
		if err != nil {
			return nil, err
		}

		dsl = append(dsl, ds)
	}

	rd := ir.RouteDestination{
		Name:     irIndexedExtServiceDestinationName(policyNamespacedName, egv1a1.KindEnvoyExtensionPolicy, extProcIdx),
		Settings: dsl,
	}

	if extProc.BackendRefs[0].Port != nil {
		authority = fmt.Sprintf(
			"%s.%s:%d",
			extProc.BackendRefs[0].Name,
			NamespaceDerefOr(extProc.BackendRefs[0].Namespace, policyNamespacedName.Namespace),
			*extProc.BackendRefs[0].Port)
	} else {
		authority = fmt.Sprintf(
			"%s.%s",
			extProc.BackendRefs[0].Name,
			NamespaceDerefOr(extProc.BackendRefs[0].Namespace, policyNamespacedName.Namespace))
	}

	extProcIR := &ir.ExtProc{
		Name:        name,
		Destination: rd,
		Authority:   authority,
	}

	if extProc.MessageTimeout != nil {
		d, err := time.ParseDuration(string(*extProc.MessageTimeout))
		if err != nil {
			return nil, fmt.Errorf("invalid ExtProc MessageTimeout value %v", extProc.MessageTimeout)
		}
		extProcIR.MessageTimeout = ptr.To(metav1.Duration{Duration: d})
	}

	if extProc.FailOpen != nil {
		extProcIR.FailOpen = extProc.FailOpen
	}

	if extProc.ProcessingMode != nil {
		if extProc.ProcessingMode.Request != nil {
			extProcIR.RequestHeaderProcessing = true
			if extProc.ProcessingMode.Request.Body != nil {
				extProcIR.RequestBodyProcessingMode = ptr.To(ir.ExtProcBodyProcessingMode(*extProc.ProcessingMode.Request.Body))
			}
		}

		if extProc.ProcessingMode.Response != nil {
			extProcIR.ResponseHeaderProcessing = true
			if extProc.ProcessingMode.Response.Body != nil {
				extProcIR.ResponseBodyProcessingMode = ptr.To(ir.ExtProcBodyProcessingMode(*extProc.ProcessingMode.Response.Body))
			}
		}
	}

	return extProcIR, err
}

// TODO: zhaohuabing differentiate extproc and wasm
func irConfigNameForEEP(policy *egv1a1.EnvoyExtensionPolicy, idx int) string {
	return fmt.Sprintf(
		"%s/%s/%d",
		strings.ToLower(egv1a1.KindEnvoyExtensionPolicy),
		utils.NamespacedName(policy).String(),
		idx)
}

func (t *Translator) buildWasms(
	policy *egv1a1.EnvoyExtensionPolicy,
	resources *Resources,
) ([]ir.Wasm, error) {
	var wasmIRList []ir.Wasm

	if policy == nil {
		return nil, nil
	}

	for idx, wasm := range policy.Spec.Wasm {
		name := irConfigNameForWasm(policy, idx)
		wasmIR, err := t.buildWasm(name, wasm, policy, idx, resources)
		if err != nil {
			return nil, err
		}
		wasmIRList = append(wasmIRList, *wasmIR)
	}
	return wasmIRList, nil
}

func (t *Translator) buildWasm(
	name string,
	config egv1a1.Wasm,
	policy *egv1a1.EnvoyExtensionPolicy,
	idx int,
	resources *Resources,
) (*ir.Wasm, error) {
	var (
		failOpen   = false
		code       *ir.HTTPWasmCode
		pullPolicy wasm.PullPolicy
		// the checksum provided by the user, it's used to validate the wasm module
		// downloaded from the original HTTP server or the OCI registry
		originalChecksum string
		egServingURL     string // the wasm module download URL from the EG HTTP server
		err              error
	)

	if config.FailOpen != nil {
		failOpen = *config.FailOpen
	}

	if config.Code.SHA256 != nil {
		originalChecksum = *config.Code.SHA256
	}

	if config.Code.PullPolicy != nil {
		switch *config.Code.PullPolicy {
		case egv1a1.ImagePullPolicyAlways:
			pullPolicy = wasm.Always
		case egv1a1.ImagePullPolicyIfNotPresent:
			pullPolicy = wasm.IfNotPresent
		default:
			pullPolicy = wasm.Unspecified
		}
	}

	switch config.Code.Type {
	case egv1a1.HTTPWasmCodeSourceType:
		// This is a sanity check, the validation should have caught this
		if config.Code.HTTP == nil {
			return nil, fmt.Errorf("missing HTTP field in Wasm code source")
		}

		http := config.Code.HTTP

		if egServingURL, _, err = t.WasmCache.Get(http.URL, wasm.GetOptions{
			Checksum:        originalChecksum,
			PullPolicy:      pullPolicy,
			ResourceName:    irConfigNameForWasm(policy, idx),
			ResourceVersion: policy.ResourceVersion,
		}); err != nil {
			return nil, err
		}

		code = &ir.HTTPWasmCode{
			EGServingURL:           egServingURL,
			OriginalDownloadingURL: http.URL,
			SHA256:                 originalChecksum,
		}

	case egv1a1.ImageWasmCodeSourceType:
		var (
			image      = config.Code.Image
			secret     *v1.Secret
			pullSecret []byte
			// the checksum of the wasm module extracted from the OCI image
			// it's different from the checksum for the OCI image
			checksum string
		)

		// This is a sanity check, the validation should have caught this
		if image == nil {
			return nil, fmt.Errorf("missing Image field in Wasm code source")
		}

		if image.PullSecretRef != nil {
			from := crossNamespaceFrom{
				group:     egv1a1.GroupName,
				kind:      KindEnvoyExtensionPolicy,
				namespace: policy.Namespace,
			}

			if secret, err = t.validateSecretRef(
				false, from, *image.PullSecretRef, resources); err != nil {
				for _, s := range resources.Secrets {
					fmt.Println(fmt.Sprintf("xxxxxx: %s/%s", s.Namespace, s.Name))
				}
				return nil, err
			}

			if data, ok := secret.Data[v1.DockerConfigJsonKey]; ok {
				pullSecret = data
			} else {
				return nil, fmt.Errorf("missing %s key in secret %s/%s", v1.DockerConfigJsonKey, secret.Namespace, secret.Name)
			}
		}

		// Wasm Cache requires the URL to be in the format "scheme://<URL>"
		imageURL := image.URL
		if !strings.HasPrefix(image.URL, ociURLPrefix) {
			imageURL = fmt.Sprintf("%s%s", ociURLPrefix, image.URL)
		}

		// If the url is an OCI image, and neither digest nor tag is provided, use the latest tag.
		if !hasDigest(imageURL) && !hasTag(imageURL) {
			imageURL += ":latest"
		}

		// The wasm checksum is different from the OCI image digest
		if egServingURL, checksum, err = t.WasmCache.Get(imageURL, wasm.GetOptions{
			Checksum:        originalChecksum,
			PullSecret:      pullSecret,
			PullPolicy:      pullPolicy,
			ResourceName:    irConfigNameForWasm(policy, idx),
			ResourceVersion: policy.ResourceVersion,
		}); err != nil {
			return nil, err
		}

		code = &ir.HTTPWasmCode{
			EGServingURL:           egServingURL,
			SHA256:                 checksum,
			OriginalDownloadingURL: imageURL,
		}
	default:
		// should never happen because of kubebuilder validation, just a sanity check
		return nil, fmt.Errorf("unsupported Wasm code source type %q", config.Code.Type)
	}

	wasmIR := &ir.Wasm{
		Name:     name,
		RootID:   config.RootID,
		WasmName: config.Name,
		Config:   config.Config,
		FailOpen: failOpen,
		Code:     code,
	}

	return wasmIR, nil
}

func hasDigest(imageURL string) bool {
	return strings.Contains(imageURL, "@")
}

func hasTag(imageURL string) bool {
	parts := strings.Split(imageURL[len(ociURLPrefix):], ":")
	// Verify that we aren't confusing a tag for a hostname with port.
	return len(parts) > 1 && !strings.Contains(parts[len(parts)-1], "/")
}

func irConfigNameForWasm(policy client.Object, index int) string {
	return fmt.Sprintf(
		"%s/wasm/%s",
		irConfigName(policy),
		strconv.Itoa(index))
}

