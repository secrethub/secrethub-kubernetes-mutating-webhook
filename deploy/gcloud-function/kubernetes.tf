resource "kubernetes_mutating_webhook_configuration" "secrethub" {
  count = var.deploy_kubernetes_resource ? 1 : 0
  metadata {
    name = "secrethub-mutating-webhook"
    labels = {
      app  = "secrethub-mutating-webhook"
      kind = "mutator"
    }
  }
  webhook {
    name                      = "secrethub-mutating-webhook.default.svc.cluster.local"
    admission_review_versions = ["v1", "v1beta1"]
    failure_policy            = "Fail"
    side_effects              = "None"
    client_config {
      url = google_cloudfunctions_function.webhook.https_trigger_url
    }
    rule {
      api_groups   = [""]
      api_versions = ["v1"]
      operations   = ["CREATE"]
      resources    = ["pods"]
    }
  }
}
