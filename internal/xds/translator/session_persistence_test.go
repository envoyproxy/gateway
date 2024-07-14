// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"
	"time"

	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	cookiev3 "github.com/envoyproxy/go-control-plane/envoy/extensions/http/stateful_session/cookie/v3"
	headerv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/http/stateful_session/header/v3"
	httpv3 "github.com/envoyproxy/go-control-plane/envoy/type/http/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"k8s.io/utils/ptr"

	"github.com/envoyproxy/gateway/internal/ir"
)

func Test_sessionPersistence_patchHCM(t *testing.T) {
	t.Parallel()
	type args struct {
		mgr        *hcmv3.HttpConnectionManager
		irListener *ir.HTTPListener
	}
	tests := []struct {
		name    string
		s       *sessionPersistence
		args    args
		wantErr bool
		wantMgr *hcmv3.HttpConnectionManager
	}{
		{
			name: "nil hcm",
			s:    &sessionPersistence{},
			args: args{
				// nil mgr
				irListener: &ir.HTTPListener{},
			},
			wantErr: true,
		},
		{
			name: "nil irListener",
			s:    &sessionPersistence{},
			args: args{
				mgr: &hcmv3.HttpConnectionManager{},
			},
			wantErr: true,
		},
		{
			// patchHCM should return early without any edit if the filter already exists
			// because patchHCM could be called multiple times for the same filter.
			name: "mgr already has the filter",
			s:    &sessionPersistence{},
			args: args{
				irListener: &ir.HTTPListener{},
				mgr: &hcmv3.HttpConnectionManager{
					HttpFilters: []*hcmv3.HttpFilter{
						{
							Name: sessionPersistenceFilter,
						},
					},
				},
			},
			wantMgr: &hcmv3.HttpConnectionManager{
				HttpFilters: []*hcmv3.HttpFilter{
					{
						Name: sessionPersistenceFilter,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "no session persistence",
			s:    &sessionPersistence{},
			args: args{
				mgr: &hcmv3.HttpConnectionManager{},
				irListener: &ir.HTTPListener{
					Routes: []*ir.HTTPRoute{
						{
							Name: "route1",
							// no session persistence config
						},
					},
				},
			},
			wantMgr: &hcmv3.HttpConnectionManager{},
		},
		{
			name: "no session persistence",
			s:    &sessionPersistence{},
			args: args{
				mgr: &hcmv3.HttpConnectionManager{},
				irListener: &ir.HTTPListener{
					Routes: []*ir.HTTPRoute{
						{
							Name: "route1",
							// no session persistence config
						},
					},
				},
			},
			wantMgr: &hcmv3.HttpConnectionManager{},
		},
		{
			name: "header-based session persistence",
			s:    &sessionPersistence{},
			args: args{
				mgr: &hcmv3.HttpConnectionManager{},
				irListener: &ir.HTTPListener{
					Routes: []*ir.HTTPRoute{
						{
							Name: "route1",
							SessionPersistence: &ir.SessionPersistence{
								SessionName: "session1",
								Header:      &ir.HeaderBasedSessionPersistence{},
							},
						},
					},
				},
			},
			wantMgr: &hcmv3.HttpConnectionManager{
				HttpFilters: []*hcmv3.HttpFilter{
					{
						Name: sessionPersistenceFilter,
						ConfigType: &hcmv3.HttpFilter_TypedConfig{
							TypedConfig: mustAnyPB(&headerv3.HeaderBasedSessionState{
								Name: "session1",
							}),
						},
					},
				},
			},
		},
		{
			name: "cookie-based session persistence (one route)",
			s:    &sessionPersistence{},
			args: args{
				mgr: &hcmv3.HttpConnectionManager{},
				irListener: &ir.HTTPListener{
					Routes: []*ir.HTTPRoute{
						{
							Name: "route1",
							SessionPersistence: &ir.SessionPersistence{
								SessionName: "session1",
								Cookie: &ir.CookieBasedSessionPersistence{
									TTL: ptr.To(time.Duration(10)),
								},
							},
						},
					},
				},
			},
			wantMgr: &hcmv3.HttpConnectionManager{
				HttpFilters: []*hcmv3.HttpFilter{
					{
						Name: sessionPersistenceFilter,
						ConfigType: &hcmv3.HttpFilter_TypedConfig{
							TypedConfig: mustAnyPB(&cookiev3.CookieBasedSessionState{
								Cookie: &httpv3.Cookie{
									Name: "session1",
									Ttl:  durationpb.New(time.Duration(10)),
									Path: "/",
								},
							}),
						},
					},
				},
			},
		},
		{
			name: "cookie-based session persistence (multiple routes)",
			s:    &sessionPersistence{},
			args: args{
				mgr: &hcmv3.HttpConnectionManager{},
				irListener: &ir.HTTPListener{
					Routes: []*ir.HTTPRoute{
						{
							Name: "route1",
							PathMatch: &ir.StringMatch{
								Prefix: ptr.To("/v1"),
							},
							SessionPersistence: &ir.SessionPersistence{
								SessionName: "session1",
								Cookie: &ir.CookieBasedSessionPersistence{
									TTL: ptr.To(time.Duration(10)),
								},
							},
						},
						{
							Name: "route2",
							PathMatch: &ir.StringMatch{
								SafeRegex: ptr.To("/v2/.*/abc"),
							},
							SessionPersistence: &ir.SessionPersistence{
								SessionName: "session1",
								Cookie: &ir.CookieBasedSessionPersistence{
									TTL: ptr.To(time.Duration(10)),
								},
							},
						},
					},
				},
			},
			wantMgr: &hcmv3.HttpConnectionManager{
				HttpFilters: []*hcmv3.HttpFilter{
					{
						Name: sessionPersistenceFilter,
						ConfigType: &hcmv3.HttpFilter_TypedConfig{
							TypedConfig: mustAnyPB(&cookiev3.CookieBasedSessionState{
								Cookie: &httpv3.Cookie{
									Name: "session1",
									Ttl:  durationpb.New(time.Duration(10)),
									Path: "/v1",
								},
							}),
						},
					},
					{
						Name: sessionPersistenceFilter,
						ConfigType: &hcmv3.HttpFilter_TypedConfig{
							TypedConfig: mustAnyPB(&cookiev3.CookieBasedSessionState{
								Cookie: &httpv3.Cookie{
									Name: "session1",
									Ttl:  durationpb.New(time.Duration(10)),
									Path: "/v2",
								},
							}),
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &sessionPersistence{}
			mgr := tt.args.mgr
			err := s.patchHCM(mgr, tt.args.irListener)
			if (err != nil) != tt.wantErr {
				t.Errorf("sessionPersistence.patchHCM() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				// No need to continue
				return
			}

			if diff := cmp.Diff(tt.wantMgr, mgr, cmpopts.IgnoreUnexported(
				hcmv3.HttpConnectionManager{},
				hcmv3.HttpFilter{},
				headerv3.HeaderBasedSessionState{},
				hcmv3.HttpFilter_TypedConfig{},
				anypb.Any{})); diff != "" {
				t.Errorf("sessionPersistence.patchHCM() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func mustAnyPB(m proto.Message) *anypb.Any {
	a, err := anypb.New(m)
	if err != nil {
		panic(err)
	}
	return a
}
