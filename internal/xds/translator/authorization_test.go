// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"fmt"
	"regexp"
	"strings"
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

// ---- buildSanStringMatcherFromEG ----
//
// Envoy's uri_san/dns_san inputs join every SAN of that type on the peer certificate
// into one comma-separated string. buildSanStringMatcherFromEG lowers Exact/Prefix/
// Suffix to a regex anchored on comma boundaries so they match one token anywhere in
// that string, per the formulas in the doc comment:
//
//	Exact X   -> (^|,)X(,|$)
//	Prefix P  -> (^|,)P.*(,|$)
//	Suffix S  -> (^|,).*S(,|$)

func sanRegexPattern(t *testing.T, sm *egv1a1.StringMatch) string {
	t.Helper()
	matcher, err := buildSanStringMatcherFromEG(sm)
	require.NoError(t, err)
	safeRegex, ok := matcher.MatchPattern.(*matcherv3.StringMatcher_SafeRegex)
	require.True(t, ok, "expected a SafeRegex match pattern, got %T", matcher.MatchPattern)
	return safeRegex.SafeRegex.Regex
}

func TestBuildSanStringMatcherFromEG_Formulas(t *testing.T) {
	cases := []struct {
		desc     string
		matchTyp egv1a1.StringMatchType
		value    string
		want     string
	}{
		{"exact", egv1a1.StringMatchExact, "foo", `(^|,)foo(,|$)`},
		{"prefix", egv1a1.StringMatchPrefix, "foo", `(^|,)foo.*(,|$)`},
		{"suffix", egv1a1.StringMatchSuffix, "foo", `(^|,).*foo(,|$)`},
		// Regex metacharacters in the literal value must be escaped, or they'd change
		// the meaning of the anchor wrapper instead of matching themselves literally.
		{"exact escapes metacharacters", egv1a1.StringMatchExact, "a.b+c", `(^|,)a\.b\+c(,|$)`},
		{"prefix escapes metacharacters", egv1a1.StringMatchPrefix, "a.b+c", `(^|,)a\.b\+c.*(,|$)`},
		{"suffix escapes metacharacters", egv1a1.StringMatchSuffix, "a.b+c", `(^|,).*a\.b\+c(,|$)`},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			sm := &egv1a1.StringMatch{Type: ptr.To(tc.matchTyp), Value: tc.value}
			require.Equal(t, tc.want, sanRegexPattern(t, sm))
		})
	}
}

func TestBuildSanStringMatcherFromEG_NilTypeDefaultsToExactFormula(t *testing.T) {
	sm := &egv1a1.StringMatch{Value: "foo"} // Type is nil
	require.Equal(t, `(^|,)foo(,|$)`, sanRegexPattern(t, sm))
}

func TestBuildSanStringMatcherFromEG_RegularExpressionPassesThroughUnwrapped(t *testing.T) {
	// A user-authored regex is not spliced into the anchor wrapper: it continues to
	// match against the raw joined SAN string exactly as buildXdsStringMatcherFromEG
	// would build it.
	sm := &egv1a1.StringMatch{Type: ptr.To(egv1a1.StringMatchRegularExpression), Value: "^foo.*bar$"}
	got, err := buildSanStringMatcherFromEG(sm)
	require.NoError(t, err)
	want, err := buildXdsStringMatcherFromEG(sm)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestBuildSanStringMatcherFromEG_UnknownTypeReturnsError(t *testing.T) {
	sm := &egv1a1.StringMatch{Type: ptr.To(egv1a1.StringMatchType("bogus")), Value: "x"}
	_, err := buildSanStringMatcherFromEG(sm)
	require.Error(t, err)
}

// TestBuildSanStringMatcherFromEG_MatchesAnyTokenInJoinedSAN is the parametric hunt for
// the bug the anchoring fix addresses: a naive Exact/Prefix/Suffix match against the
// whole joined SAN string only ever "works" by accident for the first (Prefix) or last
// (Suffix) SAN, and never for Exact once there's more than one SAN. It exercises every
// combination of 0..5 URI SANs and 0..5 DNS SANs, verifying for every token at every
// position that:
//   - Exact/Prefix/Suffix match the real joined string built from that token, and
//   - a decoy value straddling two adjacent tokens (spanning the comma that separates
//     them) never matches, which is exactly the false-negative/false-positive pattern
//     an un-anchored substring/prefix/suffix check would be fooled by.
func TestBuildSanStringMatcherFromEG_MatchesAnyTokenInJoinedSAN(t *testing.T) {
	uriToken := func(i int) string { return fmt.Sprintf("spiffe://trust-domain/ns/ns%d/sa/sa%d", i, i) }
	dnsToken := func(i int) string { return fmt.Sprintf("svc%d.example.com", i) }

	checkTokenList := func(t *testing.T, tokens []string) {
		joined := strings.Join(tokens, ",")

		for i, tok := range tokens {
			t.Run(fmt.Sprintf("token[%d]", i), func(t *testing.T) {
				mustMatch := func(desc string, sm *egv1a1.StringMatch) {
					t.Helper()
					re := regexp.MustCompile(sanRegexPattern(t, sm))
					require.True(t, re.MatchString(joined), "%s: pattern %q should match joined SAN %q", desc, re.String(), joined)
				}

				mustMatch("exact", &egv1a1.StringMatch{Type: ptr.To(egv1a1.StringMatchExact), Value: tok})

				half := len(tok) / 2
				mustMatch("prefix", &egv1a1.StringMatch{Type: ptr.To(egv1a1.StringMatchPrefix), Value: tok[:half]})
				mustMatch("suffix", &egv1a1.StringMatch{Type: ptr.To(egv1a1.StringMatchSuffix), Value: tok[half:]})
			})
		}

		t.Run("absent value does not match", func(t *testing.T) {
			re := regexp.MustCompile(sanRegexPattern(t, &egv1a1.StringMatch{
				Type: ptr.To(egv1a1.StringMatchExact), Value: "not-present-value",
			}))
			require.False(t, re.MatchString(joined))
		})

		if len(tokens) >= 2 {
			t.Run("boundary-straddling decoy does not match", func(t *testing.T) {
				// suffix of token[0] + "," + prefix of token[1]: a real substring of the
				// joined string, but not a real token — the anchors must reject it.
				a, b := tokens[0], tokens[1]
				decoy := a[len(a)-2:] + "," + b[:2]
				require.Contains(t, joined, decoy, "decoy must be a genuine substring of the joined SAN string to be a meaningful negative case")

				exact := regexp.MustCompile(sanRegexPattern(t, &egv1a1.StringMatch{Type: ptr.To(egv1a1.StringMatchExact), Value: decoy}))
				require.False(t, exact.MatchString(joined), "Exact must not match a value spanning two tokens")

				prefix := regexp.MustCompile(sanRegexPattern(t, &egv1a1.StringMatch{Type: ptr.To(egv1a1.StringMatchPrefix), Value: decoy}))
				require.False(t, prefix.MatchString(joined), "Prefix must not match a value that doesn't start at a token boundary")

				suffix := regexp.MustCompile(sanRegexPattern(t, &egv1a1.StringMatch{Type: ptr.To(egv1a1.StringMatchSuffix), Value: decoy}))
				require.False(t, suffix.MatchString(joined), "Suffix must not match a value that doesn't end at a token boundary")
			})
		}
	}

	for uriCount := 0; uriCount <= 5; uriCount++ {
		for dnsCount := 0; dnsCount <= 5; dnsCount++ {
			t.Run(fmt.Sprintf("uris=%d/dns=%d", uriCount, dnsCount), func(t *testing.T) {
				uris := make([]string, uriCount)
				for i := range uris {
					uris[i] = uriToken(i)
				}
				dns := make([]string, dnsCount)
				for i := range dns {
					dns[i] = dnsToken(i)
				}

				// uri_san and dns_san are independent inputs on the real cert/xDS side —
				// each type gets its own joined string — so check each list on its own.
				t.Run("uris", func(t *testing.T) { checkTokenList(t, uris) })
				t.Run("dns", func(t *testing.T) { checkTokenList(t, dns) })
			})
		}
	}
}

// Compile-time check: sslinput is referenced so the import is not pruned.
var _ = (*sslinput.SubjectInput)(nil)
