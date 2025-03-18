// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package utils

import (
	"context"
	"fmt"
	"net"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func RetrieveURL(c client.Client, nn types.NamespacedName, port int32, path string) (string, error) {
	host, err := ServiceHost(c, nn, port)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("http://%s%s", host, path), nil
}

func ServiceHost(c client.Client, nn types.NamespacedName, port int32) (string, error) {
	svc := corev1.Service{}
	if err := c.Get(context.Background(), nn, &svc); err != nil {
		return "", err
	}
	host := ""
	switch svc.Spec.Type {
	case corev1.ServiceTypeLoadBalancer:
		for _, ing := range svc.Status.LoadBalancer.Ingress {
			if ing.IP != "" {
				host = ing.IP
				break
			}
		}
	default:
		host = fmt.Sprintf("%s.%s.svc", nn.Name, nn.Namespace)
	}

	return net.JoinHostPort(host, strconv.Itoa(int(port))), nil
}
