package myresource

import (
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	myresourceV1Alpha1 "github.com/reshnm/k8s-sample-controller-crd/pkg/apis/samplecontroller/v1alpha1"
)

func AddControllerToManager(mgr manager.Manager) error {
	controller, err := CreateController(mgr.GetClient())
	if err != nil {
		return err
	}

	return builder.ControllerManagedBy(mgr).For(&myresourceV1Alpha1.MyResource{}).Complete(controller)
}
