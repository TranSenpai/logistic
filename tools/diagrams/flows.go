package main

// Sơ đồ cho các tài liệu trong docs/flows/ — mô tả một nghiệp vụ chảy QUA NHIỀU
// service, khác với docs/services/ vốn mô tả bên trong một service.

func matchingNotificationFlow() Diagram {
	return Diagram{
		Name:    "matching-notification-flow",
		Title:   "Luồng ghép xe và thông báo",
		Caption: "Nét liền: gọi đồng bộ. Nét đứt: sự kiện bất đồng bộ qua RabbitMQ.",
		Nodes: []Node{
			{ID: "shipper", Label: "Chủ hàng\nđăng đơn", Kind: KindClient, Col: 0, Row: 0},
			{ID: "gw", Label: "gateway_service\nPOST /matching/bids", Kind: KindGateway, Col: 1, Row: 0},
			{ID: "matching", Label: "matching_service\nSubmitBid", Kind: KindService, Col: 2, Row: 0},

			{ID: "pg", Label: "Postgres\nlưu đơn (bids)", Kind: KindStore, Col: 3, Row: 0},
			{ID: "find", Label: "FindAskForBid\ntìm tài xế phù hợp", Kind: KindNote, Col: 3, Row: 1},

			{ID: "notifier", Label: "Notifier\n(biz.Notifier)", Kind: KindNote, Col: 2, Row: 2},
			{ID: "rabbit", Label: "RabbitMQ\nlogistic.events (topic)", Kind: KindBroker, Col: 2, Row: 3},
			{ID: "queue", Label: "notification.events\nbinding: matching.#", Kind: KindBroker, Col: 2, Row: 4},
			{ID: "dlq", Label: "notification.events.dlq\nxử lý hỏng 2 lần", Kind: KindBroker, Col: 3, Row: 4},

			{ID: "consumer", Label: "notification_service\nMatchingConsumer", Kind: KindService, Col: 1, Row: 5},
			{ID: "guard", Label: "processed_events\nchống trùng (1 transaction)", Kind: KindStore, Col: 2, Row: 5},
			{ID: "pref", Label: "lọc theo\nNotificationPreference", Kind: KindNote, Col: 0, Row: 5},
			{ID: "inbox", Label: "notifications\nhộp thư từng người", Kind: KindStore, Col: 1, Row: 6},
			{ID: "unread", Label: "Redis\nbộ đếm chưa đọc", Kind: KindStore, Col: 0, Row: 6},

			{ID: "driverN", Label: "Tài xế nhận\n\"có đơn gần bạn\"", Kind: KindClient, Col: 1, Row: 7},
			{ID: "shipperN", Label: "Chủ hàng nhận\n\"đã tìm được xe\"", Kind: KindClient, Col: 0, Row: 7},
		},
		Edges: []Edge{
			{From: "shipper", To: "gw", Label: "HTTPS"},
			{From: "gw", To: "matching", Label: "gRPC"},
			{From: "matching", To: "pg", Label: "CreateBid"},
			{From: "pg", To: "find"},
			{From: "find", To: "notifier", Label: "danh sách ứng viên"},
			{From: "notifier", To: "rabbit", Label: "matching.driver.candidates_found", Dashed: true},
			{From: "rabbit", To: "queue", Label: "định tuyến", Dashed: true},
			{From: "queue", To: "dlq", Label: "NACK lần 2", Dashed: true},
			{From: "queue", To: "consumer", Label: "giao message", Dashed: true},
			{From: "consumer", To: "pref", Label: "kiểm tra"},
			{From: "consumer", To: "guard", Label: "ghi dấu event_id"},
			{From: "consumer", To: "inbox", Label: "tạo thông báo"},
			{From: "inbox", To: "unread", Label: "xoá cache"},
			{From: "inbox", To: "driverN", Label: "push"},
			{From: "inbox", To: "shipperN", Label: "push"},
		},
	}
}

func driverOnboardingFlow() Diagram {
	return Diagram{
		Name:    "driver-onboarding-flow",
		Title:   "Luồng tài xế gia nhập hệ thống",
		Caption: "Bốn cửa kiểm duyệt trước khi một chiếc xe được phép nhận đơn.",
		Nodes: []Node{
			{ID: "driver", Label: "Tài xế", Kind: KindClient, Col: 0, Row: 0},
			{ID: "reg", Label: "1. Đăng ký tài khoản\nPOST /users/register", Kind: KindGateway, Col: 1, Row: 0},
			{ID: "user", Label: "user_service\ntạo users + driver_profiles", Kind: KindService, Col: 2, Row: 0},

			{ID: "kyc", Label: "2. Nộp hồ sơ KYC\nPUT /users/{id}/driver-profile", Kind: KindGateway, Col: 1, Row: 1},
			{ID: "adminKyc", Label: "Admin duyệt KYC\nPUT /admin/kyc/{id}/review", Kind: KindClient, Col: 3, Row: 1},
			{ID: "kycOk", Label: "kyc_status = approved", Kind: KindNote, Col: 2, Row: 1},

			{ID: "veh", Label: "3. Đăng ký xe\nPOST /vehicles", Kind: KindGateway, Col: 1, Row: 2},
			{ID: "vehicle", Label: "vehicle_service\ntạo vehicles", Kind: KindService, Col: 2, Row: 2},

			{ID: "doc", Label: "4. Nộp giấy tờ xe\nPOST /vehicles/{id}/documents", Kind: KindGateway, Col: 1, Row: 3},
			{ID: "adminDoc", Label: "Admin duyệt giấy tờ\nPUT /admin/vehicles/{id}/verify", Kind: KindClient, Col: 3, Row: 3},
			{ID: "verified", Label: "verification_status\n= verified", Kind: KindNote, Col: 2, Row: 3},

			{ID: "online", Label: "5. Bật nhận đơn\nPOST /drivers/{id}/availability", Kind: KindGateway, Col: 1, Row: 4},
			{ID: "check", Label: "Kiểm tra: đã duyệt?\nkhông bảo dưỡng? toạ độ hợp lệ?", Kind: KindNote, Col: 2, Row: 4},
			{ID: "geo", Label: "Redis GEO\nxe vào chỉ mục tìm kiếm", Kind: KindStore, Col: 3, Row: 4},
			{ID: "ready", Label: "Sẵn sàng nhận đơn", Kind: KindClient, Col: 1, Row: 5},
		},
		Edges: []Edge{
			{From: "driver", To: "reg"},
			{From: "reg", To: "user", Label: "gRPC"},
			{From: "user", To: "kyc", Label: "hồ sơ rỗng"},
			{From: "kyc", To: "adminKyc", Label: "vào hàng đợi duyệt"},
			{From: "adminKyc", To: "kycOk"},
			{From: "kycOk", To: "veh"},
			{From: "veh", To: "vehicle", Label: "gRPC"},
			{From: "vehicle", To: "doc"},
			{From: "doc", To: "adminDoc", Label: "vào hàng đợi duyệt"},
			{From: "adminDoc", To: "verified"},
			{From: "verified", To: "online"},
			{From: "online", To: "check"},
			{From: "check", To: "geo", Label: "đạt"},
			{From: "geo", To: "ready"},
		},
	}
}

func shipperOrderFlow() Diagram {
	return Diagram{
		Name:    "shipper-order-flow",
		Title:   "Luồng chủ hàng đặt đơn tới khi chốt xe",
		Caption: "Vòng thương lượng: bid → offer → accept/reject. Mỗi bước đều sinh thông báo.",
		Nodes: []Node{
			{ID: "shipper", Label: "Chủ hàng", Kind: KindClient, Col: 0, Row: 0},
			{ID: "bid", Label: "1. Đăng đơn\nPOST /matching/bids", Kind: KindGateway, Col: 1, Row: 0},
			{ID: "engine", Label: "matching_service\nlưu bid, tìm ask phù hợp", Kind: KindService, Col: 2, Row: 0},
			{ID: "notif1", Label: "Thông báo tài xế\n\"có đơn gần bạn\"", Kind: KindBroker, Col: 3, Row: 0},

			{ID: "driver", Label: "Tài xế", Kind: KindClient, Col: 4, Row: 1},
			{ID: "offer", Label: "2. Báo giá\nPOST /matching/offers", Kind: KindGateway, Col: 3, Row: 1},
			{ID: "negotiate", Label: "bid.status\n= NEGOTIATING", Kind: KindNote, Col: 2, Row: 1},
			{ID: "notif2", Label: "Thông báo chủ hàng\n\"có báo giá mới\"", Kind: KindBroker, Col: 1, Row: 1},

			{ID: "decide", Label: "3. Chủ hàng quyết định", Kind: KindClient, Col: 0, Row: 2},
			{ID: "reject", Label: "Từ chối\nPOST /matching/offers/reject", Kind: KindGateway, Col: 1, Row: 3},
			{ID: "reopen", Label: "bid quay lại PENDING\ntài xế khác tiếp tục ra giá", Kind: KindNote, Col: 2, Row: 3},

			{ID: "accept", Label: "4. Chốt xe\nPOST /matching/matches/accept", Kind: KindGateway, Col: 1, Row: 4},
			{ID: "balance", Label: "wallet_service\nCheckBalance", Kind: KindService, Col: 2, Row: 4},
			{ID: "hold", Label: "Kafka wallet.hold_deposit\nđóng băng tiền cọc", Kind: KindBroker, Col: 3, Row: 4},
			{ID: "contract", Label: "matches\nhợp đồng đã chốt", Kind: KindStore, Col: 2, Row: 5},
			{ID: "notif3", Label: "Thông báo CẢ HAI\n\"đã ghép được xe\"", Kind: KindBroker, Col: 1, Row: 5},
		},
		Edges: []Edge{
			{From: "shipper", To: "bid"},
			{From: "bid", To: "engine", Label: "gRPC"},
			{From: "engine", To: "notif1", Label: "RabbitMQ", Dashed: true},
			{From: "notif1", To: "driver", Label: "push", Dashed: true},
			{From: "driver", To: "offer"},
			{From: "offer", To: "negotiate"},
			{From: "negotiate", To: "notif2", Label: "RabbitMQ", Dashed: true},
			{From: "notif2", To: "decide", Label: "push", Dashed: true},
			{From: "decide", To: "reject", Label: "không ưng"},
			{From: "reject", To: "reopen"},
			{From: "decide", To: "accept", Label: "đồng ý"},
			{From: "accept", To: "balance", Label: "gRPC"},
			{From: "balance", To: "hold", Label: "đủ tiền", Dashed: true},
			{From: "balance", To: "contract", Label: "tạo hợp đồng"},
			{From: "contract", To: "notif3", Label: "RabbitMQ", Dashed: true},
		},
	}
}

func driverLocationFlow() Diagram {
	return Diagram{
		Name:    "driver-location-flow",
		Title:   "Luồng GPS và tìm xe đang chạy",
		Caption: "Postgres là nguồn sự thật; Redis GEO là chỉ mục tìm kiếm dựng lại được.",
		Nodes: []Node{
			{ID: "app", Label: "App tài xế\nping GPS vài giây/lần", Kind: KindClient, Col: 0, Row: 0},
			{ID: "nginx", Label: "Nginx\nrate-limit riêng 60r/s", Kind: KindEdge, Col: 1, Row: 0},
			{ID: "gw", Label: "gateway_service\nPOST /vehicles/{id}/location", Kind: KindGateway, Col: 2, Row: 0},
			{ID: "vehicle", Label: "vehicle_service\nReportLocation", Kind: KindService, Col: 3, Row: 0},

			{ID: "valid", Label: "Kiểm tra toạ độ\nchặn (0,0), NaN, ngoài ngưỡng", Kind: KindNote, Col: 3, Row: 1},
			{ID: "pg", Label: "vehicle_locations\nghi đè 1 dòng / xe", Kind: KindStore, Col: 2, Row: 2},
			{ID: "avail", Label: "driver_availabilities\nđồng bộ toạ độ", Kind: KindStore, Col: 3, Row: 2},
			{ID: "online", Label: "Tài xế đang bật\nnhận đơn?", Kind: KindNote, Col: 1, Row: 2},
			{ID: "geo", Label: "Redis GEO\nsorted set xe online", Kind: KindStore, Col: 0, Row: 2},

			{ID: "search", Label: "SearchNearbyVehicles\n(matching gọi sang)", Kind: KindService, Col: 0, Row: 4},
			{ID: "geosearch", Label: "GEOSEARCH\nbán kính, gần → xa", Kind: KindNote, Col: 1, Row: 4},
			{ID: "hydrate", Label: "2 truy vấn gộp\ntránh N+1", Kind: KindNote, Col: 2, Row: 4},
			{ID: "filter", Label: "Lọc: đã duyệt · active\n· đúng loại · đủ tải", Kind: KindNote, Col: 3, Row: 4},
			{ID: "result", Label: "Danh sách xe phù hợp", Kind: KindClient, Col: 3, Row: 5},
			{ID: "fallback", Label: "Redis chết →\nquét zone_id + haversine", Kind: KindNote, Col: 0, Row: 5},
		},
		Edges: []Edge{
			{From: "app", To: "nginx"},
			{From: "nginx", To: "gw"},
			{From: "gw", To: "vehicle", Label: "gRPC"},
			{From: "vehicle", To: "valid"},
			{From: "valid", To: "pg", Label: "hợp lệ"},
			{From: "pg", To: "avail", Label: "đồng bộ"},
			{From: "pg", To: "online", Label: "sau khi ghi DB"},
			{From: "online", To: "geo", Label: "có → GEOADD"},
			{From: "search", To: "geosearch"},
			{From: "geo", To: "geosearch", Label: "đọc chỉ mục"},
			{From: "geosearch", To: "hydrate"},
			{From: "hydrate", To: "filter"},
			{From: "filter", To: "result"},
			{From: "search", To: "fallback", Label: "khi Redis lỗi", Dashed: true},
		},
	}
}

func authenticationFlow() Diagram {
	return Diagram{
		Name:    "authentication-flow",
		Title:   "Luồng xác thực và phân quyền",
		Caption: "Gateway tin header đã được lớp xác thực kiểm tra; RequireRole gắn ở cấp group.",
		Nodes: []Node{
			{ID: "user", Label: "Người dùng", Kind: KindClient, Col: 0, Row: 0},
			{ID: "login", Label: "POST /auth/login\nhoặc /auth/google/login", Kind: KindGateway, Col: 1, Row: 0},
			{ID: "auth", Label: "auth_service\nkiểm tra mật khẩu", Kind: KindService, Col: 2, Row: 0},
			{ID: "google", Label: "Google OAuth2\nconsent screen", Kind: KindExternal, Col: 3, Row: 0},

			{ID: "state", Label: "state ngẫu nhiên\n+ HttpOnly cookie", Kind: KindNote, Col: 3, Row: 1},
			{ID: "csrf", Label: "So state URL ↔ cookie\nlệch → 400", Kind: KindNote, Col: 2, Row: 1},

			{ID: "token", Label: "TokenPair\naccess + refresh", Kind: KindStore, Col: 1, Row: 1},
			{ID: "call", Label: "Gọi API kèm\nAuthorization: Bearer", Kind: KindClient, Col: 0, Row: 2},

			{ID: "identity", Label: "IdentityContext\nđọc X-User-Id / X-User-Role", Kind: KindGateway, Col: 1, Row: 3},
			{ID: "client", Label: "/api/v1/...\nnhóm client", Kind: KindGateway, Col: 0, Row: 4},
			{ID: "role", Label: "RequireRole(\"admin\")\ngắn ở CẤP GROUP", Kind: KindNote, Col: 2, Row: 3},
			{ID: "admin", Label: "/api/v1/admin/...\nnhóm quản trị", Kind: KindGateway, Col: 2, Row: 4},
			{ID: "deny", Label: "403 PERMISSION_DENIED", Kind: KindNote, Col: 3, Row: 4},
		},
		Edges: []Edge{
			{From: "user", To: "login"},
			{From: "login", To: "auth", Label: "gRPC"},
			{From: "login", To: "google", Label: "redirect 307"},
			{From: "google", To: "state", Label: "callback"},
			{From: "state", To: "csrf"},
			{From: "csrf", To: "token", Label: "khớp"},
			{From: "auth", To: "token", Label: "ký JWT"},
			{From: "token", To: "call"},
			{From: "call", To: "identity"},
			{From: "identity", To: "client", Label: "mọi vai trò"},
			{From: "identity", To: "role", Label: "đường /admin"},
			{From: "role", To: "admin", Label: "role = admin"},
			{From: "role", To: "deny", Label: "khác"},
		},
	}
}

func errorHandlingFlow() Diagram {
	return Diagram{
		Name:    "error-handling-flow",
		Title:   "Luồng lỗi xuyên tầng",
		Caption: "Một lỗi đi qua 4 lần biến đổi. Chi tiết kỹ thuật dừng lại ở log, không ra tới client.",
		Nodes: []Node{
			{ID: "db", Label: "Postgres\nduplicate key violation", Kind: KindStore, Col: 0, Row: 0},
			{ID: "ent", Label: "ent.ConstraintError\n(dao)", Kind: KindNote, Col: 1, Row: 0},
			{ID: "wrap", Label: "repo/repo_error.go\nwrapError()", Kind: KindService, Col: 2, Row: 0},
			{ID: "apperr", Label: "apperr.Error\nKind=AlreadyExists\nCode=PHONE_ALREADY_USED", Kind: KindNote, Col: 3, Row: 0},

			{ID: "biz", Label: "biz trả nguyên lỗi\nKHÔNG bọc thêm", Kind: KindService, Col: 3, Row: 1},
			{ID: "ctrl", Label: "controller trả nguyên lỗi\nKHÔNG có status.Errorf", Kind: KindService, Col: 3, Row: 2},

			{ID: "interceptor", Label: "pkg/middleware\nErrorInterceptor", Kind: KindGateway, Col: 2, Row: 3},
			{ID: "grpc", Label: "gRPC ALREADY_EXISTS\n+ ErrorInfo{Reason}", Kind: KindNote, Col: 1, Row: 3},
			{ID: "log", Label: "Log nội bộ\ngiữ câu SQL, tên bảng", Kind: KindStore, Col: 3, Row: 3},

			{ID: "resp", Label: "gateway/response\nError()", Kind: KindGateway, Col: 1, Row: 4},
			{ID: "http", Label: "HTTP 409\n{code, message, request_id}", Kind: KindClient, Col: 0, Row: 5},
			{ID: "panic", Label: "panic bất kỳ\n→ RecoveryInterceptor → 500", Kind: KindNote, Col: 2, Row: 5},
		},
		Edges: []Edge{
			{From: "db", To: "ent"},
			{From: "ent", To: "wrap"},
			{From: "wrap", To: "apperr", Label: "dịch"},
			{From: "apperr", To: "biz"},
			{From: "biz", To: "ctrl"},
			{From: "ctrl", To: "interceptor"},
			{From: "interceptor", To: "grpc", Label: "gắn ErrorInfo"},
			{From: "interceptor", To: "log", Label: "chi tiết dừng ở đây", Dashed: true},
			{From: "grpc", To: "resp", Label: "qua dây"},
			{From: "resp", To: "http", Label: "mã HTTP + code ổn định"},
		},
	}
}
