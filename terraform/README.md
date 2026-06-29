# HƯỚNG DẪN TRIỂN KHAI HẠ TẦNG VỚI TERRAFORM (INFRASTRUCTURE AS CODE)

Tài liệu này định nghĩa kiến trúc hệ thống, quy trình khai báo tài nguyên và cú pháp chuẩn (HCL) để cấp phát hạ tầng cho dự án Logistics OS. Tài liệu sử dụng các khái niệm tổng quát (Agnostic Cloud) và ví dụ cụ thể trên AWS/Cloudflare.

---

## 1. Kiến trúc Hệ thống (System Architecture)
- **Compute Provider**: Nền tảng cung cấp máy ảo (Ví dụ: AWS EC2, GCP Compute Engine, Azure VM).
- **Network & Security**: Hạ tầng mạng (VPC), Tường lửa (Firewall / Security Group).
- **DNS & Proxy Provider**: Nền tảng phân giải tên miền và bảo mật (Ví dụ: Cloudflare DNS, WAF).
- **Network Topology**: Client -> Cloud Proxy (Port 443/80) -> Public Subnet (Port 80) -> Reverse Proxy nội bộ -> App Containers.

---

## 2. Nguyên lý Đồ thị Phụ thuộc (Dependency Graph)
Terraform quản lý vòng đời tài nguyên dựa trên tính phụ thuộc. Trình tự khai báo logic yêu cầu tài nguyên độc lập phải được khởi tạo trước tài nguyên phụ thuộc:
1. **Provider**: Module thiết yếu để tương tác với Cloud API.
2. **Data Source (OS Image)**: Cung cấp metadata về hệ điều hành gốc.
3. **Firewall (Security Group)**: Tài nguyên mạng cơ sở, kiểm soát luồng giao thông.
4. **Virtual Machine (Compute Instance)**: Máy ảo. Phụ thuộc trực tiếp vào OS Image, SSH Key Pair và Firewall ID.
5. **DNS Record**: Phụ thuộc vào Public IP được cấp phát ngẫu nhiên từ Virtual Machine.

---

## 3. Cú pháp và Khai báo Tài nguyên (Resource Declaration)

### 3.1. Cấu hình Cốt lõi (Provider & Backend Configuration)
```hcl
terraform {
  required_providers {
    aws = { source = "hashicorp/aws", version = "~> 5.0" }
    cloudflare = { source = "cloudflare/cloudflare", version = "~> 4.0" }
  }
}

provider "aws" { region = "ap-southeast-1" }
provider "cloudflare" { api_token = var.cloudflare_api_token }
```
**Mục đích của Block này:**
Khai báo cho Terraform biết mã nguồn này sẽ làm việc với những nền tảng Cloud nào, tải thư viện lõi phiên bản bao nhiêu, và sử dụng thông tin xác thực (API Token) nào để đăng nhập vào hệ thống.

**Phân tích các dòng lệnh (Arguments):**
- `terraform { required_providers { ... } }`: Nơi khóa (lock) phiên bản thư viện của AWS và Cloudflare. Dấu `~>` nghĩa là cho phép tự động cập nhật các bản vá lỗi nhỏ (ví dụ từ 5.0.1 lên 5.0.2) nhưng không tự động nhảy lên phiên bản lớn (6.0) để tránh lỗi không tương thích ngược.
- `provider "aws" { region = "ap-southeast-1" }`: Chỉ định Terraform sẽ thao tác với Data Center AWS đặt tại Singapore. 
- `provider "cloudflare" { api_token = ... }`: Khởi tạo phiên làm việc với Cloudflare. Tham số `api_token` đọc giá trị từ biến bên ngoài để không bị lộ secret vào mã nguồn. Đối với AWS, Terraform tự động dò tìm credential từ file config hệ thống nên không cần truyền tay vào đây.

> **💡 Tại sao khối Provider của Cloudflare không khai báo luôn `zone_id`?**
> Bởi vì Provider đại diện cho **Cấp độ Tài khoản (Account Level)**. Một tài khoản Cloudflare có thể quản lý hàng chục tên miền (Zone) khác nhau. Nếu gán cứng `zone_id` ngay tại Provider, file code này sẽ bị "khóa chết" vào một tên miền duy nhất. Terraform thiết kế đẩy `zone_id` xuống **Cấp độ Tài nguyên (Resource Level)** để một file code có thể linh hoạt cấu hình cho nhiều tên miền (nhiều Zone ID) khác nhau.

### 3.2. Truy vấn Siêu dữ liệu (Data Source Query - OS Image)
```hcl
data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical ID
  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-*"]
  }
}
```
**Mục đích của Block này:**
Dùng để truy vấn Metadata của một tài nguyên ĐÃ TỒN TẠI sẵn trên nền tảng Cloud. Ở đây, ta dùng nó để tự động tìm ID của file cài đặt hệ điều hành (OS Image) Ubuntu 24.04 mới nhất, thay vì phải tra cứu và dán ID cứng vào code.

**Phân tích các dòng lệnh (Arguments):**
- `data "aws_ami" "ubuntu"`: Khai báo lệnh truy vấn loại `aws_ami` (Machine Image), lưu kết quả vào biến cục bộ tên là `ubuntu`.
- `most_recent = true`: Yêu cầu Terraform chỉ trả về kết quả mới nhất theo thời gian (giúp hệ thống luôn nhận được bản vá bảo mật mới nhất mỗi khi khởi tạo).
- `owners = ["099720109477"]`: Account ID của công ty Canonical (cha đẻ của Ubuntu). Đây là cơ chế bảo mật bắt buộc để không vô tình cài nhầm hệ điều hành chứa mã độc do hacker ngụy tạo trên chợ ứng dụng ảo.
- `filter { ... }`: Điều kiện lọc. Dùng Regex `*` kết hợp chuỗi tham chiếu để tìm đúng bản phân phối Ubuntu 24.04 LTS kiến trúc chip amd64 (x86_64).

> **🔗 Tra cứu các Data Source (Siêu dữ liệu) có sẵn của AWS:**
> Terraform không chỉ hỗ trợ truy vấn `aws_ami` mà còn hàng trăm khối `data` khác (như truy vấn thông tin VPC, Subnet, Database, hay IAM Roles đã có sẵn). 
> Bạn có thể tra cứu danh sách toàn bộ các "Data Source" này tại thanh menu bên trái của tài liệu AWS chính thức:
> 👉 [AWS Provider Documentation - Data Sources](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)

### 3.3. Tường lửa Ứng dụng & Mạng (Firewall / Security Group)
```hcl
resource "aws_security_group" "logistic_sg" {
  name        = "logistic-security-group"
  description = "Security rules for Logistics application"

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
```
**Mục đích của Block này:**
Tạo ra một bộ rào chắn mạng (Firewall). Nó làm nhiệm vụ chốt chặn, kiểm duyệt toàn bộ gói tin đi vào (Ingress) và đi ra (Egress) của máy ảo từ tầng Network Layer.

**Phân tích các dòng lệnh (Arguments):**
- `resource "aws_security_group" "logistic_sg"`: Ra lệnh TẠO MỚI một tài nguyên tường lửa, gán tên nội bộ là `logistic_sg` để lát nữa máy ảo gọi ra tái sử dụng.
- `ingress { ... }` (Luồng vào): 
  - `from_port` & `to_port`: Khoảng cổng (Port Range) được mở. Mở cổng 22 để quản trị viên SSH vào bảo trì. Mở cổng 80 để đón người dùng truy cập web.
  - `protocol = "tcp"`: Giao thức truyền tải ở Layer 4 (Transport Layer).
  - `cidr_blocks = ["0.0.0.0/0"]`: Dải mạng được cấp phép. `0.0.0.0/0` mang ý nghĩa là bất kỳ địa chỉ IP nào trên Internet cũng có quyền gọi vào. (Lưu ý: Với cổng 22 ở môi trường Production thực tế, bắt buộc phải đổi IP này thành dải IP nội bộ của công ty/VPN).
- `egress { ... }` (Luồng ra): 
  - Thiết lập `from_port = 0`, `to_port = 0`, `protocol = "-1"`: Đây là quy tắc wildcard cho phép máy ảo đi qua bất kỳ cổng và giao thức (TCP, UDP, ICMP) nào để ra ngoài Internet (Cực kỳ cần thiết để máy ảo tải thư viện, pull Docker image).

### 3.4. Khởi tạo Nút Điện toán (Compute Instance / VM)
```hcl
resource "aws_instance" "logistic_server" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = "t3.large"
  key_name      = "logistic-key" 
  
  vpc_security_group_ids = [aws_security_group.logistic_sg.id]

  root_block_device {
    volume_size = 30
    volume_type = "gp3"
  }

  tags = { Name = "Logistic-Production-Node" }
}
```
**Mục đích của Block này:**
Khởi tạo và cấp phát một Máy ảo (Virtual Machine) hoàn chỉnh. Đây là tiến trình lắp ghép tất cả các thành phần hạ tầng (Hệ điều hành, Tường lửa, Chìa khóa SSH, Ổ cứng) lại với nhau thành một khối điện toán hoạt động.

**Phân tích các dòng lệnh (Arguments):**
- `resource "aws_instance" "logistic_server"`: Ra lệnh cấp phát Compute Instance.
- `ami = data.aws_ami.ubuntu.id`: Không điền mã ID chết (hardcode), mà móc nối động (dynamic reference) vào kết quả của block Data Source phía trên.
- `instance_type = "t3.large"`: Lựa chọn cấu hình sức mạnh tính toán (số lượng vCPU và dung lượng RAM). Billing phí Cloud sẽ tính chủ yếu trên tham số này.
- `key_name`: Tên của Public Key (SSH Key Pair) đã tạo và cấp phép trên Cloud. Quá trình boot máy ảo (Bootstrapping) sẽ tiêm chìa khóa này vào nhân HĐH. Bỏ sót dòng này, Administrator sẽ vĩnh viễn bị chặn quyền truy cập (Locked out).
- `vpc_security_group_ids = [...]`: Gắn Tường lửa đã tạo ở trên vào máy ảo này. Việc gọi tham số `aws_security_group.logistic_sg.id` tạo ra một "Ràng buộc Phụ thuộc" (Dependency Edge). Trình biên dịch Terraform sẽ ngầm hiểu: Phải chờ API tạo Tường lửa thành công xong mới được cấp phát Máy ảo.
- `root_block_device { ... }`: Cấu hình không gian ổ đĩa (Block Storage). Nâng dung lượng lên `30` GB và khai báo chuẩn ổ SSD `gp3` (Hiệu suất IOPS/Throughput cao hơn nhưng giá rẻ hơn chuẩn `gp2` cũ).
- `tags { Name = ... }`: Định danh siêu dữ liệu (Metadata Tagging). Dùng để đặt tên cho máy ảo hiển thị trên giao diện Web UI, tiện cho quá trình thanh toán (Billing) và lọc tài nguyên.

### 3.5. Cấu trúc Phân giải Tên miền & Đảo ngược Proxy (DNS & Edge Reverse Proxy)
```hcl
resource "cloudflare_record" "api_endpoint" {
  zone_id = var.cloudflare_zone_id
  name    = "api" 
  content = aws_instance.logistic_server.public_ip
  type    = "A"
  proxied = true
}
```
**Mục đích của Block này:**
Tạo bản ghi DNS tại biên (Edge) để phân giải Tên miền văn bản (`api.glolog.dev`) về địa chỉ IP Public. Đồng thời kích hoạt tấm khiên bảo vệ lớp ứng dụng của hệ thống CDN.

**Phân tích các dòng lệnh (Arguments):**
- `zone_id`: Mã định danh phân vùng vùng quản trị Domain của bạn trên hệ thống Cloudflare.
- `name = "api"`: Cấu hình Subdomain. Hệ thống tự động ghép với tên miền gốc sẽ thành `api.domain.com`.
- `content = aws_instance.logistic_server.public_ip`: Gán địa chỉ đích (Destination Address). Tương tự như trên, Terraform bóc tách tham số `public_ip` trả về từ kết quả khởi tạo Máy ảo để chèn vào tự động.
- `type = "A"`: Khai báo Bản ghi loại A (Address Record - dùng để trỏ dải FQDN về một địa chỉ IP chuẩn IPv4).
- `proxied = true`: Cờ (Flag) thiết yếu. Bật tính năng Đám mây màu cam (Reverse Proxy Anycast). Khi cờ này kích hoạt, IP thật của máy ảo EC2 bị ẩn giấu 100%. Mọi TCP session từ bên ngoài sẽ bị Cloudflare đánh chặn (Intercept), đưa vào bộ lọc WAF quét mã độc trước khi mã hóa đóng gói đẩy về Origin Server (Máy ảo).

---

## 4. Quy trình Thực thi CLI (Execution Workflow)

- `terraform init`
  Khởi tạo không gian làm việc (working directory), phân tích code và tải về các plugins/providers tương ứng.

- `terraform fmt`
  Tự động định dạng (format) lại source code theo tiêu chuẩn HCL conventions. Khuyến nghị chạy trước khi commit code.

- `terraform validate`
  Kiểm tra tính hợp lệ của cú pháp (syntax) và xác thực cấu trúc cấu hình mà không truy cập API của cloud.

- `terraform plan`
  Phân tích đồ thị phụ thuộc và hiển thị Execution Plan (Kế hoạch thực thi). Hiển thị chi tiết tài nguyên nào sẽ được tạo (`+`), sửa (`~`) hoặc xóa (`-`).

- `terraform apply`
  Thực thi Execution Plan để cấp phát tài nguyên thực tế trên Cloud.

- `terraform destroy`
  Dọn dẹp và hủy bỏ toàn bộ các tài nguyên đang được Terraform quản lý trong State file.

---

## 5. Tham chiếu Tài liệu API Chính thức
- [AWS Provider Documentation](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [Cloudflare Provider Documentation](https://registry.terraform.io/providers/cloudflare/cloudflare/latest/docs)