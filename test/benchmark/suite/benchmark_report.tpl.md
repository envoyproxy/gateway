# Benchmark Report

Benchmark test settings:

| RPS      | Connections     | Duration (Seconds) | CPU Limits (m) | Memory Limits (MiB) |
|----------|-----------------|--------------------|----------------|---------------------|
| {{.RPS}} | {{.Connection}} | {{.Duration}}      | {{.CPU}}       | {{.Memory}}         |

## Test: ScaleHTTPRoute

Fixed one Gateway and different scales of HTTPRoutes with different portion of hostnames.


### Results

Expand to see the full results.

{{- range .Results }}
<details>
<summary>{{.Summary}}</summary>

```plaintext

{{.Text}}

```

</details>
{{- end }}

### Metrics

The CPU usage statistics of both control-plane and data-plane are the CPU usage per second over the past 30 seconds.

| Test Name | Envoy Gateway Container Memory (MiB) <br> min/max/means | Envoy Gateway Process Memory (MiB) <br> min/max/means | Envoy Gateway CPU (%) <br> min/max/means | Averaged Envoy Proxy Memory (MiB) <br> min/max/means | Averaged Envoy Proxy CPU (%) <br> min/max/means |
|-----------|---------------------------------------------------------|---------------------------------------------------------|------------------------------------------|------------------------------------------------------|-------------------------------------------------|
{{- range .Metrics }}
| {{.Name}} | {{.ControlPlaneContainerMemory}} | {{.ControlPlaneProcessMemory}} | {{.ControlPlaneCPU}} | {{.DataPlaneMemory}} | {{.DataPlaneCPU}} |  
{{- end }}

### Profiles

The profiles at different scales are stored under `/profiles` directory in report, with name `{ProfileType}.{TestCase}.pprof`.

You can visualize them in a web page by running:

```shell
go tool pprof -http=: path/to/your.pprof
```

Currently, the supported profile types are:
- heap (memory)


#### Heap

The profiles were sampled when Envoy Gateway Memory is at its maximum.


{{- range .Heaps }}
#### {{.Title}}.

![{{.Name}}]({{.URL}})

{{- end }}
