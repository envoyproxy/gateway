// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package e2e

import (
	"os"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

const (
	e2eWarnDurationEnv    = "EG_E2E_WARN_DURATION"
	defaultWarnDuration   = 1 * time.Minute
	minDurationResolution = time.Millisecond
)

type timingRecord struct {
	name     string
	duration time.Duration
	failed   bool
	skipped  bool
}

// TimingRecorder stores per-test durations for reporting.
type TimingRecorder struct {
	mu      sync.Mutex
	records []timingRecord
}

// NewTimingRecorder creates a new recorder for test durations.
func NewTimingRecorder() *TimingRecorder {
	return &TimingRecorder{}
}

// Report logs test durations sorted by time.
func (r *TimingRecorder) Report(t *testing.T) {
	if r == nil {
		return
	}
	r.mu.Lock()
	records := make([]timingRecord, len(r.records))
	copy(records, r.records)
	r.mu.Unlock()

	if len(records) == 0 {
		return
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].duration > records[j].duration
	})

	tlog.Logf(t, "E2E timing report (sorted by duration)")
	for _, rec := range records {
		status := "PASS"
		if rec.skipped {
			status = "SKIP"
		} else if rec.failed {
			status = "FAIL"
		}
		tlog.Logf(t, "%s %s (%s)", status, rec.name, rec.duration)
	}
}

// WrapConformanceTestsWithTiming logs per-test duration and warns for slow tests.
func WrapConformanceTestsWithTiming(tests []suite.ConformanceTest, recorder *TimingRecorder) []suite.ConformanceTest {
	warnAfter := parseWarnDuration()
	wrapped := make([]suite.ConformanceTest, 0, len(tests))

	for _, test := range tests {
		original := test.Test
		test.Test = func(t *testing.T, suite *suite.ConformanceTestSuite) {
			start := time.Now()
			defer func() {
				duration := time.Since(start).Round(minDurationResolution)
				tlog.Logf(t, "Test %s completed in %s", test.ShortName, duration)
				if warnAfter > 0 && duration > warnAfter {
					tlog.Logf(t, "WARNING: Test %s took %s (threshold %s)", test.ShortName, duration, warnAfter)
				}
				if recorder != nil {
					recorder.mu.Lock()
					recorder.records = append(recorder.records, timingRecord{
						name:     test.ShortName,
						duration: duration,
						failed:   t.Failed(),
						skipped:  t.Skipped(),
					})
					recorder.mu.Unlock()
				}
			}()
			original(t, suite)
		}
		wrapped = append(wrapped, test)
	}

	return wrapped
}

func parseWarnDuration() time.Duration {
	raw := strings.TrimSpace(os.Getenv(e2eWarnDurationEnv))
	if raw == "" {
		return defaultWarnDuration
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return defaultWarnDuration
	}
	if d <= 0 {
		return 0
	}
	return d
}
