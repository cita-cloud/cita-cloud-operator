package controllers

import (
	"fmt"

	citacloudv1 "github.com/cita-cloud/cita-cloud-operator/api/v1"
)

type ChainNodeService struct {
	ChainConfig *citacloudv1.ChainConfig
	ChainNode   *citacloudv1.ChainNode
}

func NewChainNodeService(chainConfig *citacloudv1.ChainConfig, chainNode *citacloudv1.ChainNode) *ChainNodeService {
	return &ChainNodeService{ChainConfig: chainConfig, ChainNode: chainNode}
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
    - journey-service`, string(cns.ChainNode.Spec.LogLevel))
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
    - journey-service`, string(cns.ChainNode.Spec.LogLevel))
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
    - journey-service`, string(cns.ChainNode.Spec.LogLevel))
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
    - journey-service`, string(cns.ChainNode.Spec.LogLevel))
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
    - journey-service`, string(cns.ChainNode.Spec.LogLevel))
}

func (cns *ChainNodeService) GenerateNodeConfig() string {
	return cns.generateNetwork() + cns.generateConsensus() + cns.generateExecutor() +
		cns.generateStorage() + cns.generateBasic() + cns.generateController() + cns.generateKms()
}

func (cns *ChainNodeService) generateNetwork() string {
	networkStr := `[network_p2p]
grpc_port = 50000
port = 40000

`

	nodeList := cns.ChainConfig.Status.NodeInfoMap
	for _, node := range nodeList {
		if node.Address == cns.ChainNode.Spec.Address {
			continue
		}
		networkStr = networkStr + fmt.Sprintf(`[[network_p2p.peers]]
address = '/dns4/%s/tcp/%d'

`, cns.ChainNode.Spec.InternalIp, cns.ChainNode.Spec.Port)
	}
	return networkStr
}

func (cns *ChainNodeService) generateConsensus() string {
	consensusStr := fmt.Sprintf(`[consensus_raft]
controller_port = 50004
grpc_listen_port = 50001
network_port = 50000
node_addr = '%s'

`, cns.ChainNode.Spec.Address)
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
`, cns.ChainConfig.Spec.PrevHash, cns.ChainConfig.Spec.Timestamp, cns.ChainConfig.Spec.AdminAddress,
		cns.ChainConfig.Spec.BlockInterval, cns.ChainConfig.Spec.BlockLimit, cns.ChainConfig.Spec.Id)
	for _, validator := range cns.ChainConfig.Spec.Validators {
		s := fmt.Sprintf(`    '%s',
`, validator)
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

`, cns.ChainNode.Spec.Address)
}

func (cns *ChainNodeService) generateKms() string {
	return fmt.Sprintf(`[kms_sm]
db_key = '%s'
kms_port = 50005
`, cns.ChainNode.Spec.KmsPassword)
}
