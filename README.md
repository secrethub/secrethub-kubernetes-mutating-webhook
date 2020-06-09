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

When the annotation is found:
- A volume which will hold the SecretHub CLI is created.
- An init container which copies the SecretHub CLI into the volume is created.

And for every container that is listed in the `secrethub.io/mutate` annotation:
- The volume is mounted to the container.
- The command is prefixed with `<path/to/volume>/secrethub run --`.

The version of the SecretHub CLI Docker image to be used can optionally be configured with `secrethub.io/version`, e.g. `secrethub.io/version: 0.39.0`. If it is not set, the `latest` version is used. A list of available versions can be found [here](https://hub.docker.com/repository/docker/secrethub/cli/tags).

## Attributions

This project is based on and heavily inspired by [Berglas's Kubernetes Mutating Webhook](https://github.com/GoogleCloudPlatform/berglas/tree/v0.5.1/examples/kubernetes).
