# ==============================================================================
# Variables & Configuration
# ==============================================================================
PROTO_DIR = api
PROTO_GEN_OUT = internal/dto/gen/go

ENT_DIR = ./ent
ENT_SCHEMA_DIR = $(ENT_DIR)/schema

# Metadata
SERVICE_NAME = goBackend

.PHONY: all help api clean-api db clean-db update-db build-all

# ==============================================================================
# Global Targets
# ==============================================================================

## all: Dọn dẹp và tái tạo toàn bộ Code (API & DB)
all: clean-api clean-db api db
	@echo "==> [$(SERVICE_NAME)] All artifacts generated successfully."

## help: Hiển thị hướng dẫn sử dụng
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

# ==============================================================================
# API & Protobuf Layer
# ==============================================================================

## api: Generate Go & gRPC code từ file .proto
api: clean-api
	@echo "==> Generating Protobuf code..."
	@mkdir -p $(PROTO_GEN_OUT)
	protoc --proto_path=$(PROTO_DIR) \
		--go_out=$(PROTO_GEN_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_GEN_OUT) --go-grpc_opt=paths=source_relative \
		$(shell find $(PROTO_DIR) -name "*.proto")
	@echo "    Protobuf generation complete."

## clean-api: Xóa các file .pb.go cũ để tránh xung đột
clean-api:
	@echo "==> Cleaning old Protobuf files..."
	@if [ -d "$(PROTO_GEN_OUT)" ]; then \
		find $(PROTO_GEN_OUT) -type f -name '*.pb.go' -delete; \
	fi
	@echo "    Cleanup complete."

# ==============================================================================
# Database & Ent Layer
# ==============================================================================

.PHONY: ent ent-new

ent:
	go generate ./ent

ent-new:
	go run -mod=mod entgo.io/ent/cmd/ent new ${name}


