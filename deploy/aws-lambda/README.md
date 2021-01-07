## Deploy in a AWS Lambda Function

You can deploy the webhook to an AWS Lambda function and activate it in your Kubernetes cluster by adding the following module to your Terraform file: 

```terraform
module "secrethub_mutating_webhook" {
  source = "github.com/secrethub/secrethub-kubernetes-mutating-webhook/deploy/aws-lambda"
}
```
