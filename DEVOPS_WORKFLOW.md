# 🚀 CẨM NANG DEVOPS TỪ ZERO ĐẾN HERO (MASTER WORKFLOW)

Tài liệu này tổng hợp toàn bộ vòng đời (Workflow) các bước từ lúc bắt đầu với hai bàn tay trắng cho đến khi dựng thành công một hệ thống hoàn chỉnh trên Cloud. Bạn có thể áp dụng luồng tư duy này cho bất kỳ nền tảng nào (AWS, Google Cloud, Azure).

---

## BƯỚC 1: TIỀN TRẠM (Chuẩn bị Vũ khí & Tài nguyên)
Trước khi đụng vào code, bạn cần chuẩn bị sẵn các "nguyên liệu" cơ bản:
1. **Mua Tên miền (Domain):** Có thể mua tại Namecheap, GoDaddy, Hostinger... Tuy nhiên, **khuyên dùng mua thẳng trên Cloudflare** để tiết kiệm thời gian cấu hình.
2. **Đăng ký / Cấu hình Cloudflare:** 
   - *Nếu mua ở nơi khác:* Phải trỏ Name Server (NS) của nhà cung cấp đó về Cloudflare để bàn giao quyền quản lý danh bạ.
   - *Nếu mua trực tiếp trên Cloudflare:* Hệ thống tự động thiết lập sẵn toàn bộ, bạn bỏ qua bước trỏ NS này.
   - Cuối cùng: Lấy **API Token** của Cloudflare để giao cho Terraform điều khiển.
3. **Đăng ký Cloud Provider (AWS / GCP / Azure):**
   - Tạo tài khoản, liên kết thẻ Visa.
   - Tạo **Chìa khóa SSH (Key Pair)** và tải file `.pem` về máy để sau này còn có chìa khóa mở cửa server.
   - Tạo IAM User (hoặc Access Key) lấy **Access Key ID** và **Secret Access Key** để cấp quyền cho máy tính ở nhà điều khiển Cloud.
4. **Kết nối máy tính với Cloud:** Mở Terminal gõ `aws configure` (hoặc lệnh tương ứng của Google/Azure) và dán Key vào.

---

## BƯỚC 2: INFRASTRUCTURE AS CODE (Xây nhà bằng Code)
Tuyệt đối KHÔNG click chuột thủ công trên web (vì sẽ quên và khó quản lý). Dùng **Terraform** để định nghĩa vạn vật:
1. **Viết file `main.tf`:**
   - **Provider:** Khai báo nhà cung cấp (AWS, Cloudflare...).
   - **Security Group (Tường lửa):** Mở cửa những cổng cần thiết (Port 22 cho SSH, Port 80/443 cho Web).
   - **Virtual Machine (Máy ảo):** Khai báo máy ảo (như EC2 của AWS hay Compute Engine của GCP). Nhớ điền cấu hình RAM/CPU, hệ điều hành (Ubuntu), dung lượng ổ cứng, và **đặc biệt không được quên khai báo tên Chìa khóa SSH (`key_name`)**.
   - **DNS Record:** Viết code nhờ Cloudflare trỏ tên miền (ví dụ `api.domain.com`) về IP Public của máy ảo vừa tạo, bật `proxied = true` để giấu IP.
2. **Thực thi pháp thuật:**
   - Gõ `terraform init` (Tải đồ nghề).
   - Gõ `terraform apply` (Tiến hành xây nhà tự động).

---

## BƯỚC 3: CONNECT & SETUP ENV (Vào nhà & Sắm nội thất)
Nhà đã xây xong, giờ là lúc bạn xách vali (mở Terminal) bước vào trong nhà.
1. **SSH vào Server:**
   - Lệnh: `ssh -i "chia-khoa.pem" ubuntu@IP_PUBLIC_CUA_MAY_AO`
   - *Lưu ý xương máu:* Vì đã dùng tính năng Proxy ẩn IP của Cloudflare, bạn KHÔNG THỂ ssh bằng tên miền (vì nó sẽ bay vào máy chủ của Cloudflare và bị văng ra). Phải SSH bằng IP thật của máy ảo.
2. **Cập nhật Hệ điều hành:**
   - Lệnh: `sudo apt update && sudo apt upgrade -y`
3. **Cài đặt hệ sinh thái:**
   - Web Server / Reverse Proxy: Cài đặt **Nginx** để đứng gác cổng, hứng traffic từ mạng vào và điều hướng.
   - Runtime / Container: Cài đặt **Docker, Podman, Git** để chạy code Backend mà không sợ rác hệ điều hành.

---

## BƯỚC 4: DEPLOY APPLICATION (Khai trương)
Giai đoạn đưa Source code (Backend, DB) lên chạy thực tế:
1. **Kéo Code về:** Dùng Git clone source code từ Github về máy ảo.
2. **Cấu hình Nginx (Reverse Proxy):** Viết file cấu hình Nginx để nó hứng yêu cầu từ Port 80/443 và đẩy ngầm về các Port nội bộ (ví dụ giấu port 8080 của Golang hay 9092 của Kafka).
3. **Chạy Services:** Dùng Podman/Docker khởi động các container:
   - Database (PostgreSQL).
   - Message Queue (Kafka KRaft).
   - Backend Service (Golang Auth Service, Logistic Service).
   - Logging System (Elasticsearch, Kibana).

---

## BƯỚC 5: CLEAN UP (Tắt điện, dọn dẹp)
Đặc quyền của việc dùng Terraform (Config as Code):
- Khi làm Lab xong hoặc cuối tuần không xài tới, chỉ cần mở Terminal trên máy tính gõ: `terraform destroy`.
- Terraform sẽ quét sạch mọi thứ (máy ảo, tường lửa, DNS) trả lại mặt bằng trống không để AWS ngừng tính tiền. 
- Lần sau muốn làm việc tiếp? Chỉ việc gõ `terraform apply`, đi pha ly cafe quay lại là nguyên hệ thống đồ sộ lại tự động mọc lên y như cũ!

---

## 🛠 CHEATSHEET: CÁC LỆNH DOCKER / PODMAN SINH TỒN
Dưới đây là các câu lệnh "cứu mạng" khi bạn thao tác với Container:

**1. "Dive" (Chui) vào bên trong Container đang chạy:**
Đây là lệnh quan trọng nhất để debug xem bên trong app có file gì, biến môi trường ra sao.
- Cú pháp chuẩn: `podman exec -it <tên_container> /bin/sh` (hoặc `/bin/bash` nếu hệ điều hành hỗ trợ).
- *Ví dụ chui vào con Auth Service (dùng Alpine nên chỉ có `/bin/sh`):*
  ```bash
  podman exec -it logistic-auth /bin/sh
  ```
- *Ví dụ chui vào con Postgres (chạy bash):*
  ```bash
  podman exec -it logistic-postgres /bin/bash
  ```
*(Vào xong gõ lệnh `ls` để xem file, `env` để xem biến môi trường. Muốn thoát ra gõ `exit`).*

**2. Xem Log (Nhật ký lỗi) của App:**
- Lệnh: `podman logs -f <tên_container>`
- *Ví dụ:* `podman logs -f logistic-auth` (Cờ `-f` giúp log tự nhảy liên tục như thật).

**3. Xem danh sách Container đang chạy:**
- Lệnh: `podman ps` (Chỉ xem những con đang chạy).
- Lệnh: `podman ps -a` (Xem cả những con đã chết/bị tắt).

**4. Khởi động / Dừng Container thủ công:**
- Lệnh: `podman start <tên_container>`
- Lệnh: `podman stop <tên_container>`

**5. Dọn dẹp Rác (Gỡ Image/Container thừa):**
- Xóa mọi thứ lơ lửng không dùng đến (Image mồ côi, cache build): `podman system prune -f`
