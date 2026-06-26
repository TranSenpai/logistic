terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.92"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 4.0"
    }
  }
  required_version = ">= 1.2"
}

provider "aws" {
  region = "ap-southeast-1"
}

provider "cloudflare" {
  api_token = var.cloudflare_api_token
}

data "aws_ami" "Ubuntu" {
  most_recent = true
  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-*"]
  }
  owners = ["099720109477"]
}

resource "aws_security_group" "lab_sg" {
  name        = "lab-security-group"
  description = "Allow SSH, HTTP, and HTTPS"

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

  ingress {
    from_port   = 443
    to_port     = 443
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
  ami           = data.aws_ami.Ubuntu.id
  instance_type = "t3.large"

  vpc_security_group_ids = [aws_security_group.lab_sg.id]

  key_name = "logistic-key"

  root_block_device {
    volume_size = 30
    volume_type = "gp3"
  }

  tags = {
    Name = "Logistic-Lab"
  }
}

resource "cloudflare_record" "api_subdomain" {
  zone_id = var.cloudflare_zone_id
  name    = "api"
  content = aws_instance.logistic_server.public_ip
  type    = "A"
  proxied = true
}
