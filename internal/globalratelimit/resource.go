package globalratelimit

import (
	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit"
)

// GetRateLimitUrl returns the URL for the rate limit service.
func GetRateLimitUrl(eg *egv1a1.EnvoyGateway, namespace, dnsDomain string) string {
	if eg != nil && eg.RateLimit != nil && eg.RateLimit.URL != nil {
		return *eg.RateLimit.URL
	}
	return ratelimit.GetServiceURL(namespace, dnsDomain)
}
