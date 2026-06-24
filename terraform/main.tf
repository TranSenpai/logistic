# Script gọi API tạo EC2 và map IP cho tên miền
// Terraform dùng HashiCorp Configuration Language (HCL) với tên file kết thúc bằng hậu tố <.tf>.
//  Ví dụ: main.tf, output.tf, variables.tf, ... .

// Khi thực hiện các lệnh cli, Terraform sẽ load hết các file cấu hình trong thư mục đang làm
// việc và tự động giải các thư việc (dependencies) trong các cấu hình

// Một file code là ghép nhiều "khối" lại với nhau
// Có 4 khối quan trọng:
//  - provider: Khai báo cho Terraform biết đang muốn làm việc với ai. Ví dụ:
//      + provider "aws"
//      + provider "cloudflare" 

//  - resource: khối quan trọng nhất. Nó ra lệnh tạo 1 thứ gì đó
//   cú pháp là: resource "<tên_nhà_cung_cấp>_<loại_tài_nguyên>" "tên_gọi_do_bạn_đặt" { ... }
//   Terraform luôn lấy phần chữ đứng trước dấu gạch dưới ĐẦU TIÊN làm tên nhà cung cấp, 
//   và TOÀN BỘ phần còn lại làm tên loại tài nguyên.
//   Ví dụ: resource "aws_instance" "server_cua_toi" { ... } (tạo máy chủ EC2)

//  - data: Dùng để truy vấn thông tin đã tồn tại thay vì tạo mới. 
//   Ví dụ tìm hệ điều hành Linux Ubuntu mới nhất trên AWS

//  - variable: Dùng để truyền dữ liệu động từ bên ngoài vào 
//   (như API Token, Zone ID) để tránh bị lộ thông tin nhạy cảm

// Muốn tự code Terraform lên Terraform Registry search 

// Terraform block cấu hình chính Terraform vì bản thân Terraform là khung rỗng 
// chưa có các Binary Plugin (được gọi là providers) để quản lý resources bằng cách call API cloud provider. 
// Các provider này được tách rời (decoupled) khỏi Terraform binary và được
// phát hành theo phiên bản, nó có thể cung cấp cho bất cứ nhà cung cấp nào.
terraform {
  // Set ràng buộc version với các nhà cung cấp
  // Terraform Registry là nơi store các provider của Terraform
  // Có thể có nhiều khối provider cùng 1 Terraform thuộc cùng 1 provider hoặc khác provider
  // Ví dụ: có cả provider "aws" và provider "cloudflare" trong cùng một file hoặc 
  // giả sử ứng dụng của bạn cần 1 server ở Singapore (ap-southeast-1) và 1 server dự phòng ở Mỹ (us-east-1)
  // có thể khai báo 2 block provider "aws" khác nhau (sử dụng thêm thuộc tính alias để phân biệt)
  // trong cùng file main.tf này.
  required_providers {
    aws = {
        source = "hashicorp/aws"
        // Không set version thì Terraform mặc định version mới nhất của provider.
        // Nhưng không khuyến khích vì nó có thể tải các version mới nhất mà DevOps chưa test.
        // Ký hiệu ~> 5.92 nghĩa là cấu hình này hỗ trợ bất cứ bản nào có verison chính là 5 và 
        // verison phụ từ 92 trở lên, còn ký hiệu >= 1.2 là từ bản 1.2 trở lên. 
        version = "~> 5.92"
    }
    cloudflare = {
        source = "cloudflare/cloudflare"
        version = "~> 4.0"
    }
  }

  required_version = ">= 1.2"
}

provider "aws" {
    // khu vực Singapore cho gần Việt Nam
    region = "ap-southeast-1"
}

// Trên AWS có hàng trăm ngàn hệ điều hành (AMI - Amazon Machine Image) khác nhau do cả AWS và cộng đồng tạo ra.
// Vì vậy nên phải filter đúng cái AMI dev muốn, đó là chức năng của filter.
data "aws_ami" "Ubuntu" {
  # Lấy bản mới nhất
  most_recent = true

  # Terraform sẽ gọi lên AWS và nói: "Hãy tìm cho tôi một cái hệ điều hành do 
  # chính chủ Canonical tạo ra (owners = ["099720109477"]), có cái tên bắt đầu 
  # bằng chữ 'ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-' (filter), 
  # và nếu có nhiều bản cập nhật thì lấy bản mới nhất cho tôi (most_recent = true)."
  filter {
    # name = "name": Báo cho Terraform biết là "Hãy lọc dựa trên thuộc tính 'Tên' của hệ điều hành".
    name = "name"
    # values = ["ubuntu/images/...-*"]: Đây là từ khóa tìm kiếm. 
    # Dấu * ở cuối đại diện cho bất kỳ ký tự nào phía sau 
    # (vì Ubuntu thường xuyên cập nhật version ngày tháng ở cuối đuôi, ví dụ ...-20240101).
    values = ["ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-*"]
  }

  // ID của công ty Canonical (cha đẻ của Ubuntu)
  owners = ["099720109477"]
}

# aws_security_group: Đây MỚI LÀ loại tài nguyên (Resource Type) kết hợp với tên nhà cung cấp (aws).
# Nó báo cho Terraform biết bạn muốn tạo một cái Tường lửa (Security Group) trên AWS.
resource "aws_security_group" "lab_sg" {
  name = "lab-security-group"
  description = "Allow SSH, HTTP, and HTTPS"

  # ingress (Luồng vào): Chúng ta cần mở port 22 (để SSH gõ lệnh) và port 80, 443 (để Nginx/Golang nhận request web).
  # egress (Luồng ra): Cho phép server gọi ra ngoài internet (để tải package, cập nhật ubuntu...). 
  # Tham số protocol = "-1" báo cho AWS biết là cho phép mọi loại kết nối đi ra ngoài.

  # Mở port 22 để có thể SSH vào server gõ lệnh
  ingress {
    from_port = 22
    to_port = 22
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Mở port 80 cho HTTP
  ingress {
    from_port = 80
    to_port = 80
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Mở port 443 cho HTTPS
  ingress {
    from_port = 443
    to_port = 443
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  } 

  # Cho phép server gọi ra ngoài internet (để cài package, update, ...)
  egress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}


provider "cloudflare" {
    api_token = var.cloudflare_api_token
}



# Phân biệt 3 loại "Tên" trong Terraform
#   Loại 1: Tên của loại tài nguyên (Resource Type)
#       Ví dụ: Chữ aws_security_group trong resource "aws_security_group".
#   Đặc điểm: Đây là tên do nhà cung cấp (AWS) quy định chết. Bạn bắt buộc 
#   phải gõ đúng từng chữ, không được phép thay đổi.

#   Loại 2: Tên gọi nội bộ trong code (Local Name)
#       Ví dụ: Chữ "lab_sg" trong đoạn code vừa rồi.
#   Đặc điểm: Đây chỉ là TÊN GỌI NỘI BỘ (Local Name) do chính bạn tự đặt ra. 
#   Tên này chỉ tồn tại và có ý nghĩa bên trong file code Terraform, 
#   dùng để phân biệt các block code với nhau. AWS không hề biết đến cái tên này.

# Loại 3: Tên thực tế hiển thị trên Đám mây (Arguments)
#       Ví dụ: Dòng name = "lab-security-group" nằm bên trong cặp ngoặc { }.
#   Đặc điểm: Đây chính là tên định danh vật lý mà Terraform sẽ đóng gói và 
#   gửi lên API của AWS. Tuy nhiên, không phải tài nguyên nào cũng dùng chữ name. 
#   Có tài nguyên dùng chữ name, nhưng có tài nguyên (như máy chủ EC2) lại bắt buộc dùng tags = { Name = "..." }.

