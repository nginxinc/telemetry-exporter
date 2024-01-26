# Example

> Running collector in a cluster is shortcut for now. This will be simplified.

Create a cluster:
```text
kind create cluster --image kindest/node:v1.28.0
```

Install collector:
```text
helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts

helm install my-opentelemetry-collector open-telemetry/opentelemetry-collector -f values.yaml
```

Port-forward:
```
kubectl port-forward <pod-name-of-collector> 4317:4317
```

Run example, you should see no errors:
```
go run github.com/nginxinc/telemetry-exporter/example
```

Check collector logs:
```text
kubectl logs <pod-name-of-collector>
. . .
2024-01-26T17:11:19.673Z        info    TracesExporter  {"kind": "exporter", "data_type": "traces", "name": "debug", "resource spans": 1, "spans": 1}
2024-01-26T17:11:19.673Z        info    ResourceSpans #0
Resource SchemaURL: https://opentelemetry.io/schemas/1.24.0
Resource attributes:
     -> service.name: Str(NGF)
     -> service.version: Str(1.0)
     -> telemetry.sdk.language: Str(go)
     -> telemetry.sdk.name: Str(opentelemetry)
     -> telemetry.sdk.version: Str(1.22.0)
ScopeSpans #0
ScopeSpans SchemaURL:
InstrumentationScope product-telemetry
Span #0
    Trace ID       : 87ed447b1f88c419bdbfdb7ea53afe0d
    Parent ID      :
    ID             : 0d152cd82d85ab0c
    Name           : report
    Kind           : Internal
    Start time     : 2024-01-26 17:11:19.649212 +0000 UTC
    End time       : 2024-01-26 17:11:19.649269667 +0000 UTC
    Status code    : Unset
    Status message :
Attributes:
     -> ProductID: Str(NGF)
     -> ResourceCount: Int(1)
        {"kind": "exporter", "data_type": "traces", "name": "debug"}
```