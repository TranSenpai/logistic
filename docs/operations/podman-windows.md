# Chạy cụm bằng Podman trên Windows

`docker-compose.yml` chạy được với Podman, nhưng Podman trên WSL2 cần một chỉnh
sửa. Tài liệu này ghi lại cấu hình đã kiểm chứng hoạt động.

## Triệu chứng nếu chưa cấu hình

```
Error response from daemon: netavark (exit code 1): nftables error:
"nft" did not return successfully while applying ruleset:
```

Container build xong, tạo xong, nhưng không khởi động được.

## Nguyên nhân

`netavark` (lớp mạng của Podman) áp ruleset nftables mà kernel WSL2 mặc định không
xử lý được trọn vẹn. Kernel **có** `CONFIG_NF_TABLES=y` và tạo được bảng NAT thủ
công, nên đây là vấn đề ở một rule cụ thể của netavark chứ không phải thiếu hỗ trợ
nftables.

Kiểm chứng nhanh: `podman run --rm --network=host alpine echo ok` chạy được, còn
với bridge network thì lỗi.

## Cấu hình đã kiểm chứng

Tắt firewall driver của netavark. Bridge network và DNS nội bộ vẫn hoạt động; NAT
ra Internet do WSL lo.

```bash
printf '[network]\nfirewall_driver = "none"\n' > /tmp/90-firewall.conf
```

```bash
cat /tmp/90-firewall.conf | podman machine ssh 'sudo tee /etc/containers/containers.conf.d/90-firewall.conf'
```

```bash
podman machine stop && podman machine start
```

> **Viết file bằng cách này, không dùng PowerShell.** Chuỗi lồng nhau trong
> PowerShell bị escape sai và tạo file hỏng ở phía Windows
> (`%APPDATA%\containers\containers.conf.d\`), khiến `podman` từ chối chạy với lỗi
> `toml: unexpected EOF`. Nếu lỡ tạo, xoá file đó đi là podman hoạt động lại.

## Phải chạy ở chế độ rootless

| Chế độ | Bridge network | Publish port ra Windows |
|---|---|---|
| rootless (mặc định) | có | **có** |
| rootful | có | không |

Với `firewall_driver = "none"`, chế độ rootful mất khả năng publish port vì không
còn rule DNAT. Rootless chuyển tiếp cổng bằng tiến trình ở không gian người dùng,
không phụ thuộc nftables — nên vẫn gọi được `http://localhost:8080` từ Windows.

Kiểm tra:

```bash
podman info --format "rootless={{.Host.Security.Rootless}}"
```

Nếu ra `false`:

```bash
podman machine stop && podman machine set --rootful=false && podman machine start
```

> **Rootless và rootful dùng kho lưu trữ riêng biệt.** Đổi qua lại là container,
> image và volume của chế độ kia không còn nhìn thấy — dữ liệu không mất, chỉ nằm
> ở namespace khác. Sau khi đổi, dữ liệu trong database phải tạo lại.

## Tài nguyên máy ảo

`podman machine list` hiển thị RAM/CPU ghi lúc tạo máy, **không phải** giá trị
thực. WSL2 tự cấp khoảng 50% RAM của máy thật. Kiểm tra thực tế:

```bash
podman machine ssh "free -h | head -2; nproc"
```

`podman machine set --cpus/--memory` **không dùng được** cho máy WSL — Podman báo
`changing CPUs not supported for WSL machines`. Muốn đổi thì sửa `%USERPROFILE%\.wslconfig`:

```ini
[wsl2]
memory=12GB
processors=8
```

Rồi `wsl --shutdown` và khởi động lại máy ảo.

## Khởi động

```bash
make auth-keys
```

```bash
podman compose up -d
```

Cụm đầy đủ khá nặng (3 Kafka broker, Elasticsearch, Kibana, 6 database). Chạy một
tập con khi chỉ cần thử luồng chính:

```bash
podman compose up -d auth-db-master auth-service gateway-service
```

## Kiểm chứng

```bash
curl http://localhost:8080/healthz
```

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@logistic.vn","full_name":"Demo","password":"Matkhau@12345","role":"shipper"}'
```

Đăng ký trả về `201` kèm `id` dạng UUID v7 (bắt đầu bằng `01a…` vì 48 bit đầu là
timestamp) nghĩa là chuỗi gateway → gRPC → auth_service → MySQL đã thông.

## Lỗi thường gặp

| Lỗi | Nguyên nhân | Cách xử lý |
|---|---|---|
| `machine ... already exists` | `podman machine init` khi máy đã tồn tại | Không cần init lại; dùng `podman machine set` |
| `cannot be destroyed` | `podman machine rm` khi máy đang chạy | `podman machine stop` trước |
| `unknown driver "mysql"` | Thiếu driver trong service | Đã sửa — auth_service import cả `go-sql-driver/mysql` và `lib/pq` |
| `connection refused` tới database | Service khởi động trước khi DB sẵn sàng | Đã sửa — `auth-db-master` có healthcheck, `auth-service` có `depends_on: service_healthy` |
| Gateway trả `504 DEADLINE_EXCEEDED` | Service đích chưa chạy | Đúng hành vi: hạn chờ 5s thay vì treo vô hạn |
