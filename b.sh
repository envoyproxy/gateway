# loop apply 100 times with name by aa + i
for i in {51..100}; do kubectl apply -f - <<EOF
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: CustomGRPCRoute
metadata:
  name: sonny-route-$i
  namespace: qat-idp
spec:
  hostnames:
    - pigw.local.experimental.geocomply.com
  parentRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: shared-gw
      namespace: qat-infra
  rules:
    - backendRefs:
        - group: ''
          kind: Service
          name: sonny
          port: 8080
          weight: 1
      matches:
        - method:
            method: POST
            service: sonny.Sonny
            type: Exact
EOF
done
```

