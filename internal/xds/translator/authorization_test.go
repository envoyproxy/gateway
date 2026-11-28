// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestBuildRBACPerRoute_Mode(t *testing.T) {
	testRule := &ir.AuthorizationRule{
		Name:   "allow-admin",
		Action: egv1a1.AuthorizationActionAllow,
		Principal: ir.Principal{
			Headers: []egv1a1.AuthorizationHeaderMatch{
				{
					Name:   "x-role",
					Values: []string{"admin"},
				},
			},
		},
	}

	testCases := []struct {
		name       string
		auth       *ir.Authorization
		wantShadow bool
	}{
		{
			name: "default mode is enforce",
			auth: &ir.Authorization{
				DefaultAction: egv1a1.AuthorizationActionDeny,
				Rules:         []*ir.AuthorizationRule{testRule},
			},
			wantShadow: false,
		},
		{
			name: "explicit enforce mode",
			auth: &ir.Authorization{
				Mode:          ptr.To(egv1a1.AuthorizationModeEnforce),
				DefaultAction: egv1a1.AuthorizationActionDeny,
				Rules:         []*ir.AuthorizationRule{testRule},
			},
			wantShadow: false,
		},
		{
			name: "shadow mode",
			auth: &ir.Authorization{
				Mode:          ptr.To(egv1a1.AuthorizationModeShadow),
				DefaultAction: egv1a1.AuthorizationActionDeny,
				Rules:         []*ir.AuthorizationRule{testRule},
			},
			wantShadow: true,
		},
		{
			name: "shadow mode with empty rules",
			auth: &ir.Authorization{
				Mode:          ptr.To(egv1a1.AuthorizationModeShadow),
				DefaultAction: egv1a1.AuthorizationActionDeny,
				Rules:         []*ir.AuthorizationRule{},
			},
			wantShadow: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rbacPerRoute, err := buildRBACPerRoute(tc.auth)
			require.NoError(t, err)
			require.NotNil(t, rbacPerRoute)
			require.NotNil(t, rbacPerRoute.Rbac)

			if tc.wantShadow {
				require.NotNil(t, rbacPerRoute.Rbac.ShadowMatcher)
				require.Nil(t, rbacPerRoute.Rbac.Matcher)
			} else {
				require.NotNil(t, rbacPerRoute.Rbac.Matcher)
				require.Nil(t, rbacPerRoute.Rbac.ShadowMatcher)
			}
		})
	}
}

func TestBuildTCPRBACConfig_Mode(t *testing.T) {
	testRule := &ir.AuthorizationRule{
		Name:   "allow-cidr",
		Action: egv1a1.AuthorizationActionAllow,
		Principal: ir.Principal{
			ClientCIDRs: []*ir.CIDRMatch{
				{
					CIDR: "10.0.0.0/8",
				},
			},
		},
	}

	testCases := []struct {
		name       string
		auth       *ir.Authorization
		wantShadow bool
	}{
		{
			name: "default mode is enforce",
			auth: &ir.Authorization{
				DefaultAction: egv1a1.AuthorizationActionDeny,
				Rules:         []*ir.AuthorizationRule{testRule},
			},
			wantShadow: false,
		},
		{
			name: "explicit enforce mode",
			auth: &ir.Authorization{
				Mode:          ptr.To(egv1a1.AuthorizationModeEnforce),
				DefaultAction: egv1a1.AuthorizationActionDeny,
				Rules:         []*ir.AuthorizationRule{testRule},
			},
			wantShadow: false,
		},
		{
			name: "shadow mode",
			auth: &ir.Authorization{
				Mode:          ptr.To(egv1a1.AuthorizationModeShadow),
				DefaultAction: egv1a1.AuthorizationActionDeny,
				Rules:         []*ir.AuthorizationRule{testRule},
			},
			wantShadow: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := buildTCPRBACConfig("test-prefix", tc.auth)
			require.NoError(t, err)
			require.NotNil(t, cfg)

			if tc.wantShadow {
				require.NotNil(t, cfg.ShadowMatcher)
				require.Nil(t, cfg.Matcher)
			} else {
				require.NotNil(t, cfg.Matcher)
				require.Nil(t, cfg.ShadowMatcher)
			}
		})
	}
}
