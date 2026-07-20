variable "cloudflare_api_token" {
  description = "Cloudflare API Token"
  type        = string
  sensitive   = true
}

variable "cloudflare_zone_id" {
  description = "Cloudflare ZoneID for Domain glolog.dev"
  type        = string
}

variable "aws_accesskey_id" {
  description = "AWS access key ID"
  type        = string
}

variable "aws_secret_access_key" {
  description = "AWS secret key"
  type        = string
}

variable "aws_region" {
  description = "Virtual compute region"
  type        = string
}
