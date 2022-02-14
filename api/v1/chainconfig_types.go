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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ChainConfigSpec defines the desired state of ChainConfig
type ChainConfigSpec struct {
	// chain id
	Id string `json:"id"`

	// 创世块的时间戳
	Timestamp int64 `json:"timestamp"`

	// 创世块的prevhash
	PrevHash string `json:"prevhash"`

	// 出块间隔
	BlockInterval int32 `json:"blockInterval"`

	// 块大小限制
	BlockLimit int32 `json:"blockLimit"`

	// 开启tls认证
	EnableTLS bool `json:"enableTls,omitempty"`

	// 共识算法
	ConsensusType ConsensusType `json:"consensusType"`

	// 期望的状态
	Action ChainStatus `json:"action,omitempty"`

	// ImageInfo
	ImageInfo `json:"imageInfo,omitempty"`
}

type ConsensusType string

const (
	BFT  ConsensusType = "BTF"
	Raft ConsensusType = "Raft"
)

type ChainStatus string

const (
	// 初始化状态
	Initialization ChainStatus = "Initialization"
	// 运行状态
	Running ChainStatus = "Running"
	// 治理维护状态
	Maintenance ChainStatus = "Maintenance"
	// 公示中状态
	Publicizing ChainStatus = "Publicizing"
	// 上线
	Online ChainStatus = "Online"
)

// ChainConfigStatus defines the observed state of ChainConfig
type ChainConfigStatus struct {
	// admin账户信息
	AdminAccount *AdminAccountInfo `json:"adminAccount,omitempty"`
	// 共识节点账户信息
	ValidatorAccountList []ValidatorAccountInfo `json:"validatorAccountList,omitempty"`
	// 状态
	Status ChainStatus `json:"status,omitempty"`
	// 详情
	Message string `json:"message,omitempty"`
	// 链下节点信息
	NodeInfoList []NodeInfo `json:"nodeInfoList,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ChainConfig is the Schema for the chainconfigs API
type ChainConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ChainConfigSpec   `json:"spec,omitempty"`
	Status ChainConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ChainConfigList contains a list of ChainConfig
type ChainConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ChainConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ChainConfig{}, &ChainConfigList{})
}
