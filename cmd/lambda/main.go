package main

import (
	"github.com/akrylysov/algnhsa"
	webhook "github.com/secrethub/secrethub-kubernetes-mutating-webhook"
)

func main() {
	algnhsa.ListenAndServe(webhook.Handler(), nil)
}
