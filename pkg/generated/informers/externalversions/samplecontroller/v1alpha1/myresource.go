// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	samplecontrollerv1alpha1 "github.com/reshnm/k8s-sample-controller-crd/pkg/apis/samplecontroller/v1alpha1"
	versioned "github.com/reshnm/k8s-sample-controller-crd/pkg/generated/clientset/versioned"
	internalinterfaces "github.com/reshnm/k8s-sample-controller-crd/pkg/generated/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/reshnm/k8s-sample-controller-crd/pkg/generated/listers/samplecontroller/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// MyResourceInformer provides access to a shared informer and lister for
// MyResources.
type MyResourceInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.MyResourceLister
}

type myResourceInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewMyResourceInformer constructs a new informer for MyResource type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewMyResourceInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredMyResourceInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredMyResourceInformer constructs a new informer for MyResource type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredMyResourceInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.SamplecontrollerV1alpha1().MyResources(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.SamplecontrollerV1alpha1().MyResources(namespace).Watch(context.TODO(), options)
			},
		},
		&samplecontrollerv1alpha1.MyResource{},
		resyncPeriod,
		indexers,
	)
}

func (f *myResourceInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredMyResourceInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *myResourceInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&samplecontrollerv1alpha1.MyResource{}, f.defaultInformer)
}

func (f *myResourceInformer) Lister() v1alpha1.MyResourceLister {
	return v1alpha1.NewMyResourceLister(f.Informer().GetIndexer())
}