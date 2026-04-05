---
title: "{{ replace .Name "-" " " | title }}"
---

## Overview

<!-- Provide a high-level introduction to the task, keeping it simple and user-friendly. What does this task accomplish? When should a user use this task, and how does it help them? -->
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

## Testing

Verify the configuration:

```shell
#Insert command to verify installation
```

<!-- 
What is the expected output after running the command above?
-->

<!-- 
Use the shortcode below to test basic configurations. For more advanced configurations, you will need to write your own tests. Please make sure to include tests for cases with and without an external load balancer. 

Take a look at site/content/en/latest/boilerplates/testing.md for an example. 
-->
{{< tabpane-include testing >}}

## Clean-Up

Remove the resources created in this task:

```shell
# Add commands to remove resources created in configuration step
```

## [Optional] Next Steps
<!-- 
Link any related pages from Envoy Gateway Docs
Example:
- [HTTPRoute Resource Reference](...)
- [Configuring a Gateway](...)
-->
Checkout the following guides:
- [INSERT PAGE NAME](INSERT_PATH_TO_FILE)
- ...
