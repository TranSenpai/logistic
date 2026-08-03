#!/bin/bash
# Lệnh set -e: Bắt buộc script phải DỪNG NGAY LẬP TỨC (exit) nếu có bất kỳ dòng lệnh nào ở dưới chạy bị lỗi. 
# Điều này giúp ta không bị tình trạng lỗi dây chuyền (lỗi tạo user nhưng vẫn báo thành công).
set -e

echo "=== CREATE USER REPLICATION FOR MASTER ==="

# 1. Tạo User Replication thông qua tiện ích psql (PostgreSQL Client)
# -v ON_ERROR_STOP=1: Cờ này báo cho psql biết nếu câu lệnh SQL bên trong bị lỗi (ví dụ sai cú pháp) 
# thì psql phải trả về mã lỗi (exit code 1) thay vì mã 0. 
# Nhờ mã lỗi này mà lệnh 'set -e' ở trên mới hoạt động được.

# --username và --dbname: Đăng nhập với tư cách là superuser (root).
# <<-EOSQL ... EOSQL: Đây gọi là kỹ thuật "Here Document" (Heredoc). 
# Nó cho phép ta truyền trực tiếp nhiều dòng lệnh SQL (nằm giữa 2 chữ EOSQL) 
# vào bên trong psql mà không cần tạo ra một file .sql trung gian ở máy tính.
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE USER $REPLICATION_USER WITH REPLICATION ENCRYPTED PASSWORD '$REPLICATION_PASSWORD';
EOSQL

# 2. Cấp quyền tường lửa (pg_hba.conf - Host Based Authentication)
# Mặc định, Postgres sẽ từ chối MỌI kết nối lạ vào database để bảo vệ an toàn.
# Dòng lệnh dưới đây sẽ "nhồi" thêm một quy tắc mạng (rule) vào cuối file cấu hình tường lửa pg_hba.conf.
# Giải thích cấu trúc rule "host replication <user> 0.0.0.0/0 md5":
# - 'host'        : Cho phép kết nối qua mạng TCP/IP.
# - 'replication' : Chỉ định loại kết nối là Replication (chứ không phải kết nối truy vấn Data bình thường).
# - '<user>'      : Tên tài khoản được phép dùng luật này.
# - '0.0.0.0/0'       : Chấp nhận mọi địa chỉ IP (vì ta đang dùng Docker network ảo, không cần chặn IP cứng).
# - 'scram-sha-256'   : Cơ chế xác thực Challenge-Response (An toàn hơn MD5, mặc định từ Postgres 14+).
#
# [CƠ CHẾ HOẠT ĐỘNG CỦA SCRAM-SHA-256 (Challenge-Response)]:
# LƯU Ý: Đây KHÔNG PHẢI là mô hình RSA (Public/Private Key). Nó hoạt động theo mô hình Thách Thức - Đáp Trả:
# 1. Replica (Client) gõ cửa: "Cho tôi đăng nhập tài khoản Replication".
# 2. Master (Server) gửi ra một lời thách thức (Challenge) gồm một chuỗi ngẫu nhiên (Salt).
# 3. Replica lấy [Password thật] trộn với [Salt của Master], rồi đem băm nát bằng thuật toán SHA-256, tạo ra một cục [Hash] gửi ngược lại cho Master.
# 4. Master tự tính toán cục [Hash] tương tự ở phía mình. Nếu 2 cục Hash khớp nhau -> Cho phép kết nối.
# => KẾT QUẢ: Mật khẩu thật KHÔNG BAO GIỜ truyền qua mạng, và Hacker không thể copy cục Hash cũ để xài lại (vì mỗi lần kết nối Master sẽ cấp một chuỗi Salt ngẫu nhiên khác nhau).
#
# Dấu '>>' mang ý nghĩa Ghi Tiếp (Append) vào cuối file, tuyệt đối không dùng dấu '>' vì nó sẽ xóa sạch file tường lửa gốc.

# 2.1 Tạo biến chứa nội dung chuỗi cấu hình (Để code dễ đọc hơn)
FIREWALL_RULE="host replication $REPLICATION_USER 0.0.0.0/0 scram-sha-256"

# 2.2 Ghi nội dung biến đó vào cuối file cấu hình
echo "$FIREWALL_RULE" >> "$PGDATA/pg_hba.conf"

echo "=== CREATE USER REPLICATION FOR MASTER SUCCESSFULLY ==="
