#!/bin/zsh

unset TYPE_SPEED

. ./demo/demo-magic.sh

alias kubectl=kubecolor
kubectl delete CustomBackendMtlsPolicy --all -A
clear

pe "kubens default"
echo ""

pe "kubectl get pods"
echo ""

pe "kubectl get services"
echo ""

pe "kubectl get httproutes"
echo ""

pe "kubectl describe httproutes"
echo ""

wait
clear
pe 'curl --header "Host: www.example.com" http://localhost:8888/get'
echo ""

wait
clear
pe 'kubectl apply -f ./config/custom-backend-mtls-policy.yaml'
echo ""

pe 'kubectl apply -f ./config/extension-manager-with-listeners-routes.yaml'
echo ""

pe 'helm upgrade --install -n envoy-gateway-system extension-server ./charts/extension-server'
echo ""

kubectl rollout restart deployments/extension-server -n envoy-gateway-system > /dev/null 2>&1

wait
clear
pe 'kubens envoy-gateway-system'
echo ""

pe 'kubectl get pods'
echo ""

kubectl rollout restart deployments/backend -n default > /dev/null 2>&1

wait
clear
pe 'kubectl logs deployment/extension-server --all-pods=true'
echo ""

wait
clear
pe 'curl --header "Host: www.example.com" http://localhost:8888/get'
echo ""
