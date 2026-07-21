// Tạo 1 instance type ami của provider aws đặt tên local là ubuntu
data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"]
  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-*"]

  }
}

// Tạo 1 security group của aws đặt tên local là logistic_sg
resource "aws_security_group" "logistic_sg" {
  name        = "logistic-security-group"
  description = "Securiry rules for Logistic application"
  vpc_id      = aws_vpc.logistic_vpc.id

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "logistic_server" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = var.aws_instance_type
  key_name      = var.aws_logistic_key
  subnet_id     = aws_subnet.logistic_public_subnet.id

  vpc_security_group_ids = [aws_security_group.logistic_sg.id]

  root_block_device {
    volume_size = 30
    volume_type = "gp3"
  }

  tags = {
    Name = "Logistic-Production-Node"
  }
}
