package biz

import "strconv"

// formatFloat dùng cho phần Details của lỗi — chỉ mang tính chẩn đoán nên
// 6 chữ số thập phân (độ chính xác ~11cm) là quá đủ.
func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 6, 64)
}
