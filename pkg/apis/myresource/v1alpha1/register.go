package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/reshnm/k8s-sample-controller-crd/pkg/apis/myresource"
)

var SchemeGroupVersion = schema.GroupVersion{
	Group: myresource.GroupName,
	Version: "v1alpha1",
}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme = SchemeBuilder.AddToScheme
)

func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

func addKnownTypes(schema *runtime.Scheme) error {
	schema.AddKnownTypes(
		SchemeGroupVersion,
		&MyResource{},
		&MyResourceList{},
	)

	metav1.AddToGroupVersion(schema, SchemeGroupVersion)
	return nil
}