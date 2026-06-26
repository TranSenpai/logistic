# Tài liệu học Terraform & DNS

File này lưu trữ toàn bộ các ghi chú quan trọng về cú pháp Terraform và các khái niệm hệ thống mạng để bạn tra cứu khi cần.

---

## Phần 1: Cấu trúc cơ bản của một file Terraform (`.tf`)
Mỗi file code Terraform là tập hợp của nhiều "Khối" (Block). Có 4 khối quan trọng nhất:

1. **`provider`**: Khai báo cho Terraform biết đang muốn làm việc với nền tảng nào (AWS, Cloudflare...).
   - *Ví dụ:* `provider "aws" { region = "ap-southeast-1" }`
2. **`resource`**: Khối quan trọng nhất, dùng để RA LỆNH tạo một tài nguyên mới trên Cloud.
   - *Cú pháp:* `resource "<tên_nhà_cung_cấp>_<loại_tài_nguyên>" "tên_gọi_do_bạn_đặt" { ... }`
   - *Ví dụ:* `resource "aws_instance" "logistic_server" { ... }`
3. **`data`**: Dùng để truy vấn thông tin đã TỒN TẠI trên Cloud thay vì tạo mới.
   - *Ví dụ:* Tìm hệ điều hành Ubuntu mới nhất do Canonical phát hành (`data "aws_ami" "Ubuntu"`).
4. **`variable`**: Dùng để truyền dữ liệu động từ bên ngoài (như API Token, Zone ID) vào code để tránh lộ thông tin nhạy cảm.

---

## Phần 2: Phân biệt 3 loại "Tên" trong Terraform
Rất dễ nhầm lẫn giữa các loại tên khi code:

1. **Tên của loại tài nguyên (Resource Type)**: Chữ ĐẦU TIÊN trong block `resource`.
   - *Ví dụ:* `aws_security_group`. Do AWS quy định chết, phải gõ đúng từng chữ.
2. **Tên gọi nội bộ (Local Name)**: Chữ THỨ HAI trong block `resource`.
   - *Ví dụ:* `lab_sg`. Do bạn tự đặt để gọi qua lại bên trong code. AWS không hề biết tên này.
3. **Tên thực tế trên Cloud (Arguments)**: Các thuộc tính nằm bên trong dấu ngoặc `{ }`.
   - *Ví dụ:* `name = "lab-security-group"` hoặc `tags = { Name = "Logistic-Lab" }`. Đây là tên thật hiển thị trên bảng điều khiển của AWS.

---

## Phần 3: Lưu ý về Security Group (Tường lửa AWS)
- **Ingress (Luồng vào)**: Cấu hình ai được phép gọi vào server.
  - Port `22` (SSH): Dùng để gõ lệnh.
  - Port `80/443` (HTTP/HTTPS): Dùng cho Web Server.
  - Cấu hình `0.0.0.0/0` nghĩa là "Bất kỳ ai trên Internet". (Với port 22 ở môi trường Production thì nên set thành IP công ty thay vì `0.0.0.0/0`).
- **Egress (Luồng ra)**: Cấu hình server được phép gọi đi đâu. Thường để `protocol = "-1"` (cho phép ra mạng tự do) để server tải package cài đặt.

---

## Phần 4: Cẩm nang các loại Bản ghi DNS (DNS Record Types)

Hệ thống DNS như một danh bạ điện thoại khổng lồ. Dưới đây là các loại bản ghi phổ biến:

1. **Bản ghi A (Address):**
   - Trỏ Tên miền về địa chỉ IPv4 (Ví dụ: `api.logistic.com -> 14.22.33.44`). Dùng để kết nối web.
2. **Bản ghi CNAME (Canonical Name):**
   - Trỏ Tên miền về một Tên miền khác (Bí danh). (Ví dụ: `www.logistic.com -> logistic.com`).
   - *Lưu ý:* CNAME chỉ ẩn được tên miền thật, **KHÔNG GIẤU ĐƯỢC IP VÀ PORT**. 
3. **Bản ghi MX (Mail Exchange):**
   - Báo cho thế giới biết máy chủ nào nhận Email cho tên miền này (Dùng khi setup Google Workspace, Outlook).
4. **Bản ghi TXT (Text):**
   - Chứa văn bản thuần túy. Thường dùng để nhét "chữ ký" chứng minh bạn là chủ sở hữu tên miền (Google/Facebook hay yêu cầu).

> **💡 Trả lời câu hỏi: Có nên dùng CNAME để giấu IP/Port không?**
> **KHÔNG.** CNAME không có khả năng giấu IP hay Port. Nếu ai đó PING vào cái CNAME của bạn, nó vẫn sẽ lòi ra cái IP thật ở đằng sau.
> **Cách duy nhất để giấu IP:** Phải sử dụng tính năng Proxy của Cloudflare (Xem chi tiết ở Phần 5).
> **Cách giấu Port:** DNS không hiểu Port. Hệ thống DNS chỉ làm nhiệm vụ tra danh bạ tìm đường, nó không biết bên trong có port gì. Để giấu Port (ví dụ giấu port 8080 của Golang), bạn phải dùng Nginx làm Reverse Proxy. Nhận port 80/443 từ Cloudflare rồi Nginx tự đẩy ngầm về port 8080. Điều này hoàn toàn ăn khớp với Master Plan Nginx + Cloudflare của bạn!

---

## Phần 5: Tính năng `proxied = true` (Đám mây màu cam của Cloudflare)

Đây là tính năng "ăn tiền" nhất của Cloudflare khi cấu hình DNS, biến Cloudflare từ một cuốn danh bạ (DNS) bình thường thành một tấm khiên bảo vệ (Reverse Proxy).

- **Khi `proxied = false` (Đám mây màu xám ☁️):**
  - Cloudflare chỉ làm đúng chức năng danh bạ (DNS Only).
  - Khi user gõ `api.logistic.com`, Cloudflare trả về đúng IP thật của server AWS (ví dụ: `13.212.x.x`).
  - User kết nối **trực tiếp** vào server AWS. Hacker biết IP thật và có thể tấn công trực tiếp.

- **Khi `proxied = true` (Đám mây màu cam ☁️):**
  - Cloudflare đứng ra làm "bia đỡ đạn" (Reverse Proxy).
  - Khi user gõ `api.logistic.com`, Cloudflare trả về một **IP ảo của Cloudflare** (ví dụ: `104.18.x.x`).
  - Toàn bộ request từ user sẽ đập vào máy chủ của Cloudflare trước. Tại đây, Cloudflare sẽ chặn DDoS, lọc SQL Injection (WAF), cache hình ảnh...
  - Những request "sạch" mới được Cloudflare âm thầm chuyển tiếp về IP thật của server AWS.
  - **Kết quả:** IP thật của máy chủ AWS được giấu kín 100%. User không bao giờ biết được máy chủ thật nằm ở đâu. Hacker cũng không thể ping hay lấy được IP thật để tấn công.

---

## Phần 6: 3 Mô hình kết nối Server với Internet (và Cloudflare)

Có 3 kỹ thuật để đưa máy chủ của bạn ra mạng Internet. Mỗi cách có mức độ bảo mật khác nhau:

### 1. Mô hình Truyền thống (Direct DNS / Đám mây Xám)
- **Cách hoạt động:** Dùng bản ghi DNS (A/CNAME) trỏ trực tiếp về IP Public của máy chủ AWS.
- **Firewall (Security Group):** Phải mở port Ingress (80, 443) cho toàn bộ thế giới (`0.0.0.0/0`).
- **Ưu điểm:** Dễ setup nhất, kết nối trực tiếp.
- **Nhược điểm:** Server bị "trần trụi" trên Internet. Lộ IP thật, dễ dính DDoS, phải tự chống chọi mọi đợt rà quét port của hacker.

### 2. Mô hình Reverse Proxy (Cloudflare Proxy / Đám mây Cam)
- **Cách hoạt động:** Bật `proxied = true`. Cloudflare đứng giữa làm bia đỡ đạn. Trình duyệt gọi Cloudflare -> Cloudflare gọi về Server của bạn. Đây chính là cách bạn đang cấu hình bằng Terraform.
- **Firewall (Security Group):** Vẫn phải mở port Ingress (80, 443). Tuy nhiên, **Best Practice** là không mở cho `0.0.0.0/0` mà chỉ cấu hình cho phép các "Dải IP của Cloudflare" được quyền đi vào, chặn đứng mọi IP khác.
- **Ưu điểm:** Giấu được IP thật của máy chủ, chống DDoS tuyệt vời, có Web Application Firewall (WAF) lọc mã độc, miễn phí chứng chỉ HTTPS/SSL.
- **Nhược điểm:** Vẫn cần cấp IP Public cho máy ảo AWS. Nếu cấu hình Firewall lỏng lẻo (vẫn để `0.0.0.0/0`), hacker vô tình dò ra được IP thật thì chúng sẽ "đi cửa sau" đâm thẳng vào server, vô hiệu hóa lớp bảo vệ Cloudflare.

### 3. Mô hình Cloudflare Tunnel (Zero Trust / Đỉnh cao bảo mật)
- **Cách hoạt động:** Cài một phần mềm tên là `cloudflared` lên thẳng máy chủ AWS. Phần mềm này sẽ chủ động đào một "đường hầm" (Tunnel) kết nối ĐI RA NGOÀI (Outbound) đến mạng lưới của Cloudflare.
- **Firewall (Security Group):** **KHÔNG CẦN MỞ BẤT KỲ CỔNG INGRESS NÀO (đóng cửa hoàn toàn port 80, 443)**. Bạn thậm chí **không cần mua Public IP** cho máy EC2!
- **Ưu điểm:** Bảo mật tối thượng. Server hoàn toàn vô hình trên Internet vì cửa đã khóa kín từ bên ngoài. Không tốn tiền duy trì Public IP. Bất chấp đứt cáp hay đổi nhà mạng vì Tunnel tự động kết nối lại.
- **Nhược điểm:** Phải cài thêm phần mềm `cloudflared` lên server. Cấu hình phức tạp hơn (phải tạo Token cho Tunnel thay vì quản lý bằng IP).

---

## Phần 7: Các thông số "xương sống" để build máy ảo EC2 (`aws_instance`)

Để tạo thành công một máy chủ EC2 (`resource "aws_instance"`), bạn cần lắp ráp các mảnh ghép (thuộc tính) cơ bản sau:

1. **`ami` (Amazon Machine Image - Hệ điều hành):** 
   - Là hệ điều hành nền tảng (ví dụ: Ubuntu, Amazon Linux, Windows).
   - *Bí kíp:* Thay vì điền một chuỗi ID chết cứng (như `ami-0xyz...`), hãy dùng block `data "aws_ami"` để Terraform tự động dò tìm bản cập nhật mới nhất của OS đó, giúp hệ thống luôn an toàn và mới nhất.

2. **`instance_type` (Cấu hình phần cứng):**
   - Quyết định số lượng CPU và RAM. 
   - Ví dụ: `t2.micro` (1 CPU, 1GB RAM - Nằm trong Free Tier), `t3.large` (2 CPU, 8GB RAM). Lựa chọn này sẽ quyết định **giá tiền** bạn phải trả cho AWS!

3. **`vpc_security_group_ids` (Tường lửa):**
   - Danh sách các Security Group gắn vào máy. Nếu không gắn, máy sẽ dùng SG mặc định (thường khóa kín). Bạn cần chỉ định ID của SG mà bạn vừa tạo để mở các cổng cần thiết như SSH (22) và HTTP/S (80, 443).

4. **`key_name` (Chìa khóa SSH - Sống còn):**
   - Tên của cặp khóa bảo mật (Key Pair - file `.pem`) đã tạo trên web AWS.
   - **Hậu quả nếu quên:** AWS vẫn tạo máy thành công, nhưng bạn sẽ **vĩnh viễn không thể đăng nhập (SSH)** vào cái máy đó được vì ổ khóa đã đóng mà bạn không có chìa. Máy ảo đó chỉ có nước đem hủy bỏ.
   - **Lưu ý CỰC KỲ QUAN TRỌNG:** Key Pair là tài sản **của riêng từng Khu vực (Region)**. Tạo Key ở Mỹ (us-east-1) thì không thể dùng để mở khóa máy ảo ở Singapore (ap-southeast-1) được.

5. **`root_block_device` (Ổ cứng chính):**
   - Mặc định AWS chỉ cấp cho hệ điều hành Linux một cái ổ cứng bé tí (8GB). Cần khai báo block này để nâng cấp dung lượng (`volume_size = 30`) và đổi sang loại ổ cứng thể rắn tốc độ cao thế hệ mới (`volume_type = "gp3"`).

6. **`tags` (Gắn nhãn định danh):**
   - Rất quan trọng khi quản lý Cloud. Trong đó thuộc tính `Name = "..."` chính là cái tên sẽ hiển thị to rõ ràng trên giao diện Web của AWS giúp bạn phân biệt máy này với máy khác.

---

## Phần 8: Các lệnh CLI cơ bản & thường dùng của Terraform

Tư duy của Terraform xoay quanh Vòng đời (Lifecycle) cực kỳ đơn giản. 
> **⚠️ LƯU Ý QUAN TRỌNG:** **Terraform KHÔNG CÓ lệnh khởi động lại (restart) máy chủ**. 
> Terraform là công cụ "Xây nhà" và "Đập nhà". Để restart máy, bạn phải dùng giao diện Web của AWS, dùng AWS CLI, hoặc SSH vào máy gõ lệnh `sudo reboot`. 

Tuy nhiên, Terraform cung cấp các lệnh quản lý vòng đời mà bạn sẽ xài mỗi ngày:

### Nhóm lệnh Vòng đời (Sống còn)

1. **`terraform init` (Khởi tạo)**
   - Dùng lần đầu tiên khi kéo code về hoặc khi thêm Provider mới. Lệnh này tải các plugin (như thư viện AWS) về máy. Không chạy lệnh này, các lệnh khác sẽ báo lỗi.

2. **`terraform plan` (Xem trước / Chạy nháp)**
   - Khuyên dùng trước khi gõ `apply`. Nó sẽ in ra bản nháp những gì nó tính làm (Tạo mới `+`, Sửa `~`, Xóa `-`) để bạn kiểm tra xem có sai sót hay xóa nhầm cái gì không mà chưa áp dụng lên hệ thống thật.

3. **`terraform apply` (Thực thi / Bật Lab)**
   - Đọc code, tự động gọi API lên Cloud để xây dựng hệ thống y chang như code viết. Nó sẽ hỏi bạn gõ `yes` trước khi thực sự làm. Đây là lệnh để **Start** dàn lab lên học.

4. **`terraform destroy` (Hủy diệt / Tắt Lab)**
   - Đọc file State và dọn dẹp, **xóa sạch sành sanh** mọi thứ nó từng tạo. Khi đi ngủ gõ lệnh này để AWS ngừng tính tiền. Sáng mai gõ `apply` là có lại hệ thống mới.

### Nhóm lệnh Thao tác & Quản lý

5. **`terraform apply -replace="aws_instance.logistic_server"` (Đập đi xây lại)**
   - Thay vì restart, nếu máy bạn cài bậy bạ bị hỏng Hệ điều hành, bạn xài lệnh này. Terraform sẽ "giết" riêng cái máy EC2 cũ và đẻ ra 1 cái máy EC2 mới tinh khôi thay vào đó. Các thành phần khác (Security Group, Cloudflare) vẫn giữ nguyên.

6. **`terraform fmt` (Làm đẹp code)**
   - Tự động canh lề, canh bằng dấu bằng `=`, thụt đầu dòng cho file `main.tf` đẹp chuẩn format y như DevOps chuyên nghiệp. Cứ code xong gõ lệnh này nhìn file cực kỳ sướng mắt.

7. **`terraform validate` (Kiểm tra lỗi syntax)**
   - Quét lỗi xem bạn gõ ngoặc nhọn `{ }` có đúng không, tên biến có bị sai chính tả không trước khi mất công chạy `apply`. Tương tự như tính năng kiểm tra lỗi (compile) của các ngôn ngữ lập trình.

# Tham khảo cú pháp chính thức của Cloudflare Record tại:
# https://registry.terraform.io/providers/cloudflare/cloudflare/latest/docs/resources/record