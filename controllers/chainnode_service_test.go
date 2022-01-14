package controllers

import (
	"testing"

	citacloudv1 "github.com/cita-cloud/cita-cloud-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestChainNodeService_GenerateNodeConfigForP2P(t *testing.T) {
	type fields struct {
		ChainConfig *citacloudv1.ChainConfig
		ChainNode   *citacloudv1.ChainNode
		Account     *citacloudv1.Account
	}

	f := fields{
		ChainConfig: &citacloudv1.ChainConfig{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-chain",
			},
			Spec: citacloudv1.ChainConfigSpec{
				Id:            "63586a3c0255f337c77a777ff54f0040b8c388da04f23ecee6bfd4953a6512b4",
				Timestamp:     1639105556777,
				PrevHash:      "0x0000000000000000000000000000000000000000000000000000000000000000",
				BlockInterval: 3,
				BlockLimit:    100,
				EnableTLS:     false,
				ConsensusType: citacloudv1.Raft,
				Action:        citacloudv1.Online,
			},
			Status: citacloudv1.ChainConfigStatus{
				AdminAccount: &citacloudv1.AdminAccountInfo{
					Name:    "admin",
					Address: "a3203aa7717ba46457d3cb373b10382adc763d2d",
				},
				ValidatorAccountMap: map[string]citacloudv1.ValidatorAccountInfo{
					"node1-account": {
						Address: "211dd8d62d6abb236a8ca72dc75489dcc6a79ce2",
					},
					"node2-account": {
						Address: "c5b606cbd200b79f05a3ed4d92ce9581a0b8ec12",
					},
				},
				Status:  citacloudv1.Online,
				Message: "",
				NodeInfoMap: map[string]citacloudv1.NodeInfo{
					"my-node-1": {
						Cluster:    "k8s-1",
						Account:    "node1-account",
						Domain:     "www.my-node-1.com",
						ExternalIp: "1.1.1.1",
						Port:       9630,
						Status:     citacloudv1.NodeInitialized,
					},
					"my-node-2": {
						Cluster:    "k8s-1",
						Account:    "node2-account",
						Domain:     "www.my-node-2.com",
						ExternalIp: "2.2.2.2",
						Port:       9630,
						Status:     citacloudv1.NodeInitialized,
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
					Account:    "node1-account",
					Domain:     "www.my-node-1.com",
					ExternalIp: "1.1.1.1",
					Port:       9630,
				},
				ChainName:        "test-chain",
				LogLevel:         citacloudv1.Info,
				StorageClassName: nil,
				StorageSize:      nil,
				Action:           citacloudv1.NodeInitialize,
			},
			Status: citacloudv1.ChainNodeStatus{Status: citacloudv1.NodeInitialized},
		},
		Account: &citacloudv1.Account{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:         "node1-account",
				GenerateName: "",
				Namespace:    "",
			},
			Spec: citacloudv1.AccountSpec{
				KmsPassword: "123456",
			},
			Status: citacloudv1.AccountStatus{
				Cert:    "",
				Address: "211dd8d62d6abb236a8ca72dc75489dcc6a79ce2",
			},
		},
	}

	var tests []struct {
		name   string
		fields fields
		want1  string
		want2  string
	}
	tests = append(tests, struct {
		name   string
		fields fields
		want1  string
		want2  string
	}{name: "my-test", fields: f, want1: `[network_p2p]
grpc_port = 50000
port = 40000

[[network_p2p.peers]]
address = '/dns4/test-chain-my-node-2-cluster-ip/tcp/40000'

[consensus_raft]
controller_port = 50004
grpc_listen_port = 50001
network_port = 50000
node_addr = '211dd8d62d6abb236a8ca72dc75489dcc6a79ce2'

[executor_evm]
executor_port = 50002

[storage_rocksdb]
kms_port = 50005
storage_port = 50003

[genesis_block]
prevhash = '0x0000000000000000000000000000000000000000000000000000000000000000'
timestamp = 1639105556777

[system_config]
admin = 'a3203aa7717ba46457d3cb373b10382adc763d2d'
block_interval = 3
block_limit = 100
chain_id = '63586a3c0255f337c77a777ff54f0040b8c388da04f23ecee6bfd4953a6512b4'
version = 0
validators = [
    '211dd8d62d6abb236a8ca72dc75489dcc6a79ce2',
    'c5b606cbd200b79f05a3ed4d92ce9581a0b8ec12',
]

[controller]
consensus_port = 50001
controller_port = 50004
executor_port = 50002
key_id = 1
kms_port = 50005
network_port = 50000
node_address = '211dd8d62d6abb236a8ca72dc75489dcc6a79ce2'
package_limit = 30000
storage_port = 50003

[kms_sm]
db_key = '123456'
kms_port = 50005
`, want2: `[network_p2p]
grpc_port = 50000
port = 40000

[[network_p2p.peers]]
address = '/dns4/test-chain-my-node-2-cluster-ip/tcp/40000'

[consensus_raft]
controller_port = 50004
grpc_listen_port = 50001
network_port = 50000
node_addr = '211dd8d62d6abb236a8ca72dc75489dcc6a79ce2'

[executor_evm]
executor_port = 50002

[storage_rocksdb]
kms_port = 50005
storage_port = 50003

[genesis_block]
prevhash = '0x0000000000000000000000000000000000000000000000000000000000000000'
timestamp = 1639105556777

[system_config]
admin = 'a3203aa7717ba46457d3cb373b10382adc763d2d'
block_interval = 3
block_limit = 100
chain_id = '63586a3c0255f337c77a777ff54f0040b8c388da04f23ecee6bfd4953a6512b4'
version = 0
validators = [
    'c5b606cbd200b79f05a3ed4d92ce9581a0b8ec12',
	'211dd8d62d6abb236a8ca72dc75489dcc6a79ce2',
]

[controller]
consensus_port = 50001
controller_port = 50004
executor_port = 50002
key_id = 1
kms_port = 50005
network_port = 50000
node_address = '211dd8d62d6abb236a8ca72dc75489dcc6a79ce2'
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
				Account:     tt.fields.Account,
			}
			if got := cns.GenerateNodeConfig(); got == tt.want1 || got == tt.want2 {

			} else {
				t.Errorf("GenerateNodeConfig() = %v, \n\n\nwant1 %v, \n\n\nwant2 %v", got, tt.want1, tt.want2)
			}
		})
	}
}

func TestChainNodeService_GenerateNodeConfigForTls(t *testing.T) {
	type fields struct {
		ChainConfig          *citacloudv1.ChainConfig
		ChainNode            *citacloudv1.ChainNode
		Account              *citacloudv1.Account
		CaSecret             *corev1.Secret
		NodeCertAndKeySecret *corev1.Secret
	}

	f := fields{
		ChainConfig: &citacloudv1.ChainConfig{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-chain",
			},
			Spec: citacloudv1.ChainConfigSpec{
				Id:            "63586a3c0255f337c77a777ff54f0040b8c388da04f23ecee6bfd4953a6512b4",
				Timestamp:     1639105556777,
				PrevHash:      "0x0000000000000000000000000000000000000000000000000000000000000000",
				BlockInterval: 3,
				BlockLimit:    100,
				EnableTLS:     true,
				ConsensusType: citacloudv1.Raft,
				Action:        citacloudv1.Online,
			},
			Status: citacloudv1.ChainConfigStatus{
				AdminAccount: &citacloudv1.AdminAccountInfo{
					Name:    "admin",
					Address: "a3203aa7717ba46457d3cb373b10382adc763d2d",
				},
				ValidatorAccountMap: map[string]citacloudv1.ValidatorAccountInfo{
					"node1-account": {
						Address: "211dd8d62d6abb236a8ca72dc75489dcc6a79ce2",
					},
					"node2-account": {
						Address: "c5b606cbd200b79f05a3ed4d92ce9581a0b8ec12",
					},
				},
				Status:  citacloudv1.Online,
				Message: "",
				NodeInfoMap: map[string]citacloudv1.NodeInfo{
					"my-node-1": {
						Cluster:    "k8s-1",
						Account:    "node1-account",
						Domain:     "www.my-node-1.com",
						ExternalIp: "1.1.1.1",
						Port:       9630,
						Status:     citacloudv1.NodeInitialized,
					},
					"my-node-2": {
						Cluster:    "k8s-1",
						Account:    "node2-account",
						Domain:     "www.my-node-2.com",
						ExternalIp: "2.2.2.2",
						Port:       9630,
						Status:     citacloudv1.NodeInitialized,
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
					Account:    "node1-account",
					Domain:     "www.my-node-1.com",
					ExternalIp: "1.1.1.1",
					Port:       9630,
				},
				ChainName:        "test-chain",
				LogLevel:         citacloudv1.Info,
				StorageClassName: nil,
				StorageSize:      nil,
				Action:           citacloudv1.NodeInitialize,
			},
			Status: citacloudv1.ChainNodeStatus{Status: citacloudv1.NodeInitialized},
		},
		Account: &citacloudv1.Account{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:         "node1-account",
				GenerateName: "",
				Namespace:    "",
			},
			Spec: citacloudv1.AccountSpec{
				KmsPassword: "123456",
			},
			Status: citacloudv1.AccountStatus{
				Cert:    "",
				Address: "211dd8d62d6abb236a8ca72dc75489dcc6a79ce2",
			},
		},
		CaSecret: &corev1.Secret{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-chain-ca-secret",
			},
			Data: map[string][]byte{CaCert: {'a', 'b'}, CaKey: {'c', 'd'}},
		},
		NodeCertAndKeySecret: &corev1.Secret{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-chain-node1-account-cert-key",
			},
			Data: map[string][]byte{NodeCert: {'a', 'b'}, NodeCsr: {'c', 'd'}, NodeKey: {'e', 'f'}},
		},
	}

	var tests []struct {
		name   string
		fields fields
		want1  string
		want2  string
	}
	tests = append(tests, struct {
		name   string
		fields fields
		want1  string
		want2  string
	}{name: "my-test-tls", fields: f, want1: `[network_tls]
ca_cert = """
ab
"""
cert = """
ab
"""
grpc_port = 50000
listen_port = 40000
priv_key = """
ef
"""
reconnect_timeout = 5

[[network_tls.peers]]
domain = 'www.my-node-2.com'
host = 'test-chain-my-node-2-cluster-ip'
port = 40000

[consensus_raft]
controller_port = 50004
grpc_listen_port = 50001
network_port = 50000
node_addr = '211dd8d62d6abb236a8ca72dc75489dcc6a79ce2'

[executor_evm]
executor_port = 50002

[storage_rocksdb]
kms_port = 50005
storage_port = 50003

[genesis_block]
prevhash = '0x0000000000000000000000000000000000000000000000000000000000000000'
timestamp = 1639105556777

[system_config]
admin = 'a3203aa7717ba46457d3cb373b10382adc763d2d'
block_interval = 3
block_limit = 100
chain_id = '63586a3c0255f337c77a777ff54f0040b8c388da04f23ecee6bfd4953a6512b4'
version = 0
validators = [
    '211dd8d62d6abb236a8ca72dc75489dcc6a79ce2',
    'c5b606cbd200b79f05a3ed4d92ce9581a0b8ec12',
]

[controller]
consensus_port = 50001
controller_port = 50004
executor_port = 50002
key_id = 1
kms_port = 50005
network_port = 50000
node_address = '211dd8d62d6abb236a8ca72dc75489dcc6a79ce2'
package_limit = 30000
storage_port = 50003

[kms_sm]
db_key = '123456'
kms_port = 50005
`, want2: `[network_tls]
ca_cert = """
ab
"""
cert = """
ab
"""
grpc_port = 50000
listen_port = 40000
priv_key = """
ef
"""
reconnect_timeout = 5

[[network_tls.peers]]
domain = 'www.my-node-2.com'
host = 'test-chain-my-node-2-cluster-ip'
port = 40000

[consensus_raft]
controller_port = 50004
grpc_listen_port = 50001
network_port = 50000
node_addr = '211dd8d62d6abb236a8ca72dc75489dcc6a79ce2'

[executor_evm]
executor_port = 50002

[storage_rocksdb]
kms_port = 50005
storage_port = 50003

[genesis_block]
prevhash = '0x0000000000000000000000000000000000000000000000000000000000000000'
timestamp = 1639105556777

[system_config]
admin = 'a3203aa7717ba46457d3cb373b10382adc763d2d'
block_interval = 3
block_limit = 100
chain_id = '63586a3c0255f337c77a777ff54f0040b8c388da04f23ecee6bfd4953a6512b4'
version = 0
validators = [
    'c5b606cbd200b79f05a3ed4d92ce9581a0b8ec12',
	'211dd8d62d6abb236a8ca72dc75489dcc6a79ce2',
]

[controller]
consensus_port = 50001
controller_port = 50004
executor_port = 50002
key_id = 1
kms_port = 50005
network_port = 50000
node_address = '211dd8d62d6abb236a8ca72dc75489dcc6a79ce2'
package_limit = 30000
storage_port = 50003

[kms_sm]
db_key = '123456'
kms_port = 50005
`})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cns := &ChainNodeService{
				ChainConfig:          tt.fields.ChainConfig,
				ChainNode:            tt.fields.ChainNode,
				Account:              tt.fields.Account,
				CaSecret:             tt.fields.CaSecret,
				NodeCertAndKeySecret: tt.fields.NodeCertAndKeySecret,
			}
			if got := cns.GenerateNodeConfig(); got == tt.want1 || got == tt.want2 {
			} else {
				t.Errorf("GenerateNodeConfig() = %v, \n\n\nwant1 %v, \n\n\nwant2 %v", got, tt.want1, tt.want2)
			}
		})
	}
}
