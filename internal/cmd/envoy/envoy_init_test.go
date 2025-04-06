// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package envoy

import "testing"

func Test_writeConfig(t *testing.T) {
	type args struct {
		configPath string
		region     string
		zone       string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				configPath: "/tmp/testdata/envoy_config.yaml",
				region:     "us-east-1",
				zone:       "us-east-1a",
			},
		}, // TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := writeConfig(tt.args.configPath, tt.args.zone); (err != nil) != tt.wantErr {
				t.Errorf("writeConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
