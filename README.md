# SecretHub Kubernetes Mutating Webhook

This mutating webhook allows you to use secret references (`secrethub://path/to/secret`) in any containers spec, without including SecretHub in the image itself.

It detects whether the container environment contains secret references (`secrethub://path/to/secret`) and when it does:
1. It creates a volume which will hold the SecretHub CLI.
1. It creates an init container which copies the SecretHub CLI into the volume.
1. It mounts the volume to the target container.
1. It prefixes the target containers command with `<path/to/volume>/secrethub run --`.

This project is based on and heavily inspired by [Berglas's Kubernetes Mutating Webhook](https://github.com/GoogleCloudPlatform/berglas/tree/v0.5.1/examples/kubernetes).

## Deploy the Webhook

The simplest method to deploy the webhook is in a serverless function. We've outlined the steps to take to [deploy the webhook to a Google Cloud Function](./deploy/gcloud-function/).
We're also [working on](https://github.com/secrethub/secrethub-kubernetes-mutating-webhook/pull/2) a way to deploy the webhook in the Kubernetes cluster itself.
