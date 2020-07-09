# SecretHub Kubernetes Mutating Webhook

[![GoDoc](https://godoc.org/github.com/secrethub/secrethub-kubernetes-mutating-webhook?status.svg)](http://godoc.org/github.com/secrethub/secrethub-kubernetes-mutating-webhook)
[![CircleCI](https://circleci.com/gh/secrethub/secrethub-kubernetes-mutating-webhook.svg?style=shield)](https://circleci.com/gh/secrethub/secrethub-kubernetes-mutating-webhook)
[![Go Report Card](https://goreportcard.com/badge/github.com/secrethub/secrethub-kubernetes-mutating-webhook)](https://goreportcard.com/report/github.com/secrethub/secrethub-kubernetes-mutating-webhook)
[![Version]( https://img.shields.io/github/release/secrethub/secrethub-kubernetes-mutating-webhook.svg)](https://github.com/secrethub/secrethub-kubernetes-mutating-webhook/releases/latest)
[![Discord](https://img.shields.io/badge/chat-on%20discord-7289da.svg?logo=discord)](https://discord.gg/5M2Fm6T)

This mutating webhook allows you to use secret references (`secrethub://path/to/secret`) in any containers spec, without including SecretHub in the image itself:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: my-app
  annotations:
    secrethub.io/mutate: my-app
spec:
  containers:
    - name: my-app
      image: my-image
      env:
        - name: STRIPE_SECRET_KEY
          value: secrethub://acme/app/prod/stripe/secret_key
        - name: PGPASSWORD
          value: secrethub://acme/app/prod/pg/password
```

You can annotate your pod spec with `secrethub.io/mutate` which expects a comma separated list of the names of the containers to mutate.

When the annotation is found, this is what will happen:
- A volume that will hold the SecretHub CLI is created.
- An init container which copies the SecretHub CLI into the volume is created.

And for every container you've specified in the `secrethub.io/mutate` annotation:
- The volume is mounted onto the container.
- The command is prefixed with `<path/to/volume>/secrethub run --`.

The version of the SecretHub CLI Docker image to be used can optionally be configured with `secrethub.io/version`, e.g. `secrethub.io/version: 0.39.0`. If isn't set, the `latest` version is used. A list of available versions can be found [here](https://hub.docker.com/repository/docker/secrethub/cli/tags).

The image that holds the CLI binary to copy is hosted on [DockerHub](https://hub.docker.com/repository/docker/secrethub/cli).
If you prefer to host your own image to source the binary from, you can use the `secrethub.io/imageOverride` annotation.
If it's set, the value of `secrethub.io/version` will get ignored.
Make sure that the image you use has the binary available at `/usr/bin/secrethub`.

## Attributions

This project is based on and heavily inspired by [Berglas's Kubernetes Mutating Webhook](https://github.com/GoogleCloudPlatform/berglas/tree/v0.5.1/examples/kubernetes).

## Deploy the Webhook

The simplest method to deploy the webhook is in a serverless function. We've outlined the steps to take to [deploy the webhook to a Google Cloud Function](./deploy/gcloud-function/).
We're also [working on](https://github.com/secrethub/secrethub-kubernetes-mutating-webhook/pull/2) a way to deploy the webhook in the Kubernetes cluster itself.
