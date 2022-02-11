package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type AdminAccountInfo struct {
	Name    string `json:"name,omitempty"`
	Address string `json:"address,omitempty"`
}

type ValidatorAccountInfo struct {
	Name    string `json:"name,omitempty"`
	Address string `json:"address,omitempty"`
	// CreationTimestamp for sort
	CreationTimestamp *metav1.Time `json:"creationTimestamp,omitempty"`
}

// ByCreationTimestampForValidatorAccount Sort from early to late
// +k8s:deepcopy-gen=false
type ByCreationTimestampForValidatorAccount []ValidatorAccountInfo

func (a ByCreationTimestampForValidatorAccount) Len() int { return len(a) }
func (a ByCreationTimestampForValidatorAccount) Less(i, j int) bool {
	return a[i].CreationTimestamp.Before(a[j].CreationTimestamp)
}
func (a ByCreationTimestampForValidatorAccount) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
