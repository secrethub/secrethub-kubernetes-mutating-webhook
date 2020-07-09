package webhook

import corev1 "k8s.io/api/core/v1"

const version = "0.1.0"

var appInfo = []corev1.EnvVar{
	{
		Name:  "SECRETHUB_APP_INFO_NAME",
		Value: "kubernetes-mutating-webhook",
	},
	{
		Name:  "SECRETHUB_APP_INFO_VERSION",
		Value: version,
	},
}
