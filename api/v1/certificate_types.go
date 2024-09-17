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

package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CertificateSpec defines the desired state of Certificate
type CertificateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// DnsName field specifies the DNSName for the Certificate.
	DNSName string `json:"dnsName"`
	// Validity field specifies the validity for the certificate.
	Validity string `json:"validity"`
	// SecretRef field references the secret where certificate will be stored.
	SecretRef v1.SecretReference `json:"secretRef"`
}

// CertificateStatus defines the observed state of Certificate
type CertificateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Conditions field represents the observations of state
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// SerialNumber field represents the serial number of certificate
	SerialNumber string `json:"serialNumber,omitempty"`

	// NotBefore field represents the time when certificate is valid
	NotBefore metav1.Time `json:"notBefore,omitempty"`

	//NotAfter field represents the time when certificate is invalid
	NotAfter metav1.Time `json:"notAfter,omitempty"`

	//Issuer field represents the issuer of certificate
	Issuer string `json:"issuer,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Certificate is the Schema for the certificates API
type Certificate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CertificateSpec   `json:"spec,omitempty"`
	Status CertificateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CertificateList contains a list of Certificate
type CertificateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Certificate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Certificate{}, &CertificateList{})
}
