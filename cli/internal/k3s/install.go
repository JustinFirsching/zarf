package k3s

import (
	"github.com/defenseunicorns/zarf/cli/config"
	"github.com/defenseunicorns/zarf/cli/internal/git"
	"github.com/defenseunicorns/zarf/cli/internal/packager"
	"github.com/defenseunicorns/zarf/cli/internal/pki"
	"github.com/defenseunicorns/zarf/cli/internal/utils"
	"github.com/sirupsen/logrus"
)

type InstallOptions struct {
	PKI        pki.PKIConfig
	Confirmed  bool
	Components string
}

func Install(options InstallOptions) {
	utils.RunPreflightChecks()

	logrus.Info("Installing K3s")

	packager.Deploy(config.PackageInitName, options.Confirmed, options.Components)

	// Install RHEL RPMs if applicable
	if utils.IsRHEL() {
		configureRHEL()
	}

	pki.HandlePKI(options.PKI)

	gitSecret := git.GetOrCreateZarfSecret()

	logrus.Info("Installation complete.  You can run \"/usr/local/bin/k9s\" to monitor the status of the deployment.")
	logrus.WithFields(logrus.Fields{
		"Gitea Username (if installed)": config.ZarfGitUser,
		"Grafana Username":              "zarf-admin",
		"Password (all)":                gitSecret,
	}).Warn("Credentials stored in ~/.git-credentials")
}
