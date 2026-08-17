#!/bin/bash
# Sinh ngẫu nhiên ELASTIC_PASSWORD và KIBANA_PASSWORD rồi ghi vào file .env ở gốc repo.
#
# VÌ SAO CẦN SCRIPT NÀY (thay vì để Elasticsearch tự sinh mật khẩu rồi đọc log)?
# Nếu không truyền biến ELASTIC_PASSWORD lúc container Elasticsearch khởi động lần đầu,
# ES sẽ TỰ SINH một mật khẩu ngẫu nhiên và chỉ in ra 1 LẦN DUY NHẤT trong log container.
# Muốn lấy lại phải "docker logs" rồi grep bằng tay, hoặc chạy "elasticsearch-reset-password"
# — cách này không tự động hoá được, dễ lỡ tay bỏ sót, không hợp để đưa vào quy trình deploy.
#
# Script này làm ngược lại: TỰ SINH mật khẩu TRƯỚC khi Elasticsearch khởi động, ghi vào .env.
# Elasticsearch đọc ELASTIC_PASSWORD từ .env để bootstrap user "elastic" ngay từ đầu (không
# đoán/không random nữa). Container "es-setup-kibana-user" trong docker-compose.yml sau đó
# tự động dùng ELASTIC_PASSWORD để gọi Security API, gán KIBANA_PASSWORD cho user hệ thống
# "kibana_system", rồi Kibana khởi động và dùng luôn giá trị đó — không ai phải copy tay.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENV_FILE="$ROOT_DIR/.env"

if [ ! -f "$ENV_FILE" ]; then
	echo "[generate-es-secrets] Không tìm thấy $ENV_FILE" >&2
	exit 1
fi

gen_pass() {
	# 24 ký tự base64, lọc bỏ các ký tự dễ gây lỗi khi nhét vào YAML/JSON/shell (/ + =).
	openssl rand -base64 32 | tr -d '/+=' | cut -c1-24
}

set_env_var() {
	local key="$1" value="$2"
	if grep -q "^${key}=" "$ENV_FILE"; then
		# In-place edit, tương thích cả GNU sed (Linux) lẫn BSD sed (macOS).
		sed -i.bak "s|^${key}=.*|${key}=${value}|" "$ENV_FILE" && rm -f "$ENV_FILE.bak"
	else
		printf '%s=%s\n' "$key" "$value" >>"$ENV_FILE"
	fi
}

current_elastic="$(grep '^ELASTIC_PASSWORD=' "$ENV_FILE" | cut -d= -f2-)"
current_kibana="$(grep '^KIBANA_PASSWORD=' "$ENV_FILE" | cut -d= -f2-)"

if [ -n "$current_elastic" ] && [ -n "$current_kibana" ]; then
	echo "[generate-es-secrets] ELASTIC_PASSWORD / KIBANA_PASSWORD đã có sẵn trong .env, bỏ qua."
	echo "[generate-es-secrets] Muốn sinh lại: xoá rỗng 2 dòng đó trong .env rồi chạy lại script."
	exit 0
fi

set_env_var "ELASTIC_PASSWORD" "$(gen_pass)"
set_env_var "KIBANA_PASSWORD" "$(gen_pass)"

echo "[generate-es-secrets] Đã sinh xong ELASTIC_PASSWORD / KIBANA_PASSWORD vào .env"
echo "[generate-es-secrets] Chạy tiếp: docker compose up -d elasticsearch es-setup-kibana-user kibana"
