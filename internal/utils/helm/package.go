// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package helm

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/kubectl/pkg/cmd/util"

	"github.com/envoyproxy/gateway/internal/cmd/options"
)

const (
	helmOperateTimeout = time.Second * 300
	egChartVersion     = "v1.0.0"
	egReleaseName      = "envoy-gateway"
	egReleaseNamespace = "envoy-gateway-system"
)

type PackageOptions struct {
	DryRun           bool
	SkipCRD          bool
	Wait             bool
	Version          string
	Timeout          time.Duration
	ReleaseName      string
	ReleaseNamespace string
	OnlyCRD          bool
	WithCRD          bool
}

type PackageTool struct {
	chartName string

	// Helm dependency objects
	envSettings     *cli.EnvSettings
	actionConfig    *action.Configuration
	actionInstall   *action.Install
	actionUninstall *action.Uninstall
	valuesOpts      *values.Options

	logger Printer
}

func NewPackageTool() *PackageTool {
	return &PackageTool{
		envSettings:  cli.New(),
		actionConfig: &action.Configuration{},
		chartName:    "oci://docker.io/envoyproxy/gateway-helm",
		valuesOpts:   &values.Options{},
	}
}

// Setup Configuration required to initialize helm action.
func (pt *PackageTool) Setup() error {

	// Since envoy-gateway uses docker's oci to store charts,
	// we need to create a registry client to make sure we can retrieve envoy-gateway chart
	registryCli, err := registry.NewClient()
	if err != nil {
		return err
	}
	pt.actionConfig = &action.Configuration{
		RegistryClient: registryCli,
	}

	kubectlFactory := options.DefaultConfigFlags
	ns, _, err := kubectlFactory.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	if err = pt.actionConfig.Init(
		kubectlFactory,
		ns,
		os.Getenv("HELM_DRIVER"),
		pt.logger.Printf,
	); err != nil {
		return err
	}

	// Build the relevant helm command action
	pt.actionInstall = action.NewInstall(pt.actionConfig)
	pt.actionUninstall = action.NewUninstall(pt.actionConfig)

	return nil
}

// SetInstallEnvSettings set the installation flags we are interested in
func (pt *PackageTool) SetInstallEnvSettings(installCmd *cobra.Command, opts *PackageOptions) {

	// add helm flags
	// we use a temporary flag to be set by helm env flags,
	// from which we can retrieve the flags we are interested
	var tmpFlags pflag.FlagSet
	pt.envSettings.AddFlags(&tmpFlags)
	tmpFlags.VisitAll(func(flag *pflag.Flag) {
		// TODO: Add more flags as needed?
		switch flag.Name {
		case "registry-config", "repository-config", "repository-cache":
			installCmd.Flags().AddFlag(flag)
		default:
		}
	})

	installCmd.Flags().DurationVar(&opts.Timeout, "timeout", helmOperateTimeout, "time to wait for any individual Kubernetes operation")
	installCmd.Flags().StringVar(&opts.Version, "version", egChartVersion, "specify a version constraint for the envoy gateway version to use")
	installCmd.Flags().StringVar(&opts.ReleaseName, "release-name", egReleaseName, "name of the envoy-gateway release to install")
	installCmd.Flags().StringVarP(&opts.ReleaseNamespace, "namespace", "n", egReleaseNamespace, "if set, specify the namespace where envoy gateway is installed")
	installCmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "console output only, make no changes")
	installCmd.Flags().BoolVar(&opts.SkipCRD, "skip-crds", false, "if set, no CRDs will be installed. By default, CRDs are installed if not already present")
	installCmd.Flags().BoolVar(&opts.OnlyCRD, "only-crds", false, "if set, only install the crds")
	installCmd.Flags().Bool("debug", false, "if set, the will output detailed execution logs")

	installCmd.Flags().StringSliceVarP(&pt.valuesOpts.ValueFiles, "values", "f", []string{}, "Specify values in a YAML file or a URL (can specify multiple)")
	installCmd.Flags().StringArrayVar(&pt.valuesOpts.Values, "set", []string{}, "Set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")

}

// SetUninstallEnvSetting set the uninstallation flags we are interested in
func (pt *PackageTool) SetUninstallEnvSetting(uninstallCmd *cobra.Command, opts *PackageOptions) {

	uninstallCmd.Flags().DurationVar(&opts.Timeout, "timeout", helmOperateTimeout, "time to wait for any individual Kubernetes operation")
	uninstallCmd.Flags().StringVar(&opts.ReleaseName, "release-name", egReleaseName, "name of the envoy-gateway release to uninstall")
	uninstallCmd.Flags().StringVarP(&opts.ReleaseNamespace, "namespace", "n", "", "if set, specify the namespace where envoy gateway is uninstalled")
	uninstallCmd.Flags().BoolVar(&opts.Wait, "wait", false, "if set, will wait until all Pods, PVCs, Services, and minimum number of Pods of a Deployment, StatefulSet, or ReplicaSet are in a ready state before marking the release as successful. It will wait for as long as --timeout")
	uninstallCmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "console output only, make no changes")
	uninstallCmd.Flags().BoolVar(&opts.WithCRD, "with-crds", false, "if set, the CRDs will also be removed")
	uninstallCmd.Flags().Bool("debug", false, "if set, the will output detailed execution logs")

}

// loadChart Load the chart instance according to the chart name and version
func (pt *PackageTool) loadChart(opts *PackageOptions) (*chart.Chart, error) {

	pt.actionInstall.Version = opts.Version
	chartName, err := pt.actionInstall.LocateChart(pt.chartName, pt.envSettings)
	if err != nil {
		return nil, err
	}

	egChart, err := loader.Load(chartName)
	if err != nil {
		return nil, err
	}

	// Before we return to the chart, we need to make sure that the chart is installed and used correctly
	if !installableChart(egChart) {
		return nil, fmt.Errorf("type of chart is not 'application' and cannot be installed")
	}
	if egChart.Metadata.Deprecated {
		return nil, fmt.Errorf("chart has been deprecated, please update chart version")
	}

	return egChart, nil
}

// extractCRDs Extract the CRDs part of the chart
func (pt *PackageTool) extractCRDs(ch *chart.Chart) ([]*resource.Info, error) {

	crdResInfo := make([]*resource.Info, 0, len(ch.CRDObjects()))

	for _, crd := range ch.CRDObjects() {
		resInfo, err := pt.actionConfig.KubeClient.Build(bytes.NewBufferString(string(crd.File.Data)), false)
		if err != nil {
			return nil, err
		}
		crdResInfo = append(crdResInfo, resInfo...)
	}

	return crdResInfo, nil
}

// RunInstall The default installation strategy we adopt is to first install Custom Resource Definitions (CRDs) separately,
// not as part of the Helm release. Subsequently, we install the Helm release without including CRDs.
// This approach ensures that when uninstalling with Helm or egctl later on, CRDs are not deleted.
// We intend for cluster administrators who understand the consequences of uninstalling CRDs to be
// responsible for their uninstallation.
// This is done to avoid garbage collection on CRs in the cluster during uninstallation,
// preventing the potential loss of crucial CR instances.
func (pt *PackageTool) RunInstall(opts *PackageOptions) error {

	if opts.Version == "v0.0.0-latest" {
		warningMarker := color.New(color.FgYellow).Add(color.Bold).Sprintf("WARNING")
		pt.logger.Println(fmt.Sprintf("%s: Currently using the latest version of envoy gateway chart, it is recommended to use the fixed version",
			warningMarker))
	}

	pt.setCommonValue()

	egChart, err := pt.loadChart(opts)
	if err != nil {
		return err
	}
	crdInfo, err := pt.extractCRDs(egChart)
	if err != nil {
		return err
	}

	// Before we install CRDs, we need to ensure that CRDs do not exist in the cluster
	// After we install CRDs, we need to ensure that the CRDs are successfully installed into the cluster
	if !opts.SkipCRD {

		if len(crdInfo) == 0 {
			return fmt.Errorf("CRDs not found in the envoy gateway chart")
		}

		if exist, err := detectExistCRDs(crdInfo); exist == nil || *exist {
			if err == nil {
				err = fmt.Errorf("found installed envoy gateway CRDs and gateway api CRDs, unable to continue installation")
			}
			return err
		}

		if err := installCRDs(crdInfo, pt.actionConfig); err != nil {
			return err
		}

		if exist, err := detectExistCRDs(crdInfo); exist == nil || !*exist {
			if err != nil {
				return fmt.Errorf("failed to install CRDs of envoy gateway")
			}
			return err
		}

		if opts.OnlyCRD {
			return nil
		}
	}

	// Merge all values flag
	providers := getter.All(pt.envSettings)
	egChartValues, err := pt.valuesOpts.MergeValues(providers)
	if err != nil {
		return err
	}

	pt.setInstallOptions(opts)
	release, err := pt.actionInstall.Run(egChart, egChartValues)
	if err != nil {
		return fmt.Errorf("failed to install envoy gateway resource: %w", err)
	}

	if opts.DryRun {
		pt.logger.Println(release.Manifest)
		return nil
	}

	successMarker := color.New(color.FgGreen).Add(color.Bold).Sprintf("SUCCESS")
	pt.logger.Println(fmt.Sprintf("%s: Envoy gateway installed", successMarker))
	return nil
}

// RunUninstall By default, we only uninstall instances of the Envoy Gateway resources.
func (pt *PackageTool) RunUninstall(opts *PackageOptions) error {

	pt.setUninstallOptions(opts)

	resp, err := pt.actionUninstall.Run(opts.ReleaseName)
	if err != nil {
		return fmt.Errorf("failed to uninstall envoy gateway: %w", err)
	}

	if opts.DryRun {
		pt.logger.Println(resp.Release.Manifest)
		return nil
	}

	if opts.WithCRD {

		if crdInfo, err := pt.extractCRDs(resp.Release.Chart); err != nil {
			return err
		} else if len(crdInfo) == 0 {
			return fmt.Errorf("CRDs not found in the envoy gateway chart")
		} else if _, errors := pt.actionConfig.KubeClient.Delete(crdInfo); len(errors) != 0 {
			return fmt.Errorf("failed to delete CRDs error: %s", util.MultipleErrors("", errors))
		}

	}

	successMarker := color.New(color.FgGreen).Add(color.Bold).Sprintf("SUCCESS")
	pt.logger.Println(fmt.Sprintf("%s: Envoy gateway uninstalled", successMarker))
	return nil
}

// setCommonValue Set the common values needed for the installation chart.
// We are not currently considering allowing users to define the following values
func (pt *PackageTool) setCommonValue() {
	pt.actionInstall.CreateNamespace = true
	pt.actionInstall.GenerateName = false
	pt.actionInstall.Description = "envoy gateway was installed using the egctl"
}

// setInstallOptions Sets the options required before install
func (pt *PackageTool) setInstallOptions(opts *PackageOptions) {

	if opts.DryRun {
		// When dry-run is set up, we do not need to connect to k8s-api server.
		// Since the kubernetes version needs to be higher than the value in the Helm library
		// for client component running, we set a fake K8s version.
		pt.actionInstall.ClientOnly = true
		pt.actionInstall.KubeVersion = createDummyK8sVersion()
	}

	// Since we already installed CRDs before installing the instance resources,
	// we skip the installation of CRDs.
	pt.actionInstall.SkipCRDs = true

	pt.actionInstall.DryRun = opts.DryRun
	pt.actionInstall.Timeout = opts.Timeout
	pt.actionInstall.ReleaseName = opts.ReleaseName
	pt.actionInstall.Namespace = opts.ReleaseNamespace

	// If '--atomic' is set, installed resources will be deleted if part of the installation fails.
	// However, after setting '--atomic', the default setting is '--wait' to wait for
	// resource installation in the foreground.
	// For the correctness of installation, we do not provide '--wait' flags and
	// always choose the foreground waiting strategy to install resources.
	pt.actionInstall.Atomic = true
}

// setUninstallOptions Sets the options required before uninstall
func (pt *PackageTool) setUninstallOptions(opts *PackageOptions) {
	pt.actionUninstall.DisableHooks = false
	pt.actionUninstall.KeepHistory = false

	pt.actionUninstall.DryRun = opts.DryRun

	if opts.Wait {
		pt.actionUninstall.Wait = opts.Wait
		pt.actionUninstall.DeletionPropagation = "foreground"
	} else {
		pt.actionUninstall.DeletionPropagation = "background"
	}
}

// SetPreRun This will perform the necessary preparations before initializing the PackageTool
func (pt *PackageTool) SetPreRun(cmd *cobra.Command) {
	existPreRunE := cmd.PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {

		pt.setPrinter(cmd)
		pt.setNamespace(cmd)

		return existPreRunE(cmd, args)
	}
}

// setNamespace We will set the namespace to EnvSettings before the PackageTool is set up
// This is to ensure that the namespace is correctly populated before we build the install/uninstall action cmd
func (pt *PackageTool) setNamespace(cmd *cobra.Command) {
	ns, err := cmd.Flags().GetString("namespace")
	if err != nil {
		ns = egReleaseNamespace
	}
	pt.envSettings.SetNamespace(ns)
}

// SetPrinter We will build the logger before the PackageTool is set up
func (pt *PackageTool) setPrinter(cmd *cobra.Command) {
	debug, err := cmd.Flags().GetBool("debug")
	if err != nil {
		debug = true
	}
	printer := NewPrinterForWriter(cmd.OutOrStdout(), debug)
	pt.logger = printer
}

// installableChart Make sure the chart can be installed
// ref: https://helm.sh/docs/topics/charts/#chart-types
func installableChart(ch *chart.Chart) bool {
	if len(ch.Metadata.Type) == 0 || ch.Metadata.Type == "application" {
		return true
	}
	return false
}

// createDummyK8sVersion Create a fake KubeVersion
func createDummyK8sVersion() *chartutil.KubeVersion {
	dummyVersion := "v9.9.9"
	sv, _ := semver.NewVersion(dummyVersion)

	return &chartutil.KubeVersion{
		Version: dummyVersion,
		Major:   strconv.FormatUint(sv.Major(), 10),
		Minor:   strconv.FormatUint(sv.Minor(), 10),
	}
}

// detectExistCRDs Check if envoy-gateway and gateway-api CRDs already exist in the cluster
func detectExistCRDs(crdResInfo []*resource.Info) (*bool, error) {

	exist := false
	existObj := make([]runtime.Object, 0, len(crdResInfo))

	for _, info := range crdResInfo {
		helper := resource.NewHelper(info.Client, info.Mapping)
		obj, err := helper.Get(info.Namespace, info.Name)
		if err != nil {
			if kerrors.IsNotFound(err) {
				continue
			}
			return nil, fmt.Errorf("failed to detect the crds resource: %w", err)
		}
		existObj = append(existObj, obj)
	}

	if len(existObj) == 0 {
		return &exist, nil
	}
	if len(existObj) == len(crdResInfo) {
		exist = true
		return &exist, nil
	}

	return nil, fmt.Errorf("expected CRDs does not match the number of CRDS that actually exist")
}

// installCRDs Install CRDs to the cluster
func installCRDs(crds []*resource.Info, actionConfig *action.Configuration) error {

	// Create all CRDs in the envoy gateway chart
	result, err := actionConfig.KubeClient.Create(crds)
	if err != nil {
		return fmt.Errorf("failed to create CRDs: %w", err)
	}

	// We invalidate the cache and let it rebuild the cache,
	// at which point no more updated CRDs exist
	client, err := actionConfig.RESTClientGetter.ToDiscoveryClient()
	if err != nil {
		return err
	}
	client.Invalidate()

	// Wait the specified amount of time for the resource to be recognized by the cluster
	if err := actionConfig.KubeClient.Wait(result.Created, 60*time.Second); err != nil {
		return err
	}
	_, err = client.ServerGroups()
	return err
}

type Printer interface {
	Printf(format string, a ...any)
	Println(string)
}

func NewPrinterForWriter(w io.Writer, debug bool) Printer {
	return &writerPrinter{writer: w, debug: debug}
}

type writerPrinter struct {
	writer io.Writer
	debug  bool
}

func (w *writerPrinter) Printf(format string, a ...any) {
	if w.debug {
		_, _ = fmt.Fprintln(w.writer, fmt.Sprintf(format, a...))
	}
}

func (w *writerPrinter) Println(str string) {
	_, _ = fmt.Fprintln(w.writer, str)
}
