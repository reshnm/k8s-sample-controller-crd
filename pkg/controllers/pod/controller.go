package pod

import (
	"context"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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
	pod := &v1.Pod{}
	err := c.client.Get(ctx, req.NamespacedName, pod)
	if err != nil {
		return reconcile.Result{}, err
	}

	ownerRef := metav1.GetControllerOf(pod)
	if ownerRef != nil {
		if ownerRef.Kind != "MyResource" {
			return reconcile.Result{}, nil
		}

		klog.V(4).Infof("handling Pod %q, phase %q", req.NamespacedName, pod.Status.Phase)

		/*
			myresource := &v1alpha1.MyResource{}
			err := c.client.Get(ctx, types.NamespacedName{Name: ownerRef.Name, Namespace: req.Namespace}, myresource)
			if err == nil {

			}
		*/
	}

	return reconcile.Result{}, nil
}
