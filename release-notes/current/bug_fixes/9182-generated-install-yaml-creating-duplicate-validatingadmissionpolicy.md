Fixed the generated `install.yaml` creating a duplicate ValidatingAdmissionPolicy and its binding which caused `kustomize build` to fail with a duplicate resource error.
