# Deploy in a Google Cloud Function

You can deploy the webhook to Google Cloud Function and activate it in your Kubernetes cluster by adding the following module to your Terraform project: 

```terraform
module "secrethub_mutating_webhook" {
  source = "github.com/secrethub/secrethub-kubernetes-mutating-webhook//deploy/gcloud-function?ref=v0.2.0"
}
```

This module requires the [Google Cloud provider](https://registry.terraform.io/providers/hashicorp/google/latest/docs) and [Kubernetes provider](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs) to be configured in your root project.
