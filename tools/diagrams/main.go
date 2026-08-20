package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// Nơi ghi kết quả. Ba định dạng phục vụ ba mục đích khác nhau:
//
//	.svg     -> nhúng vào .md, đọc tài liệu là thấy hình ngay
//	.drawio  -> mở bằng diagrams.net để chỉnh tay khi cần
//	.html    -> một trang xem tất cả, không cần công cụ gì
const (
	svgDir    = "docs/diagrams/svg"
	drawioDir = "docs/diagrams"
	htmlOut   = "docs/rendered/diagrams.html"
)

// flowNames / serviceNames dùng cho phần điều hướng của trang HTML và để
// doclint biết sơ đồ nào thuộc nhóm nào.
func flowNames() map[string]bool {
	return map[string]bool{
		"matching-notification-flow": true,
		"driver-onboarding-flow":     true,
		"shipper-order-flow":         true,
		"driver-location-flow":       true,
		"authentication-flow":        true,
		"error-handling-flow":        true,
	}
}

func serviceNames() map[string]bool {
	return map[string]bool{
		"gateway-service":      true,
		"auth-service":         true,
		"user-service":         true,
		"vehicle-service":      true,
		"matching-service":     true,
		"notification-service": true,
		"media-service":        true,
		"wallet-service":       true,
	}
}

func main() {
	ds := diagrams()

	// Bắt trùng tên sớm: hai sơ đồ cùng tên sẽ ghi đè nhau trong im lặng.
	seen := map[string]bool{}
	for _, d := range ds {
		if seen[d.Name] {
			fmt.Fprintf(os.Stderr, "diagrams: trùng tên sơ đồ %q\n", d.Name)
			os.Exit(1)
		}
		seen[d.Name] = true
	}

	for _, dir := range []string{svgDir, drawioDir, filepath.Dir(htmlOut)} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "diagrams: tạo thư mục %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	for _, d := range ds {
		svgPath := filepath.Join(svgDir, d.Name+".svg")
		if err := os.WriteFile(svgPath, []byte(renderSVG(d)), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "diagrams: ghi %s: %v\n", svgPath, err)
			os.Exit(1)
		}

		drawioPath := filepath.Join(drawioDir, d.Name+".drawio")
		if err := os.WriteFile(drawioPath, []byte(renderDrawio(d)), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "diagrams: ghi %s: %v\n", drawioPath, err)
			os.Exit(1)
		}
	}

	if err := os.WriteFile(htmlOut, []byte(renderHTMLIndex(ds)), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "diagrams: ghi %s: %v\n", htmlOut, err)
		os.Exit(1)
	}

	fmt.Printf("diagrams: đã sinh %d sơ đồ → %s/*.svg, %s/*.drawio, %s\n",
		len(ds), svgDir, drawioDir, htmlOut)
}
