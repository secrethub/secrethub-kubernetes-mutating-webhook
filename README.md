# SecretHub Kubernetes Mutating Webhook

This mutating webhook allows you to use secret references (`secrethub://path/to/secret`) in any containers spec, without including SecretHub in the image itself.

It detects whether the container environment contains secret references (`secrethub://path/to/secret`) and when it does:
1. It creates a volume which will hold the SecretHub CLI.
1. It creates an init container which copies the SecretHub CLI into the volume.
1. It mounts the volume to the target container.
1. It prefixes the target containers command with `<path/to/volume>/secrethub run --`.

This project is based on and heavily inspired by [Berglas's Kubernetes Mutating Webhook](https://github.com/GoogleCloudPlatform/berglas/tree/v0.5.1/examples/kubernetes).

## Deploy the Webhook

The simplest method to deploy the webhook is in a serverless function. Below we outline the steps to take to deploy the webhook to a Google Cloud Function.
We're also [working on](https://github.com/secrethub/secrethub-kubernetes-mutating-webhook/pull/2) a way to deploy the webhook in the Kubernetes cluster itself.

You can deploy the webhook to a Google cloud function using the following steps:

1. Deploy the webhook to a Google Cloud Function:
```sh
gcloud functions deploy secrethub-mutating-webhook --runtime go113 --entry-point F --trigger-http
```

2. Set the Google Cloud Function URL in the deploy.yaml:
```sh
URL=$(gcloud functions describe secrethub-mutating-webhook --format 'value(httpsTrigger.url)') sed -i "s|YOUR_CLOUD_FUNCTION_URL|$URL|" deploy.yaml
```

3. Enable the webhook on your Kubernetes cluster:
```sh
kubectl apply -f deploy.yaml
```
