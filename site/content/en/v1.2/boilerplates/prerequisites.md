---
---

Follow the steps from the [Quickstart](../tasks/quickstart) task to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

Verify the Gateway status:

{{< tabpane text=true >}}
{{% tab header="kubectl" %}}

```shell
kubectl get gateway/eg -o yaml
```

{{% /tab %}}
{{% tab header="egctl (experimental)" %}}

```shell
egctl x status gateway -v
```

{{% /tab %}}
{{< /tabpane >}}
