# Media Service - Logistics OS

Media Service là vi dịch vụ (microservice) chịu trách nhiệm xử lý toàn bộ các tác vụ liên quan đến File tĩnh (Hình ảnh, Giấy tờ, Báo cáo CSV/Excel) của hệ thống Logistics. Dịch vụ này sử dụng giao thức HTTP/REST (Gin) để tối ưu hóa việc truyền tải file dung lượng lớn thay vì dùng gRPC.

---

## 1. Kiến trúc luồng Upload Ảnh (Streaming & Chống OOM)

Vấn đề muôn thuở của các server khi cho phép người dùng upload file là **OOM (Out Of Memory - Tràn bộ nhớ)**. Nếu người dùng upload một file Video Camera hành trình nặng 2GB, và server cố gắng đọc toàn bộ 2GB đó vào một biến `[]byte` trên RAM, server sẽ sập ngay lập tức.

Để giải quyết, GoBackend áp dụng cơ chế **Streaming** kết hợp `multipart/form-data` và `io.Reader`.

### Cơ chế hoạt động chi tiết:
1. **Nhận dữ liệu (Tầng HTTP Delivery):**
   Client gửi request dạng `multipart/form-data`. Server hứng request bằng Gin và lấy ra cấu trúc `*multipart.FileHeader`.
   *Lưu ý:* `FileHeader` chỉ là "Phiếu thông tin" (chứa tên file, dung lượng, MIME type), nó KHÔNG chứa nội dung file. Tại thời điểm này, RAM vẫn an toàn.

2. **Mở luồng dữ liệu (Tầng Storage):**
   Khi gọi `fileHeader.Open()`, Go trả về một object implement interface `multipart.File` (về bản chất là một `io.Reader`).

3. **Múc từng gáo nước (Stream to Cloudinary):**
   Ta truyền thẳng cái `io.Reader` này vào hàm `Upload()` của SDK Cloudinary. Nhờ sức mạnh của `io.Reader` native trong Golang, dữ liệu được truyền theo kiểu "nước chảy qua ống":
   - Go đọc một cục nhỏ (chunk/bucket) khoảng vài KB từ ổ cứng tạm hoặc stream mạng.
   - Nạp cục đó thẳng xuống Tầng mạng (Network Socket) để bơm sang Server của Cloudinary.
   - Lặp lại liên tục cho đến hết file.
   => **Kết quả:** Dù file có nặng 10GB, lượng RAM mà server Golang tiêu thụ tại một thời điểm cũng chỉ loanh quanh vài chục KB. Chống OOM tuyệt đối!

---

## 2. Kết quả trả về & Chiến lược Xử lý Lỗi (Error Handling)

Sau khi stream xong lên Cloudinary, hàm `Upload` trả về 3 thông số chính và 1 error:

*   **`fileName`**: Tên ngẫu nhiên được sinh ra (kèm timestamp) giúp tránh trùng lặp. Cần lưu lại để biết file gốc tên gì.
*   **`publicID`**: Khóa chính (Primary Key) mà Cloudinary định danh cho tấm ảnh này. Đây là **chìa khóa bắt buộc phải có** nếu sau này muốn xóa (Delete) hoặc chỉnh sửa ảnh.
*   **`url` (SecureURL):** Đường dẫn HTTPS trực tiếp đến CDN của Cloudinary. Ta sẽ lưu chuỗi URL này vào CSDL (PostgreSQL) ở các service khác để Frontend lấy ra hiển thị.

### Bắt lỗi 3rd Party an toàn:
Với các dịch vụ bên thứ ba (3rd party API), đôi lúc request HTTP vẫn trả về status 200 OK, nhưng bên trong body lại chứa thông báo lỗi của dịch vụ đó.
*   Sử dụng `%w` với `fmt.Errorf` để gói (wrap) lỗi đường truyền mạng gốc, giúp giữ lại StackTrace cho việc debug.
*   Chủ động bắt `result.Error.Message`. Nếu thuộc tính này có giá trị, nghĩa là SDK Cloudinary đã báo lỗi nghiệp vụ (Vd: Sai thư mục, file bị cấm). Ta chủ động log lỗi này ra màn hình (`log.Printf`) và return lỗi tùy chỉnh để hệ thống biết mà rollback, tránh việc lưu một cái URL hỏng vào Database.

---

## 3. Luồng Xóa Ảnh (Delete Workflow)

Tương tự như Upload, API Delete nhận vào `public_id`.
*   Tầng Storage gọi hàm `Destroy` của Cloudinary SDK.
*   Gắn Context Timeout (VD: 10 giây) để tránh việc connection bị kẹt (hanging) gây rò rỉ Goroutine nếu server Cloudinary phản hồi chậm.
*   Log trạng thái trả về `result.Result` (ví dụ: `ok`, `not found`) để phục vụ tracking. Nếu Cloudinary trả về "not found" (file không tồn tại), hệ thống vẫn có thể quyết định bỏ qua vì mục tiêu cuối cùng (ảnh không còn trên cloud) đã đạt được, nhưng phải ghi log lại để rà soát.
