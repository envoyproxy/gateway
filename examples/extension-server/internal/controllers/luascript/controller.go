package luascript_controller

import (
	"context"
	"fmt"
	"time"

	v1 "github.com/exampleorg/envoygateway-extension/api/v1"
	"github.com/exampleorg/envoygateway-extension/internal/ir"

	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	RequeueAfterDuration = 15 * time.Second
)

// Handles watching and reconciling the GlobalLuaScript resources
type luaScriptController struct {
	client    client.Client
	irManager *ir.IRManager
}

// SetupWithManager watches and registers the controller to reconcile GlobalLuaScripts
func (c *luaScriptController) SetupWithManager(mgr manager.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("globalluascript").
		For(&v1.GlobalLuaScript{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 1}).
		Complete(c) // This binds the controller to the Reconciler implementation (our luaScriptController)

	// TODO: if your custom resources does things like references other secrets/custom-resources, you can add additional watches/indexers here
	// you'll also want to use a Watch() call in the above controller setup, but for this example the primary resource (GlobalLuaScript) doesn't watch or reference any other resources
}

// Reconcile implements the controller Reconcile function so that when GlobalLuaScript reconcile.Requests are
// received due to changes we can process the resources
func (c *luaScriptController) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log.Info().Str("namespace", request.Namespace).Str("name", request.Name).Msg("reconcile triggered by GlobalLuaScript")

	luaScriptKey := types.NamespacedName{
		Name:      request.Name,
		Namespace: request.Namespace,
	}

	luaScriptMetadata := ir.NewMetadata().
		AddObjectMeta(metav1.ObjectMeta{
			Name:      request.Name,
			Namespace: request.Namespace,
		}).
		AddGroupVersion(v1.GroupVersion)

	luaScript := &v1.GlobalLuaScript{}
	if err := c.client.Get(ctx, luaScriptKey, luaScript); err != nil {
		// If there is an error getting the resource, delete it from the IR
		c.irManager.DeleteLuaScript(luaScriptMetadata.ID())
		return reconcile.Result{}, nil
	}

	if !luaScript.GetDeletionTimestamp().IsZero() {
		// Remove resources from the IR on deletion
		c.irManager.DeleteLuaScript(luaScriptMetadata.ID())
		return reconcile.Result{}, nil
	}

	irLuaScript := luaScript.ToIR()

	// TODO: this example uses relatively simple conditions, but you might want to introduce a sub-reconcile func to
	// validate your resource and determine what conditions to set
	var conditions []metav1.Condition
	if irLuaScript.Lua != "" {
		conditions = []metav1.Condition{
			{
				Type:    string(v1.GlobalLuaScriptConditionReady),
				Reason:  string(v1.GlobalLuaScriptReasonGoodConfig),
				Status:  metav1.ConditionTrue,
				Message: "GlobalLuaScript is valid and ready",
			},
		}
	} else {
		conditions = []metav1.Condition{
			{
				Type:    string(v1.GlobalLuaScriptConditionReady),
				Reason:  string(v1.GlobalLuaScriptReasonBadConfig),
				Status:  metav1.ConditionFalse,
				Message: "GlobalLuaScript.spec.lua should not be an empty string",
			},
		}
	}

	luaScriptStatus := v1.GlobalLuaScriptStatus{
		Conditions: conditions,
	}

	if irLuaScript != nil {
		c.irManager.StoreLuaScript(irLuaScript)
	}

	// We don't want to update the status unless we've actually changed it otherwise we can end up
	// in an infinite reconcile loop
	newStatus, needsUpdate := getNewLuaScriptStatus(luaScript.Status, luaScriptStatus, luaScript.GetGeneration())

	// Update the status on the resource if it needs to be changed
	if needsUpdate {
		luaScript.Status = newStatus
		if err := c.client.Status().Update(ctx, luaScript); err != nil {
			return reconcile.Result{RequeueAfter: RequeueAfterDuration},
				fmt.Errorf("error updating status for GlobalLuaScript %s.%s: %w", luaScript.GetName(), luaScript.GetNamespace(), err)
		}
	}

	return reconcile.Result{}, nil
}

func NewController(client client.Client, irManager *ir.IRManager) *luaScriptController {
	return &luaScriptController{
		client:    client,
		irManager: irManager,
	}
}
