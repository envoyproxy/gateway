Fixed the Wasm module serving URL and the Wasm HTTP service cluster hardcoding `cluster.local`; they now use the configured Kubernetes cluster DNS domain (`kubernetesClusterDomain`).
