package v1

// +k8s:deepcopy-gen=false
type AdminAccountInfo struct {
	Name    string `json:"name,omitempty"`
	Address string `json:"address,omitempty"`
}

// +k8s:deepcopy-gen=false
type ValidatorAccountInfo struct {
	Address string `json:"address,omitempty"`
}
