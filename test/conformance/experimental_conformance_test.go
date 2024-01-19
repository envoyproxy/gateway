//go:build experimental
// +build experimental

// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package conformance

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/yaml"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
	confv1a1 "sigs.k8s.io/gateway-api/conformance/apis/v1alpha1"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/flags"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

var (
	cfg                 *rest.Config
	k8sClientset        *kubernetes.Clientset
	mgrClient           client.Client
	implementation      *confv1a1.Implementation
	conformanceProfiles sets.Set[suite.ConformanceProfileName]
)

func TestExperimentalConformance(t *testing.T) {
	var err error
	cfg, err = config.GetConfig()
	if err != nil {
		t.Fatalf("Error loading Kubernetes config: %v", err)
	}
	mgrClient, err = client.New(cfg, client.Options{})
	if err != nil {
		t.Fatalf("Error initializing Kubernetes client: %v", err)
	}
	k8sClientset, err = kubernetes.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("Error initializing Kubernetes REST client: %v", err)
	}

	err = v1alpha2.AddToScheme(mgrClient.Scheme())
	require.NoError(t, err)
	err = v1beta1.AddToScheme(mgrClient.Scheme())
	require.NoError(t, err)
	err = v1.AddToScheme(mgrClient.Scheme())
	require.NoError(t, err)

	// experimental conformance flags
	conformanceProfiles = sets.New(
		suite.HTTPConformanceProfileName,
		suite.TLSConformanceProfileName,
	)

	// if some conformance profiles have been set, run the experimental conformance suite...
	implementation, err = suite.ParseImplementation(
		*flags.ImplementationOrganization,
		*flags.ImplementationProject,
		*flags.ImplementationURL,
		*flags.ImplementationVersion,
		*flags.ImplementationContact,
	)
	if err != nil {
		t.Fatalf("Error parsing implementation's details: %v", err)
	}

	experimentalConformance(t)
}

func experimentalConformance(t *testing.T) {
	t.Logf("Running experimental conformance tests with %s GatewayClass\n cleanup: %t\n debug: %t\n enable all features: %t \n conformance profiles: [%v]",
		*flags.GatewayClassName, *flags.CleanupBaseResources, *flags.ShowDebug, *flags.EnableAllSupportedFeatures, conformanceProfiles)

	cSuite, err := suite.NewExperimentalConformanceTestSuite(
		suite.ExperimentalConformanceOptions{
			Options: suite.Options{
				Client:     mgrClient,
				RestConfig: cfg,
				// This clientset is needed in addition to the client only because
				// controller-runtime client doesn't support non CRUD sub-resources yet (https://github.com/kubernetes-sigs/controller-runtime/issues/452).
				Clientset:            k8sClientset,
				GatewayClassName:     *flags.GatewayClassName,
				Debug:                *flags.ShowDebug,
				CleanupBaseResources: *flags.CleanupBaseResources,
				SupportedFeatures:    suite.AllFeatures,
				SkipTests: []string{
					tests.GatewayStaticAddresses.ShortName,
				},
				ExemptFeatures: suite.MeshCoreFeatures,
			},
			Implementation:      *implementation,
			ConformanceProfiles: conformanceProfiles,
		})
	if err != nil {
		t.Fatalf("error creating experimental conformance test suite: %v", err)
	}

	cSuite.Setup(t)
	err = cSuite.Run(t, tests.ConformanceTests)
	if err != nil {
		t.Fatalf("error running conformance profile report: %v", err)
	}
	report, err := cSuite.Report()
	if err != nil {
		t.Fatalf("error generating conformance profile report: %v", err)
	}

	err = experimentalConformanceReport(t.Logf, *report, *flags.ReportOutput)
	require.NoError(t, err)
}

func experimentalConformanceReport(logf func(string, ...any), report confv1a1.ConformanceReport, output string) error {
	rawReport, err := yaml.Marshal(report)
	if err != nil {
		return err
	}

	if output != "" {
		if err = os.WriteFile(output, rawReport, 0600); err != nil {
			return err
		}
	}
	logf("Conformance report:\n%s", string(rawReport))

	return nil
}
