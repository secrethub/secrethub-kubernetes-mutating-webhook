variable "webhook_version" {
  description = "Version of the SecretHub webhook to deploy."

  # Default version will get bumped automatically on release PRs,
  # so updating the module will also update the webhook version
  # if the module doesn't override the variable.
  default = "0.2.0"
}
