package v1

// +k8s:deepcopy-gen=false
type NodeInfo struct {
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
}
