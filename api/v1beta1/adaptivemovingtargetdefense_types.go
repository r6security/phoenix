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
	corev1 "k8s.io/api/core/v1"
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
	// +kubebuilder:validation:MinItems=1
	// Define strategy that maps actions to security events (based on the security event fields)
	Strategy []ResponseStrategy `json:"strategy"`
}

type DisableAction struct{}
type DeleteAction struct{}
type QuarantineAction struct{}
type Debugger struct {

	// +kubebuilder:validation:Optional
	Name string `json:"name,omitempty"`

	// +kubebuilder:validation:Required
	Image string `json:"image,omitempty"`

	// +kubebuilder:validation:Optional
	Terminal bool `json:"terminal,omitempty"`
}

// CustomContainerPort represents a network port in a single EphemeralContainer.
type CustomContainerPort struct {
	// Number of port to expose on the pod's IP address.
	// This must be a valid port number, 0 < x < 65536.
	ContainerPort int32 `json:"containerPort"`

	// What host IP to bind the external port to.
	HostIP string `json:"hostIP,omitempty"`

	// Number of port to expose on the host.
	// If specified, this must be a valid port number, 0 < x < 65536.
	HostPort *int32 `json:"hostPort,omitempty"`

	// If specified, this must be an IANA_SVC_NAME and unique within the pod.
	// Name for the port that can be referred to by services.
	Name string `json:"name,omitempty"`

	// Protocol for port. Must be UDP, TCP, or SCTP. Defaults to "TCP".
	// +kubebuilder:validation:Enum="TCP";"UDP";"SCTP"
	// +kubebuilder:default:="TCP"
	Protocol string `json:"protocol,omitempty"`
}

// CustomAction defines the custom action with an overridden Ports field.
// controller-gen v0.19 generates for EphemeralContainers a protocol with an allOf field with two defaults which throws an error under k8@1.30.4
type CustomAction struct {
	// EphemeralContainer Fields except Ports for correct json representation.
	// Ports are forrbiden in EphemeralContainers. In this scenario Ports are a CustomContainerPort type.

	// Name of the ephemeral container specified as a DNS_LABEL.
	// This name must be unique among all containers, init containers and ephemeral containers.
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// Container image name.
	// More info: https://kubernetes.io/docs/concepts/containers/images
	Image string `json:"image,omitempty" protobuf:"bytes,2,opt,name=image"`
	// Entrypoint array. Not executed within a shell.
	// The image's ENTRYPOINT is used if this is not provided.
	// Variable references $(VAR_NAME) are expanded using the container's environment. If a variable
	// cannot be resolved, the reference in the input string will be unchanged. Double $$ are reduced
	// to a single $, which allows for escaping the $(VAR_NAME) syntax: i.e. "$$(VAR_NAME)" will
	// produce the string literal "$(VAR_NAME)". Escaped references will never be expanded, regardless
	// of whether the variable exists or not. Cannot be updated.
	// More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell
	// +optional
	// +listType=atomic
	Command []string `json:"command,omitempty" protobuf:"bytes,3,rep,name=command"`
	// Arguments to the entrypoint.
	// The image's CMD is used if this is not provided.
	// Variable references $(VAR_NAME) are expanded using the container's environment. If a variable
	// cannot be resolved, the reference in the input string will be unchanged. Double $$ are reduced
	// to a single $, which allows for escaping the $(VAR_NAME) syntax: i.e. "$$(VAR_NAME)" will
	// produce the string literal "$(VAR_NAME)". Escaped references will never be expanded, regardless
	// of whether the variable exists or not. Cannot be updated.
	// More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell
	// +optional
	// +listType=atomic
	Args []string `json:"args,omitempty" protobuf:"bytes,4,rep,name=args"`
	// Container's working directory.
	// If not specified, the container runtime's default will be used, which
	// might be configured in the container image.
	// Cannot be updated.
	// +optional
	WorkingDir string `json:"workingDir,omitempty" protobuf:"bytes,5,opt,name=workingDir"`
	// List of sources to populate environment variables in the container.
	// The keys defined within a source must be a C_IDENTIFIER. All invalid keys
	// will be reported as an event when the container is starting. When a key exists in multiple
	// sources, the value associated with the last source will take precedence.
	// Values defined by an Env with a duplicate key will take precedence.
	// Cannot be updated.
	// +optional
	// +listType=atomic
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty" protobuf:"bytes,19,rep,name=envFrom"`
	// List of environment variables to set in the container.
	// Cannot be updated.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=name
	Env []corev1.EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,7,rep,name=env"`
	// Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources
	// already allocated to the pod.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
	// Resources resize policy for the container.
	// +featureGate=InPlacePodVerticalScaling
	// +optional
	// +listType=atomic
	ResizePolicy []corev1.ContainerResizePolicy `json:"resizePolicy,omitempty" protobuf:"bytes,23,rep,name=resizePolicy"`
	// Restart policy for the container to manage the restart behavior of each
	// container within a pod.
	// This may only be set for init containers. You cannot set this field on
	// ephemeral containers.
	// +featureGate=SidecarContainers
	// +optional
	RestartPolicy *corev1.ContainerRestartPolicy `json:"restartPolicy,omitempty" protobuf:"bytes,24,opt,name=restartPolicy,casttype=ContainerRestartPolicy"`
	// Pod volumes to mount into the container's filesystem. Subpath mounts are not allowed for ephemeral containers.
	// Cannot be updated.
	// +optional
	// +patchMergeKey=mountPath
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=mountPath
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty" patchStrategy:"merge" patchMergeKey:"mountPath" protobuf:"bytes,9,rep,name=volumeMounts"`
	// volumeDevices is the list of block devices to be used by the container.
	// +patchMergeKey=devicePath
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=devicePath
	// +optional
	VolumeDevices []corev1.VolumeDevice `json:"volumeDevices,omitempty" patchStrategy:"merge" patchMergeKey:"devicePath" protobuf:"bytes,21,rep,name=volumeDevices"`
	// Probes are not allowed for ephemeral containers.
	// +optional
	LivenessProbe *corev1.Probe `json:"livenessProbe,omitempty" protobuf:"bytes,10,opt,name=livenessProbe"`
	// Probes are not allowed for ephemeral containers.
	// +optional
	ReadinessProbe *corev1.Probe `json:"readinessProbe,omitempty" protobuf:"bytes,11,opt,name=readinessProbe"`
	// Probes are not allowed for ephemeral containers.
	// +optional
	StartupProbe *corev1.Probe `json:"startupProbe,omitempty" protobuf:"bytes,22,opt,name=startupProbe"`
	// Lifecycle is not allowed for ephemeral containers.
	// +optional
	Lifecycle *corev1.Lifecycle `json:"lifecycle,omitempty" protobuf:"bytes,12,opt,name=lifecycle"`
	// Optional: Path at which the file to which the container's termination message
	// will be written is mounted into the container's filesystem.
	// Message written is intended to be brief final status, such as an assertion failure message.
	// Will be truncated by the node if greater than 4096 bytes. The total message length across
	// all containers will be limited to 12kb.
	// Defaults to /dev/termination-log.
	// Cannot be updated.
	// +optional
	TerminationMessagePath string `json:"terminationMessagePath,omitempty" protobuf:"bytes,13,opt,name=terminationMessagePath"`
	// Indicate how the termination message should be populated. File will use the contents of
	// terminationMessagePath to populate the container status message on both success and failure.
	// FallbackToLogsOnError will use the last chunk of container log output if the termination
	// message file is empty and the container exited with an error.
	// The log output is limited to 2048 bytes or 80 lines, whichever is smaller.
	// Defaults to File.
	// Cannot be updated.
	// +optional
	TerminationMessagePolicy corev1.TerminationMessagePolicy `json:"-"` /* 129-byte string literal not displayed */
	// Image pull policy.
	// One of Always, Never, IfNotPresent.
	// Defaults to Always if :latest tag is specified, or IfNotPresent otherwise.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/containers/images#updating-images
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty" protobuf:"bytes,14,opt,name=imagePullPolicy,casttype=PullPolicy"`
	// Optional: SecurityContext defines the security options the ephemeral container should be run with.
	// If set, the fields of SecurityContext override the equivalent fields of PodSecurityContext.
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty" protobuf:"bytes,15,opt,name=securityContext"`

	// Ports are not allowed for ephemeral containers.
	// +optional
	// +patchMergeKey=containerPort
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=containerPort
	// +listMapKey=protocol
	Ports []CustomContainerPort `json:"ports,omitempty" patchStrategy:"merge" patchMergeKey:"containerPort" protobuf:"bytes,6,rep,name=ports"`

	// Whether this container should allocate a buffer for stdin in the container runtime. If this
	// is not set, reads from stdin in the container will always result in EOF.
	// Default is false.
	// +optional
	Stdin bool `json:"stdin,omitempty" protobuf:"varint,16,opt,name=stdin"`
	// Whether the container runtime should close the stdin channel after it has been opened by
	// a single attach. When stdin is true the stdin stream will remain open across multiple attach
	// sessions. If stdinOnce is set to true, stdin is opened on container start, is empty until the
	// first client attaches to stdin, and then remains open and accepts data until the client disconnects,
	// at which time stdin is closed and remains closed until the container is restarted. If this
	// flag is false, a container processes that reads from stdin will never receive an EOF.
	// Default is false
	// +optional
	StdinOnce bool `json:"stdinOnce,omitempty" protobuf:"varint,17,opt,name=stdinOnce"`
	// Whether this container should allocate a TTY for itself, also requires 'stdin' to be true.
	// Default is false.
	// +optional
	TTY bool `json:"tty,omitempty" protobuf:"varint,18,opt,name=tty"`
	// If set, the name of the container from PodSpec that this ephemeral container targets.
	// The ephemeral container will be run in the namespaces (IPC, PID, etc) of this container.
	// If not set then the ephemeral container uses the namespaces configured in the Pod spec.
	//
	// The container runtime must implement support for this feature. If the runtime does not
	// support namespace targeting then the result of setting this field is undefined.
	// +optional
	TargetContainerName string `json:"targetContainerName,omitempty" protobuf:"bytes,2,opt,name=targetContainerName"`
}

// +kubebuilder:validation:MaxProperties=1
type AMTDAction struct {
	Disable      *DisableAction    `json:"disable,omitempty"`
	Delete       *DeleteAction     `json:"delete,omitempty"`
	Quarantine   *QuarantineAction `json:"quarantine,omitempty"`
	Debugger     *Debugger         `json:"debugger,omitempty"`
	CustomAction *CustomAction     `json:"customAction,omitempty"`
}

// MovingStrategy Substructure for strategy definitions
type ResponseStrategy struct {
	//TODO: use enum for the specific values of these fields

	// +kubebuilder:validation:Required
	Rule Rule `json:"rule"`

	// +kubebuilder:validation:Required
	// Action field value of the SecurityEvent that arrives
	Action AMTDAction `json:"action"`
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

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
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
