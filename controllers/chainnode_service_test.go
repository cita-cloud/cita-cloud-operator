package controllers

import (
	"testing"

	citacloudv1 "github.com/cita-cloud/cita-cloud-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestChainNodeService_GenerateNodeConfig(t *testing.T) {
	type fields struct {
		ChainConfig *citacloudv1.ChainConfig
		ChainNode   *citacloudv1.ChainNode
	}

	f := fields{
		ChainConfig: &citacloudv1.ChainConfig{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Spec: citacloudv1.ChainConfigSpec{
				Id:            "63586a3c0255f337c77a777ff54f0040b8c388da04f23ecee6bfd4953a6512b4",
				AdminAddress:  "9561d0a4f60307347c50228347b41ccda5d4b7f0",
				Validators:    []string{"0bb1af20c7532dad18f616936e26864cb8fac305", "2d5b15ef0edc69bccae1a9dbc2f4d05671b48b12"},
				Timestamp:     1639105556777,
				PrevHash:      "0x0000000000000000000000000000000000000000000000000000000000000000",
				BlockInterval: 3,
				BlockLimit:    100,
				EnableTLS:     false,
				KmsPassword:   "123456",
				ImageInfo:     citacloudv1.ImageInfo{},
			},
			Status: citacloudv1.ChainConfigStatus{
				CaCert:  "",
				CaKey:   "",
				Status:  "",
				Message: "",
				NodeInfoMap: map[string]citacloudv1.NodeInfo{
					"my-node-1": {
						Cluster:    "",
						Domain:     "",
						Address:    "0bb1af20c7532dad18f616936e26864cb8fac305",
						InternalIp: "10.10.30.98",
						ExternalIp: "",
						Port:       9999,
					},
					"my-node-2": {
						Cluster:    "",
						Domain:     "",
						Address:    "2d5b15ef0edc69bccae1a9dbc2f4d05671b48b12",
						InternalIp: "10.10.30.99",
						ExternalIp: "",
						Port:       9999,
					},
				},
			},
		},
		ChainNode: &citacloudv1.ChainNode{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:         "my-node-1",
				GenerateName: "",
				Namespace:    "",
			},
			Spec: citacloudv1.ChainNodeSpec{
				NodeInfo: citacloudv1.NodeInfo{
					Cluster:    "k8s-1",
					Domain:     "",
					Address:    "0bb1af20c7532dad18f616936e26864cb8fac305",
					InternalIp: "10.10.30.98",
					ExternalIp: "",
					Port:       9999,
				},
				ChainName:        "my-chainconfig",
				KmsPassword:      "123456",
				LogLevel:         citacloudv1.Info,
				StorageClassName: nil,
				StorageSize:      nil,
				Action:           "",
				Type:             citacloudv1.Consensus,
			},
			Status: citacloudv1.ChainNodeStatus{},
		},
	}

	var tests []struct {
		name   string
		fields fields
		want   string
	}
	tests = append(tests, struct {
		name   string
		fields fields
		want   string
	}{name: "my-test", fields: f, want: `[network_p2p]
grpc_port = 50000
port = 40000

[[network_p2p.peers]]
address = '/dns4/10.10.30.98/tcp/9999'

[consensus_raft]
controller_port = 50004
grpc_listen_port = 50001
network_port = 50000
node_addr = '0bb1af20c7532dad18f616936e26864cb8fac305'

[executor_evm]
executor_port = 50002

[storage_rocksdb]
kms_port = 50005
storage_port = 50003

[genesis_block]
prevhash = '0x0000000000000000000000000000000000000000000000000000000000000000'
timestamp = 1639105556777

[system_config]
admin = '9561d0a4f60307347c50228347b41ccda5d4b7f0'
block_interval = 3
block_limit = 100
chain_id = '63586a3c0255f337c77a777ff54f0040b8c388da04f23ecee6bfd4953a6512b4'
version = 0
validators = [
    '0bb1af20c7532dad18f616936e26864cb8fac305',
    '2d5b15ef0edc69bccae1a9dbc2f4d05671b48b12',
]

[controller]
consensus_port = 50001
controller_port = 50004
executor_port = 50002
key_id = 1
kms_port = 50005
network_port = 50000
node_address = '0bb1af20c7532dad18f616936e26864cb8fac305'
package_limit = 30000
storage_port = 50003

[kms_sm]
db_key = '123456'
kms_port = 50005
`})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cns := &ChainNodeService{
				ChainConfig: tt.fields.ChainConfig,
				ChainNode:   tt.fields.ChainNode,
			}
			if got := cns.GenerateNodeConfig(); got != tt.want {
				t.Errorf("GenerateNodeConfig() = %v, \n\n\nwant %v", got, tt.want)
			}
		})
	}
}
