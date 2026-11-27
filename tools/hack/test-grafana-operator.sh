#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# Run only against a disposable cluster: Helm installs cluster-scoped CRDs.
# Requires Go, kubectl, curl and jq. The Make target supplies a fresh kind cluster.
helm() { go tool -modfile=tools/go.mod helm "$@"; }
readonly namespace=grafana-operator-test
readonly workdir=$(mktemp -d)
port_forward_pid=""
cleanup() {
  if [[ -n "$port_forward_pid" ]]; then
    kill "$port_forward_pid" 2>/dev/null || true
    wait "$port_forward_pid" 2>/dev/null || true
  fi
  rm -rf "$workdir"
}
trap cleanup EXIT

helm install grafana-operator oci://ghcr.io/grafana/helm-charts/grafana-operator \
  --version v5.19.0 --namespace "$namespace" --create-namespace --wait --timeout 5m
kubectl apply -n "$namespace" -f - <<'EOF'
apiVersion: grafana.integreatly.org/v1beta1
kind: Grafana
metadata:
  name: test
  labels:
    dashboards: grafana
    environment: production
spec:
  config:
    auth.anonymous:
      enabled: "true"
      org_role: Viewer
    security:
      admin_user: admin
      admin_password: test-password
EOF

helm repo add prometheus-community https://prometheus-community.github.io/helm-charts --force-update
helm repo add grafana https://grafana.github.io/helm-charts --force-update
helm repo add fluent https://fluent.github.io/helm-charts --force-update
helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts --force-update
helm dependency build charts/gateway-addons-helm
helm install eg-addons charts/gateway-addons-helm --namespace "$namespace" \
  -f test/helm/gateway-addons-helm/grafana-operator.in.yaml

# Check admission, reconciliation and the imported JSON through Grafana's API.
kubectl wait -n "$namespace" --for=create deployment/test-deployment --timeout=180s
kubectl rollout status -n "$namespace" deployment/test-deployment --timeout=300s
kubectl wait grafanadashboards --all -n "$namespace" \
  --for=condition=DashboardSynchronized --timeout=300s
kubectl port-forward -n "$namespace" service/test-service :3000 >"$workdir/port-forward.log" 2>&1 &
port_forward_pid=$!
for ((attempt=0; attempt<30; attempt++)); do
  port=$(sed -n 's/.*127.0.0.1:\([0-9]*\) -> 3000/\1/p' "$workdir/port-forward.log" | head -1)
  [[ -n "$port" ]] && break
  sleep 1
done
test -n "$port"
kubectl get grafanadashboards -n "$namespace" -o json >"$workdir/dashboards.json"
jq -e '.items | length == 5' "$workdir/dashboards.json"
while IFS= read -r uid; do
  curl --fail --silent --show-error --retry 60 --retry-all-errors --retry-delay 2 \
    --max-time 5 "http://127.0.0.1:$port/api/dashboards/uid/$uid" >"$workdir/dashboard.json"
  jq -e --arg uid "$uid" '.dashboard.uid == $uid and .meta.folderTitle == "123"' "$workdir/dashboard.json"
done < <(jq -r '.items[].spec.json | fromjson | .uid' "$workdir/dashboards.json")

# Removing the expression's matching label must prevent reconciliation.
kubectl label grafana test -n "$namespace" environment=staging --overwrite
# A metadata-only update is filtered out by the Operator's watch predicate.
# Change the spec to trigger reconciliation without waiting for the default resync.
while IFS= read -r name; do
  kubectl patch grafanadashboard "$name" -n "$namespace" --type=merge \
    -p '{"spec":{"resyncPeriod":"10s"}}'
done < <(jq -r '.items[].metadata.name' "$workdir/dashboards.json")
kubectl wait grafanadashboards --all -n "$namespace" \
  --for=jsonpath='{.status.NoMatchingInstances}'=true --timeout=180s
