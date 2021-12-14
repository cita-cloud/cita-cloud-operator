package controllers

const (
	AccountVolumeName         = "account"
	AccountVolumeMountPath    = "/mnt"
	NodeConfigVolumeName      = "node-config"
	NodeConfigVolumeMountPath = "/data"
	DataVolumeName            = "datadir"
	DataVolumeMountPath       = "/data"

	NodeConfigFile          = "config.toml"
	ControllerLogConfigFile = "controller-log4rs.yaml"
	ExecutorLogConfigFile   = "executor-log4rs.yaml"
	KmsLogConfigFile        = "kms-log4rs.yaml"
	NetworkLogConfigFile    = "network-log4rs.yaml"
	StorageLogConfigFile    = "storage-log4rs.yaml"

	NetworkPort = 40000
)

//var NetworkPort int32
