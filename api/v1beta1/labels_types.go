/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Important: Run "make" to regenerate code after modifying this file

// LabelsSpec defines the desired state of Labels
type LabelsSpec struct {
	// NodeNamePatterns defines a list of node name regex patterns for which the given labels should be set.
	// String start and end anchors (^/$) will be added automatically
	NodeNamePatterns []string `json:"nodeNamePatterns"`

	// Label defines the labels which should be set if one of the node name patterns matches
	// Format of label must be domain/name=value
	Labels map[string]string `json:"labels"`
}

// LabelsStatus defines the observed state of Labels
type LabelsStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Labels is the Schema for the labels API
type Labels struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LabelsSpec   `json:"spec,omitempty"`
	Status LabelsStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LabelsList contains a list of Labels
type LabelsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Labels `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Labels{}, &LabelsList{})
}
