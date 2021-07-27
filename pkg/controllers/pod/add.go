package pod

import (
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func AddControllerToManager(mgr manager.Manager) error {
	controller, err := CreateController(mgr.GetClient())
	if err != nil {
		return err
	}

	return builder.ControllerManagedBy(mgr).For(&v1.Pod{}).Complete(controller)
}
