/*
Copyright 2021.

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
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ChainNodeSpec defines the desired state of ChainNode
type ChainNodeSpec struct {

	// 节点信息
	NodeInfo `json:"nodeInfo"`

	// 对应的链级配置名称
	ChainName string `json:"chainName,omitempty"`

	// 节点用户的kms password
	KmsPassword string `json:"kmsPassword,omitempty"`

	// 日志等级
	LogLevel string `json:"logLevel,omitempty"`

	// pvc对应的storage class, 集群中必须存在的storage class
	StorageClassName *string `json:"storageClassName,omitempty"`

	// 申请的pvc大小
	StorageSize *int64 `json:"storageSize"`
}

type NodeStatus string

const (
	Creating NodeStatus = "Creating"
	//Running  NodeStatus = "Running"
	Warning NodeStatus = "Warning"
	Error   NodeStatus = "Error"
)

// ChainNodeStatus defines the observed state of ChainNode
type ChainNodeStatus struct {
	Status NodeStatus `json:"status,omitempty"`
	// node address
	Address string `json:"address,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ChainNode is the Schema for the chainnodes API
type ChainNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ChainNodeSpec   `json:"spec,omitempty"`
	Status ChainNodeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ChainNodeList contains a list of ChainNode
type ChainNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ChainNode `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ChainNode{}, &ChainNodeList{})
}
