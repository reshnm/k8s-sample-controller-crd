package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/klog/v2"
	"path"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"time"

	apiextinstall "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/install"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	embedFSCrdRootDir = "crd"
)

//go:embed crd/*.yaml
var importedCrdFS embed.FS

type CRDManager struct {
	client       client.Client
	crdRawDataFS *embed.FS
}

func NewCrdManager(mgr manager.Manager) (*CRDManager, error) {
	apiExtensionScheme := runtime.NewScheme()
	apiextinstall.Install(apiExtensionScheme)
	kubeClient, err := client.New(mgr.GetConfig(), client.Options{Scheme: apiExtensionScheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create client for registering CRDs: %w", err)
	}

	_, err = importedCrdFS.ReadDir(embedFSCrdRootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded CRDs: %w", err)
	}

	return &CRDManager{
		client:       kubeClient,
		crdRawDataFS: &importedCrdFS,
	}, nil
}

func (m *CRDManager) EnsureCRDs() error {
	crdList, err := m.crdsFromDir()
	if err != nil {
		return err
	}

	klog.Info("registering CRDs")
	for _, crd := range crdList {
		existingCrd := &v1.CustomResourceDefinition{}
		err := m.client.Get(context.TODO(), client.ObjectKey{Name: crd.Name}, existingCrd)
		if err != nil {
			if apierrors.IsNotFound(err) {
				err = m.client.Create(context.TODO(), &crd)
				if err != nil {
					return err
				}
				klog.Infof("registered new CRD: %q", crd.Name)
				continue
			}
			return err
		}

		crd.ResourceVersion = existingCrd.ResourceVersion
		crd.UID = existingCrd.UID
		err = m.client.Patch(context.TODO(), &crd, client.MergeFrom(existingCrd))
		if err != nil {
			return err
		}
		klog.Infof("updated CRD: %q", crd.Name)
	}

	err = wait.Poll(1*time.Second, 30*time.Second, func() (done bool, err error) {
		aggregatedStatus := true

		for _, crd := range crdList {
			if !aggregatedStatus {
				return aggregatedStatus, nil
			}
			crdResult := &v1.CustomResourceDefinition{}
			err := m.client.Get(context.TODO(), client.ObjectKey{Name: crd.Name}, crdResult)
			if err != nil {
				return false, err
			}

			for _, crdCondition := range crdResult.Status.Conditions {
				switch crdCondition.Type {
				case v1.Established:
					if crdCondition.Status != v1.ConditionTrue {
						aggregatedStatus = false
					}
				case v1.NamesAccepted:
					if crdCondition.Status == v1.ConditionFalse {
						aggregatedStatus = false
					}
				}
			}
		}
		return aggregatedStatus, nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (m *CRDManager) crdsFromDir() ([]v1.CustomResourceDefinition, error) {
	crdList := make([]v1.CustomResourceDefinition, 0)
	files, err := m.crdRawDataFS.ReadDir(embedFSCrdRootDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		data, err := m.crdRawDataFS.ReadFile(path.Join(embedFSCrdRootDir, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read CRD file %q: %w", file.Name(), err)
		}

		decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 32)
		crd := &v1.CustomResourceDefinition{}
		err = decoder.Decode(crd)
		if err != nil {
			return nil, fmt.Errorf("failed to decode CRD from file %d: %w", file.Name(), err)
		}

		crdList = append(crdList, *crd)
	}

	return crdList, nil
}
