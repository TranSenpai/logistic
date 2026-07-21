# Tạo VPC 65,536 địa chỉ IP (IP v4 lấy 2 byte(16 bit đầu làm ))
resource "aws_vpc" "logistic_vpc" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "logistic-production-vpc"
  }
}

resource "aws_subnet" "logistic_public_subnet" {
  vpc_id                  = aws_vpc.logistic_vpc.id # Khai kháo subnet này thuộc vpc nào
  cidr_block              = "10.0.1.0/24"
  map_public_ip_on_launch = true
  availability_zone       = "ap-southeast-1a" # Đặt cố định ở Data Center 1a

  tags = {
    Name = "logistic-public-subnet-1a"
  }
}

resource "aws_internet_gateway" "logistic_igw" {
  vpc_id = aws_vpc.logistic_vpc.id

  tags = {
    Name = "logistic-igw"
  }
}

resource "aws_route_table" "logistic_public_rt" {
  vpc_id = aws_vpc.logistic_vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.logistic_igw.id
  }

  tags = {
    Name = "logistic-public-rt"
  }
}

resource "aws_route_table_association" "logistic_public_rta" {
  subnet_id      = aws_subnet.logistic_public_subnet.id
  route_table_id = aws_route_table.logistic_public_rt.id
}
