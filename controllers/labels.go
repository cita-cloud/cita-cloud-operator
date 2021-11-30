package controllers

func labelsForChain(name string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":     "cita-cloud",
		"app.kubernetes.io/instance": name,
	}
}
