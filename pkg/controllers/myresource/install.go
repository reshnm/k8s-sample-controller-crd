package myresource

import (
	"github.com/reshnm/k8s-sample-controller-crd/pkg/apis/samplecontroller/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

var (
	schemeBuilder = runtime.NewSchemeBuilder(
		v1alpha1.AddToScheme,
		setVersionPriority,
	)

	AddToScheme = schemeBuilder.AddToScheme
)

func setVersionPriority(scheme *runtime.Scheme) error {
	return scheme.SetVersionPriority(v1alpha1.SchemeGroupVersion)
}

func Install(scheme *runtime.Scheme) {
	utilruntime.Must(AddToScheme(scheme))
}
