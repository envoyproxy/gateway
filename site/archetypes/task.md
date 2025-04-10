---
title: "{{ replace .Name "-" " " | title }}"
---

This task provides instructions for configuring {{ replace .Name "-" " " | lower }}.

{{ replace .Name "-" " " | title }} allows you to [briefly explain what this feature does and its benefits].

## Prerequisites

{{< boilerplate prerequisites >}}

<!-- 
Use specific, action-oriented titles when writing conceptual docs. Focus on what the reader will learn or what the concept enables them to do. Use verbs and be descriptive!
Examples:
-"Applying Multiple BackendRefs"
-"Mirroring Traffic to Another Service"
-->
## [Replace Title] Configuration 

<!-- 
Briefly explain what is being configured.
-->
By applying the file below...

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
#insert stdin yaml
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
#Insert config file
```

{{% /tab %}}
{{< /tabpane >}}

## Verification

Verify the configuration:

```shell
#Insert command to verify installation
```

<!-- 
What is the expected output after running the command above?
-->
The output should include ...


## Testing

Ensure the `GATEWAY_HOST` environment variable from the [Quickstart](../../quickstart) is set. If not, follow the
Quickstart instructions to set the variable.

```shell
echo $GATEWAY_HOST
```

Send a request to the endpoint:

```shell
curl -v -H "Host: www.example.com" "http://${GATEWAY_HOST}/example"
```

You should see a successful response, confirming that your configuration is working correctly.

<!-- 
Include an example output below
-->

Expected output:
```
Insert output
...
```

## Clean-Up

Remove the resources created in this task:

```shell
# Add commands to remove resources created in configuration step
```

## Next Steps
<!-- 
Link any related pages from Envoy Gateway Docs
Example:
- [HTTPRoute Resource Reference](...)
- [Configuring a Gateway](...)
-->
Checkout the following guides:
- [INSERT PAGE NAME](INSERT_PATH_TO_FILE)
- ...
