package main

import (
	"flag"
	"os"
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
	informers.NewSharedInformerFactory(client, time.Second * 30)
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", os.Getenv("KUBECONFIG"), "Path to the kubeconfig")
}