output "vpc_id" {
  value = aws_vpc.logistic_vpc.id
}

output "subnet_id" {
  value = aws_subnet.logistic_public_subnet.id
}
