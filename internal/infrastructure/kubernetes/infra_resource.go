// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	klabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/proxy"
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
			namespaceLabel.Value(r.Namespace()),
		}
	)

	defer func() {
		if err == nil {
			resourceApplyDurationSeconds.With(labels...).Record(time.Since(startTime).Seconds())
			resourceApplyTotal.WithSuccess(labels...).Increment()
		} else {
			resourceApplyTotal.WithFailure(metrics.ReasonError, labels...).Increment()
		}

		if sa != nil {
			deleteErr := i.Client.DeleteAllExcept(ctx, &corev1.ServiceAccountList{}, client.ObjectKey{
				Namespace: sa.Namespace,
				Name:      sa.Name,
			}, &client.ListOptions{
				Namespace:     sa.Namespace,
				LabelSelector: r.LabelSelector(),
			})

			if deleteErr != nil {
				i.logger.Error(deleteErr, "failed to delete all except serviceaccount", "name", sa.Name)
			}
		}
	}()

	if sa, err = r.ServiceAccount(); err != nil {
		resourceApplyTotal.WithFailure(metrics.ReasonError, labels...).Increment()
		return err
	}

	return i.Client.ServerSideApply(ctx, sa)
}

// createOrUpdateConfigMap creates a ConfigMap in the Kube api server based on the provided
// ResourceRender, if it doesn't exist and updates it if it does.
func (i *Infra) createOrUpdateConfigMap(ctx context.Context, r ResourceRender) (err error) {
	var caCert string
	if i.EnvoyGateway.GatewayNamespaceMode() {
		caCert = i.getEnvoyGatewayCA(ctx)
	}

	var (
		cm        *corev1.ConfigMap
		startTime = time.Now()
		labels    = []metrics.LabelValue{
			kindLabel.Value("ConfigMap"),
			nameLabel.Value(r.Name()),
			namespaceLabel.Value(r.Namespace()),
		}
	)

	if cm, err = r.ConfigMap(caCert); err != nil {
		resourceApplyTotal.WithFailure(metrics.StatusFailure, labels...).Increment()
		return err
	}

	if cm == nil {
		return nil
	}

	defer func() {
		if err == nil {
			resourceApplyDurationSeconds.With(labels...).Record(time.Since(startTime).Seconds())
			resourceApplyTotal.WithSuccess(labels...).Increment()
		} else {
			resourceApplyTotal.WithFailure(metrics.ReasonError, labels...).Increment()
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
			namespaceLabel.Value(r.Namespace()),
		}
	)

	if deployment, err = r.Deployment(); err != nil {
		resourceApplyTotal.WithFailure(metrics.ReasonError, labels...).Increment()
		return err
	}

	// delete the deployment and return early
	// this handles the case where a daemonset has been
	// configured.
	if deployment == nil {
		return i.deleteDeployment(ctx, r)
	}

	defer func() {
		deleteErr := i.Client.DeleteAllExcept(ctx, &appsv1.DeploymentList{}, client.ObjectKey{
			Namespace: deployment.Namespace,
			Name:      deployment.Name,
		}, &client.ListOptions{
			Namespace:     deployment.Namespace,
			LabelSelector: r.LabelSelector(),
		})
		if deleteErr != nil {
			i.logger.Error(deleteErr, "failed to delete all except deployment", "name", r.Name())
		}

		if err == nil {
			resourceApplyDurationSeconds.With(labels...).Record(time.Since(startTime).Seconds())
			resourceApplyTotal.WithSuccess(labels...).Increment()
		} else {
			resourceApplyTotal.WithFailure(metrics.ReasonError, labels...).Increment()
		}
	}()

	old := &appsv1.Deployment{}
	err = i.Client.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, old)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// It's the deployment creation.
			return i.Client.ServerSideApply(ctx, deployment)
		}
		return err
	}

	if !equality.Semantic.DeepEqual(old.Spec.Selector, deployment.Spec.Selector) {
		// Note: Deployment created by the old gateway controller may have a different selector generated based on a custom label feature,
		// and it caused the issue that the gateway controller cannot update the deployment when users change the custom labels.
		// Therefore, we changed the gateway to always use the same selector, independent of the custom labels -
		// https://github.com/envoyproxy/gateway/issues/1818
		//
		// But, the change could break an existing deployment with custom labels initiated by the old gateway controller
		// because the selector would be different.
		//
		// Here, as a workaround, we always copy the selector from the old deployment to the new deployment
		// so that the update can be always applied successfully.
		deployment.Spec.Selector = old.Spec.Selector

		match, err := isSelectorMatch(deployment.Spec.Selector, deployment.Spec.Template.Labels)
		if err != nil {
			return err
		}
		if !match {
			// If the selector now doesn't match with labels of the pod template, return an error.
			// It could happen, for example, when users changed the custom label from {"foo": "bar"} to {"foo": "barv2"}
			// because the pod's labels have {"foo": "barv2"} while the selector keeps {"foo": "bar"}.
			// We cannot help this case, and just error it out.
			// In this case, users should recreate the envoy proxy with the new custom label, instead of upgrading it.
			// Once they recreate the envoy proxy, the envoy gateway of this version doesn't generate the selector based on the custom label,
			// and the issue won't happen again, even if they have to the custom label again.
			return fmt.Errorf("an illegal change in a custom label of EnvoyProxy is detected when updating %s/%s. The custom label config of deployment in EnvoyProxy, which is initiated with the envoy gateway of v1.1 or earlier, is immutable. Please recreate an envoy proxy with a new custom label if you need to change the custom label. This issue won't happen with the envoy proxy resource initialized by the envoygateway v1.2 or later", deployment.Namespace, deployment.Name)
		}
	}

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
			namespaceLabel.Value(r.Namespace()),
		}
	)

	if daemonSet, err = r.DaemonSet(); err != nil {
		resourceApplyTotal.WithFailure(metrics.ReasonError, labels...).Increment()
		return err
	}

	// delete the DaemonSet and return early
	// this handles the case where a deployment has been
	// configured.
	if daemonSet == nil {
		return i.deleteDaemonSet(ctx, r)
	}

	defer func() {
		deleteErr := i.Client.DeleteAllExcept(ctx, &appsv1.DaemonSetList{}, client.ObjectKey{
			Namespace: daemonSet.Namespace,
			Name:      daemonSet.Name,
		}, &client.ListOptions{
			Namespace:     daemonSet.Namespace,
			LabelSelector: r.LabelSelector(),
		})
		if deleteErr != nil {
			i.logger.Error(deleteErr, "failed to delete all except deployment", "name", r.Name())
		}

		if err == nil {
			resourceApplyDurationSeconds.With(labels...).Record(time.Since(startTime).Seconds())
			resourceApplyTotal.WithSuccess(labels...).Increment()
		} else {
			resourceApplyTotal.WithFailure(metrics.ReasonError, labels...).Increment()
		}
	}()

	old := &appsv1.DaemonSet{}
	err = i.Client.Get(ctx, types.NamespacedName{Name: daemonSet.Name, Namespace: daemonSet.Namespace}, old)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// It's the daemonset creation.
			return i.Client.ServerSideApply(ctx, daemonSet)
		}
		return err
	}

	if !equality.Semantic.DeepEqual(old.Spec.Selector, daemonSet.Spec.Selector) {
		// Note: Daemonset created by the old gateway controller may have a different selector generated based on a custom label feature,
		// and it caused the issue that the gateway controller cannot update the daemonset when users change the custom labels.
		// Therefore, we changed the gateway to always use the same selector, independent of the custom labels -
		// https://github.com/envoyproxy/gateway/issues/1818
		//
		// But, the change could break an existing daemonset with custom labels initiated by the old gateway controller
		// because the selector would be different.
		//
		// Here, as a workaround, we always copy the selector from the old daemonset to the new daemonset
		// so that the update can be always applied successfully.
		daemonSet.Spec.Selector = old.Spec.Selector
		match, err := isSelectorMatch(daemonSet.Spec.Selector, daemonSet.Spec.Template.Labels)
		if err != nil {
			return err
		}
		if !match {
			// If the selector now doesn't match with labels of the pod template, return an error.
			// It could happen, for example, when users changed the custom label from {"foo": "bar"} to {"foo": "barv2"}
			// because the pod's labels have {"foo": "barv2"} while the selector keeps {"foo": "bar"}.
			// We cannot help this case, and just error it out.
			// In this case, users should recreate the envoy proxy with the new custom label, instead of upgrading it.
			// Once they recreate the envoy proxy, the envoy gateway of this version doesn't generate the selector based on the custom label,
			// and the issue won't happen again, even if they have to the custom label again.
			return fmt.Errorf("an illegal change in a custom label of EnvoyProxy is detected when updating %s/%s. The custom label config of daemonset in EnvoyProxy, which is initiated with the envoy gateway of v1.1 or earlier, is immutable. Please recreate an envoy proxy with a new custom label if you need to change the custom label. This issue won't happen with the envoy proxy resource initialized by the envoygateway v1.2 or later", daemonSet.Namespace, daemonSet.Name)
		}
	}

	return i.Client.ServerSideApply(ctx, daemonSet)
}

func isSelectorMatch(labelselector *metav1.LabelSelector, l map[string]string) (bool, error) {
	selector, err := metav1.LabelSelectorAsSelector(labelselector)
	if err != nil {
		return false, fmt.Errorf("invalid label selector is generated: %w", err)
	}

	return selector.Matches(klabels.Set(l)), nil
}

func (i *Infra) createOrUpdatePodDisruptionBudget(ctx context.Context, r ResourceRender) (err error) {
	var (
		pdb       *policyv1.PodDisruptionBudget
		startTime = time.Now()
		labels    = []metrics.LabelValue{
			kindLabel.Value("PDB"),
			nameLabel.Value(r.Name()),
			namespaceLabel.Value(r.Namespace()),
		}
	)

	defer func() {
		if err == nil {
			resourceApplyDurationSeconds.With(labels...).Record(time.Since(startTime).Seconds())
			resourceApplyTotal.WithSuccess(labels...).Increment()
		} else {
			resourceApplyTotal.WithFailure(metrics.ReasonError, labels...).Increment()
		}

		if pdb != nil {
			deleteErr := i.Client.DeleteAllExcept(ctx, &policyv1.PodDisruptionBudgetList{}, client.ObjectKey{
				Namespace: pdb.Namespace,
				Name:      pdb.Name,
			}, &client.ListOptions{
				Namespace:     pdb.Namespace,
				LabelSelector: r.LabelSelector(),
			})
			if deleteErr != nil {
				i.logger.Error(deleteErr, "failed to delete all except PodDisruptionBudget",
					"name", r.Name(), "namespace", r.Namespace())
			}
		}
	}()

	if pdb, err = r.PodDisruptionBudget(); err != nil {
		resourceApplyTotal.WithFailure(metrics.ReasonError, labels...).Increment()
		return err
	}

	// when pdb is not set,
	// then delete the object in the kube api server if got any.
	if pdb == nil {
		return i.deletePDB(ctx, r)
	}

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
			namespaceLabel.Value(r.Namespace()),
		}
	)

	defer func() {
		if err == nil {
			resourceApplyDurationSeconds.With(labels...).Record(time.Since(startTime).Seconds())
			resourceApplyTotal.WithSuccess(labels...).Increment()
		} else {
			resourceApplyTotal.WithFailure(metrics.ReasonError, labels...).Increment()
		}

		if hpa != nil {
			deleteErr := i.Client.DeleteAllExcept(ctx, &autoscalingv2.HorizontalPodAutoscalerList{}, client.ObjectKey{
				Namespace: hpa.Namespace,
				Name:      hpa.Name,
			}, &client.ListOptions{
				Namespace:     hpa.Namespace,
				LabelSelector: r.LabelSelector(),
			})
			if deleteErr != nil {
				i.logger.Error(deleteErr, "failed to delete all except HorizontalPodAutoscaler",
					"name", r.Name(), "namespace", r.Namespace())
			}
		}
	}()

	if hpa, err = r.HorizontalPodAutoscaler(); err != nil {
		return err
	}

	// when HorizontalPodAutoscaler is not set,
	// then delete the object in the kube api server if got any.
	if hpa == nil {
		return i.deleteHPA(ctx, r)
	}

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
			namespaceLabel.Value(r.Namespace()),
		}
	)

	if svc, err = r.Service(); err != nil {
		resourceApplyTotal.WithFailure(metrics.ReasonError, labels...).Increment()
		return err
	}

	defer func() {
		deleteErr := i.Client.DeleteAllExcept(ctx, &corev1.ServiceList{}, client.ObjectKey{
			Namespace: svc.Namespace,
			Name:      svc.Name,
		}, &client.ListOptions{
			Namespace:     svc.Namespace,
			LabelSelector: r.LabelSelector(),
		})
		if deleteErr != nil {
			i.logger.Error(deleteErr, "failed to delete all except deployment", "name", r.Name())
		}

		if err == nil {
			resourceApplyDurationSeconds.With(labels...).Record(time.Since(startTime).Seconds())
			resourceApplyTotal.WithSuccess(labels...).Increment()
		} else {
			resourceApplyTotal.WithFailure(metrics.ReasonError, labels...).Increment()
		}
	}()

	return i.Client.ServerSideApply(ctx, svc)
}

// deleteServiceAccount deletes the ServiceAccount in the kube api server, if it exists.
func (i *Infra) deleteServiceAccount(ctx context.Context, r ResourceRender) (err error) {
	var (
		name, ns = r.Name(), r.Namespace()
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

	defer func() {
		if err == nil {
			resourceDeleteDurationSeconds.With(labels...).Record(time.Since(startTime).Seconds())
			resourceDeleteTotal.WithSuccess(labels...).Increment()
		} else {
			resourceDeleteTotal.WithFailure(metrics.ReasonError, labels...).Increment()
		}
	}()

	return i.Client.DeleteAllOf(ctx, sa, &client.DeleteAllOfOptions{
		ListOptions: client.ListOptions{
			Namespace:     ns,
			LabelSelector: r.LabelSelector(),
		},
	})
}

// deleteDeployment deletes the Envoy Deployment in the kube api server, if it exists.
func (i *Infra) deleteDeployment(ctx context.Context, r ResourceRender) (err error) {
	var (
		name, ns   = r.Name(), r.Namespace()
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

	defer func() {
		if err == nil {
			resourceDeleteDurationSeconds.With(labels...).Record(time.Since(startTime).Seconds())
			resourceDeleteTotal.WithSuccess(labels...).Increment()
		} else {
			resourceDeleteTotal.WithFailure(metrics.ReasonError, labels...).Increment()
		}
	}()

	return i.Client.DeleteAllOf(ctx, deployment, &client.DeleteAllOfOptions{
		ListOptions: client.ListOptions{
			Namespace:     ns,
			LabelSelector: r.LabelSelector(),
		},
	})
}

// deleteDaemonSet deletes the Envoy DaemonSet in the kube api server, if it exists.
func (i *Infra) deleteDaemonSet(ctx context.Context, r ResourceRender) (err error) {
	var (
		name, ns  = r.Name(), r.Namespace()
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

	defer func() {
		if err == nil {
			resourceDeleteDurationSeconds.With(labels...).Record(time.Since(startTime).Seconds())
			resourceDeleteTotal.WithSuccess(labels...).Increment()
		} else {
			resourceDeleteTotal.WithFailure(metrics.ReasonError, labels...).Increment()
		}
	}()

	return i.Client.DeleteAllOf(ctx, daemonSet, &client.DeleteAllOfOptions{
		ListOptions: client.ListOptions{
			Namespace:     ns,
			LabelSelector: r.LabelSelector(),
		},
	})
}

// deleteConfigMap deletes the ConfigMap in the kube api server, if it exists.
func (i *Infra) deleteConfigMap(ctx context.Context, r ResourceRender) (err error) {
	var (
		name, ns = r.Name(), r.Namespace()
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

	defer func() {
		if err == nil {
			resourceDeleteDurationSeconds.With(labels...).Record(time.Since(startTime).Seconds())
			resourceDeleteTotal.WithSuccess(labels...).Increment()
		} else {
			resourceDeleteTotal.WithFailure(metrics.ReasonError, labels...).Increment()
		}
	}()

	return i.Client.DeleteAllOf(ctx, cm, &client.DeleteAllOfOptions{
		ListOptions: client.ListOptions{
			Namespace:     ns,
			LabelSelector: r.LabelSelector(),
		},
	})
}

// deleteService deletes the Service in the kube api server, if it exists.
func (i *Infra) deleteService(ctx context.Context, r ResourceRender) (err error) {
	var (
		name, ns = r.Name(), r.Namespace()
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

	defer func() {
		if err == nil {
			resourceDeleteDurationSeconds.With(labels...).Record(time.Since(startTime).Seconds())
			resourceDeleteTotal.WithSuccess(labels...).Increment()
		} else {
			resourceDeleteTotal.WithFailure(metrics.ReasonError, labels...).Increment()
		}
	}()

	return i.Client.DeleteAllOf(ctx, svc, &client.DeleteAllOfOptions{
		ListOptions: client.ListOptions{
			Namespace:     ns,
			LabelSelector: r.LabelSelector(),
		},
	})
}

// deleteHpa deletes the Horizontal Pod Autoscaler associated to its renderer, if it exists.
func (i *Infra) deleteHPA(ctx context.Context, r ResourceRender) (err error) {
	var (
		name, ns = r.Name(), r.Namespace()
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

	defer func() {
		if err == nil {
			resourceDeleteDurationSeconds.With(labels...).Record(time.Since(startTime).Seconds())
			resourceDeleteTotal.WithSuccess(labels...).Increment()
		} else {
			resourceDeleteTotal.WithFailure(metrics.ReasonError, labels...).Increment()
		}
	}()

	return i.Client.DeleteAllOf(ctx, hpa, &client.DeleteAllOfOptions{
		ListOptions: client.ListOptions{
			Namespace:     ns,
			LabelSelector: r.LabelSelector(),
		},
	})
}

// deletePDB deletes the PodDistribution budget associated to its renderer, if it exists.
func (i *Infra) deletePDB(ctx context.Context, r ResourceRender) (err error) {
	var (
		name, ns = r.Name(), r.Namespace()
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

	defer func() {
		if err == nil {
			resourceDeleteDurationSeconds.With(labels...).Record(time.Since(startTime).Seconds())
			resourceDeleteTotal.WithSuccess(labels...).Increment()
		} else {
			resourceDeleteTotal.WithFailure(metrics.ReasonError, labels...).Increment()
		}
	}()

	return i.Client.DeleteAllOf(ctx, pdb, &client.DeleteAllOfOptions{
		ListOptions: client.ListOptions{
			Namespace:     ns,
			LabelSelector: r.LabelSelector(),
		},
	})
}

func (i *Infra) getEnvoyGatewayCA(ctx context.Context) string {
	secret := &corev1.Secret{}
	err := i.Client.Get(ctx, types.NamespacedName{
		Name:      "envoy",
		Namespace: i.ControllerNamespace,
	}, secret)
	if err != nil {
		return ""
	}
	return string(secret.Data[proxy.XdsTLSCaFileName])
}
