package k8s

import (
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/kustomize-controller/controllers"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GitopsProcess(path string) {

	kustomization := kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "replaceme",
			Namespace: "replaceme2",
		},
		Spec: kustomizev1.KustomizationSpec{
			Path: "./",
			KubeConfig: &kustomizev1.KubeConfig{
				SecretRef: meta.LocalObjectReference{
					Name: "kubeconfig",
				},
			},
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Name:      "noidea",
				Namespace: "reallynoidea",
				Kind:      sourcev1.GitRepositoryKind,
			},
			TargetNamespace: "replaceme",
			Force:           true,
		},
	}

	generator := controllers.NewGenerator(kustomization)
	generator.WriteFile(path)

}
