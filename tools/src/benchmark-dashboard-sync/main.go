// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	defaultRepoOwner = "envoyproxy"
	defaultRepoName  = "gateway"
	userAgent        = "envoy-gateway-benchmark-dashboard-sync"
)

var versionRE = regexp.MustCompile(`^v?\d+\.\d+\.\d+$`)

type options struct {
	version string
	force   bool
}

type syncer struct {
	repoRoot string
	owner    string
	repo     string
	client   *http.Client
	now      func() time.Time
	stderr   io.Writer
}

type releaseBenchmarkResult struct {
	Metadata struct {
		TestConfiguration json.RawMessage `json:"testConfiguration"`
	} `json:"metadata"`
	Results json.RawMessage `json:"results"`
}

type githubRelease struct {
	PublishedAt string `json:"published_at"`
}

type tsSuite struct {
	Metadata tsMetadata      `json:"metadata"`
	Results  json.RawMessage `json:"results"`
}

type tsMetadata struct {
	Version           string          `json:"version"`
	RunID             string          `json:"runId"`
	Date              string          `json:"date"`
	Environment       string          `json:"environment"`
	Description       string          `json:"description"`
	DownloadURL       string          `json:"downloadUrl"`
	TestConfiguration json.RawMessage `json:"testConfiguration"`
}

func main() {
	opts, err := parseFlags(os.Args[1:])
	if err != nil {
		exitf("%v\n", err)
	}

	root, err := os.Getwd()
	if err != nil {
		exitf("get working directory: %v\n", err)
	}
	root, err = findRepoRoot(root)
	if err != nil {
		exitf("%v\n", err)
	}

	s := syncer{
		repoRoot: root,
		owner:    defaultRepoOwner,
		repo:     defaultRepoName,
		client:   &http.Client{Timeout: 30 * time.Second},
		now:      func() time.Time { return time.Now().UTC() },
		stderr:   os.Stderr,
	}

	if err := s.run(context.Background(), opts); err != nil {
		exitf("%v\n", err)
	}
}

func parseFlags(args []string) (options, error) {
	fs := flag.NewFlagSet("benchmark-dashboard-sync", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var opts options
	fs.StringVar(&opts.version, "version", "", "release version in X.Y.Z or vX.Y.Z format")
	fs.BoolVar(&opts.force, "force", false, "overwrite an existing version file")

	if err := fs.Parse(args); err != nil {
		return options{}, err
	}
	if opts.version == "" {
		return options{}, errors.New("missing required --version")
	}

	return opts, nil
}

func (s syncer) run(ctx context.Context, opts options) error {
	normalized, tag, err := normalizeVersion(opts.version)
	if err != nil {
		return err
	}

	versionFilePath := filepath.Join(s.repoRoot, "site", "tools", "benchmark-dashboard", "src", "data", "versions", tag+".ts")
	if !opts.force {
		if _, statErr := os.Stat(versionFilePath); statErr == nil {
			return fmt.Errorf("version file already exists: %s (rerun with --force to overwrite)", versionFilePath)
		} else if !errors.Is(statErr, os.ErrNotExist) {
			return fmt.Errorf("stat version file: %w", statErr)
		}
	}

	benchmarkBytes, assetTimestamp, err := s.downloadBenchmarkJSON(ctx, tag)
	if err != nil {
		return err
	}

	publishedAt, fallbackSource, err := s.releasePublishedAt(ctx, tag, assetTimestamp)
	if err != nil {
		return err
	}
	if fallbackSource == "current time" {
		fmt.Fprintf(s.stderr, "warning: failed to fetch GitHub release published_at, using current UTC time %s\n", publishedAt)
	}

	output, err := generateVersionFile(normalized, tag, publishedAt, benchmarkBytes)
	if err != nil {
		return err
	}

	if err := os.WriteFile(versionFilePath, output, 0o644); err != nil {
		return fmt.Errorf("write version file: %w", err)
	}

	indexPath := filepath.Join(s.repoRoot, "site", "tools", "benchmark-dashboard", "src", "data", "index.ts")
	indexBytes, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("read index.ts: %w", err)
	}

	updatedIndex, err := updateIndex(string(indexBytes), normalized, tag, opts.force)
	if err != nil {
		return err
	}
	if err := os.WriteFile(indexPath, []byte(updatedIndex), 0o644); err != nil {
		return fmt.Errorf("write index.ts: %w", err)
	}

	fmt.Fprintf(s.stderr, "updated %s and %s\n", versionFilePath, indexPath)
	return nil
}

func normalizeVersion(version string) (string, string, error) {
	if !versionRE.MatchString(version) {
		return "", "", fmt.Errorf("invalid version %q: expected X.Y.Z or vX.Y.Z", version)
	}

	normalized := strings.TrimPrefix(version, "v")
	return normalized, "v" + normalized, nil
}

func findRepoRoot(start string) (string, error) {
	dir := start
	for {
		if repoLooksValid(dir) {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find repository root from %s", start)
		}
		dir = parent
	}
}

func repoLooksValid(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "site", "tools", "benchmark-dashboard", "src", "data", "index.ts"))
	return err == nil
}

func (s syncer) downloadBenchmarkJSON(ctx context.Context, tag string) ([]byte, string, error) {
	url := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/benchmark_result.json", s.owner, s.repo, tag)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("build benchmark download request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, "", fmt.Errorf("download %s: unexpected status %s: %s", url, resp.Status, strings.TrimSpace(string(body)))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("read %s: %w", url, err)
	}

	return data, parseLastModified(resp.Header.Get("Last-Modified")), nil
}

func (s syncer) releasePublishedAt(ctx context.Context, tag, assetTimestamp string) (string, string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", s.owner, s.repo, tag)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", fmt.Errorf("build release metadata request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := s.client.Do(req)
	if err != nil {
		ts, source := fallbackPublishedAt(assetTimestamp, s.now)
		return ts, source, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ts, source := fallbackPublishedAt(assetTimestamp, s.now)
		return ts, source, nil
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		ts, source := fallbackPublishedAt(assetTimestamp, s.now)
		return ts, source, nil
	}
	if release.PublishedAt == "" {
		ts, source := fallbackPublishedAt(assetTimestamp, s.now)
		return ts, source, nil
	}

	return release.PublishedAt, "github api", nil
}

func fallbackPublishedAt(assetTimestamp string, now func() time.Time) (string, string) {
	if assetTimestamp != "" {
		return assetTimestamp, "asset last-modified"
	}
	return now().Format(time.RFC3339), "current time"
}

func parseLastModified(value string) string {
	if value == "" {
		return ""
	}
	t, err := time.Parse(http.TimeFormat, value)
	if err != nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func generateVersionFile(normalizedVersion, tagVersion, publishedAt string, benchmarkBytes []byte) ([]byte, error) {
	var benchmark releaseBenchmarkResult
	if err := json.Unmarshal(benchmarkBytes, &benchmark); err != nil {
		return nil, fmt.Errorf("parse benchmark_result.json: %w", err)
	}
	if len(benchmark.Metadata.TestConfiguration) == 0 {
		return nil, errors.New("benchmark_result.json is missing metadata.testConfiguration")
	}
	if len(benchmark.Results) == 0 {
		return nil, errors.New("benchmark_result.json is missing results")
	}

	testConfiguration, err := normalizeTestConfiguration(benchmark.Metadata.TestConfiguration)
	if err != nil {
		return nil, err
	}

	dateValue := publishedAt
	parsedTime, err := time.Parse(time.RFC3339, publishedAt)
	if err == nil {
		dateValue = parsedTime.UTC().Format(time.RFC3339)
	}

	suite := tsSuite{
		Metadata: tsMetadata{
			Version:           normalizedVersion,
			RunID:             fmt.Sprintf("%s-release-%s", normalizedVersion, dateValue[:10]),
			Date:              dateValue,
			Environment:       "GitHub Release",
			Description:       fmt.Sprintf("Benchmark results for version %s from release artifacts", normalizedVersion),
			DownloadURL:       fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/benchmark_report.zip", defaultRepoOwner, defaultRepoName, tagVersion),
			TestConfiguration: testConfiguration,
		},
		Results: benchmark.Results,
	}

	suiteBytes, err := json.MarshalIndent(suite, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal generated suite: %w", err)
	}

	content := strings.Join([]string{
		"import { TestSuite } from '../types';",
		"",
		fmt.Sprintf("// Benchmark data extracted from release artifact for version %s", normalizedVersion),
		"// Generated from benchmark_result.json",
		"",
		"export const benchmarkData: TestSuite = " + string(suiteBytes) + ";",
		"",
	}, "\n")

	return []byte(content), nil
}

func normalizeTestConfiguration(raw json.RawMessage) (json.RawMessage, error) {
	var cfg map[string]any
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse metadata.testConfiguration: %w", err)
	}

	normalized := map[string]any{
		"rps":         mustNumber(cfg, "rps"),
		"connections": mustNumber(cfg, "connections", "connection"),
		"duration":    mustNumber(cfg, "duration"),
		"cpuLimit":    mustString(cfg, "cpuLimit", "cpu"),
		"memoryLimit": mustString(cfg, "memoryLimit", "memory"),
	}

	data, err := json.Marshal(normalized)
	if err != nil {
		return nil, fmt.Errorf("marshal normalized testConfiguration: %w", err)
	}

	return data, nil
}

func mustNumber(cfg map[string]any, keys ...string) int {
	for _, key := range keys {
		value, ok := cfg[key]
		if !ok {
			continue
		}
		switch v := value.(type) {
		case float64:
			return int(v)
		case string:
			n, err := strconv.Atoi(v)
			if err == nil {
				return n
			}
		}
	}
	return 0
}

func mustString(cfg map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := cfg[key]
		if !ok {
			continue
		}
		if s, ok := value.(string); ok {
			return s
		}
	}
	return ""
}

func updateIndex(content, normalizedVersion, tagVersion string, force bool) (string, error) {
	importAlias := fmt.Sprintf("v%sTestSuite", strings.ReplaceAll(normalizedVersion, ".", ""))
	importLine := fmt.Sprintf("import { benchmarkData as %s } from './versions/%s';", importAlias, tagVersion)
	entryLine := fmt.Sprintf("  %s,", importAlias)

	if strings.Contains(content, importLine) || strings.Contains(content, entryLine) {
		if force {
			return content, nil
		}
		return "", fmt.Errorf("index.ts already contains %s", tagVersion)
	}

	const importAnchor = "// Import all version data\n"
	idx := strings.Index(content, importAnchor)
	if idx == -1 {
		return "", errors.New("index.ts import anchor not found")
	}
	content = content[:idx] + importLine + "\n\n" + content[idx:]

	const suiteAnchor = "export const allTestSuites: TestSuite[] = [\n"
	idx = strings.Index(content, suiteAnchor)
	if idx == -1 {
		return "", errors.New("index.ts suite anchor not found")
	}
	idx += len(suiteAnchor)
	content = content[:idx] + entryLine + "\n" + content[idx:]

	return content, nil
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}
