package ir

import (
	cfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
)

// Infra defines managed infrastructure.
type Infra struct {
	// Proxy defines managed proxy infrastructure.
	Proxy *ProxyInfra
}

// ProxyInfra defines managed proxy infrastructure.
type ProxyInfra struct {
	// TODO: Figure out how to represent metadata in the IR.
	// xref: https://github.com/envoyproxy/gateway/issues/173
	//
	// Name is the name used for managed proxy infrastructure.
	Name string
	// Namespace is the namespace used for managed proxy infrastructure.
	Namespace string
	// OwnerRefs is a list of objects used to generate this object.
	OwnerRefs []ObjectReference
	// Config defines user-facing configuration of the managed proxy infrastructure.
	Config *cfgv1a1.EnvoyProxy
	// CertificateRefs is a list of object references that contain TLS certificates
	// and private keys. These certificates are used to establish secure communication
	// between the control and data plane.
	CertificateRefs []LocalObjectReference
	// Image is the container image used for the managed proxy infrastructure.
	Image string
}

// ObjectReference contains enough information to let you identify an object in
// any namespace.
type ObjectReference struct {
	// Group is the referent group. For more information, see:
	// https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md
	Group string
	// Version is the referent version. For more information, see:
	// https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md
	Version string
	// Kind is the referent kind. For more information, see:
	// https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md
	Kind string
	// Name is the name of the referent.
	Name string
	// Namespace is the namespace of the referent.
	Namespace string
}

// LocalObjectReference contains enough information to let you identify a local object.
type LocalObjectReference struct {
	// Group is the referent group. For more information, see:
	// https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md
	Group string
	// Version is the referent version. For more information, see:
	// https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md
	Version string
	// Kind is the referent kind. For more information, see:
	// https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md
	Kind string
	// Name is the name of the referent.
	Name string
}
