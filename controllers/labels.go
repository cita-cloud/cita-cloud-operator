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
