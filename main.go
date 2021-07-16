package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	samplecontrollerClientset "github.com/reshnm/k8s-sample-controller-crd/pkg/generated/clientset/versioned"
	sampleComtrollerInformers "github.com/reshnm/k8s-sample-controller-crd/pkg/generated/informers/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

var (
	kubeconfig string
)

func createClients() (*kubernetes.Clientset, *samplecontrollerClientset.Clientset) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		klog.Fatal("failed to build config", err)
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatal("failed to create kubeClient")
	}

	sampleControllerClient, err := samplecontrollerClientset.NewForConfig(config)
	if err != nil {
		klog.Fatal("failed to create sampleControllerClient", err)
	}

	return kubeClient, sampleControllerClient
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	kubeClient, sampleControllerClient := createClients()

	myresourceInformerFactory := sampleComtrollerInformers.NewSharedInformerFactoryWithOptions(
		sampleControllerClient,
		time.Second*30,
		sampleComtrollerInformers.WithNamespace(metav1.NamespaceDefault))

	podInformerFactory := informers.NewSharedInformerFactoryWithOptions(
		kubeClient,
		time.Second*30,
		informers.WithNamespace(metav1.NamespaceDefault))

	controller := CreateController(
		kubeClient,
		sampleControllerClient,
		myresourceInformerFactory.Samplecontroller().V1alpha1().MyResources(),
		podInformerFactory.Core().V1().Pods())

	stopCh := make(chan struct{})
	defer close(stopCh)
	myresourceInformerFactory.Start(stopCh)
	podInformerFactory.Start(stopCh)
	go controller.Run(stopCh)

	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, syscall.SIGTERM)
	signal.Notify(sigTerm, syscall.SIGINT)
	<-sigTerm

	klog.Info("controller stopped")
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", os.Getenv("KUBECONFIG"), "Path to the kubeconfig")
}
