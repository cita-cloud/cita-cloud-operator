package controllers

import "fmt"

// get node's config configmap name
func GetNodeConfigName(chainName string, nodeName string) string {
	return fmt.Sprintf("%s-%s-config", chainName, nodeName)
}

// get node's log config configmap name
func GetLogConfigName(chainName string, nodeName string) string {
	return fmt.Sprintf("%s-%s-log", chainName, nodeName)
}

// get node's clusterIP service name
func GetClusterIPName(chainName string, nodeName string) string {
	return fmt.Sprintf("%s-%s-cluster-ip", chainName, nodeName)
}

func GetCaSecretName(chainName string) string {
	return fmt.Sprintf("%s-ca-secret", chainName)
}

func GetAccountConfigmap(chainName, account string) string {
	return fmt.Sprintf("%s-%s", chainName, account)
}

func GetAccountCertAndKeySecretName(chainName, account string) string {
	return fmt.Sprintf("%s-%s-cert-key", chainName, account)
}
