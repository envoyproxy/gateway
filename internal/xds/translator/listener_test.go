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
	"google.golang.org/protobuf/encoding/prototext"
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

func Test_splitTCPAuthRules_and_convertPrincipals(t *testing.T) {
	rules := []*ir.AuthorizationRule{
		{
			Name:   "allow-net",
			Action: egv1a1.AuthorizationActionAllow,
			Principal: ir.Principal{
				ClientCIDRs: []*ir.CIDRMatch{{CIDR: "10.0.0.0/8"}},
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

	allow, deny := splitTCPAuthRules(rules)
	require.Len(t, allow, 1)
	require.Len(t, deny, 1)
	_, ok := allow["allow-net"]
	require.True(t, ok)
	_, ok = deny["deny-all"]
	require.True(t, ok)

	// convertPrincipals should produce at least one principal for the allow rule
	principals := convertPrincipals(rules[0].Principal)
	require.GreaterOrEqual(t, len(principals), 1)
	// GetAny() returns a bool, so check it directly.
	for _, p := range principals {
		require.True(t, p.GetAny() || p.GetRemoteIp() != nil || p.GetDirectRemoteIp() != nil,
			"expected principal to have a non-nil identifier (Any, RemoteIp, or DirectRemoteIp)")
	}
}

func Test_buildTCPRBACMatcherFromRules_Basic(t *testing.T) {
	rules := []*ir.AuthorizationRule{
		{
			Name:   "r1",
			Action: egv1a1.AuthorizationActionAllow,
			Principal: ir.Principal{
				ClientCIDRs: []*ir.CIDRMatch{{CIDR: "192.168.1.0/24"}},
			},
		},
		{
			Name:   "r2",
			Action: egv1a1.AuthorizationActionDeny,
			Principal: ir.Principal{
				ClientCIDRs: []*ir.CIDRMatch{{CIDR: "0.0.0.0/0"}},
			},
		},
	}

	rbac := buildTCPRBACMatcherFromRules(rules, egv1a1.AuthorizationActionDeny)
	require.NotNil(t, rbac)
	require.NotNil(t, rbac.Matcher)

	// If the matcher has a MatcherList, it should contain two FieldMatchers
	if ml := rbac.Matcher.GetMatcherList(); ml != nil {
		require.Len(t, ml.GetMatchers(), 2)
	}
	// OnNoMatch should be set
	require.NotNil(t, rbac.Matcher.GetOnNoMatch())
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

	// debug dump (visible with -v)
	var dump string
	mo := prototext.MarshalOptions{Multiline: true}
	if b, err := mo.Marshal(rbac); err == nil {
		dump = string(b)
	} else {
		t.Logf("failed to marshal rbac to text: %v", err)
	}
	t.Logf("rbac dump:\n%s", dump)

	// Basic shape checks
	require.NotNil(t, rbac.Matcher)
	if ml := rbac.Matcher.GetMatcherList(); ml != nil {
		require.Len(t, ml.GetMatchers(), 2, "expected two field matchers for the two rules")
	}

	// Loose textual checks
	require.Contains(t, dump, "tcp-authz-allow")
	require.Contains(t, dump, "tcp-authz-deny")
	require.Contains(t, dump, "default")

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

func Test_buildTCPRBACMatcherFromRules_DefaultActionAllow_TCP(t *testing.T) {
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

	r := buildTCPRBACMatcherFromRules(rules, egv1a1.AuthorizationActionAllow)
	require.NotNil(t, r)

	// debug dump
	mo := prototext.MarshalOptions{Multiline: true}
	if b, err := mo.Marshal(r); err == nil {
		t.Logf("tcp rbac dump:\n%s", string(b))
	}

	// Basic checks
	require.NotNil(t, r.Matcher)
	if ml := r.Matcher.GetMatcherList(); ml != nil {
		require.Len(t, ml.GetMatchers(), 2)
	}

	onNo := r.Matcher.GetOnNoMatch()
	require.NotNil(t, onNo)

	om := onNo.GetOnMatch()
	require.NotNil(t, om)

	v, ok := om.(*matcherv3.Matcher_OnMatch_Action)
	require.True(t, ok)
	require.NotNil(t, v.Action)
	require.NotNil(t, v.Action.GetTypedConfig())

	act := &rbacv3.Action{}
	require.NoError(t, v.Action.GetTypedConfig().UnmarshalTo(act))
	require.Equal(t, rbacv3.RBAC_ALLOW, act.Action)
}

func Test_EmptyActionProtoDefaultsToAllow(t *testing.T) {
	act := &rbacv3.Action{} // empty
	b, err := proto.Marshal(act)
	require.NoError(t, err)

	act2 := &rbacv3.Action{}
	require.NoError(t, proto.Unmarshal(b, act2))

	// Zero value of the enum should equal RBAC_ALLOW if RBAC_ALLOW is defined as zero.
	require.Equal(t, rbacv3.RBAC_ALLOW, act2.Action)
}

func Test_buildRBACPerRoute_DefaultActionAllow(t *testing.T) {
	auth := &ir.Authorization{
		// DefaultAction on IR is the gateway API enum type
		DefaultAction: egv1a1.AuthorizationActionAllow,
		// no rules
	}

	rp, err := buildRBACPerRoute(auth)
	require.NoError(t, err)
	require.NotNil(t, rp)

	// Debug: marshal to text for easier failure diagnosis in -v runs.
	mo := prototext.MarshalOptions{Multiline: true}
	if b, err := mo.Marshal(rp); err == nil {
		t.Logf("rbacPerRoute dump:\n%s", string(b))
	}

	// The RBAC per-route can be represented either as a Matcher (with OnNoMatch typed-config)
	// or as a simple Rules variant. Handle both shapes.
	if rp.Rbac != nil {
		// If matcher variant present:
		if m := rp.Rbac.GetMatcher(); m != nil {
			onNo := m.GetOnNoMatch()
			require.NotNil(t, onNo)
			om := onNo.GetOnMatch()
			require.NotNil(t, om)

			actVariant, ok := om.(*matcherv3.Matcher_OnMatch_Action)
			require.True(t, ok, "expected Action variant for OnMatch")
			require.NotNil(t, actVariant.Action)
			require.NotNil(t, actVariant.Action.GetTypedConfig())

			any := actVariant.Action.GetTypedConfig()
			a := &rbacv3.Action{}
			require.NoError(t, any.UnmarshalTo(a))
			require.Equal(t, rbacv3.RBAC_ALLOW, a.Action)
			return
		}

		// If Rules variant present:
		if rules := rp.Rbac.GetRules(); rules != nil {
			// Rules.Action should reflect default action
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

func Test_BuildTCPProxyHashPolicy_SourceIP(t *testing.T) {
	// ConsistentHash.SourceIP true should produce a SourceIp hash policy
	lb := &ir.LoadBalancer{
		ConsistentHash: &ir.ConsistentHash{
			SourceIP: func(b bool) *bool { return &b }(true),
		},
	}

	got := buildTCPProxyHashPolicy(lb)
	require.Len(t, got, 1, "expected one hash policy when SourceIP is true")

	hp := got[0]
	require.NotNil(t, hp, "hash policy must not be nil")

	// Expect SourceIp variant
	_, ok := hp.PolicySpecifier.(*typev3.HashPolicy_SourceIp_)
	require.True(t, ok, "expected HashPolicy_SourceIp_ variant")
	// And the SourceIp struct itself should be non-nil
	if s, ok := hp.PolicySpecifier.(*typev3.HashPolicy_SourceIp_); ok {
		require.NotNil(t, s.SourceIp, "expected SourceIp to be non-nil")
	}
}
