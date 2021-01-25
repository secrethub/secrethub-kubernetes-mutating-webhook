data "github_release" "webhook" {
  repository  = "secrethub-kubernetes-mutating-webhook"
  owner       = "secrethub"
  retrieve_by = "tag"
  release_tag = var.secrethub_mutating_webhook_version == "latest" ? "latest" : "v${var.secrethub_mutating_webhook_version}"
}

locals {
  zip_url = "https://github.com/secrethub/secrethub-kubernetes-mutating-webhook/releases/download/${data.github_release.webhook.release_tag}/secrethub-kubernetes-mutating-webhook-${data.github_release.webhook.release_tag}-lambda.zip"
}

resource "null_resource" "webhook_zip_download" {
  triggers = {
    tag = local.zip_url
  }

  provisioner "local-exec" {
    command = "curl -L -o lambda-webhook.zip ${local.zip_url}"
  }

  provisioner "local-exec" {
    when    = destroy
    command = "rm ./lambda-webhook.zip"
  }
}

resource "aws_lambda_function" "webhook" {
  function_name    = "SecretHubWebhook"
  filename         = "lambda-webhook.zip"
  handler          = "lambda-webhook"
  source_code_hash = filebase64sha256("lambda-webhook.zip")
  role             = aws_iam_role.iam_for_lambda.arn
  runtime          = "go1.x"
  memory_size      = 128
  timeout          = 10
  depends_on       = [null_resource.webhook_zip_download]
}

resource "aws_iam_role" "iam_for_lambda" {
  name               = "SecretHubKubernetesWebhookRole"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json
  description        = "SecretHub Kubernetes Webhook"
}

data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_api_gateway_rest_api" "api" {
  name = "SecretHubKubernetesWebhookAPI"
}

resource "aws_api_gateway_resource" "resource" {
  path_part   = "{proxy+}"
  parent_id   = aws_api_gateway_rest_api.api.root_resource_id
  rest_api_id = aws_api_gateway_rest_api.api.id
}

resource "aws_api_gateway_method" "method" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  resource_id   = aws_api_gateway_resource.resource.id
  http_method   = "ANY"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "lambda_integration" {
  rest_api_id             = aws_api_gateway_rest_api.api.id
  resource_id             = aws_api_gateway_resource.resource.id
  http_method             = aws_api_gateway_method.method.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.webhook.invoke_arn
}

resource "aws_api_gateway_method" "proxy_root" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  resource_id   = aws_api_gateway_rest_api.api.root_resource_id
  http_method   = "ANY"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "lambda_root" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_method.proxy_root.resource_id
  http_method = aws_api_gateway_method.proxy_root.http_method

  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.webhook.invoke_arn
}

resource "aws_lambda_permission" "apigw_lambda" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.webhook.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_api_gateway_rest_api.api.execution_arn}/*/*"
}

resource "aws_api_gateway_deployment" "webhook" {
  depends_on = [
    aws_api_gateway_integration.lambda_integration,
    aws_api_gateway_integration.lambda_root,
  ]

  rest_api_id = aws_api_gateway_rest_api.api.id
  stage_name  = "v1"
}
