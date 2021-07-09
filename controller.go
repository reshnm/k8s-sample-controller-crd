package main

import (
	clientset "github.com/reshnm/k8s-sample-controller-crd/pkg/generated/clientset/versioned"
	"github.com/reshnm/k8s-sample-controller-crd/pkg/generated/informers/externalversions/samplecontroller/v1alpha1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"time"
)

type Controller struct {
	clientSet *clientset.Clientset
	workqueue workqueue.RateLimitingInterface
	myresourceInformer v1alpha1.MyResourceInformer
	myresourceSynced cache.InformerSynced
}

func CreateController(clientset *clientset.Clientset, myresourceInformer v1alpha1.MyResourceInformer) *Controller {
	controller := &Controller{
		clientSet: clientset,
		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Myresource"),
		myresourceInformer: myresourceInformer,
		myresourceSynced: myresourceInformer.Informer().HasSynced,
	}

	myresourceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			klog.Info("Add Myresource:", key)
			if err == nil {
				controller.workqueue.Add(key)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			klog.Info("Update Myresource", key)
			if err == nil {
				controller.workqueue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			klog.Info("Delete Myresource", key)
			if err == nil {
				controller.workqueue.Add(key)
			}
		},
	})

	return controller
}

func (c *Controller) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	klog.Info("starting controller")

	if !cache.WaitForCacheSync(stopCh, c.myresourceSynced) {
		klog.Fatal("failed to sync cache")
	}

	klog.Info("cache synced")
	wait.Until(c.runWorker, time.Second, stopCh)
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}

	klog.Info("worker finished")
}

func (c *Controller) processNextWorkItem() bool {
	klog.Info("controller process next item")

	key, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	defer c.workqueue.Done(key)

	keyRaw := key.(string)
	_, _, err := c.myresourceInformer.Informer().GetIndexer().GetByKey(keyRaw)

	if err != nil {
		if c.workqueue.NumRequeues(key) < 5 {
			klog.Error("failed to process key", key, "with error, retrying ...", err)
			c.workqueue.AddRateLimited(key)
		} else {
			klog.Error("failed to process key", key, "with error, discarding ...", err)
			c.workqueue.Forget(key)
			utilruntime.HandleError(err)
		}
	} else {
		c.workqueue.Forget(key)
	}

	return true
}

