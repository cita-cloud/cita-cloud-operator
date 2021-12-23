package controllers

import "fmt"

// get admin account configmap name
func GetAdminAccountName(chainName string) string {
	return fmt.Sprintf("%s-admin", chainName)
}

// get node's account configmap name
func GetNodeAccountName(chainName string, nodeName string) string {
	return fmt.Sprintf("%s-%s-account", chainName, nodeName)
}

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

func GetNodeCertAndKeySecretName(chainName string, nodeName string) string {
	return fmt.Sprintf("%s-%s-cert-key", chainName, nodeName)
}
