package v1

// +k8s:deepcopy-gen=false
type ImageInfo struct {
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

func NewImageInfo(networkImage string, consensusImage string, executorImage string, storageImage string, controllerImage string, kmsImage string) *ImageInfo {
	return &ImageInfo{NetworkImage: networkImage, ConsensusImage: consensusImage, ExecutorImage: executorImage, StorageImage: storageImage, ControllerImage: controllerImage, KmsImage: kmsImage}
}

