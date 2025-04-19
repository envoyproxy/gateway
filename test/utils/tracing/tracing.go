package tracing

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-logfmt/logfmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

func ExpectedTraceCount(t *testing.T, suite *suite.ConformanceTestSuite, gwAddr string, expectedResponse httputils.ExpectedResponse, tags map[string]string) {
	if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true,
		func(ctx context.Context) (bool, error) {
			preCount, err := queryTraceFromTempo(t, suite.Client, tags)
			if err != nil {
				tlog.Logf(t, "failed to get trace count from tempo: %v", err)
				return false, nil
			}

			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			// looks like we need almost 15 seconds to get the trace from Tempo?
			err = wait.PollUntilContextTimeout(context.TODO(), time.Second, 15*time.Second, true, func(ctx context.Context) (done bool, err error) {
				curCount, err := queryTraceFromTempo(t, suite.Client, tags)
				if err != nil {
					tlog.Logf(t, "failed to get curCount count from tempo: %v", err)
					return false, nil
				}

				if curCount > preCount {
					return true, nil
				}

				return false, nil
			})
			if err != nil {
				tlog.Logf(t, "failed to get current count from tempo: %v", err)
				return false, nil
			}

			return true, nil
		}); err != nil {
		t.Errorf("failed to get trace from tempo: %v", err)
	}
}

// queryTraceFromTempo queries span count from tempo
func queryTraceFromTempo(t *testing.T, c client.Client, tags map[string]string) (int, error) {
	svc := corev1.Service{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "monitoring",
		Name:      "tempo",
	}, &svc); err != nil {
		return -1, err
	}
	host := ""
	for _, ing := range svc.Status.LoadBalancer.Ingress {
		if ing.IP != "" {
			host = ing.IP
			break
		}
	}

	tagsQueryParam, err := createTagsQueryParam(tags)
	if err != nil {
		return -1, err
	}

	tempoURL := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, "3100"),
		Path:   "/api/search",
	}
	query := tempoURL.Query()
	query.Add("start", fmt.Sprintf("%d", time.Now().Add(-10*time.Minute).Unix())) // query traces from last 10 minutes
	query.Add("end", fmt.Sprintf("%d", time.Now().Unix()))
	query.Add("tags", tagsQueryParam)
	tempoURL.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", tempoURL.String(), nil)
	if err != nil {
		return -1, err
	}

	tlog.Logf(t, "send request to %s", tempoURL.String())
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("failed to query tempo, url=%s, status=%s", tempoURL.String(), res.Status)
	}

	resp := &tempoResponse{}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return -1, err
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		tlog.Logf(t, "Failed to unmarshall response: %s", string(data))
		return -1, err
	}

	total := len(resp.Traces)
	tlog.Logf(t, "get response from tempo, url=%s, response=%v, total=%d", tempoURL.String(), string(data), total)
	return total, nil
}

type tempoResponse struct {
	Traces []map[string]interface{} `json:"traces,omitempty"`
}

// copy from https://github.com/grafana/tempo/blob/c0127c78c368319433c7c67ca8967adbfed2259e/cmd/tempo-query/tempo/plugin.go#L361
func createTagsQueryParam(tags map[string]string) (string, error) {
	tagsBuilder := &strings.Builder{}
	tagsEncoder := logfmt.NewEncoder(tagsBuilder)
	for k, v := range tags {
		err := tagsEncoder.EncodeKeyval(k, v)
		if err != nil {
			return "", err
		}
	}
	return tagsBuilder.String(), nil
}
