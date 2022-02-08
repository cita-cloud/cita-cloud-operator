package controllers

import "fmt"

// GetNodeConfigName get node's config configmap name
func GetNodeConfigName(nodeName string) string {
	return fmt.Sprintf("%s-config", nodeName)
}

// GetLogConfigName get node's log config configmap name
func GetLogConfigName(nodeName string) string {
	return fmt.Sprintf("%s-log", nodeName)
}

// GetNodePortServiceName get node's clusterIP service name
func GetNodePortServiceName(nodeName string) string {
	return fmt.Sprintf("%s-nodeport", nodeName)
}

func GetCaSecretName(chainName string) string {
	return fmt.Sprintf("%s-ca-secret", chainName)
}

func GetAccountConfigmap(account string) string {
	return fmt.Sprintf("%s-account", account)
}

func GetAccountCertAndKeySecretName(account string) string {
	return fmt.Sprintf("%s-cert-key", account)
}
