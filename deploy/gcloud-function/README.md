## Deploy in a Google Cloud Function

You can deploy the webhook to a Google cloud function using the following steps:

1. Clone this repository and make it your working directory:
```sh
git clone https://github.com/secrethub/secrethub-kubernetes-mutating-webhook.git && cd secrethub-kubernetes-mutating-webhook
```

2. Deploy the webhook to a Google Cloud Function:
```sh
gcloud functions deploy secrethub-mutating-webhook --runtime go113 --entry-point F --trigger-http
```

3. Set the Google Cloud Function URL in the deploy.yaml:
```sh
URL=$(gcloud functions describe secrethub-mutating-webhook --format 'value(httpsTrigger.url)') sed -i "s|YOUR_CLOUD_FUNCTION_URL|$URL|" deploy/gcloud-function/config.yaml
```

4. Enable the webhook on your Kubernetes cluster:
```sh
kubectl apply -f deploy/gcloud-function
```
