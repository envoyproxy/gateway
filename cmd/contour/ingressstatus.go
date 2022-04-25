// Copyright Project Contour Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"net"
	"strings"

	"github.com/projectcontour/contour/internal/k8s"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi_v1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// loadBalancerStatusWriter orchestrates LoadBalancer address status
// updates for HTTPProxy, Ingress and Gateway objects. Actually updating the
// address in the object status is performed by k8s.StatusAddressUpdater.
//
// The theory of operation of the loadBalancerStatusWriter is as follows:
//
// 1. On startup the loadBalancerStatusWriter waits to be elected leader.
// 2. Once elected leader, the loadBalancerStatusWriter waits to receive a
//    v1.LoadBalancerStatus value.
// 3. Once a v1.LoadBalancerStatus value has been received, the
//    cached address is updated so that it will be applied to objects
//    received in any subsequent informer events.
// 4. All Ingress, HTTPProxy and Gateway objects are listed from the informer
//    cache and an attempt is made to update their status with the new
//    address. This update may end up being a no-op in which case it
//    doesn't make an API server call.
// 5. If the worker is stopped, the informer continues but no further
//    status updates are made.
type loadBalancerStatusWriter struct {
	log                   logrus.FieldLogger
	cache                 cache.Cache
	lbStatus              chan v1.LoadBalancerStatus
	statusUpdater         k8s.StatusUpdater
	gatewayControllerName string
	gatewayRef            *types.NamespacedName
}

func (isw *loadBalancerStatusWriter) NeedLeaderElection() bool {
	return true
}

func (isw *loadBalancerStatusWriter) Start(ctx context.Context) error {
	u := &k8s.StatusAddressUpdater{
		Logger: func() logrus.FieldLogger {
			// Configure the StatusAddressUpdater logger.
			return isw.log.WithField("context", "StatusAddressUpdater")
		}(),
		Cache:                 isw.cache,
		GatewayControllerName: isw.gatewayControllerName,
		GatewayRef:            isw.gatewayRef,
		StatusUpdater:         isw.statusUpdater,
	}

	// Create informers for the types that need load balancer
	// address status. The cache should have already started
	// informers, so new informers will auto-start.
	resources := []client.Object{}

	// Only create Gateway informer if a controller or specific gateway was provided,
	// otherwise the API may not exist in the cluster.
	if len(isw.gatewayControllerName) > 0 || isw.gatewayRef != nil {
		resources = append(resources, &gatewayapi_v1alpha2.Gateway{})
	}

	for _, r := range resources {
		inf, err := isw.cache.GetInformer(context.Background(), r)
		if err != nil {
			isw.log.WithError(err).WithField("resource", r).Fatal("failed to create informer")
		}

		inf.AddEventHandler(u)
	}

	for {
		select {
		case <-ctx.Done():
			// Once started, there's no way to stop the
			// informer from here. Clear the load balancer
			// status so that subsequent informer events
			// will have no effect.
			u.Set(v1.LoadBalancerStatus{})
			return nil
		case lbs := <-isw.lbStatus:
			isw.log.WithField("loadbalancer-address", lbAddress(lbs)).
				Info("received a new address for status.loadBalancer")

			u.Set(lbs)

			// Only list Gateways if a controller or specific gateway was configured,
			// otherwise the API may not exist in the cluster.
			if len(isw.gatewayControllerName) > 0 || isw.gatewayRef != nil {
				var gatewayList gatewayapi_v1alpha2.GatewayList
				if err := isw.cache.List(context.Background(), &gatewayList); err != nil {
					isw.log.WithError(err).WithField("kind", "Gateway").Error("failed to list objects")
				} else {
					for i := range gatewayList.Items {
						u.OnAdd(&gatewayList.Items[i])
					}
				}
			}
		}
	}
}

func parseStatusFlag(status string) v1.LoadBalancerStatus {
	// Support ','-separated lists.
	var ingresses []v1.LoadBalancerIngress

	for _, item := range strings.Split(status, ",") {
		item = strings.TrimSpace(item)
		if len(item) == 0 {
			continue
		}

		// Use the parseability by net.ParseIP as a signal, since we need
		// to pass a string into the v1.LoadBalancerIngress anyway.
		if ip := net.ParseIP(item); ip != nil {
			ingresses = append(ingresses, v1.LoadBalancerIngress{
				IP: item,
			})
		} else {
			ingresses = append(ingresses, v1.LoadBalancerIngress{
				Hostname: item,
			})
		}
	}

	return v1.LoadBalancerStatus{
		Ingress: ingresses,
	}
}

// lbAddress gets the string representation of the first address, for logging.
func lbAddress(lb v1.LoadBalancerStatus) string {
	if len(lb.Ingress) == 0 {
		return ""
	}

	if lb.Ingress[0].IP != "" {
		return lb.Ingress[0].IP
	}

	return lb.Ingress[0].Hostname
}
