package gatewayapi

import (
	"fmt"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func (t *Translator) ProcessBackendTLSPolicies(backendTlsPolicies []*v1alpha2.BackendTLSPolicy,
	gateways []*GatewayContext,
	routes []RouteContext,
	xdsIR XdsIRMap) []*v1alpha2.BackendTLSPolicy {
	fmt.Println("\n flag 1 *******************************************************************")
	var res []*v1alpha2.BackendTLSPolicy
	for _, poli := range backendTlsPolicies {
		policy := poli.DeepCopy()
		res = append(res, policy)
		fmt.Println("++++++++++++++++++ ", policy.Status, "************  ", policy.Name, "############# ", policy.Namespace)
		if policy.Status.Ancestors != nil {
			for k, status := range policy.Status.Ancestors {
				fmt.Println("\n flag e *******************************************************************")

				pname := status.AncestorRef.Name
				pns := NamespaceDerefOrAlpha(status.AncestorRef.Namespace, "default")
				psec := status.AncestorRef.SectionName
				exist := false

				for _, gwc := range gateways {
					fmt.Println("\n flag q *******************************************************************")
					gw := gwc.Gateway
					if gw.Name == string(pname) && gw.Namespace == string(pns) {
						for _, lis := range gw.Spec.Listeners {
							if lis.Name == *psec {
								fmt.Println("\n flag b *******************************************************************")
								exist = true
							}
						}
					}
				}

				if !exist {
					if len(policy.Status.Ancestors) == 1 {
						policy.Status.Ancestors = []v1alpha2.PolicyAncestorStatus{}
					} else {
						fmt.Println("\n flag j *******************************************************************")
						policy.Status.Ancestors = append(policy.Status.Ancestors[:k], policy.Status.Ancestors[k+1:]...)
					}
				}
			}
		} else {
			policy.Status.Ancestors = []v1alpha2.PolicyAncestorStatus{} //nil
		}
	}

	return res
}
