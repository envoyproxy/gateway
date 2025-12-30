// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"reflect"
	"regexp"
	"strings"
	"testing"

	matcher "github.com/cncf/xds/go/xds/type/matcher/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

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

func Test_buildServerNameStringMatcher(t *testing.T) {
	tests := []struct {
		name           string
		hostname       string
		wantRegex      string
		wantExact      string
		wantIgnoreCase bool
		matches        []string
		nonMatches     []string
	}{
		{
			name:      "wildcard any",
			hostname:  "*",
			wantRegex: ".*",
			matches:   []string{"example.com", "foo.bar", "EXAMPLE.COM", "FOO.bar"},
		},
		{
			name:      "wildcard suffix mixed case",
			hostname:  "*.example.com",
			wantRegex: "(?i)^[^.]+\\.example\\.com$",
			matches:   []string{"foo.example.com", "BAR.Example.COM"},
			nonMatches: []string{
				"example.com",
				"fooexample.com",
				"foo.bar.example.com",
			},
		},
		{
			name:           "exact host ignore case",
			hostname:       "foo.example.com",
			wantExact:      "foo.example.com",
			wantIgnoreCase: true,
			matches:        []string{"FOO.EXAMPLE.COM", "foo.example.com", "FOO.example.COM"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := buildServerNameStringMatcher(tt.hostname)
			switch mt := m.MatchPattern.(type) {
			case *matcher.StringMatcher_SafeRegex:
				if tt.wantRegex == "" {
					t.Fatalf("expected exact matcher, got regex: %v", mt)
				}
				assert.Equal(t, tt.wantRegex, mt.SafeRegex.Regex)
				re, err := regexp.Compile(mt.SafeRegex.Regex)
				require.NoError(t, err)
				for _, candidate := range tt.matches {
					assert.Truef(t, re.MatchString(candidate), "expected %q to match %q", candidate, mt.SafeRegex.Regex)
				}
				for _, candidate := range tt.nonMatches {
					assert.Falsef(t, re.MatchString(candidate), "expected %q not to match %q", candidate, mt.SafeRegex.Regex)
				}
			case *matcher.StringMatcher_Exact:
				if tt.wantExact == "" {
					t.Fatalf("expected regex matcher, got exact: %v", mt)
				}
				assert.Equal(t, tt.wantExact, mt.Exact)
				assert.Equal(t, tt.wantIgnoreCase, m.IgnoreCase)
				for _, candidate := range tt.matches {
					assert.Truef(t, strings.EqualFold(tt.wantExact, candidate), "expected %q to equal %q ignoring case", candidate, tt.wantExact)
				}
			default:
				t.Fatalf("unexpected matcher type: %T", mt)
			}
		})
	}
}
