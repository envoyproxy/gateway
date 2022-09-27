// Portions of this code are based on code from Contour, available at:
// https://github.com/projectcontour/contour/blob/main/internal/k8s/status.go

package status

import (
	"context"
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/internal/provider/utils"
)

func UpdateStatus(c client.Client, newObj client.Object) error {
	name := utils.NamespacedName(newObj)
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Get the resource.
		var obj client.Object
		if err := c.Get(context.Background(), name, obj); err != nil {
			return err
		}

		if isStatusEqual(obj, newObj) {
			return nil
		}

		return c.Status().Update(context.Background(), newObj)
	}); err != nil {
		return fmt.Errorf("unable to update status for %s: %w", name, err)
	}

	return nil
}

// isStatusEqual checks if two objects have equivalent status.
//
// Supported objects:
//  GatewayClasses
//  Gateway
//  HTTPRoute
func isStatusEqual(objA, objB interface{}) bool {
	opts := cmpopts.IgnoreFields(metav1.Condition{}, "LastTransitionTime")
	switch a := objA.(type) {
	case *gwapiv1b1.GatewayClass:
		if b, ok := objB.(*gwapiv1b1.GatewayClass); ok {
			if cmp.Equal(a.Status, b.Status, opts) {
				return true
			}
		}
	case *gwapiv1b1.Gateway:
		if b, ok := objB.(*gwapiv1b1.Gateway); ok {
			if cmp.Equal(a.Status, b.Status, opts) {
				return true
			}
		}
	case *gwapiv1b1.HTTPRoute:
		if b, ok := objB.(*gwapiv1b1.HTTPRoute); ok {
			if cmp.Equal(a.Status, b.Status, opts) {
				return true
			}
		}
	}
	return false
}
