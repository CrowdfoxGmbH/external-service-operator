package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ExternalServiceHostPath struct {
	Host string `json:"host"`
	Path string `json:"path"`
}

// ExternalServiceSpec defines the desired state of ExternalService
// +k8s:openapi-gen=true
type ExternalServiceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Port           int32                     `json:"port"`
	Ips            []string                  `json:"ips"`
	Hosts          []ExternalServiceHostPath `json:"hosts"`
	ReadinessProbe corev1.Probe              `json:"readinessProbe"`
}

// ExternalServiceStatus defines the observed state of ExternalService
// +k8s:openapi-gen=true
type ExternalServiceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ExternalService is the Schema for the externalservices API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type ExternalService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExternalServiceSpec   `json:"spec,omitempty"`
	Status ExternalServiceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ExternalServiceList contains a list of ExternalService
type ExternalServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExternalService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ExternalService{}, &ExternalServiceList{})
}
