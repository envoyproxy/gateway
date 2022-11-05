// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	fileaccesslog "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/file/v3"
)

var (
	stdoutFileAccessLog = &fileaccesslog.FileAccessLog{
		Path: "/dev/stdout",
	}

	listenerAccessLogFilter = &accesslog.AccessLogFilter{
		FilterSpecifier: &accesslog.AccessLogFilter_ResponseFlagFilter{
			ResponseFlagFilter: &accesslog.ResponseFlagFilter{Flags: []string{"NR"}},
		},
	}
)
