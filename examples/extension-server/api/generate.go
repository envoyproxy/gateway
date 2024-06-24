//go:generate go run sigs.k8s.io/controller-tools/cmd/controller-gen@v0.14.0 object:headerFile="boilerplate.go.txt" rbac:roleName=manager-role crd paths="../api/..." output:crd:artifacts:config=../crds/generated
package api
