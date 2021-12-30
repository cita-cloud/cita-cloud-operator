package controllers

import (
	"fmt"

	citacloudv1 "github.com/cita-cloud/cita-cloud-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
)

type ChainNodeService struct {
	ChainConfig          *citacloudv1.ChainConfig
	ChainNode            *citacloudv1.ChainNode
	Account              *citacloudv1.Account
	CaSecret             *corev1.Secret
	NodeCertAndKeySecret *corev1.Secret
}

func NewChainNodeServiceForLog(chainConfig *citacloudv1.ChainConfig, chainNode *citacloudv1.ChainNode) *ChainNodeService {
	return &ChainNodeService{ChainConfig: chainConfig, ChainNode: chainNode}
}

func NewChainNodeServiceForP2P(chainConfig *citacloudv1.ChainConfig, chainNode *citacloudv1.ChainNode, account *citacloudv1.Account) *ChainNodeService {
	return &ChainNodeService{ChainConfig: chainConfig, ChainNode: chainNode, Account: account}
}

func NewChainNodeServiceForTls(chainConfig *citacloudv1.ChainConfig, chainNode *citacloudv1.ChainNode, account *citacloudv1.Account, caSecret, nodeCertAndKeySecret *corev1.Secret) *ChainNodeService {
	return &ChainNodeService{ChainConfig: chainConfig, ChainNode: chainNode, Account: account, CaSecret: caSecret, NodeCertAndKeySecret: nodeCertAndKeySecret}
}

type ChainNodeServiceImpl interface {
	GenerateNodeConfig() string
	GenerateControllerLogConfig() string
	GenerateExecutorLogConfig() string
	GenerateKmsLogConfig() string
	GenerateNetworkLogConfig() string
	GenerateStorageLogConfig() string
}

func (cns *ChainNodeService) GenerateControllerLogConfig() string {
	return fmt.Sprintf(`# Scan this file for changes every 30 seconds
refresh_rate: 30 seconds

appenders:
# An appender named "stdout" that writes to stdout
  stdout:
    kind: console

  journey-service:
    kind: rolling_file
    path: "logs/controller-service.log"
    policy:
      # Identifies which policy is to be used. If no kind is specified, it will
      # default to "compound".
      kind: compound
      # The remainder of the configuration is passed along to the policy's
      # deserializer, and will vary based on the kind of policy.
      trigger:
        kind: size
        limit: 50mb
      roller:
        kind: fixed_window
        base: 1
        count: 5
        pattern: "logs/controller-service.{}.gz"

# Set the default logging level and attach the default appender to the root
root:
  level: %s
  appenders:
    - stdout`, string(cns.ChainNode.Spec.LogLevel))
}

func (cns *ChainNodeService) GenerateExecutorLogConfig() string {
	return fmt.Sprintf(`# Scan this file for changes every 30 seconds
refresh_rate: 30 seconds

appenders:
# An appender named "stdout" that writes to stdout
  stdout:
    kind: console

  journey-service:
    kind: rolling_file
    path: "logs/executor-service.log"
    policy:
      # Identifies which policy is to be used. If no kind is specified, it will
      # default to "compound".
      kind: compound
      # The remainder of the configuration is passed along to the policy's
      # deserializer, and will vary based on the kind of policy.
      trigger:
        kind: size
        limit: 50mb
      roller:
        kind: fixed_window
        base: 1
        count: 5
        pattern: "logs/executor-service.{}.gz"

# Set the default logging level and attach the default appender to the root
root:
  level: %s
  appenders:
    - stdout`, string(cns.ChainNode.Spec.LogLevel))
}

func (cns *ChainNodeService) GenerateKmsLogConfig() string {
	return fmt.Sprintf(`# Scan this file for changes every 30 seconds
refresh_rate: 30 seconds

appenders:
# An appender named "stdout" that writes to stdout
  stdout:
    kind: console

  journey-service:
    kind: rolling_file
    path: "logs/kms-service.log"
    policy:
      # Identifies which policy is to be used. If no kind is specified, it will
      # default to "compound".
      kind: compound
      # The remainder of the configuration is passed along to the policy's
      # deserializer, and will vary based on the kind of policy.
      trigger:
        kind: size
        limit: 50mb
      roller:
        kind: fixed_window
        base: 1
        count: 5
        pattern: "logs/kms-service.{}.gz"

# Set the default logging level and attach the default appender to the root
root:
  level: %s
  appenders:
    - stdout`, string(cns.ChainNode.Spec.LogLevel))
}

func (cns *ChainNodeService) GenerateNetworkLogConfig() string {
	return fmt.Sprintf(`# Scan this file for changes every 30 seconds
refresh_rate: 30 seconds

appenders:
# An appender named "stdout" that writes to stdout
  stdout:
    kind: console

  journey-service:
    kind: rolling_file
    path: "logs/network-service.log"
    policy:
      # Identifies which policy is to be used. If no kind is specified, it will
      # default to "compound".
      kind: compound
      # The remainder of the configuration is passed along to the policy's
      # deserializer, and will vary based on the kind of policy.
      trigger:
        kind: size
        limit: 50mb
      roller:
        kind: fixed_window
        base: 1
        count: 5
        pattern: "logs/network-service.{}.gz"

# Set the default logging level and attach the default appender to the root
root:
  level: %s
  appenders:
    - stdout`, string(cns.ChainNode.Spec.LogLevel))
}

func (cns *ChainNodeService) GenerateStorageLogConfig() string {
	return fmt.Sprintf(`# Scan this file for changes every 30 seconds
refresh_rate: 30 seconds

appenders:
# An appender named "stdout" that writes to stdout
  stdout:
    kind: console

  journey-service:
    kind: rolling_file
    path: "logs/storage-service.log"
    policy:
      # Identifies which policy is to be used. If no kind is specified, it will
      # default to "compound".
      kind: compound
      # The remainder of the configuration is passed along to the policy's
      # deserializer, and will vary based on the kind of policy.
      trigger:
        kind: size
        limit: 50mb
      roller:
        kind: fixed_window
        base: 1
        count: 5
        pattern: "logs/storage-service.{}.gz"

# Set the default logging level and attach the default appender to the root
root:
  level: %s
  appenders:
    - stdout`, string(cns.ChainNode.Spec.LogLevel))
}

func (cns *ChainNodeService) GenerateNodeConfig() string {
	return cns.generateNetwork() + cns.generateConsensus() + cns.generateExecutor() +
		cns.generateStorage() + cns.generateBasic() + cns.generateController() + cns.generateKms()
}
func (cns *ChainNodeService) generateNetwork() string {
	if cns.ChainConfig.Spec.EnableTLS {
		return cns.generateNetworkTls()
	} else {
		return cns.generateNetworkP2P()
	}
}

func (cns *ChainNodeService) generateNetworkP2P() string {
	networkStr := `[network_p2p]
grpc_port = 50000
port = 40000

`

	nodeList := cns.ChainConfig.Status.NodeInfoMap
	for key, node := range nodeList {
		if key == cns.ChainNode.Name {
			// ignore with match name
			continue
		}
		if node.Cluster == cns.ChainNode.Spec.Cluster {
			// in the same k8s cluster
			networkStr = networkStr + fmt.Sprintf(`[[network_p2p.peers]]
address = '/dns4/%s/tcp/%d'

`, GetClusterIPName(cns.ChainConfig.Name, key), 40000)
		} else {
			networkStr = networkStr + fmt.Sprintf(`[[network_p2p.peers]]
address = '/dns4/%s/tcp/%d'

`, node.ExternalIp, node.Port)
		}
	}
	return networkStr
}

func (cns *ChainNodeService) generateNetworkTls() string {
	networkStr := fmt.Sprintf(`[network_tls]
ca_cert = """
%s
"""
cert = """
%s
"""
grpc_port = 50000
listen_port = 40000
priv_key = """
%s
"""
reconnect_timeout = 5

`, string(cns.CaSecret.Data[CaCert]), string(cns.NodeCertAndKeySecret.Data[NodeCert]), string(cns.NodeCertAndKeySecret.Data[NodeKey]))

	nodeList := cns.ChainConfig.Status.NodeInfoMap
	for key, node := range nodeList {
		if key == cns.ChainNode.Name {
			// ignore with match name
			continue
		}
		if node.Cluster == cns.ChainNode.Spec.Cluster {
			// in the same k8s cluster
			networkStr = networkStr + fmt.Sprintf(`[[network_tls.peers]]
domain = '%s'
host = '%s'
port = %d

`, node.Domain, GetClusterIPName(cns.ChainConfig.Name, key), 40000)
		} else {
			networkStr = networkStr + fmt.Sprintf(`[[network_tls.peers]]
domain = '%s'
host = '%s'
port = %d

`, node.Domain, node.ExternalIp, node.Port)
		}
	}
	return networkStr
}

func (cns *ChainNodeService) generateConsensus() string {
	consensusStr := fmt.Sprintf(`[consensus_raft]
controller_port = 50004
grpc_listen_port = 50001
network_port = 50000
node_addr = '%s'

`, cns.Account.Status.Address)
	return consensusStr
}

func (cns *ChainNodeService) generateExecutor() string {
	return `[executor_evm]
executor_port = 50002

`
}

func (cns *ChainNodeService) generateStorage() string {
	return `[storage_rocksdb]
kms_port = 50005
storage_port = 50003

`
}

func (cns *ChainNodeService) generateBasic() string {
	basicStr := fmt.Sprintf(`[genesis_block]
prevhash = '%s'
timestamp = %d

[system_config]
admin = '%s'
block_interval = %d
block_limit = %d
chain_id = '%s'
version = 0
validators = [
`, cns.ChainConfig.Spec.PrevHash, cns.ChainConfig.Spec.Timestamp, cns.ChainConfig.Status.AdminAccount.Address,
		cns.ChainConfig.Spec.BlockInterval, cns.ChainConfig.Spec.BlockLimit, cns.ChainConfig.Spec.Id)
	for _, validator := range cns.ChainConfig.Status.ValidatorAccountMap {
		s := fmt.Sprintf(`    '%s',
`, validator.Address)
		basicStr = basicStr + s
	}
	return basicStr + `]

`
}

func (cns *ChainNodeService) generateController() string {
	return fmt.Sprintf(`[controller]
consensus_port = 50001
controller_port = 50004
executor_port = 50002
key_id = 1
kms_port = 50005
network_port = 50000
node_address = '%s'
package_limit = 30000
storage_port = 50003

`, cns.Account.Status.Address)
}

func (cns *ChainNodeService) generateKms() string {
	return fmt.Sprintf(`[kms_sm]
db_key = '%s'
kms_port = 50005
`, cns.Account.Spec.KmsPassword)
}
