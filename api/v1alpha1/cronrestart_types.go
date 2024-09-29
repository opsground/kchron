package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CronRestartSpec defines the desired state of CronRestart
type CronRestartSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Namespace    string   `json:"namespace"`
	ResourceType string   `json:"resourceType"` // e.g., Deployment or StatefulSet
	Resources    []string `json:"resources"`    // List of deployment/statefulset names
	CronSchedule string   `json:"cronSchedule"` // Cron expression for scheduling
}

// CronRestartStatus defines the observed state of CronRestart
type CronRestartStatus struct {
	LastScheduleTime *metav1.Time `json:"lastScheduleTime,omitempty"`
}

// +kubebuilder:resource:path=cronrestarts,scope=Namespaced,shortName=crs
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name=Namespace,type=string,description="Namespace of the resources",JSONPath=".spec.namespace"
// +kubebuilder:printcolumn:name=ResourceType,type=string,description="Type of resource (Deployment/StatefulSet)",JSONPath=".spec.resourceType"
// +kubebuilder:printcolumn:name=Resources,type=string,description="Names of resources",JSONPath=".spec.resources"
// +kubebuilder:printcolumn:name=Schedule,type=string,description="Cron schedule for restarts",JSONPath=".spec.cronSchedule"

// CronRestart is the Schema for the cronrestarts API
type CronRestart struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CronRestartSpec   `json:"spec,omitempty"`
	Status CronRestartStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CronRestartList contains a list of CronRestart
type CronRestartList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CronRestart `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CronRestart{}, &CronRestartList{})
}
