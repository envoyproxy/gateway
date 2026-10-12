// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubejwt

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authentication/serviceaccount"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func TestValidateKubeJWTPodName(t *testing.T) {
	tests := []struct {
		name    string
		extra   map[string]authenticationv1.ExtraValue
		wantErr string
	}{
		{
			name: "credential ID without pod name",
			extra: map[string]authenticationv1.ExtraValue{
				"authentication.kubernetes.io/credential-id": {"JTI=test-credential"},
			},
			wantErr: "pod name not found in token review response",
		},
		{
			name: "nil pod names",
			extra: map[string]authenticationv1.ExtraValue{
				serviceaccount.PodNameKey: nil,
			},
			wantErr: "pod name not found in token review response",
		},
		{
			name: "empty pod names",
			extra: map[string]authenticationv1.ExtraValue{
				serviceaccount.PodNameKey: {},
			},
			wantErr: "pod name not found in token review response",
		},
		{
			name: "empty pod name",
			extra: map[string]authenticationv1.ExtraValue{
				serviceaccount.PodNameKey: {""},
			},
			wantErr: "pod name not found in token review response",
		},
		{
			name: "mismatched pod name",
			extra: map[string]authenticationv1.ExtraValue{
				serviceaccount.PodNameKey: {"pod-2"},
			},
			wantErr: "pod name mismatch: expected pod-1, got pod-2",
		},
		{
			name: "matching pod name",
			extra: map[string]authenticationv1.ExtraValue{
				serviceaccount.PodNameKey: {"pod-1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/apis/authentication.k8s.io/v1/tokenreviews", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				err := json.NewEncoder(w).Encode(&authenticationv1.TokenReview{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "authentication.k8s.io/v1",
						Kind:       "TokenReview",
					},
					Status: authenticationv1.TokenReviewStatus{
						Authenticated: true,
						User: authenticationv1.UserInfo{
							Groups: []string{"system:serviceaccounts"},
							Extra:  tt.extra,
						},
					},
				})
				assert.NoError(t, err)
			}))
			t.Cleanup(server.Close)

			clientset, err := kubernetes.NewForConfig(&rest.Config{Host: server.URL})
			require.NoError(t, err)
			interceptor := &JWTAuthInterceptor{
				clientset: clientset,
				audience:  "envoy-gateway",
			}

			err = interceptor.validateKubeJWT(context.Background(), "test-token", "pod-1")
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Equal(t, codes.Unauthenticated, status.Code(err))
				assert.Equal(t, tt.wantErr, status.Convert(err).Message())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
