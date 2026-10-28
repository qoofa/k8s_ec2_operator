/*
Copyright 2026.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Ec2instanceSpec defines the desired state of Ec2instance
type Ec2instanceSpec struct {
	InstanceType      string            `json:"instanceType"`
	AMIId             string            `json:"amiId"`
	Region            string            `json:"region"`
	AvailabilityZone  string            `json:"availabilityZone,omitempty"`
	KeyPair           string            `json:"keyPair,omitempty"`
	SecurityGroups    []string          `json:"securityGroups,omitempty"`
	Subnet            string            `json:"subnet,omitempty"`
	UserData          string            `json:"userData,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`
	Storage           StorageConfig     `json:"storage,omitempty"`
	AssociatePublicIP bool              `json:"associatePublicIp,omitempty"`
}

type StorageConfig struct {
	RootVolume        VolumeConfig   `json:"rootVolume,omitempty"`
	AdditionalVolumes []VolumeConfig `json:"additionalVolumes,omitempty"`
}

type VolumeConfig struct {
	Size       int32  `json:"size"`
	Type       string `json:"type,omitempty"`
	DeviceName string `json:"deviceName,omitempty"`
	Encrypted  bool   `json:"encrypted,omitempty"`
}

// Ec2instanceStatus defines the observed state of Ec2instance.
type Ec2instanceStatus struct {
	InstanceID string       `json:"instanceID,omitempty"`
	State      string       `json:"state,omitempty"`
	PublicIP   string       `json:"publicIp,omitempty"`
	PrivateIP  string       `json:"privateIp,omitempty"`
	PublicDNS  string       `json:"publicDNS,omitempty"`
	PrivateDNS string       `json:"privateDNS,omitempty"`
	LaunchTime *metav1.Time `json:"launchTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Ec2instance is the Schema for the ec2instances API
type Ec2instance struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of Ec2instance
	// +required
	Spec Ec2instanceSpec `json:"spec"`

	// status defines the observed state of Ec2instance
	// +optional
	Status Ec2instanceStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// Ec2instanceList contains a list of Ec2instance
type Ec2instanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []Ec2instance `json:"items"`
}

type CreatedInstanceInfo struct {
	InstanceID string `json:"instanceId"`
	PublicIP   string `json:"publicIP"`
	PrivateIP  string `json:"privateIP"`
	PublicDNS  string `json:"publicDNS"`
	PrivateDNS string `json:"privateDNS"`
	State      string `json:"state"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &Ec2instance{}, &Ec2instanceList{})
		return nil
	})
}
