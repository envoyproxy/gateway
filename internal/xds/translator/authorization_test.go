// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	matcherv3 "github.com/cncf/xds/go/xds/type/matcher/v3"
	sslinput "github.com/envoyproxy/go-control-plane/envoy/extensions/matching/common_inputs/ssl/v3"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

// ---- buildXdsStringMatcherFromEG ----

func TestBuildXdsStringMatcherFromEG_NilTypeDefaultsToExact(t *testing.T) {
	sm := &egv1a1.StringMatch{Value: "example.com"} // Type is nil
	got, err := buildXdsStringMatcherFromEG(sm)
	require.NoError(t, err)
	exact, ok := got.MatchPattern.(*matcherv3.StringMatcher_Exact)
	require.True(t, ok, "expected Exact match pattern")
	require.Equal(t, "example.com", exact.Exact)
}

func TestBuildXdsStringMatcherFromEG_UnknownTypeReturnsError(t *testing.T) {
	sm := &egv1a1.StringMatch{
		Type:  ptr.To(egv1a1.StringMatchType("bogus")),
		Value: "x",
	}
	_, err := buildXdsStringMatcherFromEG(sm)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported")
}

func TestBuildXdsStringMatcherFromEG_RegexHasGoogleRe2Engine(t *testing.T) {
	sm := &egv1a1.StringMatch{
		Type:  ptr.To(egv1a1.StringMatchRegularExpression),
		Value: "^spiffe://cluster\\.local/.*$",
	}
	got, err := buildXdsStringMatcherFromEG(sm)
	require.NoError(t, err)

	safeRegex, ok := got.MatchPattern.(*matcherv3.StringMatcher_SafeRegex)
	require.True(t, ok, "expected SafeRegex match pattern")
	require.Equal(t, "^spiffe://cluster\\.local/.*$", safeRegex.SafeRegex.Regex)
	// The GoogleRe2 engine must be set (this was the bug we fixed).
	googleRe2, ok := safeRegex.SafeRegex.EngineType.(*matcherv3.RegexMatcher_GoogleRe2)
	require.True(t, ok, "expected GoogleRe2 engine type")
	require.NotNil(t, googleRe2.GoogleRe2)
}

// ---- buildClientCertPredicate ----

// inputNameFromPredicate extracts the TypedExtensionConfig.Name from a
// SinglePredicate — it panics if the predicate is not a SinglePredicate.
func inputNameFromPredicate(p *matcherv3.Matcher_MatcherList_Predicate) string {
	sp := p.MatchType.(*matcherv3.Matcher_MatcherList_Predicate_SinglePredicate_)
	return sp.SinglePredicate.Input.Name
}

func TestBuildClientCertPredicate_SubjectOnly(t *testing.T) {
	cc := &egv1a1.ClientCertPrincipal{
		Subject: &egv1a1.StringMatch{Value: "CN=client.example.com"},
	}
	preds, err := buildClientCertPredicate(cc)
	require.NoError(t, err)
	require.Len(t, preds, 1)
	require.Equal(t, "client_cert_subject", inputNameFromPredicate(preds[0]))
}

func TestBuildClientCertPredicate_SANNonNilBothEmpty_NoError(t *testing.T) {
	// san != nil but both URIs and DNSNames are empty → no URI/DNS predicates,
	// no error; result is the empty slice (no Subject either).
	cc := &egv1a1.ClientCertPrincipal{
		Subject:         &egv1a1.StringMatch{Value: "CN=client"},
		SubjectAltNames: &egv1a1.SubjectAltNames{}, // non-nil, but empty lists
	}
	preds, err := buildClientCertPredicate(cc)
	require.NoError(t, err)
	// Subject predicate still present; no SAN predicates added.
	require.Len(t, preds, 1)
	require.Equal(t, "client_cert_subject", inputNameFromPredicate(preds[0]))
}

func TestBuildClientCertPredicate_EmailAddressesReturnsError(t *testing.T) {
	cc := &egv1a1.ClientCertPrincipal{
		SubjectAltNames: &egv1a1.SubjectAltNames{
			EmailAddresses: []egv1a1.StringMatch{{Value: "user@example.com"}},
		},
	}
	_, err := buildClientCertPredicate(cc)
	require.Error(t, err)
	require.Contains(t, err.Error(), "emailAddresses")
}

func TestBuildClientCertPredicate_SubjectAndURI_TwoPredicates(t *testing.T) {
	cc := &egv1a1.ClientCertPrincipal{
		Subject: &egv1a1.StringMatch{Value: "CN=client"},
		SubjectAltNames: &egv1a1.SubjectAltNames{
			URIs: []egv1a1.StringMatch{{Value: "spiffe://cluster.local/ns/default/sa/client"}},
		},
	}
	preds, err := buildClientCertPredicate(cc)
	require.NoError(t, err)
	require.Len(t, preds, 2)
	require.Equal(t, "client_cert_subject", inputNameFromPredicate(preds[0]))
	require.Equal(t, "client_cert_uri_san", inputNameFromPredicate(preds[1]))
}

func TestBuildClientCertPredicate_MultipleURIs_ORWrapped(t *testing.T) {
	cc := &egv1a1.ClientCertPrincipal{
		SubjectAltNames: &egv1a1.SubjectAltNames{
			URIs: []egv1a1.StringMatch{
				{Value: "spiffe://cluster.local/ns/default/sa/alice"},
				{Value: "spiffe://cluster.local/ns/default/sa/bob"},
			},
		},
	}
	preds, err := buildClientCertPredicate(cc)
	require.NoError(t, err)
	require.Len(t, preds, 1)
	// The single returned predicate must be an OR matcher.
	orMatcher, ok := preds[0].MatchType.(*matcherv3.Matcher_MatcherList_Predicate_OrMatcher)
	require.True(t, ok, "expected OrMatcher for multiple URIs")
	require.Len(t, orMatcher.OrMatcher.Predicate, 2)
	// Each inner predicate should reference the URI SAN input.
	for _, inner := range orMatcher.OrMatcher.Predicate {
		require.Equal(t, "client_cert_uri_san", inputNameFromPredicate(inner))
	}
}

func TestBuildClientCertPredicate_URIAndDNS_SingleORGroup(t *testing.T) {
	// URI and DNS SANs on the same principal group into ONE OR predicate, which is then
	// AND-combined with Subject: AND(Subject, OR(uri, dns)).
	cc := &egv1a1.ClientCertPrincipal{
		Subject: &egv1a1.StringMatch{Value: "CN=svc"},
		SubjectAltNames: &egv1a1.SubjectAltNames{
			URIs:     []egv1a1.StringMatch{{Value: "spiffe://a"}},
			DNSNames: []egv1a1.StringMatch{{Value: "svc.example.com"}},
		},
	}
	preds, err := buildClientCertPredicate(cc)
	require.NoError(t, err)
	// [Subject, OR(uri, dns)] — Subject AND'd with a single combined SAN group.
	require.Len(t, preds, 2)
	require.Equal(t, "client_cert_subject", inputNameFromPredicate(preds[0]))

	orMatcher, ok := preds[1].MatchType.(*matcherv3.Matcher_MatcherList_Predicate_OrMatcher)
	require.True(t, ok, "expected a single OR group combining URI and DNS SANs")
	require.Len(t, orMatcher.OrMatcher.Predicate, 2)
	inputs := []string{
		inputNameFromPredicate(orMatcher.OrMatcher.Predicate[0]),
		inputNameFromPredicate(orMatcher.OrMatcher.Predicate[1]),
	}
	require.ElementsMatch(t, []string{"client_cert_uri_san", "client_cert_dns_san"}, inputs)
}

func TestBuildClientCertPredicate_URIAndDNS_NoSubject_SingleORGroup(t *testing.T) {
	// Without Subject, a URI+DNS mix collapses to exactly one OR predicate (no AND).
	cc := &egv1a1.ClientCertPrincipal{
		SubjectAltNames: &egv1a1.SubjectAltNames{
			URIs:     []egv1a1.StringMatch{{Value: "spiffe://a"}, {Value: "spiffe://b"}},
			DNSNames: []egv1a1.StringMatch{{Value: "svc.example.com"}},
		},
	}
	preds, err := buildClientCertPredicate(cc)
	require.NoError(t, err)
	require.Len(t, preds, 1)
	orMatcher, ok := preds[0].MatchType.(*matcherv3.Matcher_MatcherList_Predicate_OrMatcher)
	require.True(t, ok, "expected one OR group for the URI+DNS mix")
	require.Len(t, orMatcher.OrMatcher.Predicate, 3)
}

// Compile-time check: sslinput is referenced so the import is not pruned.
var _ = (*sslinput.SubjectInput)(nil)
