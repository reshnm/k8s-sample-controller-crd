package main

import (
	"flag"
	"github.com/reshnm/k8s-sample-controller-crd/pkg/controllers/myresource"
	"github.com/reshnm/k8s-sample-controller-crd/pkg/controllers/pod"
	"github.com/reshnm/k8s-sample-controller-crd/pkg/crdmanager"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

var (
	kubeconfig string
)

func createControllerManager() manager.Manager {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), manager.Options{})
	if err != nil {
		klog.Fatal("failed to create new controller manager", err)
	}

	return mgr
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	mgr := createControllerManager()

	crdManager, err := crdmanager.CreateCrdManager(mgr)
	if err != nil {
		klog.Fatal("failed to create CRD manager: ", err)
	}

	err = crdManager.EnsureCRDs()
	if err != nil {
		klog.Fatal("failed to ensure CRDs: ", err)
	}

	myresource.Install(mgr.GetScheme())
	err = myresource.AddControllerToManager(mgr)
	if err != nil {
		klog.Fatalf("error creating MyResource controller: %w", err)
	}
	pod.AddControllerToManager(mgr)

	klog.Info("starting the controller")

	err = mgr.Start(signals.SetupSignalHandler())
	if err != nil {
		klog.Fatalf("error starting the controller: %w", err)
	}

	klog.Info("controller stopped")
}
