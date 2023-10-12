/*
 * Copyright (C) 2023 R6 Security, Inc.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the Server Side Public License, version 1,
 * as published by MongoDB, Inc.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * Server Side Public License for more details.
 *
 * You should have received a copy of the Server Side Public License
 * along with this program. If not, see
 * <http://www.mongodb.com/licensing/server-side-public-license>.
 */

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SecurityEventSpec defines the desired state of SecurityEvent
type SecurityEventSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Required
	// Targets contains the list of affected pods, each item in the form of "namespace/name" or "/name"
	Targets []string `json:"targets"`

	// +kubebuilder:validation:Required
	Rule Rule `json:"rule"`

	// +kubebuilder:validation:Required
	// Description of the security threat
	Description string `json:"description"`
}

// SecurityEventStatus defines the observed state of SecurityEvent
type SecurityEventStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// SecurityEvent is the Schema for the securityevents API
type SecurityEvent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecurityEventSpec   `json:"spec,omitempty"`
	Status SecurityEventStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SecurityEventList contains a list of SecurityEvent
type SecurityEventList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecurityEvent `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SecurityEvent{}, &SecurityEventList{})
}
