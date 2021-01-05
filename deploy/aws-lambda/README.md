## Deploy in a AWS Lambda Function

You can deploy the webhook to a AWS Lambda function in two ways:
1. by using the provided Terraform module (with or without cloning the repository)

2. by manually making the serversless application with AWS Lambda and API Gateway

### Deploy using the Terraform module

A. Without cloning the repository

1. Add the module to your terraform file:
```terraform
module "secrethub_mutating_webhook" {
  source = "github.com/secrethub/secrethub-kubernetes-mutating-webhook/deploy/aws-lambda"
}
```

2. When your file is ready, run `terraform apply` and you will have a functioning webhook with AWS Lambda.

B. With cloning the repository

1. Clone this repository and go to `deploy/aws-lambda` directory:
```sh
git clone https://github.com/secrethub/secrethub-kubernetes-mutating-webhook.git && cd secrethub-kubernetes-mutating-webhook/deploy/aws-lambda
```

2. Initialize the Terraform module and then apply it:
```sh
terraform init && terraform apply
```

3. Copy the link that is provided as an output from previous step and set the AWS Lambda Function URL in the `config.yaml`:
```sh
URL="<your-resulted-URL>" sed -i "s|YOUR_AWS_API_GATEWAY_URL|$URL|" deploy/aws-lambda/config.yaml
```

4. Enable the webhook on your Kubernetes cluster:
```sh
kubectl apply -f deploy/aws-lambda
```

### Deploy by manually making the serveless application with AWS Lambda and API Gateway

1. Clone this repository and make it your working directory:
```sh
git clone https://github.com/secrethub/secrethub-kubernetes-mutating-webhook.git && cd secrethub-kubernetes-mutating-webhook
```

2. Create an IAM role that will be used for the lambda function:
```sh
aws iam create-role --role-name secrethub-webhook-role --assume-role-policy-document '{"Version": "2012-10-17","Statement": [{ "Effect": "Allow", "Principal": {"Service": "lambda.amazonaws.com"}, "Action": "sts:AssumeRole"}]}'
```

3. Create the deployment package

    a. Make the binary that will be used for the deployment package:
    ```sh
    go build -o lambda-webhook cmd/lambda
    ```

    b. Create a deployment package:
    ```sh
    zip lambda-webhook.zip lambda-webhook
    ```

4. Create the Lambda function. Replace `123456789012` in the role ARN with your account ID.
```sh
aws lambda create-function --function-name secrethub-mutating-webhook \
--zip-file fileb://lambda-webhook.zip --handler lambda-webhook --runtime go1.x \
--role arn:aws:iam::123456789012:role/secrethub-webhook-role
```

Note your Lambda function ARN. you will need it at a later step.

5. Set up a Lambda proxy integration using API Gateway

    a. Create a Rest API:
    ```sh
    aws apigateway create-rest-api --name 'SecretHubWebhookApi'
    ```

    Note the resulting API's `id` value. It will be needed for the next steps.

    b. Get the root resource `id`:
    ```sh
    aws apigateway get-resources --rest-api-id <your-API-id>
    ```

    Note the resulting root resource's `id` value. It will be needed for the next steps as well.

    c. Create an API Gateway Resource:
    ```sh
    aws apigateway create-resource --rest-api-id <your-API-id> \
      --parent-id <your-root-resource-id> \
      --path-part {proxy+}
    ``` 
    Note the resulting `{proxy+}` resource's `id` value. It will be needed for the next step.

    d. Create an `ANY` method request for the proxy:
    ```sh
    aws apigateway put-method --rest-api-id <your-API-id> \
       --resource-id <your-proxy-resource-id> \
       --http-method ANY \
       --authorization-type "NONE" 
    ```

    e. Set up the integration of the method created in the previous step. For this step you need the ARN of the Lambda function created in step 3. Don't forget to replace `123456789012` with your account ID.
    ```sh
    aws apigateway put-integration \
        --rest-api-id <your-API-id \
        --resource-id <your-proxy-resource-id> \
        --http-method POST \
        --type AWS_PROXY \
        --integration-http-method POST \
        --uri arn:aws:apigateway:<your-region>:lambda:path/2015-03-31/functions/<your-Lambda-function-ARN>/invocations \
        --credentials arn:aws:iam::123456789012:role/apigAwsProxyRole
    ```

    f. Now you will need to perfrom the previous 2 steps for the root as well:
    ```sh
    aws apigateway put-method --rest-api-id <your-API-id> \
       --resource-id <your-root-resource-id> \
       --http-method ANY \
       --authorization-type "NONE"
    ```
    ```sh
    aws apigateway put-integration \
        --rest-api-id <your-API-id \
        --resource-id <your-root-resource-id> \
        --http-method POST \
        --type AWS_PROXY \
        --integration-http-method POST \
        --uri arn:aws:apigateway:us-east-1:lambda:path/2015-03-31/functions/<your-Lambda-function-ARN>/invocations \
        --credentials arn:aws:iam::123456789012:role/apigAwsProxyRole
    ```

    f. Deploy the API:
    ```sh
    aws apigateway create-deployment --rest-api-id <your-API-id> --stage-name v1
    ```

> The function is configured to allow unauthenticated requests. The function doesn't give access to any resources or data. It only allows you to mutate provided data.

6. Set the AWS Lambda Function URL in the `config.yaml`:
```sh
URL="https://<your-API-id>.execute-api.$(aws configure get default.region).amazonaws.com/v1" sed -i "s|YOUR_AWS_API_GATEWAY_URL|$URL|" deploy/aws-lambda/config.yaml
```

7. Enable the webhook on your Kubernetes cluster:
```sh
kubectl apply -f deploy/aws-lambda
```
