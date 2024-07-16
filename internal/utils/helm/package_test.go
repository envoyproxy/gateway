// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package helm

import (
	"testing"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
)

func TestPackageTool_Setup(t *testing.T) {
	type fields struct {
		chartName       string
		envSettings     *cli.EnvSettings
		actionConfig    *action.Configuration
		actionInstall   *action.Install
		actionUninstall *action.Uninstall
		valuesOpts      *values.Options
		logger          Printer
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "shouldSetup",
			fields: fields{
				chartName: "mychart",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pt := &PackageTool{
				chartName:       tt.fields.chartName,
				envSettings:     tt.fields.envSettings,
				actionConfig:    tt.fields.actionConfig,
				actionInstall:   tt.fields.actionInstall,
				actionUninstall: tt.fields.actionUninstall,
				valuesOpts:      tt.fields.valuesOpts,
				logger:          tt.fields.logger,
			}
			if err := pt.Setup(); (err != nil) != tt.wantErr {
				t.Errorf("Setup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPackageTool_setInstallOptions(t *testing.T) {
	type fields struct {
		chartName       string
		envSettings     *cli.EnvSettings
		actionConfig    *action.Configuration
		actionInstall   *action.Install
		actionUninstall *action.Uninstall
		valuesOpts      *values.Options
		logger          Printer
	}
	type args struct {
		opts *PackageOptions
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "shouldSetInstallOptions",
			fields: fields{
				chartName: "mychart",
			},
			args: args{
				opts: &PackageOptions{
					Version: "1.0.2",
					Timeout: 1 * time.Second,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pt := &PackageTool{
				chartName:       tt.fields.chartName,
				envSettings:     tt.fields.envSettings,
				actionConfig:    tt.fields.actionConfig,
				actionInstall:   tt.fields.actionInstall,
				actionUninstall: tt.fields.actionUninstall,
				valuesOpts:      tt.fields.valuesOpts,
				logger:          tt.fields.logger,
			}
			if err := pt.Setup(); err != nil {
				t.Errorf("Setup() error = %v", err)
			}
			pt.setInstallOptions(tt.args.opts)
		})
	}
}
