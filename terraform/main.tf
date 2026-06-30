// Khai báo terraform sẽ làm việc với provider nào để terrform biết 
// mà cài các API call của provider đó. Trong case này là AWS và Cloudflare
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 4.0"
    }
  }
}

// Khai báo provider compute là AWS
provider "aws" {
  region     = "ap-southeast-1"
  access_key = var.aws_accesskey_id
  secret_key = var.aws_secret_access_key
}

// Khai báo provider network là Cloudflare
provider "cloudflare" {
  api_token = var.cloudflare_api_token
}

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
  instance_type = "t3.large"
  key_name      = "logistic-key"

  vpc_security_group_ids = [aws_security_group.logistic_sg.id]

  root_block_device {
    volume_size = 30
    volume_type = "gp3"
  }

  tags = {
    Name = "Logistic-Production-Node"
  }
}

resource "cloudflare_record" "api_endpoint" {
  zone_id = var.cloudflare_zone_id
  name    = "api"
  content = aws_instance.logistic_server.public_ip
  type    = "A"
  proxied = true
}
