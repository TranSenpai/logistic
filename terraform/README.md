# HƯỚNG DẪN TRIỂN KHAI HẠ TẦNG VỚI TERRAFORM (INFRASTRUCTURE AS CODE)

Tài liệu này định nghĩa kiến trúc hệ thống, quy trình khai báo tài nguyên và cú pháp chuẩn (HCL) để cấp phát hạ tầng cho dự án Logistics OS. Tài liệu sử dụng các khái niệm tổng quát (Agnostic Cloud) và ví dụ cụ thể trên AWS/Cloudflare.

---

## 1. Kiến trúc Hệ thống (System Architecture)

### 1.1. Các thành phần cốt lõi (Core Components)

#### 1.1.1. Các khối cú pháp chính trong Terraform:
- **Provider (Nền tảng kết nối):** Đóng vai trò cầu nối. Terraform sẽ thao tác gọi API tới các Provider này (ví dụ: AWS cho máy ảo, Cloudflare cho tên miền) để cấp phát tài nguyên thực tế.
- **Data (Truy vấn dữ liệu):** Dùng để lấy thông tin của các tài nguyên **đã tồn tại sẵn** do Provider cung cấp. Việc này giúp đóng gói thông tin (ví dụ: ID của bản OS mới nhất), tăng tính tái sử dụng và tránh sai sót do cấu hình thủ công.
- **Resource (Tài nguyên):** Khối khai báo các tài nguyên **cần được quản lý (tạo mới, cập nhật, xóa)** trên Cloud (như máy ảo, mạng, tường lửa). Các tài nguyên này có thể được cấu hình độc lập hoặc nhận dữ liệu tham chiếu từ khối `Data`.

#### 1.1.2. Các khối tài nguyên Hạ tầng quan trọng:
- **Network & Security (Mạng & Bảo mật):** Đây là phần cốt lõi bảo vệ hệ thống:
  1. **Hạ tầng mạng ảo (VPC):** Khoanh vùng mạng độc lập cho dự án. Vì Cloud là môi trường dùng chung phần cứng vật lý (Multi-tenant), VPC giúp cô lập hệ thống của bạn với các khách hàng khác, ngăn chặn triệt để nguy cơ bị rà quét IP/Port hay bị "nghe lén" gói tin. 
  2. **Kiểm soát truy cập:** Áp dụng nguyên tắc "đặc quyền tối thiểu". Chỉ mở những Port/IP thực sự cần thiết. Càng ít đường vào, hệ thống càng an toàn, giúp DevOps dễ dàng kiểm soát và truy vết khi có sự cố.
  3. **Phân mảnh mạng (Subnetting):** Tiếp tục chia nhỏ VPC thành các mạng nội bộ cho từng cụm service/database. Nếu một cụm bị tấn công, sự cố sẽ không thể lây lan (Lateral Movement) sang vùng khác. Khi khác dải mạng, luồng dữ liệu buộc phải đi qua Router và sẽ bị Tường lửa đánh chặn hoặc DevOps có thể cấu hình cách ly nóng ngay lập tức.
- **DNS & Proxy Provider (Phân giải tên miền & Reverse Proxy):** Hệ thống phân giải DNS và bảo vệ tại biên (Ví dụ: Cloudflare). Cơ chế **Reverse Proxy (Đảo ngược Proxy)** hoạt động như sau:
  1. **Nhận Request:** Khi người dùng truy cập Domain, trình duyệt phân giải DNS và gửi Request tới máy chủ của Cloudflare (thay vì server gốc).
  2. **Đánh chặn & Quét (WAF):** Cloudflare chặn Request lại để lọc DDoS và quét các cuộc tấn công web (SQL Injection, XSS...).
  3. **Đại diện kết nối:** Nếu Request an toàn, Cloudflare đối chiếu bản ghi để tìm IP Public thực sự của Server AWS. Sau đó nó đứng ra làm đại diện gửi Request đó tới AWS.
  4. **Trả kết quả:** Server AWS xử lý xong trả Response về cho Cloudflare, Cloudflare cầm kết quả trả lại cho Client. Nhờ vậy, IP thật của AWS được ẩn giấu và bảo vệ 100%.
- **Network Topology (Mô hình Định tuyến):** Tổng quan luồng đi của dữ liệu (Traffic Flow) sẽ diễn ra theo trình tự khép kín:
  1. `Client`: Thiết bị của người dùng (Mobile App, Web Browser) khởi tạo Request.
  2. `Cloud Proxy (Port 443/80)`: Trạm gác Cloudflare. Tiếp nhận Request, giải mã HTTPS (Port 443), lọc mã độc, và mã hóa lại trước khi gửi đi.
  3. `Public Subnet (Port 80)`: Vùng đệm mạng trên AWS (DMZ). Chứa điểm tiếp nhận (như Load Balancer hoặc Nginx Edge Server).
  4. `Internal Reverse Proxy`: Máy chủ Nginx nội bộ. Nhận Request từ vùng đệm, phân tích URL (ví dụ `/api/auth` hay `/api/media`) để điều phối (Routing) luồng dữ liệu.
  5. `App Containers`: Các container Docker chứa mã nguồn xử lý logic kinh doanh (Auth, Matching, Media). Nằm sâu trong vùng Private, không bao giờ lộ mặt ra Internet.

#### 1.1.3. Giao thức Quản trị (SSH) & Lựa chọn Hệ điều hành Server
- **SSH (Secure Shell) là gì? Tại sao phải dùng nó?**
  - **Bản chất (Nó giải quyết bài toán gì?):** Các máy chủ Cloud (đặc biệt là Linux) thường **KHÔNG CÓ Giao diện Đồ họa (GUI)**. Bạn không thể cắm dây cáp màn hình vào máy ảo đặt tại Singapore được. Bài toán là: *Làm sao để một kỹ sư ở Việt Nam gõ một lệnh trên Terminal/Powershell của máy tính cá nhân, và lệnh đó được thực thi ngay trên máy chủ ở Singapore?* SSH sinh ra để làm "sợi dây cáp tàng hình" đó. Nó cho phép bạn mở Terminal ở máy nhà, nhưng thực chất là đang điều khiển máy chủ từ xa.
  - **Bảo mật (Chữ "Secure" trong SSH nghĩa là gì?):** 
    - *Chuyện gì xảy ra nếu không dùng SSH?* Ngày xưa, người ta dùng giao thức **Telnet** (gõ lệnh từ xa không mã hóa). Nếu dùng Terminal thuần bằng Telnet và bạn gõ `password: 123456`, đoạn text đó truyền qua hàng trăm trạm trung chuyển Internet dưới dạng chữ thô (Plain Text). Bất kỳ hacker nào dùng tool "bắt gói tin" (Wireshark) đều có thể đọc được mật khẩu của bạn dễ như ăn kẹo.
    - *Cách SSH bảo vệ:* SSH áp dụng **Mật mã bất đối xứng (Asymmetric Cryptography)** với Public Key và Private Key. Khi bạn gõ `123456`, SSH lập tức băm nó thành chuỗi vô nghĩa `hj!@#dsd890` trước khi gửi đi. Hacker bắt được gói tin này trên đường truyền cũng vô dụng vì không có Private Key (chỉ được lưu ẩn trên máy tính cá nhân của bạn) để giải mã.
  - **Ưu điểm:**
    - **Bảo mật tuyệt đối:** Chống nghe lén (Sniffing) và chống giả mạo máy chủ (Man-in-the-middle).
    - **Đăng nhập không cần Password:** SSH hỗ trợ xác thực bằng Key Pair (file `.pem` bạn hay tải từ AWS). Nếu dùng Key, dù hacker có dò mật khẩu (Brute-force) hàng triệu lần cũng không thể vào được server vì Server đã cấm hoàn toàn việc gõ password.
    - **Đa dụng:** SSH không chỉ để gõ lệnh. Nó còn tạo "Hầm bí mật" (SSH Tunneling) để chui vào các vùng mạng Private nội bộ một cách an toàn.
  - **Nhược điểm:**
    - Yêu cầu người dùng phải quen xài dòng lệnh (CLI).
    - **Rủi ro mất Key:** Nếu bạn làm mất file Private Key, bạn sẽ mất quyền truy cập vĩnh viễn vào server đó (vì không cho gõ password nữa). Nếu lộ Key vào tay kẻ gian, hệ thống coi như sập.
    - Vì cổng mặc định của SSH là 22, nó luôn là mục tiêu bị các con Botnet trên mạng càn quét 24/7.

  - **Cú pháp Lệnh kết nối thực tế:**
    Để remote vào máy chủ EC2 (Ubuntu) trên AWS, bạn mở Terminal tại thư mục chứa file `.pem` và gõ:
    ```bash
    # Bước 1: Siết chặt quyền của file key (Bắt buộc trên Mac/Linux để chống lỗi "Unprotected Private Key")
    chmod 400 logistic-key.pem 
    
    # Bước 2: Kích hoạt kết nối SSH
    ssh -i logistic-key.pem ubuntu@<IP_PUBLIC_CUA_EC2>
    ```
    *(Ghi chú: Thay `<IP_PUBLIC_CUA_EC2>` bằng địa chỉ IP thực tế mà Terraform in ra sau khi chạy lệnh apply. User mặc định của hệ điều hành Ubuntu luôn là `ubuntu`).*

- **So sánh OS (Hệ điều hành) Server: Tại sao Linux là vua?**
  1. **Linux (Ubuntu, CentOS, Alpine...):** 
     - *Ưu điểm:* Mã nguồn mở, hoàn toàn **Miễn phí** (không tốn phí bản quyền License). Siêu nhẹ: một server Linux không có giao diện chỉ tốn khoảng 100MB-200MB RAM để chạy hệ điều hành (dành toàn bộ tài nguyên còn lại cho ứng dụng). Độ ổn định cực kỳ cao (chạy hàng năm trời không cần Restart). Là môi trường gốc rễ sản sinh ra Docker và Kubernetes.
     - *Nhược điểm:* Khó học với người mới, đòi hỏi DevOps phải thuộc lòng các câu lệnh (CLI commands) thay vì click chuột.
  2. **Windows Server:**
     - *Ưu điểm:* Có giao diện đồ họa (GUI) trực quan y hệt máy cá nhân, dễ làm quen bằng cách click chuột. Phù hợp nếu công ty đang bị khóa chặt (Vendor Lock-in) vào hệ sinh thái của Microsoft (như code C# .NET đời cũ, dùng SQL Server, quản lý bằng Active Directory).
     - *Nhược điểm:* **Chi phí cực đắt** (phải đóng phí bản quyền cho Microsoft theo từng nhân CPU). Quá nặng nề (HĐH tốn vài GB RAM và hàng chục GB ổ cứng chỉ để render cái giao diện hình ảnh đồ họa). Thường xuyên phải khởi động lại (Restart) mỗi khi cập nhật (Update), gây gián đoạn dịch vụ.
  3. **macOS Server:**
     - *Ưu điểm:* Môi trường Unix tuyệt vời. **Bắt buộc** phải dùng nếu công ty bạn làm luồng CI/CD (tự động hóa build app) cho ứng dụng iOS/macOS (vì Apple cấm build code iOS trên HĐH khác).
     - *Nhược điểm:* Apple không bán HĐH rời cho các nhà cung cấp Cloud. Để có macOS trên Cloud, bạn phải thuê nguyên một chiếc máy tính "Mac Mini vật lý" đặt ở Data Center (Dedicated Host). Chi phí đắt khủng khiếp và cực kỳ khó nhân bản (scale) so với việc khởi tạo máy ảo linh hoạt.
  - **Kết luận:** Người ta luôn chọn **Linux** làm tiêu chuẩn vàng cho Server/Cloud deployment vì 3 tiêu chí cốt lõi: **Miễn phí, Tiết kiệm tài nguyên (Siêu nhẹ), và Độ ổn định cao.** Windows chỉ dùng khi bất đắc dĩ bị ép framework, còn macOS chỉ dùng làm máy build app iOS.

### 1.2. Terraform State (`terraform.tfstate`)
- **Bản đồ Hạ tầng (Mapping & Metadata):** Là nơi lưu trữ trạng thái hiện tại của hạ tầng và thứ tự phụ thuộc của tài nguyên (cái nào tạo trước, cái nào tạo sau). Nhờ đó, Terraform biết chính xác cần thêm, xóa, sửa tài nguyên nào và theo trình tự nào.
- **Lưu trữ sự ràng buộc (Bindings):** Ví dụ, khi bạn tạo máy ảo bằng lệnh `resource "aws_instance" "web" {}`, Terraform sẽ ghi vào file state rằng: *block code "web" tương ứng với máy ảo có ID `i-1234567890abcdef0` trên AWS*. Lần sau nếu bạn sửa cấu hình của block "web", Terraform biết chính xác phải sửa ID nào trên thực tế.
- **Tối ưu Hiệu năng (Performance):** Nếu không có file state, Terraform sẽ bị "mù". Mỗi lần chạy lệnh, nó sẽ phải gọi API bắt AWS liệt kê toàn bộ tài nguyên của tài khoản ra để dò tìm. Việc này cực kỳ chậm và dễ bị chặn (Rate Limit). State giúp Terraform tra cứu trực tiếp bằng ID siêu nhanh.
- **Cơ chế hoạt động (Refresh ➡️ Plan ➡️ Apply):** Trước khi quyết định thay đổi hạ tầng, Terraform thực hiện quy trình sau:
  1. **Refresh (Cập nhật thực tại):** Đọc file state để lấy danh sách ID, sau đó gọi API lên Cloud kiểm tra hiện trạng thực tế của các ID này (đề phòng trường hợp ai đó lén sửa bằng tay trên giao diện Web). Nếu có sai lệch, nó ngầm cập nhật lại file state.
  2. **Plan (Lập kế hoạch):** Lấy source code (`.tf`) so sánh với file state (vừa được refresh) để phát hiện thay đổi, từ đó in ra màn hình Kế hoạch thực thi.
  3. **Apply (Áp dụng):** Chờ người dùng (DevOps) review. Nếu OK, Terraform mới thực sự gọi API lên Cloud để áp dụng các sửa đổi.
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
- **VPC (Virtual Private Cloud) - Bản chất Backbone của Điện toán Đám mây:** 
  - **1. Nó là gì? (Bản chất):** VPC là một không gian mạng ảo (Software-Defined Networking) độc lập và bị cô lập hoàn toàn về mặt logic trên cơ sở hạ tầng vật lý của AWS. Nó chính là "Bao bì" (Network Perimeter) gói gọn toàn bộ kiến trúc (EC2, RDS, Load Balancer) vào một Data Center riêng tư ảo của bạn trên Cloud.
  - **2. Tại sao các hệ thống Cloud lại sinh ra khái niệm này?** Thời kỳ đầu (trước 2009), AWS sử dụng mạng "EC2-Classic": Máy ảo của bạn, của đối thủ cạnh tranh, và của Hacker đều bị ném chung vào một mạng lưới phẳng khổng lồ (Flat Network). Các khách hàng chia sẻ chung cơ sở hạ tầng mạng (Multi-tenant) dẫn đến nguy cơ bảo mật cực kỳ khủng khiếp (Bị rà quét IP chéo nhau). Để cứu vãn lỗ hổng chết người này, AWS buộc phải đẻ ra VPC, sử dụng công nghệ ảo hóa mạng để "chia lô", đảm bảo máy ảo của công ty A vĩnh viễn không thể "nhìn thấy" gói tin mạng của công ty B.
  - **3. Nó tồn tại để giải quyết bài toán gì?** Bài toán "Tính riêng tư trên một phần cứng công cộng" (Privacy on Public Cloud). VPC cho phép bạn mang nguyên xi cấu trúc mạng LAN ở dưới mặt đất (IP nội bộ, Router, Subnet, Firewall) lên thẳng Cloud mà không sợ bị trùng lặp IP hay lộ lọt dữ liệu ra ngoài Internet. Nhờ áp dụng dải IP Private (`10.0.0.0/16` theo chuẩn RFC 1918), hệ thống hoàn toàn vô hình trước các máy quét trên Internet.
  - **4. Thiếu nó sẽ ra sao?** Hệ thống của bạn sẽ phơi mình 100% ra Public Internet giống như một chiếc máy tính ngoài quán Net. Không có ranh giới mạng, không có khả năng ẩn giấu Database (Private Subnet), mọi EC2 khởi tạo ra đều buộc phải gánh IP Public và chịu những đợt tấn công DDoS, Brute-force vô tận từ Hacker 24/7.
  - **5. Cái giá phải trả (Trade-off) của VPC là gì?** 
    - *Độ phức tạp (Complexity):* Chuyển gánh nặng quản trị mạng từ tay AWS sang tay kỹ sư DevOps. Bạn phải tự quy hoạch Routing Table, NAT Gateway, Internet Gateway, Subnetting. Sai một ly (cấu hình sai Route) là toàn bộ máy chủ bị ngắt kết nối (Network Outage) dẫn tới sập hệ thống.
    - *Chi phí ẩn (Hidden Cost):* Bản thân VPC thì miễn phí, nhưng các thành phần giải quyết bài toán "đi xuyên qua VPC" lại đắt đỏ. Ví dụ: NAT Gateway (công cụ để Private Subnet chui ra Internet tải báo cáo/update) tính phí theo mỗi GB dữ liệu đi qua. Càng bảo mật, chi phí truyền tải càng cao.
  - **6. Tại sao lại dùng Custom VPC mà cạch mặt Default VPC?** AWS luôn cấp sẵn một "Default VPC" cho tài khoản mới. Nhưng nó được thiết kế quá dễ dãi: TẤT CẢ các Subnet bên trong đều tự động nối thẳng ra Internet (Tất cả đều là sân trước). Dùng Default VPC cho hệ thống Enterprise chứa thông tin khách hàng (Logistic) là một thảm họa bảo mật. Việc tự code ra Custom VPC đảm bảo nguyên lý tối thượng **Zero Trust**: Bắt đầu từ số 0, mọi cánh cửa mặc định là khóa kín tuyệt đối, cho đến khi kỹ sư chủ động cấu hình mở ra từng lỗ một.

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