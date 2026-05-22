{{- if .Values.gatewayAPI.validatingAdmissionPolicy.enabled }}
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
{{- /*
Render the Gateway API safe-upgrade ValidatingAdmissionPolicy only when
.Values.gatewayAPI.validatingAdmissionPolicy.enabled is true. Keep the
lookups behind that same guard so disabling the policy does not require RBAC
for admissionregistration resources.

During upgrades, also require existing cluster-scoped policy resources to be
absent or already owned by this Helm release so Helm does not adopt or
overwrite resources managed by another installation.
*/ -}}
{{- if or (not .Release.IsUpgrade) (and $vapOwnedOrAbsent $vapBindingOwnedOrAbsent) }}
