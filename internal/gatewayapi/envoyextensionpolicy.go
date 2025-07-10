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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/luavalidator"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
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
	resources *resource.Resources,
	xdsIR resource.XdsIRMap,
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

	handledPolicies := make(map[types.NamespacedName]*egv1a1.EnvoyExtensionPolicy)

	// Translate
	// 1. First translate Policies targeting xRoutes
	// 2. Finally, the policies targeting Gateways

	// Process the policies targeting xRoutes
	for _, currPolicy := range envoyExtensionPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := getPolicyTargetRefs(currPolicy.Spec.PolicyTargetReferences, routes)
		for _, currTarget := range targetRefs {
			if currTarget.Kind != resource.KindGateway {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = currPolicy.DeepCopy()
					res = append(res, policy)
					handledPolicies[policyName] = policy
				}

				// Negative statuses have already been assigned so its safe to skip
				route, resolveErr := resolveEEPolicyRouteTargetRef(policy, currTarget, routeMap)
				if route == nil {
					continue
				}

				// Find the Gateway that the route belongs to and add it to the
				// gatewayRouteMap and ancestor list, which will be used to check
				// policy overrides and populate its ancestor status.
				parentRefs := GetParentReferences(route)
				ancestorRefs := make([]gwapiv1a2.ParentReference, 0, len(parentRefs))
				for _, p := range parentRefs {
					if p.Kind == nil || *p.Kind == resource.KindGateway {
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
				status.SetAcceptedForPolicyAncestors(&policy.Status, ancestorRefs, t.GatewayControllerName, policy.Generation)
			}
		}
	}

	// Process the policies targeting Gateways
	for _, currPolicy := range envoyExtensionPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := getPolicyTargetRefs(currPolicy.Spec.PolicyTargetReferences, gateways)
		for _, currTarget := range targetRefs {
			if currTarget.Kind == resource.KindGateway {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = currPolicy.DeepCopy()
					res = append(res, policy)
					handledPolicies[policyName] = policy
				}

				// Negative statuses have already been assigned so its safe to skip
				gateway, resolveErr := resolveEEPolicyGatewayTargetRef(policy, currTarget, gatewayMap)
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
				if err := t.translateEnvoyExtensionPolicyForGateway(policy, currTarget, gateway, xdsIR, resources); err != nil {
					status.SetTranslationErrorForPolicyAncestors(&policy.Status,
						ancestorRefs,
						t.GatewayControllerName,
						policy.Generation,
						status.Error2ConditionMsg(err),
					)
				}

				// Set Accepted condition if it is unset
				status.SetAcceptedForPolicyAncestors(&policy.Status, ancestorRefs, t.GatewayControllerName, policy.Generation)

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
	}

	return res
}

func resolveEEPolicyGatewayTargetRef(policy *egv1a1.EnvoyExtensionPolicy, target gwapiv1a2.LocalPolicyTargetReferenceWithSectionName, gateways map[types.NamespacedName]*policyGatewayTargetContext) (*GatewayContext, *status.PolicyResolveError) {
	// Check if the gateway exists
	key := types.NamespacedName{
		Name:      string(target.Name),
		Namespace: policy.Namespace,
	}
	gateway, ok := gateways[key]

	// Gateway not found
	if !ok {
		return nil, nil
	}

	// Check if another policy targeting the same Gateway exists
	if gateway.attached {
		message := fmt.Sprintf("Unable to target Gateway %s, another EnvoyExtensionPolicy has already attached to it",
			string(target.Name))

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

func resolveEEPolicyRouteTargetRef(policy *egv1a1.EnvoyExtensionPolicy, target gwapiv1a2.LocalPolicyTargetReferenceWithSectionName, routes map[policyTargetRouteKey]*policyRouteTargetContext) (RouteContext, *status.PolicyResolveError) {
	// Check if the route exists
	key := policyTargetRouteKey{
		Kind:      string(target.Kind),
		Name:      string(target.Name),
		Namespace: policy.Namespace,
	}

	route, ok := routes[key]
	// Route not found
	if !ok {
		return nil, nil
	}

	// Check if another policy targeting the same xRoute exists
	if route.attached {
		message := fmt.Sprintf("Unable to target %s %s, another EnvoyExtensionPolicy has already attached to it",
			string(target.Kind), string(target.Name))

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
	xdsIR resource.XdsIRMap,
	resources *resource.Resources,
) error {
	var (
		wasms     []ir.Wasm
		luas      []ir.Lua
		err, errs error
	)

	if wasms, err = t.buildWasms(policy, resources); err != nil {
		err = perr.WithMessage(err, "Wasm")
		errs = errors.Join(errs, err)
	}

	// Apply IR to all relevant routes
	prefix := irRoutePrefix(route)
	parentRefs := GetParentReferences(route)
	for _, p := range parentRefs {
		parentRefCtx := GetRouteParentContext(route, p)
		gtwCtx := parentRefCtx.GetGateway()
		if gtwCtx == nil {
			continue
		}

		if luas, err = t.buildLuas(policy, resources, gtwCtx.envoyProxy); err != nil {
			err = perr.WithMessage(err, "Lua")
			errs = errors.Join(errs, err)
		}

		var extProcs []ir.ExtProc
		if extProcs, err = t.buildExtProcs(policy, resources, gtwCtx.envoyProxy); err != nil {
			err = perr.WithMessage(err, "ExtProc")
			errs = errors.Join(errs, err)
		}
		irKey := t.getIRKey(gtwCtx.Gateway)
		for _, listener := range parentRefCtx.listeners {
			irListener := xdsIR[irKey].GetHTTPListener(irListenerName(listener))
			if irListener != nil {
				for _, r := range irListener.Routes {
					if strings.HasPrefix(r.Name, prefix) {
						// return 500 and do not configure EnvoyExtensions in this case
						if errs != nil {
							r.DirectResponse = &ir.CustomResponse{
								StatusCode: ptr.To(uint32(500)),
							}
							continue
						}
						r.EnvoyExtensions = &ir.EnvoyExtensionFeatures{
							ExtProcs: extProcs,
							Wasms:    wasms,
							Luas:     luas,
						}
					}
				}
			}
		}
	}

	return errs
}

func (t *Translator) translateEnvoyExtensionPolicyForGateway(
	policy *egv1a1.EnvoyExtensionPolicy,
	target gwapiv1a2.LocalPolicyTargetReferenceWithSectionName,
	gateway *GatewayContext,
	xdsIR resource.XdsIRMap,
	resources *resource.Resources,
) error {
	var (
		extProcs  []ir.ExtProc
		wasms     []ir.Wasm
		luas      []ir.Lua
		err, errs error
	)

	if extProcs, err = t.buildExtProcs(policy, resources, gateway.envoyProxy); err != nil {
		err = perr.WithMessage(err, "ExtProc")
		errs = errors.Join(errs, err)
	}
	if wasms, err = t.buildWasms(policy, resources); err != nil {
		err = perr.WithMessage(err, "Wasm")
		errs = errors.Join(errs, err)
	}
	if luas, err = t.buildLuas(policy, resources, gateway.envoyProxy); err != nil {
		err = perr.WithMessage(err, "Lua")
		errs = errors.Join(errs, err)
	}

	irKey := t.getIRKey(gateway.Gateway)
	// Should exist since we've validated this
	x := xdsIR[irKey]

	policyTarget := irStringKey(policy.Namespace, string(target.Name))

	for _, http := range x.HTTP {
		gatewayName := http.Name[0:strings.LastIndex(http.Name, "/")]
		if t.MergeGateways && gatewayName != policyTarget {
			continue
		}

		// A Policy targeting the most specific scope(xRoute) wins over a policy
		// targeting a lesser specific scope(Gateway).
		for _, r := range http.Routes {
			// if already set - there's a route level policy, so skip
			if r.EnvoyExtensions != nil {
				continue
			}

			// return 500 and do not configure EnvoyExtensions in this case
			if errs != nil {
				r.DirectResponse = &ir.CustomResponse{
					StatusCode: ptr.To(uint32(500)),
				}
				continue
			}

			r.EnvoyExtensions = &ir.EnvoyExtensionFeatures{
				ExtProcs: extProcs,
				Wasms:    wasms,
				Luas:     luas,
			}
		}
	}

	return errs
}

func (t *Translator) buildLuas(policy *egv1a1.EnvoyExtensionPolicy, resources *resource.Resources, envoyProxy *egv1a1.EnvoyProxy) ([]ir.Lua, error) {
	var luaIRList []ir.Lua

	if policy == nil {
		return nil, nil
	}

	for idx, ep := range policy.Spec.Lua {
		name := irConfigNameForLua(policy, idx)
		luaIR, err := t.buildLua(name, policy, ep, resources, envoyProxy)
		if err != nil {
			return nil, err
		}
		luaIRList = append(luaIRList, *luaIR)
	}
	return luaIRList, nil
}

func (t *Translator) buildLua(
	name string,
	policy *egv1a1.EnvoyExtensionPolicy,
	lua egv1a1.Lua,
	resources *resource.Resources,
	envoyProxy *egv1a1.EnvoyProxy,
) (*ir.Lua, error) {
	var luaCode *string
	var luaValidation egv1a1.LuaValidation
	var err error
	if lua.Type == egv1a1.LuaValueTypeValueRef {
		luaCode, err = getLuaBodyFromLocalObjectReference(lua.ValueRef, resources, policy.Namespace)
	} else {
		luaCode = lua.Inline
	}
	if err != nil {
		return nil, err
	}
	if envoyProxy != nil && envoyProxy.Spec.LuaValidation != nil {
		luaValidation = *envoyProxy.Spec.LuaValidation
	} else {
		luaValidation = egv1a1.LuaValidationStrict
	}
	if err = luavalidator.NewLuaValidator(*luaCode, luaValidation).Validate(); err != nil {
		return nil, fmt.Errorf("validation failed for lua body in policy with name %v: %w", name, err)
	}
	return &ir.Lua{
		Name: name,
		Code: luaCode,
	}, nil
}

// getLuaBodyFromLocalObjectReference assumes the local object reference points to a Kubernetes ConfigMap
func getLuaBodyFromLocalObjectReference(valueRef *gwapiv1.LocalObjectReference, resources *resource.Resources, policyNs string) (*string, error) {
	cm := resources.GetConfigMap(policyNs, string(valueRef.Name))
	if cm != nil {
		b, dataOk := cm.Data["lua"]
		switch {
		case dataOk:
			return &b, nil
		case len(cm.Data) > 0: // Fallback to the first key if lua is not found
			for _, value := range cm.Data {
				b = value
				break
			}
			return &b, nil
		default:
			return nil, fmt.Errorf("can't find the key lua in the referenced configmap %s", valueRef.Name)
		}

	} else {
		return nil, fmt.Errorf("can't find the referenced configmap %s in namespace %s", valueRef.Name, policyNs)
	}
}

func (t *Translator) buildExtProcs(policy *egv1a1.EnvoyExtensionPolicy, resources *resource.Resources, envoyProxy *egv1a1.EnvoyProxy) ([]ir.ExtProc, error) {
	var extProcIRList []ir.ExtProc

	if policy == nil {
		return nil, nil
	}

	for idx, ep := range policy.Spec.ExtProc {
		name := irConfigNameForExtProc(policy, idx)
		extProcIR, err := t.buildExtProc(name, policy, ep, idx, resources, envoyProxy)
		if err != nil {
			return nil, err
		}
		extProcIRList = append(extProcIRList, *extProcIR)
	}
	return extProcIRList, nil
}

func (t *Translator) buildExtProc(
	name string,
	policy *egv1a1.EnvoyExtensionPolicy,
	extProc egv1a1.ExtProc,
	extProcIdx int,
	resources *resource.Resources,
	envoyProxy *egv1a1.EnvoyProxy,
) (*ir.ExtProc, error) {
	var (
		rd        *ir.RouteDestination
		authority string
		err       error
	)

	if rd, err = t.translateExtServiceBackendRefs(policy, extProc.BackendRefs, ir.GRPC, resources, envoyProxy, "extproc", extProcIdx); err != nil {
		return nil, err
	}

	if extProc.BackendRefs[0].Port != nil {
		authority = fmt.Sprintf(
			"%s.%s:%d",
			extProc.BackendRefs[0].Name,
			NamespaceDerefOr(extProc.BackendRefs[0].Namespace, policy.Namespace),
			*extProc.BackendRefs[0].Port)
	} else {
		authority = fmt.Sprintf(
			"%s.%s",
			extProc.BackendRefs[0].Name,
			NamespaceDerefOr(extProc.BackendRefs[0].Namespace, policy.Namespace))
	}

	traffic, err := translateTrafficFeatures(extProc.BackendSettings)
	if err != nil {
		return nil, err
	}

	extProcIR := &ir.ExtProc{
		Name:        name,
		Destination: *rd,
		Traffic:     traffic,
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

			if extProc.ProcessingMode.Request.Attributes != nil {
				extProcIR.RequestAttributes = append(extProcIR.RequestAttributes, extProc.ProcessingMode.Request.Attributes...)
			}
		}

		if extProc.ProcessingMode.Response != nil {
			extProcIR.ResponseHeaderProcessing = true
			if extProc.ProcessingMode.Response.Body != nil {
				extProcIR.ResponseBodyProcessingMode = ptr.To(ir.ExtProcBodyProcessingMode(*extProc.ProcessingMode.Response.Body))
			}

			if extProc.ProcessingMode.Response.Attributes != nil {
				extProcIR.ResponseAttributes = append(extProcIR.ResponseAttributes, extProc.ProcessingMode.Response.Attributes...)
			}
		}
		extProcIR.AllowModeOverride = extProc.ProcessingMode.AllowModeOverride
	}

	if extProc.Metadata != nil {
		if extProc.Metadata.AccessibleNamespaces != nil {
			extProcIR.ForwardingMetadataNamespaces = append(extProcIR.ForwardingMetadataNamespaces,
				extProc.Metadata.AccessibleNamespaces...)
		}

		if extProc.Metadata.WritableNamespaces != nil {
			extProcIR.ReceivingMetadataNamespaces = append(extProcIR.ReceivingMetadataNamespaces,
				extProc.Metadata.WritableNamespaces...)
		}
	}

	return extProcIR, err
}

func irConfigNameForExtProc(policy *egv1a1.EnvoyExtensionPolicy, index int) string {
	return fmt.Sprintf(
		"%s/extproc/%s",
		irConfigName(policy),
		strconv.Itoa(index))
}

func irConfigNameForLua(policy *egv1a1.EnvoyExtensionPolicy, index int) string {
	return fmt.Sprintf(
		"%s/lua/%s",
		irConfigName(policy),
		strconv.Itoa(index))
}

func (t *Translator) buildWasms(
	policy *egv1a1.EnvoyExtensionPolicy,
	resources *resource.Resources,
) ([]ir.Wasm, error) {
	var wasmIRList []ir.Wasm

	if len(policy.Spec.Wasm) == 0 {
		return wasmIRList, nil
	}

	if t.WasmCache == nil {
		return nil, fmt.Errorf("wasm cache is not initialized")
	}

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
	resources *resource.Resources,
) (*ir.Wasm, error) {
	var (
		failOpen   = false
		code       *ir.HTTPWasmCode
		pullPolicy wasm.PullPolicy
		// the checksum provided by the user, it's used to validate the wasm module
		// downloaded from the original HTTP server or the OCI registry
		originalChecksum string
		servingURL       string // the wasm module download URL from the EG HTTP server
		err              error
	)

	if config.FailOpen != nil {
		failOpen = *config.FailOpen
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
		var checksum string

		// This is a sanity check, the validation should have caught this
		if config.Code.HTTP == nil {
			return nil, fmt.Errorf("missing HTTP field in Wasm code source")
		}

		if config.Code.HTTP.SHA256 != nil {
			originalChecksum = *config.Code.HTTP.SHA256
		}

		http := config.Code.HTTP

		if servingURL, checksum, err = t.WasmCache.Get(http.URL, wasm.GetOptions{
			Checksum:        originalChecksum,
			PullPolicy:      pullPolicy,
			ResourceName:    irConfigNameForWasm(policy, idx),
			ResourceVersion: policy.ResourceVersion,
		}); err != nil {
			return nil, err
		}

		code = &ir.HTTPWasmCode{
			ServingURL:  servingURL,
			OriginalURL: http.URL,
			SHA256:      checksum,
		}

	case egv1a1.ImageWasmCodeSourceType:
		var (
			image      = config.Code.Image
			secret     *corev1.Secret
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
				kind:      resource.KindEnvoyExtensionPolicy,
				namespace: policy.Namespace,
			}

			if secret, err = t.validateSecretRef(
				false, from, *image.PullSecretRef, resources); err != nil {
				return nil, err
			}

			if data, ok := secret.Data[corev1.DockerConfigJsonKey]; ok {
				pullSecret = data
			} else {
				return nil, fmt.Errorf("missing %s key in secret %s/%s", corev1.DockerConfigJsonKey, secret.Namespace, secret.Name)
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

		if config.Code.Image.SHA256 != nil {
			originalChecksum = *config.Code.Image.SHA256
		}

		// The wasm checksum is different from the OCI image digest.
		// The original checksum in the EEP is used to match the digest of OCI image.
		// The returned checksum from the cache is the checksum of the wasm file
		// extracted from the OCI image, which is used by the envoy to verify the wasm file.
		if servingURL, checksum, err = t.WasmCache.Get(imageURL, wasm.GetOptions{
			Checksum:        originalChecksum,
			PullSecret:      pullSecret,
			PullPolicy:      pullPolicy,
			ResourceName:    irConfigNameForWasm(policy, idx),
			ResourceVersion: policy.ResourceVersion,
		}); err != nil {
			return nil, err
		}

		code = &ir.HTTPWasmCode{
			ServingURL:  servingURL,
			SHA256:      checksum,
			OriginalURL: imageURL,
		}
	default:
		// should never happen because of kubebuilder validation, just a sanity check
		return nil, fmt.Errorf("unsupported Wasm code source type %q", config.Code.Type)
	}

	wasmName := name
	if config.Name != nil {
		wasmName = *config.Name
	}
	wasmIR := &ir.Wasm{
		Name:     name,
		RootID:   config.RootID,
		WasmName: wasmName,
		Config:   config.Config,
		FailOpen: failOpen,
		Code:     code,
	}

	if config.Env != nil && len(config.Env.HostKeys) > 0 {
		wasmIR.HostKeys = config.Env.HostKeys
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
