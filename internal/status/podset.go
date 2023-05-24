// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	appsv1 "k8s.io/api/apps/v1"
)

// A PodSet is an abstraction over Deployment and DaemonSet.
type PodSet interface {
	// Kind returns the resource kind.
	Kind() string

	// AvailableReplicas returns how many instances are actually available.
	AvailableReplicas() int32

	// DesiredReplicas returns how many instances should be running.
	DesiredReplicas() int32
}

// DeploymentPodSet is a PodSet based on a Deployment.
type DeploymentPodSet appsv1.Deployment

func (pset *DeploymentPodSet) Kind() string { return (*appsv1.Deployment)(pset).TypeMeta.Kind }

func (pset *DeploymentPodSet) AvailableReplicas() int32 {
	return (*appsv1.Deployment)(pset).Status.AvailableReplicas
}
func (pset *DeploymentPodSet) DesiredReplicas() int32 {
	return (*appsv1.Deployment)(pset).Status.Replicas
}

// DaemonSetPodSet is a PodSet based on a DaemonSet.
type DaemonSetPodSet appsv1.DaemonSet

func (pset *DaemonSetPodSet) Kind() string { return (*appsv1.DaemonSet)(pset).TypeMeta.Kind }

func (pset *DaemonSetPodSet) AvailableReplicas() int32 {
	return (*appsv1.DaemonSet)(pset).Status.NumberAvailable
}
func (pset *DaemonSetPodSet) DesiredReplicas() int32 {
	return (*appsv1.DaemonSet)(pset).Status.DesiredNumberScheduled
}
