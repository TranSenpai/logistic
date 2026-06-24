# Khai báo các biến như Region, Instance Type
variable "cloudflare_api_token" {
  description = "Cloudflare API Token"
  type = string
  # Biến này có chứa thông tin nhạy cảm và cần được ẩn trong giao diện người dùng hay không.
  sensitive = true
}

variable "cloudflare_zone_id" {
  description = "Cloudflare Zone ID cho domain glolog.dev"
  type = string
}
