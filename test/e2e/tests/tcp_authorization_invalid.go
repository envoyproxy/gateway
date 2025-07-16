//go:build e2e

package tests

import (
	"context"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func init() {
	ConformanceTests = append(ConformanceTests, TCPAuthorizationInvalidTest)
}

var TCPAuthorizationInvalidTest = suite.ConformanceTest{
	ShortName:   "TCPAuthzInvalid",
	Description: "TCP authorization with invalid configurations should be rejected",
	Manifests:   []string{"testdata/tcp-authorization-invalid.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"

		t.Run("SecurityPolicy with JWT should be rejected for TCP", func(t *testing.T) {
			policyNN := types.NamespacedName{Name: "tcp-authorization-invalid-jwt", Namespace: ns}

			// Wait for the SecurityPolicy to be rejected
			waitErr := wait.PollImmediate(1*time.Second, 30*time.Second, func() (bool, error) {
				policy := &egv1a1.SecurityPolicy{}
				err := suite.Client.Get(context.Background(), policyNN, policy)
				if err != nil {
					return false, err
				}

				// Check if the policy has been rejected
				// The status structure may vary, check what's actually available
				if len(policy.Status.Ancestors) > 0 {
					for _, ancestor := range policy.Status.Ancestors {
						for _, condition := range ancestor.Conditions {
							if condition.Type == string(gwapiv1a2.PolicyConditionAccepted) &&
								condition.Status == metav1.ConditionFalse {
								t.Logf("Policy correctly rejected: %s", condition.Message)
								return true, nil
							}
						}
					}
				}
				return false, nil
			})

			if waitErr != nil {
				t.Errorf("Expected SecurityPolicy with JWT to be rejected for TCP route, but it wasn't: %v", waitErr)
			}
		})

		t.Run("SecurityPolicy with headers should be rejected for TCP", func(t *testing.T) {
			policyNN := types.NamespacedName{Name: "tcp-authorization-invalid-headers", Namespace: ns}

			// Similar validation logic for headers
			waitErr := wait.PollImmediate(1*time.Second, 30*time.Second, func() (bool, error) {
				policy := &egv1a1.SecurityPolicy{}
				err := suite.Client.Get(context.Background(), policyNN, policy)
				if err != nil {
					return false, err
				}

				// Check if the policy has been rejected
				if len(policy.Status.Ancestors) > 0 {
					for _, ancestor := range policy.Status.Ancestors {
						for _, condition := range ancestor.Conditions {
							if condition.Type == string(gwapiv1a2.PolicyConditionAccepted) &&
								condition.Status == metav1.ConditionFalse {
								t.Logf("Policy correctly rejected: %s", condition.Message)
								return true, nil
							}
						}
					}
				}
				return false, nil
			})

			if waitErr != nil {
				t.Errorf("Expected SecurityPolicy with headers to be rejected for TCP route, but it wasn't: %v", waitErr)
			}
		})
	},
}
