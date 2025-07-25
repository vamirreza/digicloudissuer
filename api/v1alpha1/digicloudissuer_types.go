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
	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  This is scaffolding for you to own.
// NOTE: json tags are required.  Any new fields you add must have json:"-" or json:"fieldName" tags for the fields to be serialized.

// DigicloudIssuerSpec defines the desired state of DigicloudIssuer
type DigicloudIssuerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Provisioner contains the provisioner configuration for the issuer
	Provisioner DigicloudIssuerProvisioner `json:"provisioner"`
}

// DigicloudIssuerProvisioner contains the configuration for the Digicloud DNS provider
type DigicloudIssuerProvisioner struct {
	// APIBaseURL is the base URL for the Digicloud API
	// +kubebuilder:default="https://api.digicloud.ir"
	APIBaseURL string `json:"apiBaseUrl,omitempty"`

	// APITokenSecretRef is a reference to a secret containing the Digicloud API token
	APITokenSecretRef SecretKeySelector `json:"apiTokenSecretRef"`

	// TTL is the time-to-live for DNS records in seconds
	// +kubebuilder:default=300
	// +kubebuilder:validation:Minimum=60
	// +kubebuilder:validation:Maximum=86400
	TTL *int `json:"ttl,omitempty"`

	// PropagationTimeout is the maximum time to wait for DNS propagation
	// +kubebuilder:default="5m"
	PropagationTimeout *metav1.Duration `json:"propagationTimeout,omitempty"`

	// PollingInterval is the interval between DNS propagation checks
	// +kubebuilder:default="10s"
	PollingInterval *metav1.Duration `json:"pollingInterval,omitempty"`
}

// SecretKeySelector is a reference to a secret key
type SecretKeySelector struct {
	// Name is the name of the secret
	Name string `json:"name"`

	// Key is the key within the secret
	Key string `json:"key"`
}

// DigicloudIssuerStatus defines the observed state of DigicloudIssuer
type DigicloudIssuerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Conditions represent the latest available observations of the issuer's state
	Conditions []cmapi.IssuerCondition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status"
//+kubebuilder:printcolumn:name="Reason",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].reason"
//+kubebuilder:printcolumn:name="Message",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message"

// DigicloudIssuer is the Schema for the digicloudissuers API
type DigicloudIssuer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DigicloudIssuerSpec   `json:"spec,omitempty"`
	Status DigicloudIssuerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DigicloudIssuerList contains a list of DigicloudIssuer
type DigicloudIssuerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DigicloudIssuer `json:"items"`
}

// DigicloudClusterIssuerSpec defines the desired state of DigicloudClusterIssuer
type DigicloudClusterIssuerSpec struct {
	// Provisioner contains the provisioner configuration for the cluster issuer
	Provisioner DigicloudIssuerProvisioner `json:"provisioner"`
}

// DigicloudClusterIssuerStatus defines the observed state of DigicloudClusterIssuer
type DigicloudClusterIssuerStatus struct {
	// Conditions represent the latest available observations of the cluster issuer's state
	Conditions []cmapi.IssuerCondition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status"
//+kubebuilder:printcolumn:name="Reason",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].reason"
//+kubebuilder:printcolumn:name="Message",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message"

// DigicloudClusterIssuer is the Schema for the digicloudclusterissuers API
type DigicloudClusterIssuer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DigicloudClusterIssuerSpec   `json:"spec,omitempty"`
	Status DigicloudClusterIssuerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DigicloudClusterIssuerList contains a list of DigicloudClusterIssuer
type DigicloudClusterIssuerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DigicloudClusterIssuer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DigicloudIssuer{}, &DigicloudIssuerList{})
	SchemeBuilder.Register(&DigicloudClusterIssuer{}, &DigicloudClusterIssuerList{})
}
