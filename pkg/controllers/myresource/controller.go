package myresource

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	v1alpha1 "github.com/reshnm/k8s-sample-controller-crd/pkg/apis/samplecontroller/v1alpha1"
)

type Controller struct {
	client client.Client
}

func CreateController(client client.Client) (reconcile.Reconciler, error) {
	controller := Controller{
		client: client,
	}
	return &controller, nil
}

func (c *Controller) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	myresource := &v1alpha1.MyResource{}
	err := c.client.Get(ctx, req.NamespacedName, myresource)
	if err != nil {
		return reconcile.Result{}, err
	}

	klog.V(4).Infof("handling MyResource %q, message=%q",
		req.NamespacedName,
		myresource.Spec.Message)

	podName := fmt.Sprintf("%s-pod", myresource.Name)
	pod := &corev1.Pod{}
	err = c.client.Get(ctx, types.NamespacedName{Name: podName, Namespace: req.Namespace}, pod)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.Infof("creating pod '%q'", podName)
			err := c.client.Create(ctx, newPod(myresource, podName))
			if err != nil {
				return reconcile.Result{}, err
			}

			if !metav1.IsControlledBy(pod, myresource) {
				return reconcile.Result{}, fmt.Errorf("pod %q is not controlled by samplecontroller", podName)
			}

			err = c.updateMyResource(ctx, myresource, podName)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
	}

	return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
}

func (c *Controller) updateMyResource(ctx context.Context, myresource *v1alpha1.MyResource, podName string) error {
	myresource.Status.PodName = podName
	err := c.client.Status().Update(ctx, myresource)

	if err != nil {
		klog.Errorf("failed to update status of MyResource '%q'", myresource.Name)
		return err
	}

	klog.Infof("updated status of MyResource '%s' with podName '%q'", myresource.Name, podName)
	return nil
}

func newPod(myresource *v1alpha1.MyResource, podName string) *corev1.Pod {
	labels := map[string]string{
		"app":        "echoserver",
		"controller": myresource.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: myresource.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(myresource, v1alpha1.SchemeGroupVersion.WithKind("MyResource")),
			},
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "echoserver",
					Image: "reshnm/echoserver:latest",
					Env: []corev1.EnvVar{
						{
							Name:  "ECHO_MESSAGE",
							Value: myresource.Spec.Message,
						},
					},
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 80,
						},
					},
				},
			},
		},
	}
}
