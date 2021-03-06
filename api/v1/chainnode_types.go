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
	v1 "k8s.io/api/core/v1"
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

	// 日志等级
	LogLevel LogLevel `json:"logLevel,omitempty"`

	// pvc对应的storage class, 集群中必须存在的storage class
	StorageClassName *string `json:"storageClassName,omitempty"`

	// 申请的pvc大小
	StorageSize *int64 `json:"storageSize"`

	// 期望的状态
	// +kubebuilder:validation:Enum=Initialize;Stop;Start
	Action NodeAction `json:"action"`

	// ImageInfo
	ImageInfo `json:"imageInfo,omitempty"`

	// Resources describes the compute resource request and limit (include cpu、memory)
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
}

type LogLevel string

const (
	Info LogLevel = "info"
	Warn LogLevel = "warn"
)

type NodeAction string

const (
	NodeInitialize NodeAction = "Initialize"
	NodeStop       NodeAction = "Stop"
	NodeStart      NodeAction = "Start"
)

type NodeStatus string

const (
	NodeWaitChainOnline NodeStatus = "WaitChainOnline"
	NodeInitialized     NodeStatus = "Initialized"
	NodeStarting        NodeStatus = "Starting"
	NodeRunning         NodeStatus = "Running"
	NodeWarning         NodeStatus = "Warning"
	NodeError           NodeStatus = "Error"
	NodeUpdating        NodeStatus = "Updating"
	// NodeNeedRestart if chainnode's config modified, chainnode should restart
	NodeNeedRestart NodeStatus = "NeedRestart"
	NodeStopping    NodeStatus = "Stopping"
	NodeStopped     NodeStatus = "Stopped"
)

// ChainNodeStatus defines the observed state of ChainNode
type ChainNodeStatus struct {
	Status NodeStatus `json:"status,omitempty"`
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
