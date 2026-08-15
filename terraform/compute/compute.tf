// Tạo 1 instance type ami của provider aws đặt tên local là ubuntu
data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"]
  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-*"]

  }
}

data "terraform_remote_state" "network_logistic" {
  backend = "s3"
  config = {
    bucket = "chuong-logistic-bucket"
    key    = "logistic/dev/network/terraform.tfstate"
    region = var.aws_region
  }
}

// Tạo 1 security group của aws đặt tên local là logistic_sg
resource "aws_security_group" "logistic_sg" {
  name        = "logistic-security-group"
  description = "Securiry rules for Logistic application"
  vpc_id      = data.terraform_remote_state.network_logistic.outputs.vpc_id

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
  subnet_id     = data.terraform_remote_state.network_logistic.outputs.subnet_id

  vpc_security_group_ids = [aws_security_group.logistic_sg.id]

  root_block_device {
    volume_size = 30
    volume_type = "gp3"
  }

  tags = {
    Name = "Logistic-Production-Node"
  }

  # Ép Terraform tự động xóa EC2 cũ và tạo EC2 mới mỗi khi đoạn script user_data bên dưới bị sửa
  user_data_replace_on_change = true

  user_data = <<-EOF
              #!/bin/bash
              # 1. Ghi một script vào thư mục per-boot. Tất cả script ở đây sẽ TỰ ĐỘNG CHẠY MỖI KHI MÁY KHỞI ĐỘNG (Bất kể là bật lần đầu hay Stop/Start lại).
              cat << 'SCRIPT' > /var/lib/cloud/scripts/per-boot/01-setup-podman.sh
              #!/bin/bash
              
              # Kiểm tra xem Podman đã được cài chưa?
              if ! command -v podman &> /dev/null; then
                  echo "Podman is not install. Starting install Podman ..."
                  
                  # Gỡ sạch Docker cũ để tránh xung đột (nếu có)
                  apt-get purge -y docker-engine docker docker.io docker-ce docker-ce-cli
                  apt-get autoremove -y --purge
                  rm -rf /var/lib/docker /var/lib/containerd /etc/docker /var/run/docker.sock
                  
                  # Cài đặt Podman
                  apt-get update -y
                  apt-get install -y podman podman-compose
                  echo "Install Podman successfully!"
              else
                  echo "Podman is installed !"
              fi
              SCRIPT

              # 2. Cấp quyền thực thi cho script vừa tạo
              chmod +x /var/lib/cloud/scripts/per-boot/01-setup-podman.sh

              # 3. Kích hoạt chạy script luôn cho lần boot đầu tiên này
              /var/lib/cloud/scripts/per-boot/01-setup-podman.sh
              EOF

}
