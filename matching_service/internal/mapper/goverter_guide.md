# Hướng Dẫn Sử Dụng Goverter (Type-Safe Mapping)

Trong kiến trúc phần mềm, việc chuyển đổi dữ liệu qua lại giữa các tầng (Ví dụ: Từ `ent.Asks` của tầng Database sang `entity.Ask` của tầng Domain) là một tác vụ cực kỳ phổ biến.

Tài liệu này hướng dẫn cách sử dụng `goverter` để thực hiện Data Mapping, đồng thời giải thích lý do tại sao phương pháp này vượt trội hơn hẳn so với thư viện `jinzhu/copier` truyền thống.

## 1. Vấn đề của `jinzhu/copier` (Reflection & Runtime)

Trước đây, nhiều dự án Go sử dụng thư viện [jinzhu/copier](https://github.com/jinzhu/copier) để copy dữ liệu từ Struct A sang Struct B. 

```go
// Ví dụ dùng jinzhu/copier
var dest entity.Ask
copier.Copy(&dest, &sourceEntAsk)
```

> [!WARNING]
> Mặc dù code trông rất ngắn gọn, nhưng `jinzhu/copier` ẩn chứa 3 điểm yếu chí mạng:
> 
> 1. **Hiệu năng cực thấp (Runtime Reflection):** `copier` sử dụng kỹ thuật Reflection của Go (giống cơ chế `json.Marshal/Unmarshal`) để soi từng cột dữ liệu lúc chương trình đang chạy. Quá trình này ngốn rất nhiều CPU và Memory.
> 2. **Lỗi tiềm ẩn (Không Type-Safe):** Nếu bạn đổi tên một cột ở Struct A nhưng quên đổi ở Struct B, code của bạn VẪN BIÊN DỊCH THÀNH CÔNG. `copier` sẽ âm thầm bỏ qua cột đó lúc chạy, dẫn đến việc dữ liệu bị thiếu hụt (Data Loss) mà bạn không hề hay biết cho đến khi crash production.
> 3. **Khó debug:** Khi mapping bị sai, bạn rất khó trace xem cột nào đang gán giá trị cho cột nào vì mọi thứ bị giấu trong hộp đen của Reflection.

---

## 2. Giải pháp của `goverter` (Type-Safe & Compile-time CodeGen)

[Goverter](https://github.com/jmattheis/goverter) đi theo một triết lý hoàn toàn khác, tương tự như triết lý của `Ent`: **Code Generation**.

Thay vì mò mẫm dữ liệu lúc chạy, Goverter sẽ đọc Interface của bạn và **Tự Động Sinh Ra Code Go thuần** (như việc bạn tự gõ tay `dest.A = source.A`).

> [!TIP]
> Lợi ích tuyệt đối của Goverter:
> - **Hiệu năng tối đa (Zero-Cost):** Vì nó đẻ ra code gán tay trực tiếp, không hề dùng Reflection, nên tốc độ chạy nhanh ngang ngửa với code do chính bạn gõ.
> - **Type-Safe 100%:** Bất kỳ sự lệch pha nào về Tên Cột hoặc Kiểu Dữ Liệu đều bị phát hiện ngay lập tức lúc Compile (khi gõ `go generate`). Trình biên dịch sẽ báo lỗi đỏ lòm, không cho phép bạn deploy một hệ thống bị lỗi mapping.
> - **Dễ dàng tùy biến:** Bạn có thể viết các hàm Custom Logic để xử lý những cột phức tạp.

---

## 3. Cách sử dụng Goverter trong Project của chúng ta

### Bước 1: Khai báo Interface (Bản thiết kế)

Bạn tạo file `mapper.go` và định nghĩa một Interface `Converter`. Sử dụng các comment đặc biệt `// goverter:map` để chỉ định luật lệ mapping.

```go
// goverter:converter
// go:generate go run github.com/jmattheis/goverter/cmd/goverter@latest gen ./
type Converter interface {
    // 1. Chỉ định Map 1-1 qua Custom Function
    // goverter:map ID | UUIDToString
    
    // 2. Gom nhiều cột thành 1 object
    // goverter:map . CurrentLocation | MapAskCurrentLocation
    
    // 3. Bỏ qua các cột không cần thiết
    // goverter:ignore DriverID
    EntAskToEntityAsk(source *ent.Asks) entity.Ask
}
```

### Bước 2: Viết các Custom Function

Khi 2 kiểu dữ liệu không khớp nhau (Ví dụ: `uuid.UUID` và `string`), Goverter không tự đoán được. Bạn phải viết hàm hướng dẫn nó cách chuyển đổi:

```go
func UUIDToString(id uuid.UUID) string {
    return id.String()
}

// Hàm này nhận Toàn bộ object source (ent.Asks) để gom các cột Flat thành 1 object Location
func MapAskCurrentLocation(source *ent.Asks) entity.Location {
    if source == nil {
        return entity.Location{}
    }
    return entity.Location{
        Latitude:  source.OriginLat,
        Longitude: source.OriginLng,
        ZoneID:    source.ZoneID,
    }
}
```

### Bước 3: Chạy máy đúc (Generate)

Bạn mở Terminal tại thư mục chứa file `mapper.go` và chạy:

```bash
go generate ./
```

Goverter sẽ đọc file `mapper.go` của bạn và tự động đẻ ra một file tên là `generated.go` nằm ngay bên cạnh. Nếu bạn mở file này ra, bạn sẽ thấy nó tự động gõ hàng ngàn dòng code gán biến `dest.X = source.Y` cực kỳ tối ưu cho bạn!

### Bước 4: Sử dụng trong code nghiệp vụ

```go
// Ở tầng Usecase / Controller
import "matching_service/internal/mapper"

func HandleAsk() {
    var entAsk *ent.Asks = GetFromDB()
    
    // Khởi tạo Mapper
    conv := mapper.ConverterImpl{}
    
    // Convert 1 phát ăn ngay, an toàn tuyệt đối
    entityAsk := conv.EntAskToEntityAsk(entAsk)
}
```
