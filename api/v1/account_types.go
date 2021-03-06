/*
 * Copyright Rivtower Technologies LLC.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Role string

const (
	Admin     Role = "Admin"
	Consensus Role = "Consensus"
	Ordinary  Role = "Ordinary"
)

// AccountSpec defines the desired state of Account
type AccountSpec struct {
	// Corresponding chain name
	Chain string `json:"chain"`
	// Role type
	// +kubebuilder:validation:Enum=Admin;Consensus;Ordinary
	Role Role `json:"role,omitempty"`
	// kms password
	KmsPassword string `json:"kmsPassword"`
	// if role is Ordinary, you should set this field
	Domain string `json:"domain,omitempty"`
	// Address if address set, controller will not generate a new account address
	Address string `json:"address,omitempty"`
}

// AccountStatus defines the observed state of Account
type AccountStatus struct {
	// Currently only the admin user will set this value
	Cert string `json:"cert,omitempty"`
	// Address
	Address string `json:"address,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Account is the Schema for the accounts API
type Account struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AccountSpec   `json:"spec,omitempty"`
	Status AccountStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AccountList contains a list of Account
type AccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Account `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Account{}, &AccountList{})
}
