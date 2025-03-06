---
title: "Fuzzing "
---

## Overview

This design document introduces **Fuzz Testing** in Envoy Gateway.
Its goal is to detect unexpected crashes, memory leaks, and undefined behaviors in the translation of Gateway 
API resources and configuration parsing, areas that may not be fully covered by unit tests. 

Additionally, [OSS-Fuzz](https://github.com/google/oss-fuzz) 
provides a continuous fuzzing infrastructure for popular open-source projects. By writing fuzz tests, 
we can leverage OSS-Fuzz to continuously fuzz Envoy Gateway against a wide range of inputs, 
improving its resilience and reliability.

**Note:** This work is sponsored by the 
[Linux Foundation Mentorship](https://mentorship.lfx.linuxfoundation.org/project/44020e81-1218-49aa-95e0-ee3e03998eb3) 
program.

## Goals

* Identify fuzz targets considering the tradeoff between realism, efficiency, and maintenance effort.
* Develop a set of initial fuzzers.
* Integrate with OSS-Fuzz for continuous, automated fuzz testing.
* Iteratively refine, update, and enhance fuzzers based on findings and evolving requirements.


## Non Goals

* Introducing new features to Envoy Gateway.
* Achieving 100% code coverage.

## Implementation

As the fuzzers will be integrated with OSS-Fuzz, the implementation will follow best practices 
outlined in [OSS-Fuzz Ideal Integration](https://google.github.io/oss-fuzz/advanced-topics/ideal-integration/) page.
Fuzzers will be developed native Go fuzzing library [go-fuzz](https://go.dev/blog/fuzz-beta).

### Example
Here is an example of a simple fuzzer that tests the translation of Gateway API resource to Intermediate 
representation (IR).

```go
func FuzzGatewayAPIToIRWithGatewayClass(f *testing.F) {
	f.Add("valid-name", "valid-namespace", "gateway.envoyproxy.io/gatewayclass-controller")
	f.Fuzz(func(t *testing.T, name, namespace, controllerName string) {
		resources := &resource.Resources{
			GatewayClass: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: gwapiv1.GatewayController(controllerName),
				},
			},
		}

		result, err := translateGatewayAPIToIR(resources)
		require.NoError(t, err, "Error should be nil")
		require.NotNil(t, result, "Result should not be nil")
		require.NotNil(t, result.Resources, "Resources should not be nil")
		require.NotNil(t, result.XdsIR, "XdsIR should not be nil")
		require.NotNil(t, result.InfraIR, "InfraIR should not be nil")
	})
}
```

## Design Decisions
* Fuzzing library: [go-fuzz](https://go.dev/blog/fuzz-beta).
* Fuzzers directory: **TODO:**

## Conclusion

Fuzz testing is an ongoing process. Once the initial fuzzers are developed, 
crashes reported by OSS-Fuzz will be continuously monitored, and the fuzzers will be iteratively refined.
Additionally, future efforts will be directed toward exploring the integration of fuzz testing into the CI pipeline using GitHub Actions.