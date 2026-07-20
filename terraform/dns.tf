resource "cloudflare_record" "api_endpoint" {
  zone_id = var.cloudflare_zone_id
  name    = "api"
  content = aws_instance.logistic_server.public_ip
  type    = "A"
  proxied = true
}
