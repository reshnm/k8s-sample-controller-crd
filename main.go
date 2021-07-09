package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	clientset "github.com/reshnm/k8s-sample-controller-crd/pkg/generated/clientset/versioned"
	informers "github.com/reshnm/k8s-sample-controller-crd/pkg/generated/informers/externalversions"
)

var (
	kubeconfig string
)

func createClient() *clientset.Clientset {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		klog.Fatal("failed to build config", err)
	}

	client, err := clientset.NewForConfig(config)
	if err != nil {
		klog.Fatal("failed to create client", err)
	}

	klog.Info("client created")
	return client
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	client := createClient()
	myresourceInformerFactory := informers.NewSharedInformerFactory(client, time.Second * 30)

	controller := CreateController(client, myresourceInformerFactory.Samplecontroller().V1alpha1().MyResources())

	stopCh := make(chan struct{})
	defer close(stopCh)
	myresourceInformerFactory.Start(stopCh)
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