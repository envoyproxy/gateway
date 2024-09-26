// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build celvalidation
// +build celvalidation

package celvalidation

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestHTTPRouteFilter(t *testing.T) {
	ctx := context.Background()
	baseHTTPRouteFilter := egv1a1.HTTPRouteFilter{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hrf",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: egv1a1.HTTPRouteFilterSpec{},
	}

	cases := []struct {
		desc         string
		mutate       func(httproutefilter *egv1a1.HTTPRouteFilter)
		mutateStatus func(httproutefilter *egv1a1.HTTPRouteFilter)
		wantErrors   []string
	}{
		{
			desc: "Valid RegexHTTPPathModifier",
			mutate: func(httproutefilter *egv1a1.HTTPRouteFilter) {
				httproutefilter.Spec = egv1a1.HTTPRouteFilterSpec{
					URLRewrite: &egv1a1.HTTPURLRewriteFilter{
						Path: &egv1a1.HTTPPathModifier{
							Type: egv1a1.RegexHTTPPathModifier,
							ReplaceRegexMatch: &egv1a1.ReplaceRegexMatch{
								Pattern:      "foo",
								Substitution: "bar",
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "invalid RegexHTTPPathModifier missing settings",
			mutate: func(httproutefilter *egv1a1.HTTPRouteFilter) {
				httproutefilter.Spec = egv1a1.HTTPRouteFilterSpec{
					URLRewrite: &egv1a1.HTTPURLRewriteFilter{
						Path: &egv1a1.HTTPPathModifier{
							Type: egv1a1.RegexHTTPPathModifier,
						},
					},
				}
			},
			wantErrors: []string{"spec.urlRewrite.path: Invalid value: \"object\": If HTTPPathModifier type is ReplaceRegexMatch, replaceRegexMatch field needs to be set."},
		},
		{
			desc: "invalid RegexHTTPPathModifier missing pattern and substitution",
			mutate: func(httproutefilter *egv1a1.HTTPRouteFilter) {
				httproutefilter.Spec = egv1a1.HTTPRouteFilterSpec{
					URLRewrite: &egv1a1.HTTPURLRewriteFilter{
						Path: &egv1a1.HTTPPathModifier{
							Type: egv1a1.RegexHTTPPathModifier,
							ReplaceRegexMatch: &egv1a1.ReplaceRegexMatch{
								Pattern:      "",
								Substitution: "",
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.urlRewrite.path.replaceRegexMatch.pattern: Invalid value: \"\": spec.urlRewrite.path.replaceRegexMatch.pattern in body should be at least 1 chars long",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			httpRouteFilter := baseHTTPRouteFilter.DeepCopy()
			httpRouteFilter.Name = fmt.Sprintf("hrf-%v", time.Now().UnixNano())

			if tc.mutate != nil {
				tc.mutate(httpRouteFilter)
			}
			err := c.Create(ctx, httpRouteFilter)

			if tc.mutateStatus != nil {
				tc.mutateStatus(httpRouteFilter)
				err = c.Status().Update(ctx, httpRouteFilter)
			}

			if (len(tc.wantErrors) != 0) != (err != nil) {
				t.Fatalf("Unexpected response while creating HTTPRouteFilter; got err=\n%v\n;want error=%v", err, tc.wantErrors)
			}

			var missingErrorStrings []string
			for _, wantError := range tc.wantErrors {
				if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(wantError)) {
					missingErrorStrings = append(missingErrorStrings, wantError)
				}
			}
			if len(missingErrorStrings) != 0 {
				t.Errorf("Unexpected response while creating HTTPRouteFilter; got err=\n%v\n;missing strings within error=%q", err, missingErrorStrings)
			}
		})
	}
}
