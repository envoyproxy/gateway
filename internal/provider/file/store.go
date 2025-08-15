// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package file

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"sort"
	"time"

	"github.com/go-logr/logr"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/message"
)

const (
	GatewayDeletionOrder = 3
)

type ResourcesStore struct {
	name      string
	keys      sets.Set[storeKey]
	client    client.Client
	resources *message.ProviderResources
	reconcile chan int64

	logger logr.Logger
}

type storeKey struct {
	schema.GroupVersionKind
	types.NamespacedName

	// deletionOrder is used to determine the order in which a resource is deleted.
	// The larger the value, the earlier it is deleted.
	deletionOrder int
}

func (s storeKey) String() string {
	return fmt.Sprintf("%s/%s/%d",
		s.GroupVersionKind.String(), s.NamespacedName.String(), s.deletionOrder)
}

func NewResourcesStore(name string, client client.Client, resources *message.ProviderResources, logger logr.Logger) *ResourcesStore {
	return &ResourcesStore{
		name:      name,
		keys:      sets.New[storeKey](),
		client:    client,
		resources: resources,
		reconcile: make(chan int64),
		logger:    logger,
	}
}

func newStoreKey(obj client.Object) storeKey {
	return storeKey{
		GroupVersionKind: obj.GetObjectKind().GroupVersionKind(),
		NamespacedName:   client.ObjectKeyFromObject(obj),
	}
}

// ReloadAll loads and stores all resources from all given files and directories.
func (r *ResourcesStore) ReloadAll(ctx context.Context, files, dirs []string) error {
	// TODO(sh2): add arbitrary number of resources support for load function.
	resources, err := loadFromFilesAndDirs(files, dirs)
	if err != nil {
		return err
	}

	var errList error
	currentKeys, err := r.Store(ctx, resources, false)
	if err != nil {
		errList = errors.Join(errList, err)
	}

	// If no resources were created or updated, stop reconciling.
	if errList != nil && len(currentKeys) == 0 {
		return errList
	}

	// Remove the resources that no longer exist.
	rn := 0
	deletedKeys := r.keys.Difference(currentKeys)
	for _, k := range deletionOrderKeyList(deletedKeys) {
		delObj := makeUnstructuredObjectFromKey(k)
		if err := r.client.Delete(ctx, delObj); err != nil {
			errList = errors.Join(errList, err)
			// Insert back if the object is not be removed.
			currentKeys.Insert(k)
		} else if k.deletionOrder <= GatewayDeletionOrder {
			// Reconcile once if gateway got deleted, this may be able to
			// remove the finalizer on gatewayclass.
			r.reconcile <- generateReconcileID()
			rn++
		}
	}

	r.keys = currentKeys
	r.reconcile <- generateReconcileID()
	rn++

	r.logger.Info("reload resources finished",
		"reload_resources_num", len(r.keys), "reconcile_times", rn, "time", time.Now())
	return errList
}

// Store stores resources via offline gateway-api client.
// For file provider, storeService set to false, means all gateway-api resources will be stored except:
// - Service
// - ServiceImport
// - EndpointSlices
// Becasues these resources has no effects on the host infra layer.
func (r *ResourcesStore) Store(ctx context.Context, re *resource.LoadResources, storeService bool) (sets.Set[storeKey], error) {
	if re == nil {
		return nil, nil
	}

	var (
		errs        error
		collectKeys = sets.New[storeKey]()
	)

	for _, obj := range re.GatewayClasses {
		if err := r.storeObjectWithKeys(ctx, obj, collectKeys); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	for _, obj := range re.EnvoyProxies {
		if err := r.storeObjectWithKeys(ctx, obj, collectKeys); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	for _, obj := range re.Gateways {
		if err := r.storeObjectWithKeys(ctx, obj, collectKeys); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	for _, obj := range re.HTTPRoutes {
		if err := r.storeObjectWithKeys(ctx, obj, collectKeys); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	for _, obj := range re.GRPCRoutes {
		if err := r.storeObjectWithKeys(ctx, obj, collectKeys); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	for _, obj := range re.TLSRoutes {
		if err := r.storeObjectWithKeys(ctx, obj, collectKeys); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	for _, obj := range re.TCPRoutes {
		if err := r.storeObjectWithKeys(ctx, obj, collectKeys); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	for _, obj := range re.UDPRoutes {
		if err := r.storeObjectWithKeys(ctx, obj, collectKeys); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	for _, obj := range re.ReferenceGrants {
		if err := r.storeObjectWithKeys(ctx, obj, collectKeys); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	for _, obj := range re.Namespaces {
		if err := r.storeObjectWithKeys(ctx, obj, collectKeys); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	for _, obj := range re.Secrets {
		if err := r.storeObjectWithKeys(ctx, obj, collectKeys); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	for _, obj := range re.ConfigMaps {
		if err := r.storeObjectWithKeys(ctx, obj, collectKeys); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	for _, obj := range re.EnvoyPatchPolicies {
		if err := r.storeObjectWithKeys(ctx, obj, collectKeys); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	for _, obj := range re.ClientTrafficPolicies {
		if err := r.storeObjectWithKeys(ctx, obj, collectKeys); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	for _, obj := range re.BackendTrafficPolicies {
		if err := r.storeObjectWithKeys(ctx, obj, collectKeys); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	for _, obj := range re.SecurityPolicies {
		if err := r.storeObjectWithKeys(ctx, obj, collectKeys); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	for _, obj := range re.BackendTLSPolicies {
		if err := r.storeObjectWithKeys(ctx, obj, collectKeys); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	for _, obj := range re.EnvoyExtensionPolicies {
		if err := r.storeObjectWithKeys(ctx, obj, collectKeys); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	for _, obj := range re.Backends {
		if err := r.storeObjectWithKeys(ctx, obj, collectKeys); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	for _, obj := range re.HTTPRouteFilters {
		if err := r.storeObjectWithKeys(ctx, obj, collectKeys); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	if storeService {
		for _, obj := range re.Services {
			if err := r.storeObjectWithKeys(ctx, obj, collectKeys); err != nil {
				errs = errors.Join(errs, err)
			}
		}

		for _, obj := range re.ServiceImports {
			if err := r.storeObjectWithKeys(ctx, obj, collectKeys); err != nil {
				errs = errors.Join(errs, err)
			}
		}

		for _, obj := range re.EndpointSlices {
			if err := r.storeObjectWithKeys(ctx, obj, collectKeys); err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}

	return collectKeys, errs
}

// storeObjectWithKeys stores object while collecting its key.
func (r *ResourcesStore) storeObjectWithKeys(ctx context.Context, obj client.Object, keys sets.Set[storeKey]) error {
	key, err := r.storeObject(ctx, obj)
	if err != nil && key != nil {
		return fmt.Errorf("failed to store %s %s: %w", key.Kind, key.NamespacedName.String(), err)
	} else if err != nil {
		return fmt.Errorf("failed to store object: %w", err)
	}

	if key != nil {
		keys.Insert(*key)
	}

	return nil
}

// storeObject will do create for non-exist object and update for existing object.
func (r *ResourcesStore) storeObject(ctx context.Context, obj client.Object) (*storeKey, error) {
	if obj == nil || reflect.ValueOf(obj).IsNil() {
		return nil, nil
	}

	var (
		err    error
		key    = newStoreKey(obj)
		oldObj = makeUnstructuredObjectFromKey(key)
	)

	if err = r.client.Get(ctx, key.NamespacedName, oldObj); err == nil {
		return &key, r.client.Patch(ctx, obj, client.Merge)
	}
	if kerrors.IsNotFound(err) {
		return &key, r.client.Create(ctx, obj)
	}

	return nil, err
}

func makeUnstructuredObjectFromKey(key storeKey) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(key.GroupVersionKind)
	obj.SetNamespace(key.Namespace)
	obj.SetName(key.Name)
	return obj
}

// deletionOrderKeyList returns a list sorted in descending order by deletionOrder in its key.
func deletionOrderKeyList(keys sets.Set[storeKey]) []storeKey {
	out := keys.UnsortedList()
	for i, k := range out {
		switch k.Kind {
		case resource.KindNamespace, resource.KindReferenceGrant,
			resource.KindConfigMap, resource.KindSecret:
			out[i].deletionOrder = GatewayDeletionOrder - 3

		case resource.KindEnvoyProxy:
			out[i].deletionOrder = GatewayDeletionOrder - 2

		case resource.KindGatewayClass:
			out[i].deletionOrder = GatewayDeletionOrder - 1

		case resource.KindGateway:
			out[i].deletionOrder = GatewayDeletionOrder

		case resource.KindHTTPRoute, resource.KindGRPCRoute,
			resource.KindTLSRoute, resource.KindTCPRoute, resource.KindUDPRoute,
			resource.KindSecurityPolicy, resource.KindClientTrafficPolicy, resource.KindBackendTrafficPolicy,
			resource.KindEnvoyPatchPolicy, resource.KindEnvoyExtensionPolicy, resource.KindBackendTLSPolicy:
			out[i].deletionOrder = GatewayDeletionOrder + 1

		case resource.KindBackend, resource.KindHTTPRouteFilter:
			out[i].deletionOrder = GatewayDeletionOrder + 2

		default:
			out[i].deletionOrder = GatewayDeletionOrder + 3
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].deletionOrder > out[j].deletionOrder
	})
	return out
}

func generateReconcileID() int64 {
	n, _ := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	return n.Int64()
}
