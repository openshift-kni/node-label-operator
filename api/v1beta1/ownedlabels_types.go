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

// OwnedLabelsSpec defines the desired state of OwnedLabels
type OwnedLabelsSpec struct {
	// Domain defines the label domain which is owned by this operator
	// If a node label
	// - matches this domain AND
	// - matches the namePattern if given AND
	// - no label rule matches
	// then the label will be removed
	Domain *string `json:"domain,omitempty"`

	// NamePattern defines the label name pattern which is owned by this operator
	// If a node label
	// - matches this name pattern AND
	// - matches the domain if given AND
	// - no label rule matches
	// then the label will be removed
	// String start and end anchors (^/$) will be added automatically
	NamePattern *string `json:"namePattern,omitempty"`
}

// OwnedLabelsStatus defines the observed state of OwnedLabels
type OwnedLabelsStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// OwnedLabels is the Schema for the ownedlabels API
// They define which node labels are owned by this operator and can be removed in case no label rule matches
type OwnedLabels struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OwnedLabelsSpec   `json:"spec,omitempty"`
	Status OwnedLabelsStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OwnedLabelsList contains a list of OwnedLabels
type OwnedLabelsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OwnedLabels `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OwnedLabels{}, &OwnedLabelsList{})
}
