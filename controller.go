package main

import (
	"fmt"
	clientset "github.com/reshnm/k8s-sample-controller-crd/pkg/generated/clientset/versioned"
	"github.com/reshnm/k8s-sample-controller-crd/pkg/generated/informers/externalversions/samplecontroller/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
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
			var key string
			var err error
			if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
				utilruntime.HandleError(err)
				return
			}
			klog.Info("Add Myresource: ", key)
			controller.workqueue.Add(key)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			var key string
			var err error
			if key, err = cache.MetaNamespaceKeyFunc(newObj); err != nil {
				utilruntime.HandleError(err)
				return
			}
			klog.Info("Update Myresource: ", key)
			controller.workqueue.Add(key)
		},
		DeleteFunc: func(obj interface{}) {
			var key string
			var err error
			if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
				utilruntime.HandleError(err)
				return
			}
			klog.Info("Delete Myresource: ", key)
			controller.workqueue.Add(key)
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

	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	defer c.workqueue.Done(obj)

	key, ok := obj.(string)
	if !ok {
		c.workqueue.Forget(obj)
		utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
		return false
	}

	err := c.syncHandler(key)
	if err != nil {
		c.workqueue.AddRateLimited(obj)
	} else {
		c.workqueue.Forget(obj)
	}

	return true
}

func (c *Controller) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid key: %s", key))
		return nil
	}

	myresource, err := c.myresourceInformer.Lister().MyResources(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("MyResource '%s' in workqueue doesn't exist", key))
			return nil
		}
		return err
	}

	klog.Infof("handling MyResource '%s', message='%s', someValue='%d'",
		key,
		myresource.Spec.Message,
		*myresource.Spec.SomeValue)

	return nil
}

