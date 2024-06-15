// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/envoyproxy/gateway/internal/metrics"
)

// createOrUpdateServiceAccount creates a ServiceAccount in the kube api server based on the
// provided ResourceRender, if it doesn't exist and updates it if it does.
func (i *Infra) createOrUpdateServiceAccount(ctx context.Context, r ResourceRender) (err error) {
	var (
		sa        *corev1.ServiceAccount
		startTime = time.Now()
		labels    = []metrics.LabelValue{
			kindLabel.Value("ServiceAccount"),
			nameLabel.Value(r.Name()),
			namespaceLabel.Value(i.Namespace),
		}
	)

	resourceApplyTotal.With(labels...).Increment()

	if sa, err = r.ServiceAccount(); err != nil {
		resourceApplyFailed.With(labels...).Increment()

		return err
	}

	defer func() {
		if err == nil {
			resourceApplyDurationSeconds.With(labels...).Record(time.Since(startTime).Seconds())
			resourceApplySuccess.With(labels...).Increment()
		} else {
			resourceApplyFailed.With(labels...).Increment()
		}
	}()

	return i.Client.ServerSideApply(ctx, sa)
}

// createOrUpdateConfigMap creates a ConfigMap in the Kube api server based on the provided
// ResourceRender, if it doesn't exist and updates it if it does.
func (i *Infra) createOrUpdateConfigMap(ctx context.Context, r ResourceRender) (err error) {
	var (
		cm        *corev1.ConfigMap
		startTime = time.Now()
		labels    = []metrics.LabelValue{
			kindLabel.Value("ConfigMap"),
			nameLabel.Value(r.Name()),
			namespaceLabel.Value(i.Namespace),
		}
	)

	resourceApplyTotal.With(labels...).Increment()

	if cm, err = r.ConfigMap(); err != nil {
		resourceApplyFailed.With(labels...).Increment()

		return err
	}

	if cm == nil {
		return nil
	}

	defer func() {
		if err == nil {
			resourceApplyDurationSeconds.With(labels...).Record(time.Since(startTime).Seconds())
			resourceApplySuccess.With(labels...).Increment()
		} else {
			resourceApplyFailed.With(labels...).Increment()
		}
	}()

	return i.Client.ServerSideApply(ctx, cm)
}

// createOrUpdateDeployment creates a Deployment in the kube api server based on the provided
// ResourceRender, if it doesn't exist and updates it if it does.
func (i *Infra) createOrUpdateDeployment(ctx context.Context, r ResourceRender) (err error) {
	var (
		deployment *appsv1.Deployment
		startTime  = time.Now()
		labels     = []metrics.LabelValue{
			kindLabel.Value("Deployment"),
			nameLabel.Value(r.Name()),
			namespaceLabel.Value(i.Namespace),
		}
	)

	resourceApplyTotal.With(labels...).Increment()

	if deployment, err = r.Deployment(); err != nil {
		resourceApplyFailed.With(labels...).Increment()

		return err
	}

	// delete the deployment and return early
	// this handles the case where a daemonset has been
	// configured.
	if deployment == nil {
		return i.deleteDeployment(ctx, r)
	}

	defer func() {
		if err == nil {
			resourceApplyDurationSeconds.With(labels...).Record(time.Since(startTime).Seconds())
			resourceApplySuccess.With(labels...).Increment()
		} else {
			resourceApplyFailed.With(labels...).Increment()
		}
	}()

	return i.Client.ServerSideApply(ctx, deployment)
}

// createOrUpdateDaemonSet creates a DaemonSet in the kube api server based on the provided
// ResourceRender, if it doesn't exist and updates it if it does.
func (i *Infra) createOrUpdateDaemonSet(ctx context.Context, r ResourceRender) (err error) {
	var (
		daemonSet *appsv1.DaemonSet
		startTime = time.Now()
		labels    = []metrics.LabelValue{
			kindLabel.Value("DaemonSet"),
			nameLabel.Value(r.Name()),
			namespaceLabel.Value(i.Namespace),
		}
	)

	resourceApplyTotal.With(labels...).Increment()

	if daemonSet, err = r.DaemonSet(); err != nil {
		resourceApplyFailed.With(labels...).Increment()

		return err
	}

	// delete the daemonset and return early
	// this handles the case where a deployment has been
	// configured.
	if daemonSet == nil {
		return i.deleteDaemonSet(ctx, r)
	}

	defer func() {
		if err == nil {
			resourceApplyDurationSeconds.With(labels...).Record(time.Since(startTime).Seconds())
			resourceApplySuccess.With(labels...).Increment()
		} else {
			resourceApplyFailed.With(labels...).Increment()
		}
	}()

	return i.Client.ServerSideApply(ctx, daemonSet)
}

func (i *Infra) createOrUpdatePodDisruptionBudget(ctx context.Context, r ResourceRender) (err error) {
	var (
		pdb       *policyv1.PodDisruptionBudget
		startTime = time.Now()
		labels    = []metrics.LabelValue{
			kindLabel.Value("PDB"),
			nameLabel.Value(r.Name()),
			namespaceLabel.Value(i.Namespace),
		}
	)

	resourceApplyTotal.With(labels...).Increment()

	if pdb, err = r.PodDisruptionBudget(); err != nil {
		resourceApplyFailed.With(labels...).Increment()
		return err
	}

	// when pdb is not set,
	// then delete the object in the kube api server if got any.
	if pdb == nil {
		return i.deletePDB(ctx, r)
	}

	defer func() {
		if err == nil {
			resourceApplyDurationSeconds.With(labels...).Record(time.Since(startTime).Seconds())
			resourceApplySuccess.With(labels...).Increment()
		} else {
			resourceApplyFailed.With(labels...).Increment()
		}
	}()

	return i.Client.ServerSideApply(ctx, pdb)
}

// createOrUpdateHPA creates HorizontalPodAutoscaler object in the kube api server based on
// the provided ResourceRender, if it doesn't exist and updates it if it does,
// and delete hpa if not set.
func (i *Infra) createOrUpdateHPA(ctx context.Context, r ResourceRender) (err error) {
	var (
		hpa       *autoscalingv2.HorizontalPodAutoscaler
		startTime = time.Now()
		labels    = []metrics.LabelValue{
			kindLabel.Value("HPA"),
			nameLabel.Value(r.Name()),
			namespaceLabel.Value(i.Namespace),
		}
	)

	resourceApplyTotal.With(labels...).Increment()

	if hpa, err = r.HorizontalPodAutoscaler(); err != nil {
		resourceApplyFailed.With(labels...).Increment()

		return err
	}

	// when HorizontalPodAutoscaler is not set,
	// then delete the object in the kube api server if got any.
	if hpa == nil {
		return i.deleteHPA(ctx, r)
	}

	defer func() {
		if err == nil {
			resourceApplyDurationSeconds.With(labels...).Record(time.Since(startTime).Seconds())
			resourceApplySuccess.With(labels...).Increment()
		} else {
			resourceApplyFailed.With(labels...).Increment()
		}
	}()

	return i.Client.ServerSideApply(ctx, hpa)
}

// createOrUpdateRateLimitService creates a Service in the kube api server based on the provided ResourceRender,
// if it doesn't exist or updates it if it does.
func (i *Infra) createOrUpdateService(ctx context.Context, r ResourceRender) (err error) {
	var (
		svc       *corev1.Service
		startTime = time.Now()
		labels    = []metrics.LabelValue{
			kindLabel.Value("Service"),
			nameLabel.Value(r.Name()),
			namespaceLabel.Value(i.Namespace),
		}
	)

	resourceApplyTotal.With(labels...).Increment()

	if svc, err = r.Service(); err != nil {
		resourceApplyFailed.With(labels...).Increment()

		return err
	}

	defer func() {
		if err == nil {
			resourceApplyDurationSeconds.With(labels...).Record(time.Since(startTime).Seconds())
			resourceApplySuccess.With(labels...).Increment()
		} else {
			resourceApplyFailed.With(labels...).Increment()
		}
	}()

	return i.Client.ServerSideApply(ctx, svc)
}

// deleteServiceAccount deletes the ServiceAccount in the kube api server, if it exists.
func (i *Infra) deleteServiceAccount(ctx context.Context, r ResourceRender) (err error) {
	var (
		name, ns = r.Name(), i.Namespace
		sa       = &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      name,
			},
		}
		startTime = time.Now()
		labels    = []metrics.LabelValue{
			kindLabel.Value("ServiceAccount"),
			nameLabel.Value(name),
			namespaceLabel.Value(ns),
		}
	)

	resourceDeleteTotal.With(labels...).Increment()

	defer func() {
		if err == nil {
			resourceDeleteDurationSeconds.With(labels...).Record(time.Since(startTime).Seconds())
			resourceDeleteSuccess.With(labels...).Increment()
		} else {
			resourceDeleteFailed.With(labels...).Increment()
		}
	}()

	return i.Client.Delete(ctx, sa)
}

// deleteDeployment deletes the Envoy Deployment in the kube api server, if it exists.
func (i *Infra) deleteDeployment(ctx context.Context, r ResourceRender) (err error) {
	var (
		name, ns   = r.Name(), i.Namespace
		deployment = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      name,
			},
		}
		startTime = time.Now()
		labels    = []metrics.LabelValue{
			kindLabel.Value("Deployment"),
			nameLabel.Value(name),
			namespaceLabel.Value(ns),
		}
	)

	resourceDeleteTotal.With(labels...).Increment()

	defer func() {
		if err == nil {
			resourceDeleteDurationSeconds.With(labels...).Record(time.Since(startTime).Seconds())
			resourceDeleteSuccess.With(labels...).Increment()
		} else {
			resourceDeleteFailed.With(labels...).Increment()
		}
	}()

	return i.Client.Delete(ctx, deployment)
}

// deleteDaemonSet deletes the Envoy DaemonSet in the kube api server, if it exists.
func (i *Infra) deleteDaemonSet(ctx context.Context, r ResourceRender) (err error) {
	var (
		name, ns  = r.Name(), i.Namespace
		daemonSet = &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      name,
			},
		}
		startTime = time.Now()
		labels    = []metrics.LabelValue{
			kindLabel.Value("DaemonSet"),
			nameLabel.Value(name),
			namespaceLabel.Value(ns),
		}
	)

	resourceDeleteTotal.With(labels...).Increment()

	defer func() {
		if err == nil {
			resourceDeleteDurationSeconds.With(labels...).Record(time.Since(startTime).Seconds())
			resourceDeleteSuccess.With(labels...).Increment()
		} else {
			resourceDeleteFailed.With(labels...).Increment()
		}
	}()

	return i.Client.Delete(ctx, daemonSet)
}

// deleteConfigMap deletes the ConfigMap in the kube api server, if it exists.
func (i *Infra) deleteConfigMap(ctx context.Context, r ResourceRender) (err error) {
	var (
		name, ns = r.Name(), i.Namespace
		cm       = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      name,
			},
		}
		startTime = time.Now()
		labels    = []metrics.LabelValue{
			kindLabel.Value("ConfigMap"),
			nameLabel.Value(name),
			namespaceLabel.Value(ns),
		}
	)

	resourceDeleteTotal.With(labels...).Increment()

	defer func() {
		if err == nil {
			resourceDeleteDurationSeconds.With(labels...).Record(time.Since(startTime).Seconds())
			resourceDeleteSuccess.With(labels...).Increment()
		} else {
			resourceDeleteFailed.With(labels...).Increment()
		}
	}()

	return i.Client.Delete(ctx, cm)
}

// deleteService deletes the Service in the kube api server, if it exists.
func (i *Infra) deleteService(ctx context.Context, r ResourceRender) (err error) {
	var (
		name, ns = r.Name(), i.Namespace
		svc      = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      name,
			},
		}
		startTime = time.Now()
		labels    = []metrics.LabelValue{
			kindLabel.Value("Service"),
			nameLabel.Value(name),
			namespaceLabel.Value(ns),
		}
	)

	resourceDeleteTotal.With(labels...).Increment()

	defer func() {
		if err == nil {
			resourceDeleteDurationSeconds.With(labels...).Record(time.Since(startTime).Seconds())
			resourceDeleteSuccess.With(labels...).Increment()
		} else {
			resourceDeleteFailed.With(labels...).Increment()
		}
	}()

	return i.Client.Delete(ctx, svc)
}

// deleteHpa deletes the Horizontal Pod Autoscaler associated to its renderer, if it exists.
func (i *Infra) deleteHPA(ctx context.Context, r ResourceRender) (err error) {
	var (
		name, ns = r.Name(), i.Namespace
		hpa      = &autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      name,
			},
		}
		startTime = time.Now()
		labels    = []metrics.LabelValue{
			kindLabel.Value("HPA"),
			nameLabel.Value(name),
			namespaceLabel.Value(ns),
		}
	)

	resourceDeleteTotal.With(labels...).Increment()

	defer func() {
		if err == nil {
			resourceDeleteDurationSeconds.With(labels...).Record(time.Since(startTime).Seconds())
			resourceDeleteSuccess.With(labels...).Increment()
		} else {
			resourceDeleteFailed.With(labels...).Increment()
		}
	}()

	return i.Client.Delete(ctx, hpa)
}

// deletePDB deletes the PodDistribution budget associated to its renderer, if it exists.
func (i *Infra) deletePDB(ctx context.Context, r ResourceRender) (err error) {
	var (
		name, ns = r.Name(), i.Namespace
		pdb      = &policyv1.PodDisruptionBudget{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      name,
			},
		}
		startTime = time.Now()
		labels    = []metrics.LabelValue{
			kindLabel.Value("PDB"),
			nameLabel.Value(name),
			namespaceLabel.Value(ns),
		}
	)

	resourceDeleteTotal.With(labels...).Increment()

	defer func() {
		if err == nil {
			resourceDeleteDurationSeconds.With(labels...).Record(time.Since(startTime).Seconds())
			resourceDeleteSuccess.With(labels...).Increment()
		} else {
			resourceDeleteFailed.With(labels...).Increment()
		}
	}()

	return i.Client.Delete(ctx, pdb)
}
