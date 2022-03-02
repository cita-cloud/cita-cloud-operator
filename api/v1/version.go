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

import corev1 "k8s.io/api/core/v1"

var VERSION_MAP = map[string]ImageInfo{
	// v6.3.2
	VERSION632_TLS_BFT: {
		PullPolicy:      corev1.PullIfNotPresent,
		NetworkImage:    "citacloud/network_tls:v6.3.0",
		ConsensusImage:  "citacloud/consensus_bft:v6.3.2-alpha.1",
		ExecutorImage:   "citacloud/executor_evm:latest",
		StorageImage:    "citacloud/storage_rocksdb:v6.3.0",
		ControllerImage: "citacloud/controller:v6.3.2-alpha.1",
		KmsImage:        "citacloud/kms_sm:v6.3.1",
	},
	VERSION632_TLS_RAFT: {
		PullPolicy:      corev1.PullIfNotPresent,
		NetworkImage:    "citacloud/network_tls:v6.3.0",
		ConsensusImage:  "citacloud/consensus_raft:v6.3.0",
		ExecutorImage:   "citacloud/executor_evm:latest",
		StorageImage:    "citacloud/storage_rocksdb:v6.3.0",
		ControllerImage: "citacloud/controller:v6.3.2-alpha.1",
		KmsImage:        "citacloud/kms_sm:v6.3.1",
	},
	VERSION632_P2P_BFT: {
		PullPolicy:      corev1.PullIfNotPresent,
		NetworkImage:    "citacloud/network_p2p:v6.3.0",
		ConsensusImage:  "citacloud/consensus_bft:v6.3.2-alpha.1",
		ExecutorImage:   "citacloud/executor_evm:latest",
		StorageImage:    "citacloud/storage_rocksdb:v6.3.0",
		ControllerImage: "citacloud/controller:v6.3.2-alpha.1",
		KmsImage:        "citacloud/kms_sm:v6.3.1",
	},
	VERSION632_P2P_RAFT: {
		PullPolicy:      corev1.PullIfNotPresent,
		NetworkImage:    "citacloud/network_p2p:v6.3.0",
		ConsensusImage:  "citacloud/consensus_raft:v6.3.0",
		ExecutorImage:   "citacloud/executor_evm:latest",
		StorageImage:    "citacloud/storage_rocksdb:v6.3.0",
		ControllerImage: "citacloud/controller:v6.3.2-alpha.1",
		KmsImage:        "citacloud/kms_sm:v6.3.1",
	},
}
