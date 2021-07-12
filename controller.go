package main

import (
	"context"
	"fmt"
	"time"

	"github.com/reshnm/k8s-sample-controller-crd/pkg/apis/samplecontroller/v1alpha1"
	samplecontrollerClientset "github.com/reshnm/k8s-sample-controller-crd/pkg/generated/clientset/versioned"
	samplecontroller "github.com/reshnm/k8s-sample-controller-crd/pkg/generated/informers/externalversions/samplecontroller/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	informercorev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

type Controller struct {
	kubeClient *kubernetes.Clientset
	samplecontrollerClient *samplecontrollerClientset.Clientset
	workqueue workqueue.RateLimitingInterface
	myresourceInformer samplecontroller.MyResourceInformer
	myresourceSynced cache.InformerSynced
	podInformer informercorev1.PodInformer
	podsSynced cache.InformerSynced
}

func CreateController(
	kubeClient *kubernetes.Clientset,
	samplecontrollerClient *samplecontrollerClientset.Clientset,
	myresourceInformer samplecontroller.MyResourceInformer,
	podInformer informercorev1.PodInformer) *Controller {

	controller := &Controller{
		kubeClient: kubeClient,
		samplecontrollerClient: samplecontrollerClient,
		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Myresource"),
		myresourceInformer: myresourceInformer,
		myresourceSynced: myresourceInformer.Informer().HasSynced,
		podInformer: podInformer,
		podsSynced: podInformer.Informer().HasSynced,
	}

	myresourceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueMyResource,
		UpdateFunc: func(oldObj, newObj interface{}) {
			controller.enqueueMyResource(newObj)
		},
	})

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handlePod,
		UpdateFunc: func(oldObj, newObj interface{}) {
			newPod := newObj.(*corev1.Pod)
			oldPod := oldObj.(*corev1.Pod)
			if newPod.ResourceVersion != oldPod.ResourceVersion {
				controller.handlePod(newObj)
			}
		},
		DeleteFunc: controller.handlePod,
	})

	return controller
}

func (c *Controller) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	klog.Info("starting controller")

	if !cache.WaitForCacheSync(stopCh, c.myresourceSynced, c.podsSynced) {
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

	klog.Infof("handling MyResource '%s', message='%s'",
		key,
		myresource.Spec.Message)

	podName := fmt.Sprintf("%s-pod", myresource.Name)
	pod, err := c.kubeClient.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})

	if errors.IsNotFound(err) {
		klog.Infof("creating pod '%s'", podName)
		pod, err = c.kubeClient.CoreV1().Pods(namespace).Create(
			context.TODO(),
			newPod(myresource, podName),
			metav1.CreateOptions{})

		if err != nil {
			return err
		}

		if !metav1.IsControlledBy(pod, myresource) {
			return fmt.Errorf("pod '%s' is not controlled by samplecontroller", podName)
		}

		err = c.updateMyResource(myresource, pod)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) updateMyResource(myresource *v1alpha1.MyResource, pod *corev1.Pod) error {
	myresourceCopy := myresource.DeepCopy()
	myresourceCopy.Status.PodName = pod.Name
	_, err := c.samplecontrollerClient.SamplecontrollerV1alpha1().MyResources(myresource.Namespace).Update(
		context.TODO(),
		myresourceCopy,
		metav1.UpdateOptions{})

	if err != nil {
		klog.Infof("failed to update status of MyResource '%s'", myresource.Name)
		return err
	}

	klog.Infof("updated status of MyResource '%s' with podName '%s'", myresource.Name, pod.Name)
	return nil
}

func (c *Controller) enqueueMyResource(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	klog.Info("enqueue Myresource: ", key)
	c.workqueue.Add(key)
}

func (c *Controller) handlePod(obj interface{}) {
	pod, ok := obj.(metav1.Object)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding pod, invalid type"))
			return
		}
		pod, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding pod tombstone, invalid type"))
			return
		}
		klog.Infof("recovered deleted pod '%s' from tombstone", pod.GetName())
	}

	klog.Infof("handling pod '%s'", pod.GetName())
	ownerRef := metav1.GetControllerOf(pod)
	if ownerRef != nil {
		if ownerRef.Kind != "MyResource" {
			return
		}

		myresource, err := c.myresourceInformer.Lister().MyResources(pod.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			klog.Infof("ignoring orphaned pod '%s' of MyResource '%s'", pod.GetSelfLink(), ownerRef.Name)
			return
		}

		c.enqueueMyResource(myresource)
	}
}

func newPod(myresource *v1alpha1.MyResource, podName string) *corev1.Pod {
	labels := map[string]string{
		"app": "echoserver",
		"controller": myresource.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
			Namespace: myresource.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(myresource, v1alpha1.SchemeGroupVersion.WithKind("MyResource")),
			},
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "echoserver",
					Image: "reshnm/echoserver:latest",
					Env: []corev1.EnvVar{
						corev1.EnvVar{
							Name:      "ECHO_MESSAGE",
							Value:     myresource.Spec.Message,
						},
					},
					Ports: []corev1.ContainerPort{
						corev1.ContainerPort{
							ContainerPort: 80,
						},
					},
				},
			},
		},
	}
}
