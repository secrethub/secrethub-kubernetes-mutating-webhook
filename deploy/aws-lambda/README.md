# Deploy to AWS Lambda

You can deploy the webhook to AWS Lambda and activate it in your Kubernetes cluster by adding the following module to your Terraform project: 

```terraform
module "secrethub_mutating_webhook" {
  source = "github.com/secrethub/secrethub-kubernetes-mutating-webhook?ref=v0.2.0/deploy/aws-lambda"
}
```

This module requires the [AWS provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs) and [Kubernetes provider](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs) to be configured in your root project.
