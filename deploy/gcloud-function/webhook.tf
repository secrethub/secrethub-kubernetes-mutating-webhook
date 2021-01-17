data "github_release" "webhook" {
  repository  = "secrethub-kubernetes-mutating-webhook"
  owner       = "secrethub"
  retrieve_by = "tag"
  release_tag = var.secrethub_mutating_webhook_version
}

locals {
  zip_url = "https://github.com/secrethub/secrethub-kubernetes-mutating-webhook/releases/download/${data.github_release.webhook.release_tag}/secrethub-kubernetes-mutating-webhook-${data.github_release.webhook.release_tag}-gcloud.zip"
}

resource "null_resource" "webhook_zip_download" {
  triggers = {
    tag = local.zip_url
  }

  provisioner "local-exec" {
    command = "curl -L -o gcp-webhook.zip ${local.zip_url}"
  }

  provisioner "local-exec" {
    when    = destroy
    command = "rm ./gcp-webhook.zip"
  }
}

module "project_id" {
  source  = "matti/resource/shell"
  command = "gcloud config list --format 'value(core.project)'"
}

module "region" {
  source  = "matti/resource/shell"
  command = "gcloud config list --format 'value(compute.zone)'"
}

resource "google_cloudfunctions_function" "webhook" {
  name                  = "secrethub-mutating-webhook"
  entry_point           = "F"
  runtime               = "go113"
  available_memory_mb   = 128
  timeout               = 61
  project               = module.project_id.stdout
  region                = substr(module.region.stdout, 0, length(module.region.stdout) - 2)
  trigger_http          = true
  source_archive_bucket = google_storage_bucket.bucket.name
  source_archive_object = google_storage_bucket_object.archive.name
  labels = {
    deployment_name = "secrethub-mutating-webhook"
  }
}

resource "google_cloudfunctions_function_iam_member" "invoker" {
  project        = google_cloudfunctions_function.webhook.project
  region         = google_cloudfunctions_function.webhook.region
  cloud_function = google_cloudfunctions_function.webhook.name

  role   = "roles/cloudfunctions.invoker"
  member = "allUsers"
}

resource "google_storage_bucket" "bucket" {
  name    = "cloudfunction-deploy-secrethub-mutating-webhook"
  project = module.project_id.stdout
}

resource "google_storage_bucket_object" "archive" {
  name       = "gcp-webhook.zip"
  bucket     = google_storage_bucket.bucket.name
  source     = "gcp-webhook.zip"
  depends_on = [null_resource.webhook_zip_download]
}
