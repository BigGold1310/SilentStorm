/*
Copyright 2024.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AlertmanagerSpec defines the desired state of Alertmanager
type AlertmanagerSpec struct {
	// Address configures the HTTP API endpoint of the Alertmanager.
	Address string `json:"address"`
	// SilenceSelector selects which to which Alertmanager the silence should apply.
	SilenceSelector metav1.LabelSelector `json:"silenceSelector"`
	// Authentication configures how to authenticate with the Alertmanager. If omited, no authentication is performed.
	Authentication Authentication `json:"authentication,omitempty"`
}

// Authentication defines the possible authentication options.
type Authentication struct {
	// ServiceAccountRef enables bearer token authentication with the ServiceAccount token.
	// The ServiceAccountRef is expected to be in the same namespace as this Alertmanager.
	ServiceAccountRef string `json:"serviceAccountRef,omitempty"`
}

// AlertmanagerStatus defines the observed state of Alertmanager
type AlertmanagerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Alertmanager is the Schema for the alertmanagers API
type Alertmanager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AlertmanagerSpec   `json:"spec,omitempty"`
	Status AlertmanagerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AlertmanagerList contains a list of Alertmanager
type AlertmanagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Alertmanager `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Alertmanager{}, &AlertmanagerList{})
}
