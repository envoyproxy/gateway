package v1

import (
	"github.com/exampleorg/envoygateway-extension/internal/ir"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GlobalLuaScript defines a lua script that should be run on all requests
//
// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=gateway
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=luascript
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].reason`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type GlobalLuaScript struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired user configuration for a GlobalLuaScript
	// +kubebuilder:validation:Required
	Spec GlobalLuaScriptSpec `json:"spec"`

	// Status defines the current state of a lua script that provides user feedback on
	// whether it is configured correctly or not.
	Status GlobalLuaScriptStatus `json:"status,omitempty"`
}

// List of global lua scripts
//
// +kubebuilder:object:root=true
type GlobalLuaScriptList struct {
	metav1.TypeMeta `json:""`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GlobalLuaScript `json:"items"`
}

// List of conditions for a lua script that describe its state
type GlobalLuaScriptStatus struct {
	// Conditions describe the current state of the lua script
	//
	// +kubebuilder:validation:Optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=8
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// GlobalLuaScriptConditionReason is the type the lua script's condition
type GlobalLuaScriptConditionType string

func (t GlobalLuaScriptConditionType) String() string {
	return string(t)
}

// GlobalLuaScriptConditionReason is the reason for a condition on a lua script
type GlobalLuaScriptConditionReason string

func (r GlobalLuaScriptConditionReason) String() string {
	return string(r)
}

const (
	// If you have other conditions/reasons, etc. add them here
	GlobalLuaScriptConditionReady   GlobalLuaScriptConditionType   = "Ready"
	GlobalLuaScriptReasonGoodConfig GlobalLuaScriptConditionReason = "ValidConfiguration"
	GlobalLuaScriptReasonBadConfig  GlobalLuaScriptConditionReason = "InvalidConfiguration"
)

// Defines the lua code to be run against all requests
type GlobalLuaScriptSpec struct {
	// Raw lua code to be run on all requests
	// +kubebuilder:validation:Required
	Lua string `json:"lua"`
}

func (lf *GlobalLuaScript) ToIR() *ir.GlobalLuaScript {
	return &ir.GlobalLuaScript{
		Metadata: ir.NewMetadata().
			AddObjectMeta(lf.ObjectMeta).
			AddGroupVersion(GroupVersion),
		Lua: lf.Spec.Lua,
	}
}
