// Package v1 contains API Schema definitions for the exampleorg.io API group.
//
// +kubebuilder:object:generate=true
// +groupName=exampleorg.io
package v1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	GroupVersion  = schema.GroupVersion{Group: "exampleorg.io", Version: "v1"}
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}
	AddToScheme   = SchemeBuilder.AddToScheme
)

func init() {
	SchemeBuilder.Register(&GlobalLuaScript{}, &GlobalLuaScriptList{})
}
