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

// GetClusterIPName get node's clusterIP service name
func GetClusterIPName(nodeName string) string {
	return fmt.Sprintf("%s-cluster-ip", nodeName)
}

func GetCaSecretName(chainName string) string {
	return fmt.Sprintf("%s-ca-secret", chainName)
}

func GetAccountConfigmap(chainName, account string) string {
	return fmt.Sprintf("%s-%s", chainName, account)
}

func GetAccountCertAndKeySecretName(account string) string {
	return fmt.Sprintf("%s-cert-key", account)
}
