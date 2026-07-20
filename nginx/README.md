# 🚀 Cẩm Nang Nginx: Kỹ Thuật Reverse Proxy (Lễ Tân Phân Luồng)

Nginx là một trong những Web Server mạnh mẽ và phổ biến nhất thế giới. Trong kiến trúc Microservices, Nginx thường đóng vai trò là **Reverse Proxy** (Đứng mũi chịu sào, phân luồng giao thông).

## 1. Tư duy cốt lõi của Nginx Reverse Proxy
Hãy tưởng tượng Nginx là một cô Lễ Tân tòa nhà:
- **`listen`**: Cô ấy đứng gác ở cửa số mấy? (Thường là cửa `80` cho HTTP, `443` cho HTTPS bảo mật).
- **`server_name`**: Cô ấy phục vụ cho công ty nào? (Nhận diện Tên miền, ví dụ `api.glolog.dev`).
- **`location`**: Khách hỏi đi bộ phận nào? (Đường dẫn Route, ví dụ khách muốn đi `/auth` hay `/media`).
- **`proxy_pass`**: Cô ấy chỉ đường cho khách vào phòng nào bên trong tòa nhà? (Chuyển tiếp ngầm tới `http://localhost:8080`).

---

## 2. Cú pháp (Syntax) Kinh điển

Dưới đây là bộ khung chuẩn mực mà bạn có thể tham khảo để tự tay gõ lại từ tờ giấy trắng cho bất kỳ dự án nào:

```nginx
server {
    # 1. Đứng gác ở cổng 80 (HTTP)
    listen 80;
    
    # 2. Đón khách từ tên miền này (Nếu test ở máy tính thì đổi thành localhost)
    server_name api.glolog.dev; 

    # 3. Phân luồng cho khách truy cập đường dẫn /auth/
    location /auth/ {
        # Đẩy khách vào cổng 8080 của con App chạy bằng Docker bên trong
        proxy_pass http://localhost:8080/;
        
        # 4 DÒNG BẮT BUỘC: Truyền IP thật của Khách hàng vào cho Backend
        # Nếu không có, Backend sẽ tưởng Nginx (localhost) là người truy cập!
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

---

## 3. Các lệnh vòng đời Nginx trên Linux (Ubuntu Server)
Khi làm việc trên máy ảo EC2, đây là các lệnh bạn sẽ gõ mòn bàn phím:
- **Tạo/Sửa code Nginx:** `sudo nano /etc/nginx/sites-available/<tên_file>`
- **Bật cấu hình (Tạo Link):** `sudo ln -s /etc/nginx/sites-available/<tên_file> /etc/nginx/sites-enabled/`
- **Tắt cấu hình (Dọn rác):** `sudo rm /etc/nginx/sites-enabled/<tên_file>`
- **Kiểm tra lỗi chính tả (CỰC QUAN TRỌNG):** `sudo nginx -t` (Phải gõ cái này trước khi khởi động lại).
- **Áp dụng cấu hình mới (Không làm sập web):** `sudo systemctl reload nginx`
- **Khởi động lại toàn bộ:** `sudo systemctl restart nginx`
- **Xem trạng thái sống chết:** `sudo systemctl status nginx`

---

## 4. Tài liệu chính chủ (Nguồn tu luyện)
Khi bạn quên cú pháp, đừng cố nhớ, hãy tra cứu tại các trang gốc này:
1. **Tài liệu Reverse Proxy chính thức của Nginx (Rất dễ hiểu):**
   👉 https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy/
2. **Cấu hình Nginx Boilerplate (Mẫu chuẩn thế giới):**
   👉 https://github.com/h5bp/server-configs-nginx (Kho tàng các cấu hình tối ưu sẵn về bảo mật, nén file, chống cache mà các pháp sư DevOps hay copy dùng).

---

## 5. Mẹo Test Nginx tại máy tính (Local Windows)
Bạn không cần thiết phải cài Nginx native lên máy Windows (vì nó rất khác với môi trường Linux và dễ sinh lỗi). Hãy tận dụng chính Podman để tạo một cô Lễ Tân Nginx ngay trên Windows của bạn:

```bash
# Đứng ở thư mục chứa file logistic.conf (ví dụ thư mục nginx) và gõ lệnh sau:
podman run --name nginx-test -p 80:80 -v ./logistic.conf:/etc/nginx/conf.d/default.conf -d docker.io/library/nginx:alpine
```
Sau đó mở trình duyệt gõ `http://localhost/auth` để xem Nginx có điều hướng đúng vào con App Backend đang chạy ở máy bạn không nhé!
