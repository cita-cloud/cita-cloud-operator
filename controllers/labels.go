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

func LabelsForChain(name string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/component":  "cita-cloud",
		"app.kubernetes.io/chain-name": name,
	}
}

func LabelsForNode(chainName, nodeName string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/chain-name": chainName,
		"app.kubernetes.io/chain-node": nodeName,
	}
}

// MergeLabels merges all labels together and returns a new label.
func MergeLabels(allLabels ...map[string]string) map[string]string {
	lb := make(map[string]string)

	for _, label := range allLabels {
		for k, v := range label {
			lb[k] = v
		}
	}

	return lb
}
