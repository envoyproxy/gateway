package kubernetes

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakeclock "k8s.io/utils/clock/testing"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/log"
)

func TestGatewayHasMatchingController(t *testing.T) {
	match := &gwapiv1b1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "matched",
		},
		Spec: gwapiv1b1.GatewayClassSpec{
			ControllerName: v1alpha1.GatewayControllerName,
		},
	}

	nonMatch := &gwapiv1b1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "non-matched",
		},
		Spec: gwapiv1b1.GatewayClassSpec{
			ControllerName: "not.configured/controller-name",
		},
	}

	testCases := []struct {
		name   string
		obj    client.Object
		expect bool
	}{
		{
			name: "matching object type, gatewayclass, and controller name",
			obj: &gwapiv1b1.Gateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Gateway",
					APIVersion: fmt.Sprintf("%s/%s", gwapiv1b1.GroupName, gwapiv1b1.GroupVersion.Version),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: gwapiv1b1.GatewaySpec{
					GatewayClassName: gwapiv1b1.ObjectName(match.Name),
				},
			},
			expect: true,
		},
		{
			name: "matching object type but gatewayclass doesn't exist",
			obj: &gwapiv1b1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: gwapiv1b1.GatewaySpec{
					GatewayClassName: "non-existent-gc",
				},
			},
			expect: false,
		},
		{
			name: "matching object type and gatewayclass but not controller name",
			obj: &gwapiv1b1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: gwapiv1b1.GatewaySpec{
					GatewayClassName: gwapiv1b1.ObjectName(nonMatch.Name),
				},
			},
			expect: false,
		},
		{
			name: "gatewayclass name match but object type doesn't match",
			obj: &gwapiv1b1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: gwapiv1b1.GatewayClassSpec{
					ControllerName: gwapiv1b1.GatewayController(v1alpha1.GatewayControllerName),
				},
			},
			expect: false,
		},
	}

	// Create the reconciler.
	logger, err := log.NewLogger()
	require.NoError(t, err)
	r := gatewayReconciler{
		classController: v1alpha1.GatewayControllerName,
		log:             logger,
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r.client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(match, nonMatch, tc.obj).Build()
			actual := r.hasMatchingController(tc.obj)
			require.Equal(t, tc.expect, actual)
		})
	}
}

func TestIsAccepted(t *testing.T) {
	testCases := []struct {
		name   string
		gc     *gwapiv1b1.GatewayClass
		expect bool
	}{
		{
			name: "gatewayclass accepted condition",
			gc: &gwapiv1b1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: gwapiv1b1.GatewayClassSpec{
					ControllerName: gwapiv1b1.GatewayController(v1alpha1.GatewayControllerName),
				},
				Status: gwapiv1b1.GatewayClassStatus{
					Conditions: []metav1.Condition{
						{
							Type:   string(gwapiv1b1.GatewayClassConditionStatusAccepted),
							Status: metav1.ConditionTrue,
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "gatewayclass not accepted condition",
			gc: &gwapiv1b1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: gwapiv1b1.GatewayClassSpec{
					ControllerName: gwapiv1b1.GatewayController(v1alpha1.GatewayControllerName),
				},
				Status: gwapiv1b1.GatewayClassStatus{
					Conditions: []metav1.Condition{
						{
							Type:   string(gwapiv1b1.GatewayClassConditionStatusAccepted),
							Status: metav1.ConditionFalse,
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "no gatewayclass accepted condition type",
			gc: &gwapiv1b1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: gwapiv1b1.GatewayClassSpec{
					ControllerName: gwapiv1b1.GatewayController(v1alpha1.GatewayControllerName),
				},
				Status: gwapiv1b1.GatewayClassStatus{
					Conditions: []metav1.Condition{
						{
							Type:   "SomeOtherType",
							Status: metav1.ConditionTrue,
						},
					},
				},
			},
			expect: false,
		},
		{
			name:   "nil gatewayclass",
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := isAccepted(tc.gc)
			require.Equal(t, tc.expect, actual)
		})
	}
}

func TestGatewaysOfClass(t *testing.T) {
	gc := &gwapiv1b1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
	}
	testCases := []struct {
		name   string
		gws    []gwapiv1b1.Gateway
		expect int
	}{
		{
			name: "no matching gateways",
			gws: []gwapiv1b1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Spec: gwapiv1b1.GatewaySpec{
						GatewayClassName: gwapiv1b1.ObjectName("no-match"),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Spec: gwapiv1b1.GatewaySpec{
						GatewayClassName: gwapiv1b1.ObjectName("no-match2"),
					},
				},
			},
			expect: 0,
		},
		{
			name: "one of two matching gateways",
			gws: []gwapiv1b1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Spec: gwapiv1b1.GatewaySpec{
						GatewayClassName: gwapiv1b1.ObjectName(gc.Name),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "test",
					},
					Spec: gwapiv1b1.GatewaySpec{
						GatewayClassName: gwapiv1b1.ObjectName("no-match"),
					},
				},
			},
			expect: 1,
		},
		{
			name: "two of two matching gateways",
			gws: []gwapiv1b1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Spec: gwapiv1b1.GatewaySpec{
						GatewayClassName: gwapiv1b1.ObjectName(gc.Name),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "test",
					},
					Spec: gwapiv1b1.GatewaySpec{
						GatewayClassName: gwapiv1b1.ObjectName(gc.Name),
					},
				},
			},
			expect: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gwList := &gwapiv1b1.GatewayList{Items: tc.gws}
			actual := gatewaysOfClass(gc, gwList)
			require.Equal(t, tc.expect, len(actual))
		})
	}
}

func TestOldestGateway(t *testing.T) {
	// Create a fake clock and set different times for gateway CreationTimestamp.
	fakeClock := fakeclock.NewFakeClock(time.Time{})
	now := fakeClock.Now()
	later := now.Add(1 * time.Minute)
	latest := now.Add(2 * time.Minute)

	testCases := []struct {
		name   string
		in     []gwapiv1b1.Gateway
		expect *gwapiv1b1.Gateway
	}{
		{
			name: "one gateway",
			in: []gwapiv1b1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "first",
						Namespace:         "test",
						CreationTimestamp: metav1.NewTime(now),
					},
				},
			},
			expect: &gwapiv1b1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "first",
					Namespace: "test",
				},
			},
		},
		{
			name: "two gateways with different times",
			in: []gwapiv1b1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "second",
						Namespace:         "test",
						CreationTimestamp: metav1.NewTime(later),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "first",
						Namespace:         "test",
						CreationTimestamp: metav1.NewTime(now),
					},
				},
			},
			expect: &gwapiv1b1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "first",
					Namespace: "test",
				},
			},
		},
		{
			name: "three gateways with different times",
			in: []gwapiv1b1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "first",
						Namespace:         "test",
						CreationTimestamp: metav1.NewTime(latest),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "second",
						Namespace:         "test",
						CreationTimestamp: metav1.NewTime(later),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "third",
						Namespace:         "test",
						CreationTimestamp: metav1.NewTime(now),
					},
				},
			},
			expect: &gwapiv1b1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "third",
					Namespace: "test",
				},
			},
		},
		{
			name: "three gateways with same time",
			in: []gwapiv1b1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "third",
						Namespace:         "test",
						CreationTimestamp: metav1.NewTime(now),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "second",
						Namespace:         "test",
						CreationTimestamp: metav1.NewTime(now),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "first",
						Namespace:         "test",
						CreationTimestamp: metav1.NewTime(now),
					},
				},
			},
			expect: &gwapiv1b1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "first",
					Namespace: "test",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := oldestGateway(tc.in)
			require.Equal(t, tc.expect.Name, actual.Name)
			require.Equal(t, tc.expect.Namespace, actual.Namespace)
		})
	}
}
