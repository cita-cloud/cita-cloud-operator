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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type NodeInfo struct {
	// Name
	Name string `json:"name,omitempty"`
	// 所属的k8s集群
	Cluster string `json:"cluster,omitempty"`

	// Domain
	Domain string `json:"domain,omitempty"`

	// 账号，不为空
	Account string `json:"account"`

	// 节点对外public ip
	ExternalIp string `json:"externalIp,omitempty"`

	// 节点暴露的端口号
	Port int32 `json:"port,omitempty"`

	// node status
	Status NodeStatus `json:"status,omitempty"`

	// CreationTimestamp
	CreationTimestamp *metav1.Time `json:"creationTimestamp,omitempty"`
}

// ByCreationTimestampForNode Sort from early to late
// +k8s:deepcopy-gen=false
type ByCreationTimestampForNode []NodeInfo

func (a ByCreationTimestampForNode) Len() int { return len(a) }
func (a ByCreationTimestampForNode) Less(i, j int) bool {
	return a[i].CreationTimestamp.Before(a[j].CreationTimestamp)
}
func (a ByCreationTimestampForNode) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
