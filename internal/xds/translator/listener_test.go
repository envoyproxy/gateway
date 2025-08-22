// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"reflect"
	"testing"

	matcherv3 "github.com/cncf/xds/go/xds/type/matcher/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	rbacv3 "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func Test_toNetworkFilter(t *testing.T) {
	tests := []struct {
		name    string
		proto   proto.Message
		wantErr error
	}{
		{
			name: "valid filter",
			proto: &hcmv3.HttpConnectionManager{
				StatPrefix: "stats",
				RouteSpecifier: &hcmv3.HttpConnectionManager_RouteConfig{
					RouteConfig: &routev3.RouteConfiguration{
						Name: "route",
					},
				},
			},
			wantErr: nil,
		},
		{
			name:    "invalid proto msg",
			proto:   &hcmv3.HttpConnectionManager{},
			wantErr: errors.New("invalid HttpConnectionManager.StatPrefix: value length must be at least 1 runes; invalid HttpConnectionManager.RouteSpecifier: value is required"),
		},
		{
			name:    "nil proto msg",
			proto:   nil,
			wantErr: errors.New("empty message received"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := toNetworkFilter("name", tt.proto)
			if tt.wantErr != nil {
				assert.Containsf(t, err.Error(), tt.wantErr.Error(), "toNetworkFilter(%v)", tt.proto)
			} else {
				assert.NoErrorf(t, err, "toNetworkFilter(%v)", tt.proto)
			}
		})
	}
}

func Test_buildTCPProxyHashPolicy(t *testing.T) {
	tests := []struct {
		name string
		lb   *ir.LoadBalancer
		want []*typev3.HashPolicy
	}{
		{
			name: "Nil LoadBalancer",
			lb:   nil,
			want: nil,
		},
		{
			name: "Nil ConsistentHash in LoadBalancer",
			lb:   &ir.LoadBalancer{},
			want: nil,
		},
		{
			name: "ConsistentHash without hash policy",
			lb:   &ir.LoadBalancer{ConsistentHash: &ir.ConsistentHash{}},
			want: nil,
		},
		{
			name: "ConsistentHash with SourceIP set to false",
			lb:   &ir.LoadBalancer{ConsistentHash: &ir.ConsistentHash{SourceIP: new(bool)}}, // *new(bool) defaults to false
			want: nil,
		},
		{
			name: "ConsistentHash with SourceIP set to true",
			lb:   &ir.LoadBalancer{ConsistentHash: &ir.ConsistentHash{SourceIP: func(b bool) *bool { return &b }(true)}},
			want: []*typev3.HashPolicy{{PolicySpecifier: &typev3.HashPolicy_SourceIp_{SourceIp: &typev3.HashPolicy_SourceIp{}}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildTCPProxyHashPolicy(tt.lb)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildTCPProxyHashPolicy() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildTCPFilterChain_RBACPresenceByRules(t *testing.T) {
	// CASE A: Authorization present with >=1 rule -> RBAC filter should be added
	routeWithRule := &ir.TCPRoute{
		Name: "r-with-rule",
		Destination: &ir.RouteDestination{
			Name: "u1",
		},
		Security: &ir.SecurityFeatures{
			Authorization: &ir.Authorization{
				Rules: []*ir.AuthorizationRule{
					{
						Name:   "allow-10/8",
						Action: egv1a1.AuthorizationActionAllow,
						Principal: ir.Principal{
							ClientCIDRs: []*ir.CIDRMatch{{CIDR: "10.0.0.0/8"}},
						},
					},
				},
			},
		},
	}

	// buildTCPFilterChain signature requires: (irRoute, clusterName, statPrefix, accesslog, timeout, connection)
	fcA, err := buildTCPFilterChain(routeWithRule, "cluster-a", "stats_tcp", nil, nil, nil)
	require.NoError(t, err, "building filter chain (rules present)")
	require.NotNil(t, fcA)

	require.True(t, hasNetworkFilter(fcA, "envoy.filters.network.rbac"),
		"expected RBAC network filter to be present when rules >= 1")

	// CASE B: Authorization nil/empty -> RBAC filter should NOT be added
	routeNoRules := &ir.TCPRoute{
		Name: "r-no-rules",
		Destination: &ir.RouteDestination{
			Name: "u1",
		},
		// Security absent OR Security.Authorization present but Rules empty should both skip RBAC
		// Uncomment either variant you want to exercise:
		// Security: nil,
		Security: &ir.SecurityFeatures{Authorization: &ir.Authorization{Rules: nil}},
	}

	fcB, err := buildTCPFilterChain(routeNoRules, "cluster-b", "stats_tcp", nil, nil, nil)
	require.NoError(t, err, "building filter chain (no rules)")
	require.NotNil(t, fcB)

	require.False(t, hasNetworkFilter(fcB, "envoy.filters.network.rbac"),
		"did not expect RBAC network filter when rules are nil/empty")
}

func hasNetworkFilter(fc *listenerv3.FilterChain, name string) bool {
	if fc == nil {
		return false
	}
	for _, nf := range fc.GetFilters() {
		if nf.GetName() == name {
			return true
		}
	}
	return false
}

func Test_buildTCPRBACMatcherFromRules_DefaultActionAllow(t *testing.T) {
	rules := []*ir.AuthorizationRule{
		{
			Name:   "allow-local",
			Action: egv1a1.AuthorizationActionAllow,
			Principal: ir.Principal{
				ClientCIDRs: []*ir.CIDRMatch{{CIDR: "192.168.254.0/24"}},
			},
		},
		{
			Name:   "deny-all",
			Action: egv1a1.AuthorizationActionDeny,
			Principal: ir.Principal{
				ClientCIDRs: []*ir.CIDRMatch{{CIDR: "0.0.0.0/0"}},
			},
		},
	}

	rbac := buildTCPRBACMatcherFromRules(rules, egv1a1.AuthorizationActionAllow)
	require.NotNil(t, rbac)

	// Basic shape checks
	require.NotNil(t, rbac.Matcher)
	if ml := rbac.Matcher.GetMatcherList(); ml != nil {
		require.Len(t, ml.GetMatchers(), 2, "expected two field matchers for the two rules")
	}

	// Verify default action typed-config decodes to ALLOW (zero value)
	onNo := rbac.Matcher.GetOnNoMatch()
	require.NotNil(t, onNo, "expected OnNoMatch to be set")
	om := onNo.GetOnMatch()
	require.NotNil(t, om, "expected OnNoMatch.OnMatch to be set")

	v, ok := om.(*matcherv3.Matcher_OnMatch_Action)
	require.True(t, ok, "expected OnMatch to be Action variant")
	require.NotNil(t, v.Action, "expected TypedExtensionConfig on OnMatch.Action")
	require.NotNil(t, v.Action.GetTypedConfig(), "expected TypedConfig on OnMatch.Action")

	act := &rbacv3.Action{}
	require.NoError(t, v.Action.GetTypedConfig().UnmarshalTo(act))
	require.Equal(t, rbacv3.RBAC_ALLOW, act.Action)
}

func Test_buildRBACPerRoute_DefaultActionAllow(t *testing.T) {
	auth := &ir.Authorization{
		DefaultAction: egv1a1.AuthorizationActionAllow,
	}

	rp, err := buildRBACPerRoute(auth)
	require.NoError(t, err)
	require.NotNil(t, rp)

	if rp.Rbac != nil {
		if m := rp.Rbac.GetMatcher(); m != nil {
			onNo := m.GetOnNoMatch()
			require.NotNil(t, onNo)
			om := onNo.GetOnMatch()
			require.NotNil(t, om)

			actVariant, ok := om.(*matcherv3.Matcher_OnMatch_Action)
			require.True(t, ok)
			require.NotNil(t, actVariant.Action)
			require.NotNil(t, actVariant.Action.GetTypedConfig())

			a := &rbacv3.Action{}
			require.NoError(t, actVariant.Action.GetTypedConfig().UnmarshalTo(a))
			require.Equal(t, rbacv3.RBAC_ALLOW, a.Action)
			return
		}
		if rules := rp.Rbac.GetRules(); rules != nil {
			require.Equal(t, rbacv3.RBAC_ALLOW, rules.GetAction())
			return
		}
	}
	t.Fatalf("unexpected RBAC per-route shape: %+v", rp)
}

func Test_ConvertPrincipals_DirectRemoteIP(t *testing.T) {
	// Build IR principal with a CIDR
	pr := ir.Principal{
		ClientCIDRs: []*ir.CIDRMatch{{CIDR: "10.0.0.0/8"}},
	}

	principals := convertPrincipals(pr)
	require.GreaterOrEqual(t, len(principals), 1, "expected at least one principal")

	// The implementation creates DirectRemoteIp entries for CIDRs
	p0 := principals[0]
	require.NotNil(t, p0, "principal must not be nil")

	dr := p0.GetDirectRemoteIp()
	require.NotNil(t, dr, "expected DirectRemoteIp to be set")

	// AddressPrefix should be the parsed IP of the CIDR
	// ParseCIDR("10.0.0.0/8") yields ip.String() == "10.0.0.0"
	require.Equal(t, "10.0.0.0", dr.AddressPrefix)
}

func Test_buildTCPFilterChain_IPAwareRoute(t *testing.T) {
	route := &ir.TCPRoute{
		Name: "ip-aware-route",
		Destination: &ir.RouteDestination{
			Name: "upstream-1",
		},
		Security: &ir.SecurityFeatures{
			Authorization: &ir.Authorization{
				Rules: []*ir.AuthorizationRule{
					{
						Name:   "allow-10-8",
						Action: egv1a1.AuthorizationActionAllow,
						Principal: ir.Principal{
							ClientCIDRs: []*ir.CIDRMatch{
								{CIDR: "10.0.0.0/8"},
							},
						},
					},
				},
			},
		},
		LoadBalancer: &ir.LoadBalancer{
			ConsistentHash: &ir.ConsistentHash{
				SourceIP: func(b bool) *bool { return &b }(true),
			},
		},
	}

	fc, err := buildTCPFilterChain(route, "cluster-a", "stats_tcp", nil, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, fc)

	// chain name should match the route name
	require.Equal(t, route.Name, fc.Name)

	// RBAC filter present
	require.True(t, hasNetworkFilter(fc, "envoy.filters.network.rbac"))

	// TCP proxy present and typed-config references v3 TcpProxy
	require.True(t, hasNetworkFilter(fc, "envoy.filters.network.tcp_proxy"))
	foundTCPProxy := false
	for _, f := range fc.GetFilters() {
		if f.GetName() == "envoy.filters.network.tcp_proxy" {
			foundTCPProxy = true
			tc := f.GetTypedConfig()
			require.NotNil(t, tc)
			require.Contains(t, tc.GetTypeUrl(), "envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy")
		}
	}
	require.True(t, foundTCPProxy, "tcp proxy filter not found")
}

// SingleAction = all rules share the same action (all ALLOW or all DENY) -> classic rules-based RBAC (Rules field set, Matcher nil)

func Test_buildTCPRBACMatcherFromRules_SingleActionAllow(t *testing.T) {
	rules := []*ir.AuthorizationRule{
		{
			Name:   "allow-a",
			Action: egv1a1.AuthorizationActionAllow,
			Principal: ir.Principal{
				ClientCIDRs: []*ir.CIDRMatch{{CIDR: "10.1.0.0/16"}},
			},
		},
		{
			Name:   "allow-b",
			Action: egv1a1.AuthorizationActionAllow,
			Principal: ir.Principal{
				ClientCIDRs: []*ir.CIDRMatch{{CIDR: "10.2.0.0/16"}},
			},
		},
	}

	r := buildTCPRBACMatcherFromRules(rules, egv1a1.AuthorizationActionDeny)
	require.NotNil(t, r)
	require.NotNil(t, r.Rules, "expected classic rules-based RBAC (SingleAction ALLOW)")
	require.Nil(t, r.Matcher, "matcher should be nil for SingleAction set")
	require.Equal(t, rbacv3.RBAC_ALLOW, r.Rules.Action)
	require.Len(t, r.Rules.Policies, 2)
}

func Test_buildTCPRBACMatcherFromRules_SingleActionDeny(t *testing.T) {
	rules := []*ir.AuthorizationRule{
		{
			Name:   "deny-a",
			Action: egv1a1.AuthorizationActionDeny,
			Principal: ir.Principal{
				ClientCIDRs: []*ir.CIDRMatch{{CIDR: "0.0.0.0/0"}},
			},
		},
		{
			Name:   "deny-b",
			Action: egv1a1.AuthorizationActionDeny,
			Principal: ir.Principal{
				ClientCIDRs: []*ir.CIDRMatch{{CIDR: "192.0.2.0/24"}},
			},
		},
	}

	r := buildTCPRBACMatcherFromRules(rules, egv1a1.AuthorizationActionAllow)
	require.NotNil(t, r)
	require.NotNil(t, r.Rules, "expected classic rules-based RBAC (SingleAction DENY)")
	require.Nil(t, r.Matcher)
	require.Equal(t, rbacv3.RBAC_DENY, r.Rules.Action)
	require.Len(t, r.Rules.Policies, 2)
}

func Test_buildTCPRBACMatcherFromRules_DefaultActionDeny_Mixed(t *testing.T) {
	rules := []*ir.AuthorizationRule{
		{
			Name:   "allow-subnet",
			Action: egv1a1.AuthorizationActionAllow,
			Principal: ir.Principal{
				ClientCIDRs: []*ir.CIDRMatch{{CIDR: "198.51.100.0/24"}},
			},
		},
		{
			Name:   "deny-host",
			Action: egv1a1.AuthorizationActionDeny,
			Principal: ir.Principal{
				ClientCIDRs: []*ir.CIDRMatch{{CIDR: "203.0.113.42/32"}},
			},
		},
	}

	r := buildTCPRBACMatcherFromRules(rules, egv1a1.AuthorizationActionDeny)
	require.NotNil(t, r)
	require.NotNil(t, r.Matcher, "expected matcher for mixed actions")
	require.Nil(t, r.Rules, "rules variant should be nil for mixed action set")

	// Validate OnNoMatch default DENY action
	onNo := r.Matcher.GetOnNoMatch()
	require.NotNil(t, onNo)

	onNoActionWrapper, ok := onNo.OnMatch.(*matcherv3.Matcher_OnMatch_Action)
	require.True(t, ok)
	require.NotNil(t, onNoActionWrapper.Action)
	require.NotNil(t, onNoActionWrapper.Action.TypedConfig)

	act := &rbacv3.Action{}
	require.NoError(t, onNoActionWrapper.Action.TypedConfig.UnmarshalTo(act))
	require.Equal(t, rbacv3.RBAC_DENY, act.Action, "default OnNoMatch should be DENY when defaultAction is DENY")
}

func Test_sanitizeMatcherActionName(t *testing.T) {
	cases := map[string]string{
		"":                "unnamed",
		"Allow-Net":       "allow_net",
		"deny.ALL":        "deny_all",
		"MiXeD_123":       "mixed_123",
		"spaces and tabs": "spaces_and_tabs",
		"Ünicode*Chars!":  "ünicode_chars_", // preserves unicode letters, replaces others
	}
	for in, want := range cases {
		got := sanitizeMatcherActionName(in)
		require.Equal(t, want, got, "input=%q", in)
	}
}
