# Cẩm Nang Apache Kafka

Tài liệu này ghi chép lại một cách sâu sắc và chi tiết nhất về các thành phần cốt lõi của Kafka, các edge cases (trường hợp góc) và các thông số cấu hình mang tính sống còn khi vận hành hệ thống phân tán.

---

## 1. Thành Phần Cấu Trúc (Architecture Components)

### 1.1 Cấu tạo của một Kafka Message (Anatomy of a Message)
Mỗi tin nhắn (Message/Record) được Producer tạo ra và gửi đi bao gồm các phần cốt lõi sau:
- **Key (Binary - Optional):** Dùng để định tuyến (Routing). Mọi message có cùng Key chắc chắn sẽ bay vào cùng một Partition. (Ví dụ: `BidID`).
- **Value (Binary - Optional):** Nội dung thực sự của tin nhắn (thường là chuỗi JSON hoặc Protobuf đã bị nén thành mảng bytes). Có thể để `null` (thường dùng trong khái niệm Tombstone để xóa dữ liệu).
- **Compression Type:** Loại nén dữ liệu (None, Gzip, Snappy, LZ4, Zstd). Giúp giảm băng thông mạng và dung lượng lưu trữ trên đĩa.
- **Headers (Optional):** Chứa các cặp Key-Value siêu dữ liệu (metadata) giống như HTTP Header (Ví dụ: Chứa Trace ID để phục vụ Distributed Tracing).
- **Partition + Offset:** Thông tin này do Broker gắn vào sau khi tin nhắn được ghi xuống đĩa thành công.
- **Timestamp (System or User set):** Thời gian tạo ra tin nhắn (do Producer set) hoặc thời gian tin nhắn được Broker ghi nhận (LogAppendTime).

### 1.2 Tuần tự hóa (Serialization) và Giải tuần tự hóa (Deserialization)
Bản thân Kafka Broker là những "gã mù mờ", nó **CHỈ** nhìn thấy và lưu trữ mảng Bytes (Binary raw). Nó không hề hiểu Object (như `Struct` trong Go hay `Class` trong Java) là gì.

- **Serialization (Ở Producer):** Là quá trình biến đổi Object thành chuỗi Bytes. 
  - Ví dụ: `Key Object (123) -> IntegerSerializer -> Bytes`.
  - `Value Object ("hello world") -> StringSerializer -> Bytes`.
- **Deserialization (Ở Consumer):** Là quá trình biến chuỗi Bytes ngược lại thành Object để ứng dụng đọc được. (Ví dụ dùng `IntegerDeserializer` cho Key và `StringDeserializer` cho Value).
- **Common Serializers/Deserializers:** String (bao gồm JSON), Int, Float, Avro, Protobuf. 

**Câu hỏi: Làm sao Serializer và Deserializer "hiểu" được type của Key và Data?**
1. **Thực tế phũ phàng:** Kafka Broker không lưu và không truyền bất kỳ metadata nào về Type của Data (nếu không dùng Header). Nên bản thân Deserializer **KHÔNG THỂ TỰ HIỂU**.
2. **Thỏa thuận ngầm (Contract):** Lập trình viên ở 2 đầu (Producer và Consumer) phải **tự cấu hình cứng (hardcode) khớp với nhau**. Nếu Producer dùng `StringSerializer` nhưng Consumer lại ngây thơ dùng `IntegerDeserializer` để đọc, Consumer sẽ bị văng lỗi (Crash / SerializationException) ngay lập tức.
3. **Quy tắc bất di bất dịch (Topic Lifecycle):** Bạn KHÔNG BAO GIỜ được phép thay đổi kiểu Serialization trong vòng đời của một Topic. Nếu ngày xưa dùng JSON, bây giờ muốn chuyển sang Protobuf, **hãy tạo một Topic hoàn toàn mới**. Cố tình thay đổi sẽ làm sập toàn bộ các Consumer đang chạy.
4. *(Nâng cao: Để giải quyết triệt để bài toán này ở quy mô tập đoàn, người ta dùng thêm hệ thống **Schema Registry** (vd: Confluent Schema Registry). Producer sẽ gửi ID của Schema kèm với mảng Bytes, Consumer đọc Bytes, lấy ID đi hỏi Schema Registry để tải cấu trúc (Schema) về rồi mới Deserialize).*

### 1.3 Broker
- **Bản chất vật lý:** Là một tiến trình (process) Java chạy trên một Server/Pod. Nhiệm vụ chính là nhận data từ Producer, ghi xuống ổ cứng (Disk I/O) và đẩy data cho Consumer.
- **Tại sao Kafka lại nhanh dù ghi xuống ổ cứng?**
  - Khác với Database thông thường ghi ngẫu nhiên (Random Access), Kafka ghi dữ liệu theo kiểu **Sequential I/O** (Ghi tuần tự vào cuối file). Tốc độ ghi tuần tự của ổ cứng HDD/SSD gần bằng tốc độ RAM.
  - Sử dụng **Zero-Copy (sendfile system call)**: Data đi thẳng từ Disk Buffer qua Network Socket mà không cần copy ngược lên User Space (Application).
- **Zookeeper vs KRaft Mode:**
  - *Cũ (Zookeeper):* Dùng để lưu trữ metadata (Topic nào nằm ở Broker nào, ai làm Leader). Điểm yếu là chậm khi có quá nhiều Topic và dễ bị split-brain (lỗi đồng đồng thuận).
  - *Mới (KRaft):* Zookeeper bị loại bỏ. Bản thân các Broker sẽ tự bầu ra một Quorum Controller bằng giao thức Raft để quản lý metadata. Nhanh hơn, nhẹ hơn, dễ deploy bằng Docker hơn.

### 1.4 Topic, Partition & Offset (Luồng dữ liệu cốt lõi)

**A. Topic (Chủ đề dữ liệu - Logic)**
- **Định nghĩa:** Topic đại diện cho một luồng dữ liệu (data stream) cụ thể. Nó khá giống với khái niệm "Table" (Bảng) trong Database, nhưng không hề có ràng buộc (constraints) hay schema khắt khe.
- **Đặc tính:** 
  - Bạn có thể tạo vô số Topic trong một Cluster Kafka (VD: `logs`, `purchases`, `trucks_gps`).
  - Nhận mọi định dạng dữ liệu (JSON, Avro, Text...).
  - **Không thể Query:** Khác với Database có thể dùng lệnh `SELECT * WHERE`, bạn không thể query trực tiếp vào Topic. Cách duy nhất để lấy dữ liệu là dùng Kafka Consumer để đọc (đọc tuần tự như đọc băng cassette).

**B. Partition (Phân mảnh - Vật lý)**
- Mỗi Topic sẽ được chia nhỏ thành nhiều Partition (VD: Topic `trucks_gps` có thể chia làm 100 partitions).
- **Mục đích:** Để phân tán dữ liệu ra nhiều Server/Broker khác nhau, giúp tăng tốc độ ghi và đọc song song (Parallelism) lên mức tối đa. Bạn có thể có bao nhiêu Partition tùy thích.
- **Phân phối:** Nếu tin nhắn không có Key, dữ liệu sẽ được Kafka rải ngẫu nhiên (randomly assigned) vào các Partition.

**C. Offset (Chỉ số định danh)**
- Bên trong mỗi Partition, mỗi khi có một tin nhắn mới được ghi vào, nó sẽ được cấp một số ID tăng dần liên tục, gọi là **Offset** (Bắt đầu từ 0, 1, 2, 3...).
- **Những lưu ý SỐNG CÒN về Offset & Partition:**
  1. **Tính Bất Biến (Immutability):** Một khi dữ liệu đã được ghi vào một Partition, nó **KHÔNG THỂ BỊ THAY ĐỔI** (không có lệnh UPDATE hay DELETE từng dòng như SQL).
  2. **Thứ Tự (Ordering):** Thứ tự của tin nhắn **CHỈ ĐƯỢC ĐẢM BẢO TRONG CÙNG 1 PARTITION**. Hoàn toàn không có khái niệm thứ tự global trên toàn bộ Topic.
  3. **Độc lập Offset:** Offset số `3` ở Partition `0` mang dữ liệu hoàn toàn khác biệt với Offset số `3` ở Partition `1`. Offset chỉ có ý nghĩa khi đi kèm với một Partition cụ thể.
  4. **Không Tái Sử Dụng:** Kể cả khi các tin nhắn cũ bị xóa (do hết hạn Retention 7 ngày), Kafka vẫn tiếp tục tăng mã Offset lên (Ví dụ từ 1000 lên 1001), các số Offset cũ đã xóa sẽ không bao giờ được tái sử dụng.

### 1.5 Leader, Follower & ISR (In-Sync Replicas)
- Nếu Topic có Replication Factor = 3, nghĩa là có 1 bản chính (Leader) và 2 bản phụ (Follower) nằm rải rác ở 3 Broker khác nhau.
- Mọi thao tác Đọc/Ghi đều phải chĩa vào **Leader**. Follower chỉ có nhiệm vụ hì hục pull data từ Leader về để làm backup.
- **ISR (In-Sync Replicas):** Là danh sách các Follower đang bắt kịp tốc độ của Leader (không bị lag quá thời gian `replica.lag.time.max.ms`). Nếu một Follower bị chậm, nó bị đá khỏi danh sách ISR. Nếu Leader chết, Kafka CHỈ bầu những Follower nằm trong ISR lên làm Leader mới để tránh mất data.

---

## 2. Đi Sâu Vào Producer (Nhà Sản Xuất)

Producer không hề ngốc nghếch, nó rất thông minh và làm nhiều việc ngầm:

### 2.1 Khả năng tự nhận thức (Cluster Topology Awareness)
- Producer không gửi tin nhắn một cách mù quáng. Nó biết chính xác **Partition nào đang nằm ở Broker nào**. 
- Trong trường hợp một Broker bị chết (Failures), Producer sẽ tự động phục hồi (Auto Recover) và tìm ra Broker mới đang nắm giữ Leader Partition để tiếp tục gửi data.
- Quá trình chia tải (Load Balancing) được Producer chủ động tính toán trước khi đẩy data sang mạng, giúp cụm Broker không bị quá tải cục bộ.

### 2.2 Thuật toán Phân luồng (Producer Partitioner Logic & Key Hashing)
Kafka Partitioner là một khối logic mã nguồn nằm bên trong Producer, có nhiệm vụ tiếp nhận một Record (tin nhắn) và quyết định chính xác xem tin nhắn này sẽ được đẩy vào Partition nào. Quá trình ánh xạ (mapping) từ một Key ra một Partition cụ thể được gọi là **Key Hashing** (Băm khóa).

Trong Partitioner mặc định (Default Partitioner) của Kafka, thuật toán băm được sử dụng là **murmur2**.
- **Murmur2 là gì?** Đây là một hàm băm phi mật mã (non-cryptographic hash function). Nó không được sinh ra để bảo mật như SHA-256 hay MD5, mà được thiết kế tối thượng cho **tốc độ thực thi cực nhanh** và **khả năng phân phối đồng đều (good distribution)**. Nhờ đó, dữ liệu được rải rất mượt mà trên toàn bộ các Partitions, tránh tình trạng "nút cổ chai" cục bộ (hotspots).

**Công thức định tuyến (Routing Formula):**
Mã nguồn thực tế ẩn dưới nắp capo của Producer để tính toán Partition mục tiêu là:
`targetPartition = Math.abs(Utils.murmur2(keyBytes)) % (numPartitions - 1)`

**Chiến lược cụ thể khi gọi lệnh `Send(Message)`:**
1. **Chỉ định tường minh (Explicit):** Nếu lập trình viên truyền cứng `Partition ID`, Producer sẽ bỏ qua thuật toán băm và gửi thẳng vào Partition đó.
2. **Có Key (Keyed Message):** Áp dụng công thức `murmur2` ở trên. Mọi tin nhắn có cùng Key (VD: `BidID=123`) sẽ luôn tạo ra cùng một giá trị Hash, do đó chắc chắn rơi vào cùng một Partition. Điều này là cốt lõi để bảo toàn **Strict Ordering** (Thứ tự tuyệt đối) cho các event có tính nhân quả.
3. **Không có Key (Key = null):** Sử dụng chiến lược **Sticky Partitioner** (gom các tin nhắn thành một Batch lớn gửi vào 1 Partition ngẫu nhiên cho đến khi đầy Batch, sau đó "dính" sang Partition khác) nhằm tối ưu hóa băng thông mạng, hiệu quả hơn rất nhiều so với Round-Robin truyền thống.

### 2.3 Cấu hình `acks` (Sự đánh đổi Mạng sống và Tốc độ)
- **`acks=0` (Fire and Forget):** Ném vào network là Producer báo thành công, không cần đợi Broker rep. Tốc độ bàn thờ, nhưng mất mạng là bay màu tin nhắn.
- **`acks=1` (Leader Ack):** Chỉ cần Leader ghi xuống đĩa thành công là báo OK. (Nguy cơ: Leader báo OK xong tẻo luôn, Follower chưa kịp chép về -> Vẫn mất data).
- **`acks=all` (hoặc `-1`):** Leader phải đợi tất cả các thằng trong danh sách ISR chép xong thì mới báo OK cho Producer. Chậm nhất nhưng an toàn tuyệt đối.

### 2.4 Idempotent Producer & Retries (Chống gửi đúp)
- Nếu Producer gửi tin nhắn, Broker nhận được nhưng lúc báo Ack về thì rớt mạng. Producer tưởng gửi xịt nên gửi lại -> **Dữ liệu bị đúp (Duplicate)**.
- **Giải pháp:** Bật `enable.idempotence=true`. Lúc này Producer sẽ dán cho mỗi tin nhắn 1 cái `PID` (Producer ID) và 1 `Sequence Number`. Nếu Broker thấy `Seq No` bị trùng, nó tự động vứt tin nhắn thứ 2 đi -> Đảm bảo **Exactly-Once Semantics (EOS)** ở cấp độ Producer.

---

## 3. Đi Sâu Vào Consumer & Consumer Group

### 3.0 Đặc tính cốt lõi của Consumer
- **Mô hình Kéo (Pull Model):** Khác với RabbitMQ (Broker chủ động đẩy - Push data cho worker), ở Kafka, Consumer tự mình chủ động gọi lệnh (pull) để xin dữ liệu từ Topic. Điều này giúp Consumer không bao giờ bị "ngộp" (overwhelmed) vì nó tự quyết định tốc độ đọc.
- **Nhận thức không gian:** Khi kết nối vào cụm, Consumer tự động "biết" chính xác phải đọc Partition nào và phải kết nối đến Broker vật lý nào (nhờ cơ chế Discovery).
- **Tự phục hồi (Recovery):** Nếu Broker mà Consumer đang đọc bị sập (Failures), Consumer đủ thông minh để tự động chuyển hướng kết nối sang Broker khác đang nắm giữ bản sao Leader mới.
- **Thứ tự (Ordering):** Dữ liệu luôn được đọc theo thứ tự nghiêm ngặt từ Offset thấp (cũ nhất) đến Offset cao (mới nhất) **TRONG PHẠM VI TỪNG PARTITION**.

### 3.1 Cân bằng tải (Load Balancing) & Starvation
- **Rule thép:** 1 Partition CHỈ được đọc bởi 1 Consumer trong 1 Group.
- Nếu Số Consumer > Số Partition -> Các Consumer dư thừa sẽ ở trạng thái Idle (Ngồi chơi).
- Nếu Số Consumer < Số Partition -> 1 Consumer sẽ phải cõng (đọc) nhiều Partition.

### 3.2 Rebalance (Tái cân bằng)
Xảy ra khi: Có Consumer chết, Consumer mới join vào, hoặc Topic tăng thêm Partition.
- **Stop-The-World:** Toàn bộ Consumer trong Group dừng xử lý, giải phóng Partition, Coordinator (Broker đóng vai trò lớp trưởng) sẽ gán lại Partition mới cho từng Consumer.
- Quá trình này có thể gây gián đoạn vài giây đến vài chục giây. Dùng chiến lược **Cooperative Sticky Assignor** để giảm thiểu Stop-The-World (chỉ thu hồi những Partition cần chia lại, những cái cũ giữ nguyên).

### 3.3 Offset Commit (Lưu điểm nhớ)
Khi Consumer đọc xong, nó phải báo cho Kafka biết "Tao đọc đến dòng số mấy rồi" bằng cách Commit Offset. Kafka lưu cái này vào một topic ẩn tên là `__consumer_offsets`.
- **Auto Commit (`enable.auto.commit=true`):** Cứ mỗi 5 giây, thư viện tự động báo Offset. (Cực nguy hiểm: Nếu app crash khi vừa đọc xong nhưng chưa kịp xử lý DB, lúc sau app khởi động lại nó sẽ bỏ qua tin nhắn đó -> **Mất data**).
- **Manual Commit (`enable.auto.commit=false`):** Ứng dụng tự gọi lệnh `Commit()`. 
  - *Commit trước khi xử lý:* Giống Auto Commit. Nguy cơ mất data (At-most-once).
  - *Commit sau khi xử lý xong DB:* An toàn. Nếu app crash khi xử lý xong DB nhưng chưa kịp Commit, lúc khởi động lại nó sẽ đọc lại tin đó -> **Bị đúp data (At-least-once)**.
  - **Best Practice:** Dùng Manual Commit (At-least-once) + Xây dựng tính năng **Idempotent (Chống trùng lặp)** ở tầng Database (VD: Dùng khóa ngoại (Unique Key) hoặc bảng cờ để check xem tin nhắn có ID này đã xử lý chưa).

---

## 4. Retention (Dọn Dẹp Rác)
Kafka không xóa tin nhắn ngay khi Consumer đọc xong (khác với RabbitMQ). Nó giữ lại theo cấu hình:
- **Time-based (`log.retention.hours`):** Mặc định 168 giờ (7 ngày). Quá 7 ngày tự xóa.
- **Size-based (`log.retention.bytes`):** Giữ tối đa X Gigabytes cho mỗi Partition. Đầy thì xóa file cũ nhất.
- **Log Compaction:** Xóa dữ liệu cũ nhưng **Dựa trên Key**. Nếu Topic có nhiều message cùng Key (VD: Cập nhật tọa độ tài xế ID=5 liên tục), Kafka chỉ giữ lại message *mới nhất* của tài xế ID=5. (Cực kỳ hữu ích để lưu trạng thái System State).

---

## 5. Kết Luận cho Dự Án Matching Service
- **Topic:** `matching.events`
- **Partition:** Mặc định 3 (để sau này có thể scale lên 3 Consumer Pods).
- **Producer Config:** `acks=all`, `enable.idempotence=true` (Đảm bảo không rớt và không đúp lệnh Match).
- **Consumer Config:** Dùng Manual Commit. Lấy message -> Gửi Email/Push Notification -> Gọi `Commit()`.
