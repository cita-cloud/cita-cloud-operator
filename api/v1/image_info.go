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

import v1 "k8s.io/api/core/v1"

// +k8s:deepcopy-gen=false
type ImageInfo struct {
	// PullPolicy
	PullPolicy v1.PullPolicy `json:"pullPolicy,omitempty"`

	// network微服务镜像
	NetworkImage string `json:"networkImage,omitempty"`

	// consensus微服务镜像
	ConsensusImage string `json:"consensusImage,omitempty"`

	// executor微服务镜像
	ExecutorImage string `json:"executorImage,omitempty"`

	// storage微服务镜像
	StorageImage string `json:"storageImage,omitempty"`

	// controller微服务镜像
	ControllerImage string `json:"controllerImage,omitempty"`

	// kms微服务镜像
	KmsImage string `json:"kmsImage,omitempty"`
}
