{{- /*
Render Gateway API supporting resources only when
.Values.gatewayAPI.supportingResources.enabled is true. Supporting resources
are non-CRD resources shipped with the Gateway API bundle, such as the
safe-upgrades ValidatingAdmissionPolicy and binding.
*/ -}}
{{- if .Values.gatewayAPI.supportingResources.enabled }}
{{- $renderSafeUpgradePolicy := true -}}
{{- /*
Require existing Gateway API policy resources to be absent or already owned by this Helm release 
so Helm does not overwrite resources managed by another installation or by the cluster provider.
*/ -}}
{{- $safeUpgradePolicyName := "safe-upgrades.gateway.networking.k8s.io" -}}
{{- $vap := lookup "admissionregistration.k8s.io/v1" "ValidatingAdmissionPolicy" "" $safeUpgradePolicyName -}}
{{- $vapBinding := lookup "admissionregistration.k8s.io/v1" "ValidatingAdmissionPolicyBinding" "" $safeUpgradePolicyName -}}
{{- $vapOwned := and $vap
  (eq (dig "metadata" "labels" "app.kubernetes.io/managed-by" "" $vap) "Helm")
  (eq (dig "metadata" "annotations" "meta.helm.sh/release-name" "" $vap) .Release.Name)
  (eq (dig "metadata" "annotations" "meta.helm.sh/release-namespace" "" $vap) .Release.Namespace)
 -}}
{{- $vapBindingOwned := and $vapBinding
  (eq (dig "metadata" "labels" "app.kubernetes.io/managed-by" "" $vapBinding) "Helm")
  (eq (dig "metadata" "annotations" "meta.helm.sh/release-name" "" $vapBinding) .Release.Name)
  (eq (dig "metadata" "annotations" "meta.helm.sh/release-namespace" "" $vapBinding) .Release.Namespace)
 -}}
{{- $vapOwnedOrAbsent := or (not $vap) $vapOwned -}}
{{- $vapBindingOwnedOrAbsent := or (not $vapBinding) $vapBindingOwned -}}
{{- $renderSafeUpgradePolicy = and $vapOwnedOrAbsent $vapBindingOwnedOrAbsent -}}
{{- if $renderSafeUpgradePolicy }}
