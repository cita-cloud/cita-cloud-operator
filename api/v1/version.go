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
	// v6.3.3
	VERSION633_TLS_BFT: {
		PullPolicy:      corev1.PullIfNotPresent,
		NetworkImage:    "citacloud/network_tls:v6.3.3",
		ConsensusImage:  "citacloud/consensus_bft:v6.3.3",
		ExecutorImage:   "citacloud/executor_evm:v6.3.3",
		StorageImage:    "citacloud/storage_rocksdb:v6.3.3",
		ControllerImage: "citacloud/controller:v6.3.3",
		KmsImage:        "citacloud/kms_sm:v6.3.3",
	},
	VERSION633_TLS_RAFT: {
		PullPolicy:      corev1.PullIfNotPresent,
		NetworkImage:    "citacloud/network_tls:v6.3.3",
		ConsensusImage:  "citacloud/consensus_raft:v6.3.3",
		ExecutorImage:   "citacloud/executor_evm:v6.3.3",
		StorageImage:    "citacloud/storage_rocksdb:v6.3.3",
		ControllerImage: "citacloud/controller:v6.3.3",
		KmsImage:        "citacloud/kms_sm:v6.3.3",
	},
	VERSION633_P2P_BFT: {
		PullPolicy:      corev1.PullIfNotPresent,
		NetworkImage:    "citacloud/network_p2p:v6.3.3",
		ConsensusImage:  "citacloud/consensus_bft:v6.3.3",
		ExecutorImage:   "citacloud/executor_evm:v6.3.3",
		StorageImage:    "citacloud/storage_rocksdb:v6.3.3",
		ControllerImage: "citacloud/controller:v6.3.3",
		KmsImage:        "citacloud/kms_sm:v6.3.3",
	},
	VERSION633_P2P_RAFT: {
		PullPolicy:      corev1.PullIfNotPresent,
		NetworkImage:    "citacloud/network_p2p:v6.3.3",
		ConsensusImage:  "citacloud/consensus_raft:v6.3.3",
		ExecutorImage:   "citacloud/executor_evm:v6.3.3",
		StorageImage:    "citacloud/storage_rocksdb:v6.3.3",
		ControllerImage: "citacloud/controller:v6.3.3",
		KmsImage:        "citacloud/kms_sm:v6.3.3",
	},
	// v6.4.0
	VERSION640_TLS_BFT: {
		PullPolicy:      corev1.PullIfNotPresent,
		NetworkImage:    "citacloud/network_tls:v6.4.0",
		ConsensusImage:  "citacloud/consensus_bft:v6.4.0",
		ExecutorImage:   "citacloud/executor_evm:v6.4.0",
		StorageImage:    "citacloud/storage_rocksdb:v6.4.0",
		ControllerImage: "citacloud/controller:v6.4.0",
		KmsImage:        "citacloud/kms_sm:v6.4.0",
	},
	VERSION640_TLS_RAFT: {
		PullPolicy:      corev1.PullIfNotPresent,
		NetworkImage:    "citacloud/network_tls:v6.4.0",
		ConsensusImage:  "citacloud/consensus_raft:v6.4.0",
		ExecutorImage:   "citacloud/executor_evm:v6.4.0",
		StorageImage:    "citacloud/storage_rocksdb:v6.4.0",
		ControllerImage: "citacloud/controller:v6.4.0",
		KmsImage:        "citacloud/kms_sm:v6.4.0",
	},
	VERSION640_P2P_BFT: {
		PullPolicy:      corev1.PullIfNotPresent,
		NetworkImage:    "citacloud/network_p2p:v6.4.0",
		ConsensusImage:  "citacloud/consensus_bft:v6.4.0",
		ExecutorImage:   "citacloud/executor_evm:v6.4.0",
		StorageImage:    "citacloud/storage_rocksdb:v6.4.0",
		ControllerImage: "citacloud/controller:v6.4.0",
		KmsImage:        "citacloud/kms_sm:v6.4.0",
	},
	VERSION640_P2P_RAFT: {
		PullPolicy:      corev1.PullIfNotPresent,
		NetworkImage:    "citacloud/network_p2p:v6.4.0",
		ConsensusImage:  "citacloud/consensus_raft:v6.4.0",
		ExecutorImage:   "citacloud/executor_evm:v6.4.0",
		StorageImage:    "citacloud/storage_rocksdb:v6.4.0",
		ControllerImage: "citacloud/controller:v6.4.0",
		KmsImage:        "citacloud/kms_sm:v6.4.0",
	},
	// latest
	LATEST_TLS_BFT: {
		PullPolicy:      corev1.PullIfNotPresent,
		NetworkImage:    "citacloud/network_tls:latest",
		ConsensusImage:  "citacloud/consensus_bft:latest",
		ExecutorImage:   "citacloud/executor_evm:latest",
		StorageImage:    "citacloud/storage_rocksdb:latest",
		ControllerImage: "citacloud/controller:latest",
		KmsImage:        "citacloud/kms_sm:latest",
	},
	LATEST_TLS_RAFT: {
		PullPolicy:      corev1.PullIfNotPresent,
		NetworkImage:    "citacloud/network_tls:latest",
		ConsensusImage:  "citacloud/consensus_raft:latest",
		ExecutorImage:   "citacloud/executor_evm:latest",
		StorageImage:    "citacloud/storage_rocksdb:latest",
		ControllerImage: "citacloud/controller:latest",
		KmsImage:        "citacloud/kms_sm:latest",
	},
	LATEST_P2P_BFT: {
		PullPolicy:      corev1.PullIfNotPresent,
		NetworkImage:    "citacloud/network_p2p:latest",
		ConsensusImage:  "citacloud/consensus_bft:latest",
		ExecutorImage:   "citacloud/executor_evm:latest",
		StorageImage:    "citacloud/storage_rocksdb:latest",
		ControllerImage: "citacloud/controller:latest",
		KmsImage:        "citacloud/kms_sm:latest",
	},
	LATEST_P2P_RAFT: {
		PullPolicy:      corev1.PullIfNotPresent,
		NetworkImage:    "citacloud/network_p2p:latest",
		ConsensusImage:  "citacloud/consensus_raft:latest",
		ExecutorImage:   "citacloud/executor_evm:latest",
		StorageImage:    "citacloud/storage_rocksdb:latest",
		ControllerImage: "citacloud/controller:latest",
		KmsImage:        "citacloud/kms_sm:latest",
	},
}
