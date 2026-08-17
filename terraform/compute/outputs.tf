output "server_logistic_ip" {
  value = aws_instance.logistic_server.public_ip
}

output "security_group_id" {
  value = aws_security_group.logistic_sg.id
}
