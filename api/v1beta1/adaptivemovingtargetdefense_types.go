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

// AdaptiveMovingTargetDefenseSpec defines the desired state of AdaptiveMovingTargetDefense
type AdaptiveMovingTargetDefenseSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Required
	// PodSelector is the selector of the Kubernetes Pods on which the user desires to enable moving target defense
	PodSelector map[string]string `json:"podSelector"`

	// +kubebuilder:validation:Required
	// Define strategy that maps actions to security events (based on the security event fields)
	Strategy []ResponseStrategy `json:"strategy"`
}

// MovingStrategy Substructure for strategy definitions
type ResponseStrategy struct {
	//TODO: use enum for the specific values of these fields
	//TODO: enforce that at least one strategy response definition is required

	// +kubebuilder:validation:Required
	Rule Rule `json:"rule"`

	// +kubebuilder:validation:Required
	// Action field value of the SecurityEvent that arrives
	Action string `json:"action"`
}

type Rule struct {
	// +kubebuilder:validation:Optional
	// Type field value of the SecurityEvent that arrives
	Type string `json:"type"`

	// +kubebuilder:validation:Optional
	// ThreatLevel field value of the SecurityEvent that arrives
	ThreatLevel string `json:"threatLevel"`

	// +kubebuilder:validation:Optional
	// Source field value of the SecurityEvent that arrives
	Source string `json:"source"`
}

// AdaptiveMovingTargetDefenseStatus defines the observed state of AdaptiveMovingTargetDefense
type AdaptiveMovingTargetDefenseStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AdaptiveMovingTargetDefense is the Schema for the adaptivemovingtargetdefenses API
type AdaptiveMovingTargetDefense struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AdaptiveMovingTargetDefenseSpec   `json:"spec,omitempty"`
	Status AdaptiveMovingTargetDefenseStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AdaptiveMovingTargetDefenseList contains a list of AdaptiveMovingTargetDefense
type AdaptiveMovingTargetDefenseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AdaptiveMovingTargetDefense `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AdaptiveMovingTargetDefense{}, &AdaptiveMovingTargetDefenseList{})
}
