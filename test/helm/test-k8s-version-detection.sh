#!/usr/bin/env bash
# Copyright Envoy Gateway Authors
# SPDX-License-Identifier: Apache-2.0
# The full text of the Apache license is available in the LICENSE file at
# the root of the repo.

set -euo pipefail

CHART_DIR="charts/gateway-helm"
CHART_NAME="envoy-gateway"

echo "Testing K8s version detection for topology injector..."
echo ""

# Test 1: K8s 1.34 - topology injector should be enabled by default
echo "Test 1: K8s 1.34 - topology injector enabled by default"
OUTPUT=$(helm template ${CHART_NAME} ${CHART_DIR} --kube-version 1.34.0)
if echo "$OUTPUT" | grep -q "kind: MutatingWebhookConfiguration"; then
  echo "✓ PASS: Webhook found on K8s 1.34"
else
  echo "✗ FAIL: Webhook not found on K8s 1.34"
  exit 1
fi
echo ""

# Test 2: K8s 1.35 - topology injector should be auto-disabled
echo "Test 2: K8s 1.35 - topology injector auto-disabled"
OUTPUT=$(helm template ${CHART_NAME} ${CHART_DIR} --kube-version 1.35.0)
if ! echo "$OUTPUT" | grep -q "kind: MutatingWebhookConfiguration"; then
  echo "✓ PASS: Webhook not found on K8s 1.35 (auto-disabled)"
else
  echo "✗ FAIL: Webhook found on K8s 1.35 (should be auto-disabled)"
  exit 1
fi
echo ""

# Test 3: K8s 1.34 with explicit disable
echo "Test 3: K8s 1.34 with explicit topologyInjector.enabled=false"
OUTPUT=$(helm template ${CHART_NAME} ${CHART_DIR} --kube-version 1.34.0 --set topologyInjector.enabled=false)
if ! echo "$OUTPUT" | grep -q "kind: MutatingWebhookConfiguration"; then
  echo "✓ PASS: Webhook not found when explicitly disabled on K8s 1.34"
else
  echo "✗ FAIL: Webhook found when explicitly disabled on K8s 1.34"
  exit 1
fi
echo ""

# Test 4: Verify EnvoyGateway config has correct proxyTopologyInjector.disabled setting on K8s 1.35
echo "Test 4: Verify EnvoyGateway config on K8s 1.35"
OUTPUT=$(helm template ${CHART_NAME} ${CHART_DIR} --kube-version 1.35.0 --show-only templates/envoy-gateway-config.yaml)
if echo "$OUTPUT" | grep -q "disabled: true"; then
  echo "✓ PASS: proxyTopologyInjector.disabled found in config on K8s 1.35"
else
  echo "✗ FAIL: proxyTopologyInjector.disabled not found in config on K8s 1.35"
  exit 1
fi
echo ""

# Test 5: Verify webhook port not in deployment on K8s 1.35
echo "Test 5: Verify webhook port excluded from deployment on K8s 1.35"
OUTPUT=$(helm template ${CHART_NAME} ${CHART_DIR} --kube-version 1.35.0 --show-only templates/envoy-gateway-deployment.yaml)
if ! echo "$OUTPUT" | grep -q "containerPort: 9443"; then
  echo "✓ PASS: Webhook port 9443 not in deployment on K8s 1.35"
else
  echo "✗ FAIL: Webhook port 9443 found in deployment on K8s 1.35"
  exit 1
fi
echo ""

# Test 6: Verify webhook port is in deployment on K8s 1.34
echo "Test 6: Verify webhook port included in deployment on K8s 1.34"
OUTPUT=$(helm template ${CHART_NAME} ${CHART_DIR} --kube-version 1.34.0 --show-only templates/envoy-gateway-deployment.yaml)
if echo "$OUTPUT" | grep -q "containerPort: 9443"; then
  echo "✓ PASS: Webhook port 9443 found in deployment on K8s 1.34"
else
  echo "✗ FAIL: Webhook port 9443 not found in deployment on K8s 1.34"
  exit 1
fi
echo ""

# Test 7: Verify certgen args include --disable-topology-injector on K8s 1.35
echo "Test 7: Verify certgen args on K8s 1.35"
OUTPUT=$(helm template ${CHART_NAME} ${CHART_DIR} --kube-version 1.35.0 --show-only templates/certgen.yaml)
if echo "$OUTPUT" | grep -q "\\- --disable-topology-injector"; then
  echo "✓ PASS: Certgen args include --disable-topology-injector on K8s 1.35"
else
  echo "✗ FAIL: Certgen args missing --disable-topology-injector on K8s 1.35"
  exit 1
fi
echo ""

echo "========================================="
echo "All tests passed! ✓"
echo "========================================="
