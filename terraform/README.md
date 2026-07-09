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

### 3.3. Quy hoạch Mạng Nội bộ (Custom VPC & Subnet)
```hcl
resource "aws_vpc" "logistic_vpc" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
}

resource "aws_subnet" "logistic_public_subnet" {
  vpc_id                  = aws_vpc.logistic_vpc.id
  cidr_block              = "10.0.1.0/24"
  map_public_ip_on_launch = true
}
```
**Bản chất Kiến trúc chuyên sâu (Deep Architecture Essence):**
- **VPC (Virtual Private Cloud - Network Perimeter):** 
  - *Định nghĩa:* Là một vùng mạng riêng ảo độc lập hoàn toàn trên cơ sở hạ tầng AWS. VPC đóng vai trò là Network Perimeter (Vòng đai mạng) cô lập hoàn toàn hệ thống Logistic khỏi phần còn lại của Internet.
  - *Tại sao phải tạo Custom VPC?* AWS cung cấp sẵn một Default VPC với cấu hình định tuyến rất lỏng lẻo (tất cả Subnet đều Public). Đối với hệ thống Enterprise đòi hỏi tính bảo mật dữ liệu cao, ta phải tuân thủ nguyên tắc **Zero Trust Network**. Việc tự định nghĩa Custom VPC từ con số 0 đảm bảo không một kết nối nào được phép ra/vào trừ khi được khai báo tường minh.
  - *CIDR Block*: Cấu hình `10.0.0.0/16` áp dụng **RFC 1918** (Dải IP Private, không định tuyến được trên Internet toàn cầu), giúp hệ thống hoàn toàn ẩn danh. Prefix `/16` cung cấp không gian 65,536 IP, đáp ứng hoàn hảo khả năng Scale-out (Mở rộng theo chiều ngang) cho các cụm Microservices/Containers sau này.

- **Public Subnet (DMZ / Public-Facing Subnet):** 
  - *Định nghĩa:* Quá trình phân đoạn mạng (Network Segmentation). Phân mảnh không gian `/16` khổng lồ thành các vùng nhỏ hơn (Subnet) để gán Routing Policies (Chính sách định tuyến) chuyên biệt. Dải `10.0.1.0/24` cấp phát 256 IP cho cụm máy chủ Public.
  - *Tại sao gọi là Public Subnet?* Vì Subnet này sẽ được liên kết trực tiếp với **Internet Gateway (IGW)** thông qua Route Table, đóng vai trò như một vùng phi quân sự (DMZ), nơi đặt các Application Load Balancer hoặc Edge Proxy Server để hứng Inbound Traffic từ người dùng cuối.
  - *Resource Dependency (Ràng buộc tài nguyên)*: Khai báo `vpc_id = aws_vpc.logistic_vpc.id` tạo ra một Implicit Dependency trong luồng thực thi Terraform. Nó ép trình biên dịch (Terraform Core) phải hoàn thành việc gọi API tạo VPC trước, lấy được định danh VPC ID, rồi mới truyền xuống để tạo Subnet.

### 3.4. Tường lửa Ứng dụng & Mạng (Firewall / Security Group)
```hcl
resource "aws_security_group" "logistic_sg" {
  name        = "logistic-security-group"
  description = "Security rules for Logistics application"
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
```
**Bản chất Kiến trúc chuyên sâu (Deep Architecture Essence):**
- **Security Group (Stateful Firewall / L4 Packet Filtering):** 
  - *Định nghĩa:* Hoạt động như một Tường lửa trạng thái (Stateful Firewall) kiểm soát Inbound/Outbound traffic ở cấp độ Máy ảo (Instance-level), thuộc Layer 4 (Transport Layer) trong mô hình OSI. Khác với NACL (Network Access Control List) hoạt động ở cấp độ Subnet, Security Group bám sát vào từng Network Interface (ENI).
  - *Tại sao phải tạo?* Việc đặt Compute Node trong Public Subnet khiến nó bộc lộ bề mặt tấn công (Attack Surface) ra Internet. Security Group áp dụng chính sách **Default Deny** (Chặn toàn bộ Inbound traffic mặc định), ngăn chặn triệt để các cuộc rà quét cổng (Port Scanning). Tham số `vpc_id` khóa chặt tập luật này vào đúng vùng mạng Logistic VPC.

- **Ingress Rules (Luồng mạng đi vào):**
  - Chỉ cho phép Mở luồng Explicit (Tường minh):
  - `Port 22 (TCP)`: Giới hạn riêng cho Administrator kết nối SSH (Secure Shell) để Provisioning (Cấu hình) máy chủ. (Trong môi trường Production, `0.0.0.0/0` nên được siết chặt lại thành dải IP VPN của công ty).
  - `Port 80 (TCP)`: Tiếp nhận HTTP traffic từ end-user để phục vụ Web Application.
  - Bất kỳ TCP SYN packet nào nhắm vào các cổng không khai báo (ví dụ: 3306, 6379) đều bị "Drop" (Loại bỏ) ngay tại Network Layer, giúp hệ điều hành bên trong giải phóng chu kỳ CPU (không phải mất công phản hồi RST packet), giảm thiểu nguy cơ SYN Flood DDoS.

- **Egress Rules (Luồng mạng đi ra):**
  - Cấu hình wildcard `protocol = "-1"` cho phép Node nội bộ khởi tạo kết nối đến mọi đích đến trên Internet. Điều này là bắt buộc trong giai đoạn Bootstrapping để hệ điều hành có thể phân giải DNS, chạy `apt update` (tải Linux packages), kết nối GitHub để clone source code, hoặc giao tiếp với Docker Hub để pull images.

### 3.5. Khởi tạo Nút Điện toán (Compute Instance / VM)
```hcl
resource "aws_instance" "logistic_server" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = "t3.large"
  key_name      = "logistic-key" 
  subnet_id     = aws_subnet.logistic_public_subnet.id
  
  vpc_security_group_ids = [aws_security_group.logistic_sg.id]

  root_block_device {
    volume_size = 30
    volume_type = "gp3"
  }

  tags = { Name = "Logistic-Production-Node" }
}
```
**Bản chất Kiến trúc chuyên sâu (Deep Architecture Essence):**
- **EC2 Instance (Compute Node / Raw Compute Capacity):** 
  - *Định nghĩa:* Elastic Compute Cloud (EC2) cung cấp năng lực tính toán cốt lõi (vCPU, RAM, Block Storage). Ở trạng thái khởi tạo ban đầu (Vanilla State), nó chỉ là một hạt nhân Linux thuần túy chưa có Application Stack (Docker, Go Runtime, Database).
  - *Thứ tự biên dịch (Dependency Graph)*: Compute Node bắt buộc phải được tạo cuối cùng trong chuỗi cung ứng Hạ tầng Mạng. Terraform Core xây dựng Dependency Graph nội bộ và nhận diện rõ EC2 phải đợi VPC, Subnet và Security Group đạt trạng thái `Available` thì mới được kích hoạt API `RunInstances`.

- **Network Interface Placement (Subnet Binding):**
  - Tham số `subnet_id` quyết định vị trí Vùng khả dụng (Availability Zone) và gán Elastic Network Interface (ENI) của máy ảo vào đích danh môi trường DMZ (Public Subnet). Khuyết thiếu cấu hình này sẽ gây ra tình trạng Drift Configuration, khiến AWS tự động đặt Node vào Default VPC.

- **Traffic Filtering Enforcement (Security Binding):** 
  - Tham số `vpc_security_group_ids` đính kèm tập luật Stateful Firewall vào trực tiếp ENI của EC2 Instance. Khẳng định cơ chế an ninh phân tán: Việc thanh lọc Packet xảy ra ngay tại Network Interface của từng máy ảo độc lập, thay vì dồn về một Gateway trung tâm.

- **Access Provisioning (Key Injection / Cloud-Init):** 
  - Khai báo `key_name = "logistic-key"`. Khi máy ảo được cung cấp năng lượng lần đầu (First Boot), tiến trình `cloud-init` của AWS sẽ tự động lấy đoạn văn bản Public Key (RSA) tương ứng tiêm thẳng vào nhân HĐH (tại đường dẫn `~/.ssh/authorized_keys`). 
  - Đảm bảo cơ chế Asymmetric Cryptography (Mật mã Bất đối xứng) cho phiên SSH đầu tiên. Nếu lược bỏ tham số này, môi trường sẽ rơi vào trạng thái "Blackbox" (Chạy thành công nhưng Administrator hoàn toàn mất quyền truy cập SSH để thực thi Configuration Management).

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