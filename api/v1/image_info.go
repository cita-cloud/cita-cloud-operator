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
