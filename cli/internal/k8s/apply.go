package k8s

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/releaseutil"
)

const applyTimeout = time.Minute * 2
const waitInterval = 5 * time.Second

func ApplyManifest(manifest releaseutil.Manifest) {
	logContext := logrus.WithFields(logrus.Fields{
		"path": manifest.Name,
		"kind": manifest.Head.Kind,
		"name": manifest.Head.Metadata.Name,
	})

	logContext.Info("Applying K8s resource")

	_, kubeClient := connect()

	manifestContent := strings.NewReader(manifest.Content)

	for syncCount := 0; syncCount < 20; syncCount++ {

		logrus.Infof("Manifest apply attempt %d of 20", syncCount+1)

		// Be slightly elastic in how the delays are handled
		if syncCount > 2 {
			logContext.Info("Sleeping 5 seconds before processing manifest again")
			time.Sleep(waitInterval)
		} else {
			time.Sleep(1 * time.Second)
		}

		resources, err := kubeClient.Build(manifestContent, true)
		if err != nil {
			// Expect this is related to a lack of a CRD being ready or some cluster connection issue
			logContext.Debug(err)
			logContext.Warn("Unable to process the manifest")
			continue
		}

		// Attempt to create/update the manifest
		result, err := kubeClient.Update(resources, resources, true)
		logContext.Debug(result)

		if err != nil {
			logContext.Debug(err)
			logContext.Warn("Unable to apply the manifest file")
			continue
		}

		// Only wait if there is something to wait for
		if len(result.Created) > 0 || len(result.Updated) > 0 {
			logContext.Info("Waiting for resources to be created")
			if waitErr := kubeClient.WatchUntilReady(resources, applyTimeout); waitErr != nil {
				logContext.Debug(waitErr)
				logContext.Warn("Problem waiting for manifest to apply")
				continue
			}
		}

		break
	}
}

func debug(format string, v ...interface{}) {
	format = fmt.Sprintf("%s\n", format)
	log.Output(1, fmt.Sprintf(format, v...))
}
