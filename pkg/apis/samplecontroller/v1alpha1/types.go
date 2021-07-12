package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MyResource struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec MyResourceSpec `json:"spec"`
	Status MyResourceStatus `json:"status"`
}

type MyResourceSpec struct {
	Message string `json:"message"`
}

type MyResourceStatus struct {
	PodName string `json:"podName"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MyResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []MyResource `json:"items"`
}