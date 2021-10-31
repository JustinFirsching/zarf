package k8s

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/argoproj/gitops-engine/pkg/cache"
	"github.com/argoproj/gitops-engine/pkg/engine"
	"github.com/argoproj/gitops-engine/pkg/sync"
	"github.com/argoproj/gitops-engine/pkg/utils/kube"
	"github.com/defenseunicorns/zarf/cli/internal/utils"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	annotationGCMark = "gitops-agent.argoproj.io/gc-mark"
)

type resourceInfo struct {
	gcMark string
}

type settings struct {
	path string
}

func (syncSettings *settings) getGCMark(key kube.ResourceKey) string {
	h := sha256.New()
	_, _ = h.Write([]byte(syncSettings.path))
	_, _ = h.Write([]byte(strings.Join([]string{key.Group, key.Kind, key.Name}, "/")))
	return "sha256." + base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

func (syncSettings *settings) parseManifests() ([]*unstructured.Unstructured, error) {
	var res []*unstructured.Unstructured

	manifests := utils.RecursiveFileList(syncSettings.path)

	for _, manifest := range manifests {
		if ext := strings.ToLower(filepath.Ext(manifest)); ext == ".yml" || ext == ".yaml" {
			// Load the file contents
			data, err := ioutil.ReadFile(manifest)
			if err != nil {
				logrus.Fatal(err)
			}
			// Split the k8s resources
			items, err := kube.SplitYAML(data)
			if err != nil {
				logrus.Fatal(err)
			}
			res = append(res, items...)

		}
	}

	for i := range res {
		annotations := res[i].GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}
		annotations[annotationGCMark] = syncSettings.getGCMark(kube.GetResourceKey(res[i]))
		res[i].SetAnnotations(annotations)
	}
	return res, nil
}

func GitopsProcess(path string, revision string) {

	syncSettings := settings{path}

	namespace := ""
	prune := true
	restConfig := getRestConfig()

	clusterCache := cache.NewClusterCache(restConfig,
		cache.SetPopulateResourceInfoHandler(func(un *unstructured.Unstructured, isRoot bool) (info interface{}, cacheManifest bool) {
			// store gc mark of every resource
			gcMark := un.GetAnnotations()[annotationGCMark]
			info = &resourceInfo{gcMark: un.GetAnnotations()[annotationGCMark]}
			// cache resources that has that mark to improve performance
			cacheManifest = gcMark != ""
			return
		}),
	)
	gitOpsEngine := engine.NewEngine(restConfig, clusterCache)

	cleanup, _ := gitOpsEngine.Run()

	defer cleanup()

	for syncCount := 0; syncCount < 20; syncCount++ {

		target, err := syncSettings.parseManifests()
		if err != nil {
			logrus.Error(err, "Failed to parse target state")
			time.Sleep(3 * time.Second)
			continue
		}

		result, err := gitOpsEngine.Sync(context.Background(), target, func(r *cache.Resource) bool {
			return r.Info.(*resourceInfo).gcMark == syncSettings.getGCMark(r.ResourceKey())
		}, revision, namespace, sync.WithPrune(prune))
		if err != nil {
			logrus.Error(err, "Failed to synchronize cluster state")
			time.Sleep(3 * time.Second)
			continue
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintf(w, "RESOURCE\tRESULT\n")
		for _, res := range result {
			_, _ = fmt.Fprintf(w, "%s\t%s\n", res.ResourceKey.String(), res.Message)
		}
		_ = w.Flush()
		break
	}

}
