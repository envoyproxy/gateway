package ir

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Metadata struct {
	Name       string
	Namespace  string
	APIVersion string
}

func NewMetadata() *Metadata {
	return &Metadata{}
}

func (m *Metadata) ID() string {
	return fmt.Sprintf("%s.%s", m.Namespace, m.Name)
}

func (m *Metadata) AddObjectMeta(meta metav1.ObjectMeta) *Metadata {
	m.Name = meta.Name
	m.Namespace = meta.Namespace
	return m
}

func (m *Metadata) AddGroupVersion(gv schema.GroupVersion) *Metadata {
	m.APIVersion = gv.Identifier()
	return m
}
