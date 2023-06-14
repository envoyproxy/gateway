# Fluent-bit


```
helm repo add fluent https://fluent.github.io/helm-charts
helm repo update
helm upgrade --install fluent-bit fluent/fluent-bit -f examples/fluent-bit/helm-values.yaml -n monitoring --create-namespace --version 0.30.4
```