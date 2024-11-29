package translator

import (
	"testing"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

func Test_determineIPFamily(t *testing.T) {
	tests := []struct {
		name     string
		settings []*ir.DestinationSetting
		want     *v1alpha1.IPFamily
	}{
		{
			name:     "nil settings should return nil",
			settings: nil,
			want:     nil,
		},
		{
			name:     "empty settings should return nil",
			settings: []*ir.DestinationSetting{},
			want:     nil,
		},
		{
			name: "single IPv4 setting",
			settings: []*ir.DestinationSetting{
				{IPFamily: ptr.To(v1alpha1.IPv4)},
			},
			want: ptr.To(v1alpha1.IPv4),
		},
		{
			name: "single IPv6 setting",
			settings: []*ir.DestinationSetting{
				{IPFamily: ptr.To(v1alpha1.IPv6)},
			},
			want: ptr.To(v1alpha1.IPv6),
		},
		{
			name: "single DualStack setting",
			settings: []*ir.DestinationSetting{
				{IPFamily: ptr.To(v1alpha1.DualStack)},
			},
			want: ptr.To(v1alpha1.DualStack),
		},
		{
			name: "mixed IPv4 and IPv6 should return DualStack",
			settings: []*ir.DestinationSetting{
				{IPFamily: ptr.To(v1alpha1.IPv4)},
				{IPFamily: ptr.To(v1alpha1.IPv6)},
			},
			want: ptr.To(v1alpha1.DualStack),
		},
		{
			name: "DualStack with IPv4 should return DualStack",
			settings: []*ir.DestinationSetting{
				{IPFamily: ptr.To(v1alpha1.DualStack)},
				{IPFamily: ptr.To(v1alpha1.IPv4)},
			},
			want: ptr.To(v1alpha1.DualStack),
		},
		{
			name: "DualStack with IPv6 should return DualStack",
			settings: []*ir.DestinationSetting{
				{IPFamily: ptr.To(v1alpha1.DualStack)},
				{IPFamily: ptr.To(v1alpha1.IPv6)},
			},
			want: ptr.To(v1alpha1.DualStack),
		},
		{
			name: "mixed with nil IPFamily should be ignored",
			settings: []*ir.DestinationSetting{
				{IPFamily: ptr.To(v1alpha1.IPv4)},
				{IPFamily: nil},
				{IPFamily: ptr.To(v1alpha1.IPv6)},
			},
			want: ptr.To(v1alpha1.DualStack),
		},
		{
			name: "multiple IPv4 settings should return IPv4",
			settings: []*ir.DestinationSetting{
				{IPFamily: ptr.To(v1alpha1.IPv4)},
				{IPFamily: ptr.To(v1alpha1.IPv4)},
			},
			want: ptr.To(v1alpha1.IPv4),
		},
		{
			name: "multiple IPv6 settings should return IPv6",
			settings: []*ir.DestinationSetting{
				{IPFamily: ptr.To(v1alpha1.IPv6)},
				{IPFamily: ptr.To(v1alpha1.IPv6)},
			},
			want: ptr.To(v1alpha1.IPv6),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineIPFamily(tt.settings)
			assert.Equal(t, tt.want, got)
		})
	}
}
