package v1

// +k8s:deepcopy-gen=false
type NodeInfo struct {
	// 所属的k8s集群
	Cluster string `json:"cluster,omitempty"`

	// 对应节点名称
	Name string `json:"name,omitempty"`

	// Domain
	Domain string `json:"domain,omitempty"`

	// 节点地址
	Address string `json:"address,omitempty"`

	// 节点集群内部ip
	InternalIp string `json:"internalIp,omitempty"`

	// 节点对外public ip
	ExternalIp string `json:"externalIp,omitempty"`

	// 节点暴露的端口号
	Port string `json:"port,omitempty"`
}
