package helm

import (
	"os"

	"github.com/defenseunicorns/zarf/cli/config"
	"github.com/defenseunicorns/zarf/cli/internal/git"
	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"

	"strings"

	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

type ChartOptions struct {
	BasePath      string
	Chart         config.ZarfChart
}

func InstallChart(options ChartOptions) {

	logContext := logrus.WithFields(logrus.Fields{
		"Chart":   options.Chart.Name,
		"URL":     options.Chart.Url,
		"Version": options.Chart.Version,
	})
	logrus.SetLevel(logrus.TraceLevel)
	logContext.Info("Processing helm chart")

	// Get the path the temporary helmchart tarball
	sourceTarball := StandardName(options.BasePath+"/charts", options.Chart) + ".tgz"

	// Initialize helm SDK
	actionConfig := new(action.Configuration)
	settings := cli.New()

	// Setup K8s connection
	if err := actionConfig.Init(settings.RESTClientGetter(), options.Chart.Namespace, "", logrus.Debugf); err != nil {
		logContext.Debug(err)
		logContext.Fatal("Unable to initialize the K8s client")
	}

	// Bind the helm action
	client := action.NewInstall(actionConfig)

	// Must be unique per-namespace and < 53 characters. @todo: restrict helm chart name to this
	client.ReleaseName = options.Chart.Name

	// Namespace must be specified
	client.Namespace = options.Chart.Namespace

	// Need to create the namespace if it doesn't exist
	client.CreateNamespace = true

	// Load the chart tarball
	chart, err := loader.Load(sourceTarball)
	if err != nil {
		logContext.Debug(err)
		logContext.Fatal("Unable to load the helm chart")
	}

	// @todo: merge chart values from ZarfChart
	// Perform the chart installation
	releaser, err := client.Run(chart, chart.Values)
	logContext.Debug(releaser)
	if err != nil {
		logContext.Debug(err)
		logContext.Fatal("Unable to install the helm chart")
	}

}

func DownloadChartFromGit(chart config.ZarfChart, destination string) {
	logContext := logrus.WithFields(logrus.Fields{
		"Chart":   chart.Name,
		"URL":     chart.Url,
		"Version": chart.Version,
	})

	logContext.Info("Processing git-based helm chart")
	client := action.NewPackage()

	// Get the git repo
	tempPath := git.DownloadRepoToTemp(chart.Url)

	// Switch to the correct tag
	git.CheckoutTag(tempPath, chart.Version)

	// Tell helm where to save the archive and create the package
	client.Destination = destination
	client.Run(tempPath+"/chart", nil)

	_ = os.RemoveAll(tempPath)
}

func DownloadPublishedChart(chart config.ZarfChart, destination string) {
	logContext := logrus.WithFields(logrus.Fields{
		"Chart":   chart.Name,
		"URL":     chart.Url,
		"Version": chart.Version,
	})

	logContext.Info("Processing published helm chart")

	var out strings.Builder

	// Setup the helm pull config
	pull := action.NewPull()
	pull.Settings = cli.New()

	// Setup the chart downloader
	downloader := downloader.ChartDownloader{
		Out:     &out,
		Verify:  downloader.VerifyNever,
		Getters: getter.All(pull.Settings),
	}

	// @todo: process OCI-based charts

	// Perform simple chart download
	chartURL, err := repo.FindChartInRepoURL(chart.Url, chart.Name, chart.Version, pull.CertFile, pull.KeyFile, pull.CaFile, getter.All(pull.Settings))
	if err != nil {
		logContext.Debug(err)
		logContext.Fatal("Unable to pull the helm chart")
	}

	// Download the file (we don't control what name helm creates here)
	saved, _, err := downloader.DownloadTo(chartURL, pull.Version, destination)
	if err != nil {
		logContext.Debug(err)
		logContext.Fatal("Unable to download the helm chart")
	}

	// Ensure the name is consistent for deployments
	destinationTarball := StandardName(destination, chart) + ".tgz"
	os.Rename(saved, destinationTarball)
}

// StandardName generates a predictable full path for a helm chart for Zarf
func StandardName(destination string, chart config.ZarfChart) string {
	return destination + "/" + chart.Name + "-" + chart.Version
}
