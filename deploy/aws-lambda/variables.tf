variable "deploy_kubernetes_resource" {
  description = "Whether to also create the Kubernetes resource right away. Requires the Kubernetes provider to be configured in the root module."
  default     = true
}
