/*
Copyright 2024 Abhijeet Rokade.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RollingUpdateSpec defines the desired state of RollingUpdate
type RollingUpdateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// MatchLabels specifies a set of {key, value} pairs used to select specific resources based on labels.
	// If specified, only resources matching all key-value pairs will be considered for rollout.
	// If MatchLabels is not specified, all resources in the namespace will be considered for rollout.
	// Each {key, value} pair in MatchLabels is equivalent to a label selector requirement using the "In" operator,
	// where the requirement's key field matches the key, the operator is "In", and the values array contains only the value.
	// The requirements are ANDed together.
	// +optional
	MatchLabels map[string]string `json:"matchLabels,omitempty"`

	// Interval specifies the time interval between rollouts.
	// It must be a valid duration string, such as "12h" or "30m".
	// +optional
	// +kubebuilder:validation:Pattern=`^[0-9]+(m|h|d|w)?$`
	// +kubebuilder:default="24h"
	Interval string `json:"interval,omitempty"`
}

// RollingUpdateStatus defines the observed state of RollingUpdate
type RollingUpdateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// LastRolloutTime indicates the timestamp of the last rolling restart or rollout operation.
	// If not set, it indicates that no rolling restart or rollout has been performed yet.
	// +optional
	LastRolloutTime metav1.Time `json:"lastRolloutTime,omitempty"`

	// Deployments stores the list of deployments that were restarted by this RollingUpdate CR.
	// This allows for back tracing to identify which deployments were affected by a particular
	// rolling restart or rollout operation initiated by this RollingUpdate custom resource.
	// Each entry in the list represents the name of a deployment in the format "name".
	// The namespace of the deployment can be inferred as the namespace where this RollingUpdate CR
	// is installed, which can be retrieved from the metadata section of this custom resource.
	// +optional
	Deployments []string `json:"deployments,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// RollingUpdate is the Schema for the rollingupdates API
type RollingUpdate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RollingUpdateSpec   `json:"spec,omitempty"`
	Status RollingUpdateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RollingUpdateList contains a list of RollingUpdate
type RollingUpdateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RollingUpdate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RollingUpdate{}, &RollingUpdateList{})
}
