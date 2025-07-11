// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"sort"

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
)

type XdsIRRoutes []*ir.HTTPRoute

func (x XdsIRRoutes) Len() int      { return len(x) }
func (x XdsIRRoutes) Swap(i, j int) { x[i], x[j] = x[j], x[i] }
func (x XdsIRRoutes) Less(i, j int) bool {
	// 1. Sort based on path match type
	// Exact > RegularExpression > PathPrefix
	if x[i].PathMatch != nil && x[i].PathMatch.Exact != nil {
		if x[j].PathMatch != nil {
			if x[j].PathMatch.SafeRegex != nil {
				return false
			}
			if x[j].PathMatch.Prefix != nil {
				return false
			}
		}
	}
	if x[i].PathMatch != nil && x[i].PathMatch.SafeRegex != nil {
		if x[j].PathMatch != nil {
			if x[j].PathMatch.Exact != nil {
				return true
			}
			if x[j].PathMatch.Prefix != nil {
				return false
			}
		}
	}
	if x[i].PathMatch != nil && x[i].PathMatch.Prefix != nil {
		if x[j].PathMatch != nil {
			if x[j].PathMatch.Exact != nil {
				return true
			}
			if x[j].PathMatch.SafeRegex != nil {
				return true
			}
		}
	}
	// Equal case

	// 2. Sort based on characters in a matching path.
	pCountI := pathMatchCount(x[i].PathMatch)
	pCountJ := pathMatchCount(x[j].PathMatch)
	if pCountI < pCountJ {
		return true
	}
	if pCountI > pCountJ {
		return false
	}
	// Equal case

	// 3. Sort based on the number of Header matches.
	// When the number is same, sort based on number of Exact Header matches.
	hCountI := len(x[i].HeaderMatches)
	hCountJ := len(x[j].HeaderMatches)
	if hCountI < hCountJ {
		return true
	}
	if hCountI > hCountJ {
		return false
	}

	hExtNumberI := numberOfExactMatches(x[i].HeaderMatches)
	hExtNumberJ := numberOfExactMatches(x[j].HeaderMatches)
	if hExtNumberI < hExtNumberJ {
		return true
	}
	if hExtNumberI > hExtNumberJ {
		return false
	}
	// Equal case

	// 4. Sort based on the number of Query param matches.
	// When the number is same, sort based on number of Exact Query param matches.
	qCountI := len(x[i].QueryParamMatches)
	qCountJ := len(x[j].QueryParamMatches)
	if qCountI < qCountJ {
		return true
	}
	if qCountI > qCountJ {
		return false
	}

	qExtNumberI := numberOfExactMatches(x[i].QueryParamMatches)
	qExtNumberJ := numberOfExactMatches(x[j].QueryParamMatches)
	if qExtNumberI < qExtNumberJ {
		return true
	}
	if qExtNumberI > qExtNumberJ {
		return false
	}
	// Equal case

	// 5. Default sort logic based on route name (Alphabetically).
	// ns1/route1/rule/0/match/0/domain > ns1/route2/rule/0/match/0/domain
	return x[i].Name > x[j].Name
}

// sortXdsIR sorts the xdsIR based on the match precedence
// defined in the Gateway API spec.
// https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.HTTPRouteRule
func sortXdsIRMap(xdsIR resource.XdsIRMap) {
	for _, irItem := range xdsIR {
		for _, http := range irItem.HTTP {
			if !http.PreserveRouteOrder {
				// descending order
				sort.Sort(sort.Reverse(XdsIRRoutes(http.Routes)))
			}
		}
	}
}

func pathMatchCount(pathMatch *ir.StringMatch) int {
	if pathMatch != nil {
		if pathMatch.Exact != nil {
			return len(*pathMatch.Exact)
		}
		if pathMatch.SafeRegex != nil {
			return len(*pathMatch.SafeRegex)
		}
		if pathMatch.Prefix != nil {
			return len(*pathMatch.Prefix)
		}
	}
	return 0
}

func numberOfExactMatches(stringMatches []*ir.StringMatch) int {
	var cnt int
	for _, stringMatch := range stringMatches {
		if stringMatch != nil && stringMatch.Exact != nil {
			cnt++
		}
	}
	return cnt
}
