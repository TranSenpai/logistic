# QUY TRÌNH TRIỂN KHAI VẬN HÀNH (DEVOPS MASTER WORKFLOW)

Tài liệu này chuẩn hóa quy trình (Workflow) triển khai hạ tầng từ số không (Zero) đến khi vận hành toàn diện một hệ thống phần mềm phân tán (Distributed System) trên nền tảng Cloud. Quy trình được thiết kế theo tư duy Cơ sở hạ tầng dưới dạng Mã (Infrastructure as Code - IaC).

---

## GIAI ĐOẠN 1: KHỞI TẠO TIỀN ĐỀ (PRE-REQUISITES)
Giai đoạn chuẩn bị định danh, quyền hạn và khóa mã hóa trước khi tương tác với Cloud API.

1. **Quản trị Tên miền (Domain Management):** 
   - Đăng ký tên miền (Ví dụ: `glolog.dev`) và cấu hình Name Server (NS) chuyển giao quyền quản lý về Cloudflare.
   - Cấp phát **API Token** từ Cloudflare Dashboard với quyền hạn can thiệp vào DNS Record để phục vụ Terraform.
2. **Quản trị Cloud Provider (Ví dụ: AWS / GCP):**
   - Khởi tạo **Key Pair (Chìa khóa SSH)** định dạng `.pem`. Đây là phương thức xác thực phi đối xứng (Asymmetric Cryptography) bắt buộc để truy cập máy chủ (Vô hiệu hóa xác thực bằng mật khẩu truyền thống).
   - Thiết lập IAM User, cấp quyền Programmatic Access và tải bộ Credentials (`Access Key ID` & `Secret Access Key`).
3. **Cấu hình Môi trường Local:**
   - Liên kết môi trường phát triển cục bộ với Cloud API thông qua CLI (`aws configure`).

---

## GIAI ĐOẠN 2: CẤP PHÁT HẠ TẦNG TỰ ĐỘNG (INFRASTRUCTURE PROVISIONING)
Sử dụng Terraform để định nghĩa hạ tầng dưới dạng mã nguồn (IaC). Trình tự logic quy hoạch tài nguyên yêu cầu tuân thủ nguyên lý Đồ thị phụ thuộc (Directed Acyclic Graph - DAG).

1. **Định nghĩa Kiến trúc Mạng (Networking & Security):**
   - Thiết lập Data Source để truy vấn tự động ID của OS Image (AMI) mới nhất (Ví dụ: Ubuntu 24.04 LTS).
   - Khởi tạo Security Group (Firewall). Giới hạn luồng Ingress tại Port 22 (Quản trị mạng riêng) và Port 80/443 (Giao thông Web).
2. **Khởi tạo Nút Điện toán (Compute Instance):**
   - Cấp phát Máy ảo. Khai báo ràng buộc chặt chẽ (Binding) với ID của Security Group và tên của Key Pair (SSH).
   - Định nghĩa quy mô Block Storage (Ví dụ: 30GB, chuẩn `gp3`).
3. **Phân giải Tên miền (DNS Configuration):**
   - Tạo DNS Record (A Record) trên hệ thống Cloudflare.
   - Ánh xạ động Public IP của máy ảo vừa được Terraform cấp phát sang Subdomain. Kích hoạt cờ Proxy (Edge Reverse Proxy).
4. **Thực thi Pipeline (CI/CD Local):**
   - Lệnh `terraform init`: Tải thư viện Provider API.
   - Lệnh `terraform plan`: Phân tích DAG và in ra Execution Plan.
   - Lệnh `terraform apply`: Thực thi gọi Cloud API để cấp phát toàn bộ hạ tầng thực tế.

---

## GIAI ĐOẠN 3: CẤU HÌNH MÔI TRƯỜNG NODE (CONFIGURATION MANAGEMENT)
Sau khi phần cứng và mạng lưới đã được Provider cấp phát thành công, tiến hành truy cập máy chủ để cấu hình hệ điều hành và cài đặt Runtime Dependencies.

1. **Xác thực Mạng Khách (SSH Authentication):**
   - Kết nối tới Public IP gốc (Tuyệt đối không kết nối qua Tên miền do rào cản TLS Termination từ Cloudflare Proxy).
   - Cú pháp: `ssh -i <path_to_pem_file> ubuntu@<PUBLIC_IP>`
2. **Cập nhật Nhân Hệ điều hành:**
   - Quét và cập nhật các bản vá lỗi bảo mật (Security Patches): `sudo apt update && sudo apt upgrade -y`
3. **Thiết lập Môi trường Chạy ứng dụng (Runtime Setup):**
   - Cài đặt **Nginx**: Cấu hình Service để đóng vai trò Reverse Proxy nội bộ chặn tại Port 80.
   - Cài đặt **Podman/Docker**: Nền tảng Containerization để cô lập các tiến trình Microservices khỏi hệ điều hành vật lý.

---

## GIAI ĐOẠN 4: TRIỂN KHAI ỨNG DỤNG (APPLICATION DEPLOYMENT)
Quy trình khởi động hệ thống phần mềm (Backend, Message Queue, Database) trên hạ tầng mạng lưới đã chuẩn bị.

1. **Triển khai Mã nguồn:** Đồng bộ Source Code từ Git Repository về máy chủ. Khởi tạo tệp tin `.env` chứa cấu hình môi trường và Secrets.
2. **Điều hướng Lưu lượng (Reverse Proxy Routing):**
   - Cấu hình file `nginx.conf` (hoặc `/etc/nginx/sites-available`).
   - Thiết lập nguyên lý: Nginx lắng nghe tại Port 80, bóc tách Request và điều hướng `proxy_pass` tới các Port của Container Backend (Ví dụ: 8080 cho Auth Service, 8082 cho Media Service).
3. **Vận hành Container (Orchestration):**
   - Triển khai toàn bộ cụm Microservices thông qua cấu hình `docker-compose.yml` (hoặc `podman-compose up -d`).
   - Mạng nội bộ (Private Bridge Network): Các Container giao tiếp chéo với nhau qua Bridge Network nhưng bị cô lập mạng (Network Isolation) hoàn toàn với thế giới Internet.

---

## GIAI ĐOẠN 5: BẢO TRÌ & THU HỒI TÀI NGUYÊN (MAINTENANCE & TEARDOWN)
- Khi hoàn tất phiên làm việc, giai đoạn kiểm thử, hoặc cần hủy bỏ hạ tầng, thực thi thu hồi toàn bộ tài nguyên bằng lệnh `terraform destroy`.
- Trình biên dịch Terraform sẽ đối chiếu với tệp tin State (Trạng thái), tự động đảo ngược cấu trúc phụ thuộc và hủy bỏ (De-provision) toàn bộ máy ảo, tường lửa, và bản ghi DNS để triệt tiêu chi phí Cloud.

---

## PHỤ LỤC: CHEATSHEET LỆNH CONTAINER CLI (PODMAN/DOCKER)
Danh sách các câu lệnh chẩn đoán (Diagnostics) cốt lõi dành cho DevOps Engineer:

- **Thâm nhập Không gian Container (Exec/Dive):**
  - Mục đích: Chẩn đoán file hệ thống, cấu hình biến môi trường bên trong Namespace cách ly.
  - Cú pháp: `podman exec -it <container_name> /bin/sh` (hoặc `/bin/bash`).
- **Theo dõi Nhật ký Giao dịch (Tail Logs):**
  - Theo dõi realtime: `podman logs -f <container_name>`
- **Kiểm tra trạng thái Vòng đời (Process Status):**
  - Xem Container đang hoạt động: `podman ps`
  - Xem toàn bộ lịch sử Container: `podman ps -a`
- **Giải phóng Bộ nhớ & Rác Hệ thống (Pruning):**
  - Dọn dẹp cache build, Image mồ côi: `podman system prune -f`
