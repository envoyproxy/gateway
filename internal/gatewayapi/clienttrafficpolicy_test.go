package gatewayapi

import (
	"testing"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/utils/ptr"
	"github.com/stretchr/testify/assert"
)

func TestTLSVersionToNumericVersion(t *testing.T) {
	tests := []struct {
		name  string
		input *egv1a1.TLSVersion
		want  int
		err   error
	}{
		{
			name:  "nil",
			input: nil,
			want:  -1,
			err:   nil,
		},
		{
			name:  "auto",
			input: ptr.To(egv1a1.TLSVersion("TLS_Auto")),
			want:  -1,
			err:   nil,
		},
		{
			name:  "bad",
			input: ptr.To(egv1a1.TLSVersion("not a version")),
			want:  -1,
			err:   ErrInvalidTLSProtocolVersion,
		},
		{
			name:  "v1.3",
			input: ptr.To(egv1a1.TLSVersion("TLSv1_3")),
			want:  3,
			err:   nil,
		},
		{
			name:  "v1.2",
			input: ptr.To(egv1a1.TLSVersion("TLSv1_2")),
			want:  2,
			err:   nil,
		},
		{
			name:  "v1.1",
			input: ptr.To(egv1a1.TLSVersion("TLSv1_1")),
			want:  1,
			err:   nil,
		},
		{
			name:  "v1.0",
			input: ptr.To(egv1a1.TLSVersion("TLSv1_0")),
			want:  0,
			err:   nil,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			actual, err := tlsVersionToNumericVersion(test.input)
			assert.Equal(t, test.err, err)
			assert.Equal(t, test.want, actual)
		})
	}

}
