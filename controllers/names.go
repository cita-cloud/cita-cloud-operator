/*
 * Copyright Rivtower Technologies LLC.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

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
