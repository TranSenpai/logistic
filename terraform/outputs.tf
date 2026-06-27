# Dùng để in ra IP của server sau khi tạo thành công

output "server_ip_de_ssh" {
  value = aws_instance.logistic_server.public_ip
}
