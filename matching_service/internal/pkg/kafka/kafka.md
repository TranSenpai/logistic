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

### 1.3 Kafka Brokers & Cluster (Máy chủ và Cụm)
- **Bản chất vật lý:** Là một tiến trình (process) Java chạy trên một Server/Pod. Một **Kafka Cluster (Cụm Kafka)** được tạo thành từ nhiều Broker gộp lại.
- **Định danh (ID):** Mỗi Broker trong cụm được định danh bằng một số nguyên (Integer ID), ví dụ Broker 101, 102, 103 (con số thường chọn ngẫu nhiên bắt đầu từ 100 cho dễ phân biệt).
- **Quy mô (Scale):** Để chạy Production an toàn (có tính Availability cao), con số khởi điểm lý tưởng là **3 Brokers**. Các cụm Kafka khổng lồ của các tập đoàn có thể lên tới hơn 100 Brokers.

**A. Cơ chế kết nối thông minh (Broker Discovery / Bootstrap Server)**
- Mọi Broker trong cụm Kafka đều đóng vai trò là một **"Bootstrap Server"**.
- Điểm làm nên tên tuổi của Kafka Client (Producer/Consumer) là nó rất **Thông minh (Smart Clients)**:
  - Bạn **CHỈ CẦN** cung cấp IP của 1 Broker duy nhất (ví dụ Broker 101) vào cấu hình kết nối.
  - Khi Client kết nối vào Broker 101 và gửi `Metadata Request`, Broker 101 (do nắm toàn bộ thông tin của Cluster) sẽ trả về danh sách tất cả các Broker còn lại, cùng với sơ đồ vị trí của các Topic/Partition.
  - Sau đó Client sẽ tự động biết cách mở kết nối ngang hàng tới các Broker chứa đúng Partition mà nó cần đọc/ghi (Broker 102, 103...).

**B. Sự phân bố dữ liệu (Brokers and Topics)**
- Các Partition của một Topic sẽ được phân tán (distributed) rải rác trên các Broker.
- **Nguyên tắc cốt lõi:** Một Broker **CHỈ** nắm giữ dữ liệu mà nó được phân công, nó không bắt buộc phải chứa toàn bộ Topic của hệ thống.
  - Ví dụ: `Topic-A` có 3 Partitions -> Có thể chia đều cho Broker 101 (P0), Broker 102 (P1), Broker 103 (P2).
  - Nhưng `Topic-B` chỉ có 2 Partitions -> Nó có thể chỉ nằm ở Broker 101 (P1) và Broker 102 (P0). Lúc này Broker 103 sẽ **không chứa một giọt data nào** của Topic-B cả (chuyện này hết sức bình thường).

**C. Tại sao Kafka lại siêu tốc dù ghi xuống ổ cứng?**
- Khác với Database thông thường ghi ngẫu nhiên (Random Access), Kafka ghi dữ liệu theo kiểu **Sequential I/O** (Ghi tuần tự vào đít file). Tốc độ ghi tuần tự của ổ cứng gần tương đương với tốc độ RAM.
- Sử dụng **Zero-Copy (sendfile system call)**: Data đi thẳng từ Disk Buffer qua Network Socket bay ra ngoài mạng mà không cần copy ngược lên User Space (Application memory), tiết kiệm cực kỳ nhiều CPU.

**D. Zookeeper vs KRaft Mode (KIP-500)**
- *Cũ (Zookeeper):* Phải chạy một cụm Zookeeper riêng biệt để lưu trữ metadata.
  - Nhược điểm chí mạng: Nút thắt cổ chai (Scaling issues) khi cụm Kafka vượt quá 100,000 Partitions.
- *Mới (KRaft - Kafka Raft Architecture):* Khởi động từ 2020, Kafka tự chạy giao thức đồng thuận Raft để bầu ra Quorum Controller, chính thức gỡ bỏ hoàn toàn sự phụ thuộc vào Zookeeper.
  - **Lợi ích khổng lồ:** Khả năng scale lên **hàng triệu Partitions**. Giảm thời gian shutdown và recovery. Quản lý chung một hệ thống bảo mật (Single security model). Khởi chạy chỉ với 1 process duy nhất (nhẹ nhàng hơn).
  - **Lộ trình:** Production-ready từ bản 3.3.1. Kể từ **Kafka 4.0**, KRaft là lựa chọn DUY NHẤT (Zookeeper bị khai tử hoàn toàn).

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

### 1.5 Replication Factor, Leader & ISR (Nhân bản & Chống sập)

**A. Phân biệt Rạch ròi: Partition vs Replication Factor (Cực kỳ dễ nhầm lẫn)**
Rất nhiều người mới học lầm tưởng "Partition 1 là bản copy của Partition 0". **ĐIỀU NÀY LÀ SAI HOÀN TOÀN.**
- **Partition (Phân mảnh - Scale Horizontal):** Dùng để CHIA NHỎ dữ liệu. Ví dụ Topic có 100GB, chia làm 2 Partition thì P0 chứa 50GB data đầu, P1 chứa 50GB data sau (2 data này hoàn toàn khác nhau). Mục đích là để nhiều Consumer cùng đọc song song cho nhanh.
- **Replication Factor (Nhân bản - Persistence/Durability):** Dùng để COPY nguyên si dữ liệu nhằm chống mất mát (phòng hờ server cháy). Nếu Replication Factor = 2, thì toàn bộ 50GB của P0 sẽ được nhân bản thành một "P0 Replica" nằm ở một Broker khác. **Tức là bản sao của Partition 0 CHÍNH LÀ Partition 0, chứ KHÔNG PHẢI là Partition 1.**

**B. Nguyên lý Nhân bản (Replication) & Độ bền (Durability)**
- Một Topic LUÔN nên thiết lập Replication Factor > 1 (thông thường là 2 hoặc 3, chuẩn công nghiệp là 3).
- **Mục đích (Failover):** Giả sử P0 có bản chính ở Broker 101 và bản sao ở Broker 102. Nếu Broker 102 đột ngột chết (mất điện), Broker 101 và 103 vẫn đứng ra gánh vác việc phục vụ data bình thường, hệ thống không hề có downtime.
- **Công thức tính độ bền (Rule of thumb):** Với Replication Factor là `N`, hệ thống có thể chịu đựng được việc mất đi vĩnh viễn tối đa `N - 1` Broker mà không bị mất dữ liệu. (VD: Replication Factor = 3 thì có thể chết vĩnh viễn 2 Broker cùng lúc vẫn an toàn).

**C. Concept of Leader (Ai làm đại ca?)**
- Tại bất kỳ thời điểm nào, đối với một Partition cụ thể (ví dụ P0), **CHỈ CÓ DUY NHẤT 1 Broker** được nắm chức **Leader**.
- Các Broker chứa bản sao còn lại được gọi là **Follower** (hoặc Replica).
- **Quy tắc Đọc/Ghi (Behavior):**
  - Kafka Producers **CHỈ** được phép bắn dữ liệu vào Broker đang làm Leader của Partition đó.
  - Các Follower đứng ngoài rìa hì hục copy (replicate) data từ Leader về.
  - Kafka Consumers **theo mặc định** sẽ chỉ kéo dữ liệu từ Broker đang làm Leader.
  - *(Ngoại lệ)* **Consumer Replica Fetching (Từ Kafka v2.4+):** Giờ đây bạn có thể cấu hình cho Consumer đọc dữ liệu từ một Follower (ISR) gần nó nhất thay vì Leader. Điều này giúp giảm thiểu độ trễ (latency) và tiết kiệm cực nhiều chi phí mạng (network costs) nếu chạy trên môi trường Cloud đắt đỏ (cross-AZ data transfer).

**D. ISR (In-Sync Replicas - Bản sao đồng bộ)**
- Bất kỳ Follower nào đang sao chép dữ liệu và bắt kịp tốc độ của Leader (không bị trễ quá thời gian `replica.lag.time.max.ms`) thì được liệt kê vào danh sách **ISR**.
- Nếu một Follower bị chậm (do đứt cáp, ổ cứng bad sector...), nó sẽ bị Leader đá văng khỏi danh sách ISR.
- Nếu Leader chết đột ngột, Kafka **CHỈ ĐỀ CỬ** những Follower đang nằm trong danh sách ISR lên làm Leader mới để đảm bảo dữ liệu không bị thất thoát.

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

**5 Chiến lược phân luồng (Partitioner Strategies) của Producer:**
1. **Key Hashing (Mặc định khi có Key):** Áp dụng thuật toán `murmur2`. Mọi tin nhắn có cùng Key (VD: `BidID=123`) sẽ luôn bay vào cùng một Partition. Cực kỳ quan trọng để bảo toàn **Strict Ordering** (Thứ tự tuyệt đối).
2. **Sticky Partitioner (Mặc định khi Key = null, từ Kafka 2.4+):** Cực kỳ tối ưu! Producer gom các tin nhắn (dù khác loại) thành một Batch lớn gửi vào 1 Partition ngẫu nhiên cho đến khi đầy Batch (hoặc hết thời gian `linger.ms`), sau đó mới đổi sang Partition khác. Giảm triệt để độ trễ mạng.
3. **RoundRobin Partitioner (Legacy khi Key = null, Kafka cũ):** Phân phối kiểu chia bài (Msg 1 vào P0, Msg 2 vào P1, Msg 3 vào P2). Nghe có vẻ công bằng nhưng **CỰC KỲ TỆ** vì nó tạo ra vô số các Batch nhỏ xíu tốn băng thông. Hiện tại gần như không ai dùng.
4. **Explicit Partitioner (Chỉ định tường minh):** Lập trình viên truyền cứng `Partition ID` vào record, Producer sẽ nhắm mắt ném thẳng vào Partition đó, bỏ qua mọi logic.
5. **Custom Partitioner:** Tự code logic phân luồng (Implement interface `Partitioner`). Ví dụ: Khách Vip ném vào P0 (server xịn), khách thường ném vào P1, P2.

### 2.3 Cấu hình `acks` (Producer Acknowledgements - Đánh đổi Mạng sống và Tốc độ)
Producer có quyền tự quyết định xem nó cần mức độ xác nhận (acknowledgment) nào từ Broker sau khi ghi dữ liệu:
- **`acks=0` (Fire and Forget):** Producer ném data vào mạng là xong, hoàn toàn **KHÔNG ĐỢI** xác nhận từ Broker. Tốc độ bàn thờ nhưng rủi ro **mất dữ liệu (Data Loss)** rất cao nếu Broker chưa kịp nhận mà rớt mạng.
- **`acks=1` (Leader Ack):** Producer đợi xác nhận từ Leader. Chỉ cần Leader ghi xuống đĩa thành công là báo OK. Rủi ro **mất dữ liệu hữu hạn (Limited Data Loss)** nếu Leader báo OK xong tẻo luôn trước khi Follower kịp chép về.
- **`acks=all` (hoặc `-1`):** Producer đợi xác nhận từ CẢ Leader và các Replicas. Leader phải đợi tất cả các thằng trong danh sách ISR chép xong thì mới báo OK. Chậm nhất nhưng **Không bao giờ mất dữ liệu (No Data Loss)**.

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
- **Rule thép:** 1 Partition CHỈ được đọc bởi TỐI ĐA 1 Consumer trong CÙNG 1 Group.
- Nếu Số Consumer > Số Partition -> Các Consumer dư thừa sẽ ở trạng thái Idle (Ngồi chơi - Starvation).
- Nếu Số Consumer < Số Partition -> 1 Consumer sẽ phải cõng (đọc) nhiều Partition.

### 3.2 Đa Consumer Group (Multiple Consumer Groups trên 1 Topic)
- **Hoàn toàn hợp lệ (Acceptable):** Kafka được thiết kế lõi theo mô hình Pub/Sub (Fan-out), cho phép rất nhiều Consumer Group cùng subscribe và đọc dữ liệu trên cùng 1 Topic mà không tranh giành hay triệt tiêu data của nhau.
- **Cách phân tách:** Chỉ cần cấu hình thuộc tính `group.id` khác nhau cho mỗi cụm ứng dụng (Ví dụ: `group.id=consumer-group-application-1`, `group.id=consumer-group-application-2`).
- **Cơ chế hoạt động:**
  - Giả sử Topic-A có 3 Partition.
  - **Group 1** (có 2 Consumer): Sẽ chia nhau đọc 3 Partition (vd: C1 đọc P0, P1; C2 đọc P2).
  - **Group 2** (có 3 Consumer): Mỗi Consumer đọc đúng 1 Partition (C1->P0, C2->P1, C3->P2).
  - **Group 3** (có 1 Consumer): Mình nó ôm "show" đọc cả 3 Partition.
- **Điểm ăn tiền:** Mỗi Consumer Group duy trì một con trỏ Offset hoàn toàn độc lập. Ứng dụng này đọc xong, commit Offset rồi thì ứng dụng kia vẫn có thể bắt đầu đọc lại từ Offset số 0.

### 3.3 Rebalance (Tái cân bằng)

**A. Bản chất của Rebalance là gì?**
- Việc thuyên chuyển (di dời) quyền đọc các Partition từ Consumer này sang Consumer khác được gọi là **Rebalance** (Tái cân bằng).
- **Khi nào xảy ra?**
  1. Khi có một Consumer rời khỏi nhóm (bị crash, tắt máy) hoặc join mới vào nhóm (scale up).
  2. Khi quản trị viên hệ thống (Administrator) add thêm Partition mới vào Topic.

**B. Sự khác biệt giữa 2 trường phái Rebalance:**
1. **Trường phái Eager Rebalance (Cực kỳ tốn kém):**
   - **Tất cả Consumer phải DỪNG HOẠT ĐỘNG (Stop-The-World).**
   - Mọi người phải tự nguyện từ bỏ (give up) quyền đọc tất cả các Partition đang giữ.
   - Tham gia lại vào Group và nhận lệnh phân công mới từ Coordinator.
   - Nhược điểm: Cả một khoảng thời gian ngắn, toàn bộ Group bị tê liệt không xử lý data. Khi nhận lại, chưa chắc Consumer được chia lại đúng Partition mà nó vừa thả ra.
2. **Trường phái Cooperative Rebalance / Incremental (Khuyên dùng):**
   - Chỉ thu hồi (revoke) một **tập hợp nhỏ** các Partition thực sự cần phải dời đi.
   - **Đột phá:** Những Consumer không bị dính líu đến các Partition bị thu hồi vẫn tiếp tục xử lý data bình thường (Không bị ngắt quãng).
   - Quá trình này diễn ra qua vài vòng lặp (iterations) để đạt được trạng thái ổn định (incremental).
   - Chấm dứt hoàn toàn hiện tượng Stop-the-world.

**C. Cấu hình phân chia (partition.assignment.strategy):**
*(Ghi chú: Mặc định (Default Assignor) của Kafka Client là một danh sách `[RangeAssignor, CooperativeStickyAssignor]`. Hệ thống sẽ ưu tiên dùng RangeAssignor, nhưng thiết kế này cho phép nâng cấp mượt mà lên CooperativeStickyAssignor chỉ với một lần **rolling bounce** (khởi động lại xoay vòng) bằng cách xóa RangeAssignor ra khỏi danh sách).*
*(Lưu ý đối với hệ sinh thái Kafka:)*
*- **Kafka Connect:** Cooperative Rebalance đã được tích hợp và bật mặc định.*
*- **Kafka Streams:** Cooperative Rebalance cũng được bật mặc định thông qua thuật toán `StreamsPartitionAssignor`.*

*(Giả sử hệ thống có 2 Topic: Topic-A (3 Partition A0, A1, A2) và Topic-B (3 Partition B0, B1, B2). Group ban đầu có 2 Consumer: C1, C2)*

*▶ Nhóm Eager Rebalance (Gây Stop-The-World):*

**1. Range Assignor (Mặc định cũ)**
- **Khái niệm:** Phân chia Partition trên cơ sở tính toán độc lập cho từng Topic (per-topic basis).
- **Các bước hoạt động:**
  - Bước 1: Sắp xếp các Partition của Topic đang xét theo thứ tự tăng dần.
  - Bước 2: Sắp xếp các Consumer trong Group theo danh pháp (bảng chữ cái hoặc ID).
  - Bước 3: Lấy tổng số Partition chia cho tổng số Consumer.
  - Bước 4: Cắt lát mảng Partition theo tỷ lệ. Các Consumer đứng đầu danh sách sẽ phải gánh thêm phần dư (nếu có). Lặp lại quy trình với các Topic tiếp theo.
- **Ví dụ chạy thực tế:** Xử lý Topic A: C1 nhận (A0, A1), C2 nhận (A2). Xử lý Topic B: C1 nhận (B0, B1), C2 nhận (B2). Kết quả: C1 quản lý 4 Partition, C2 quản lý 2 Partition.
- **Ưu điểm:** Logic đơn giản, dễ triển khai. Các Partition của cùng 1 Topic có xu hướng hội tụ (co-locate) vào cùng một Consumer.
- **Nhược điểm:** Gây mất cân bằng tải trầm trọng (Imbalance) khi số lượng Topic lớn. Consumer đầu tiên (C1) sẽ liên tục phải chịu tải cao hơn, dẫn đến nguy cơ quá tải cục bộ.

**2. RoundRobin Assignor**
- **Khái niệm:** Loại bỏ ranh giới giữa các Topic. Hợp nhất tất cả Partition và phân bổ tuần tự (round-robin fashion) cho toàn bộ Consumer.
- **Các bước hoạt động:**
  - Bước 1: Tổng hợp tất cả Partition của các Topic mà Group đang theo dõi thành một danh sách duy nhất.
  - Bước 2: Sắp xếp danh sách Partition và danh sách Consumer.
  - Bước 3: Phân bổ tuần tự từng Partition cho từng Consumer cho đến khi duyệt hết danh sách.
- **Ví dụ chạy thực tế:** Tập hợp Partition: (A0, A1, A2, B0, B1, B2). Vòng 1: C1 nhận A0, C2 nhận A1. Vòng 2: C1 nhận A2, C2 nhận B0. Vòng 3: C1 nhận B1, C2 nhận B2.
- **Ưu điểm:** Đạt được độ cân bằng tải tối ưu (Optimal balance). Mỗi Consumer quản lý chính xác 3 Partition.
- **Nhược điểm:** Khi cấu trúc Group thay đổi (ví dụ C3 tham gia), trình tự phân bổ bị xáo trộn hoàn toàn. C1 và C2 có nguy cơ bị thu hồi toàn bộ Partition hiện tại để nhận các Partition mới, gây lãng phí tài nguyên và thời gian tái thiết lập bộ nhớ đệm (State/Cache).

**3. Sticky Assignor**
- **Khái niệm:** Một chiến lược cân bằng tải dựa trên nguyên tắc của RoundRobin, nhưng bổ sung cơ chế "bám dính" (sticky) nhằm tối thiểu hóa số lượng Partition phải di dời (partition movements) mỗi khi có sự kiện thay đổi thành viên trong Group.
- **Các bước hoạt động:**
  - Bước 1: Ở trạng thái ban đầu, hệ thống phân bổ cân bằng tương tự RoundRobin.
  - Bước 2: Khi có Consumer mới (C3) gia nhập, Coordinator kích hoạt Rebalance. Giai đoạn này thuộc nhóm Eager Rebalance, do đó **TẤT CẢ Consumer phải tạm dừng xử lý dữ liệu (Stop-The-World)** và thu hồi quyền kiểm soát Partition.
  - Bước 3: Coordinator tính toán phương án phân bổ mới, với ưu tiên hàng đầu là gán lại các Partition cũ cho chính Consumer đang xử lý nó (C1 và C2).
  - Bước 4: Sau khi bảo toàn tối đa cấu trúc cũ, Coordinator sẽ bóc tách những Partition dư thừa (do mất cân bằng) để gán cho Consumer mới (C3).
- **Ví dụ chạy thực tế (Chi tiết từng bước):**
  - *Trạng thái ban đầu:* Hệ thống có Topic-A (A0, A1, A2) và Topic-B (B0, B1, B2). Group hiện có 2 Consumer. Phân bổ đang tối ưu: C1 xử lý (A0, A2, B1), C2 xử lý (A1, B0, B2).
  - *Sự kiện phát sinh:* C3 khởi động và gửi yêu cầu gia nhập Group.
  - *Bước 1 (Stop-The-World):* Coordinator phát lệnh dừng toàn hệ thống. C1 ngừng đọc A0, A2, B1. C2 ngừng đọc A1, B0, B2. Mọi Partition bị thu hồi tạm thời về trạng thái vô chủ.
  - *Bước 2 (Bảo toàn State):* Coordinator tính toán phân chia lại 6 Partition cho 3 Consumer (mỗi Consumer sẽ nhận 2 Partition). Nó ưu tiên giữ lại các kết nối cũ: Trả lại A0, A2 cho C1; trả lại A1, B0 cho C2.
  - *Bước 3 (Gán phần dư):* Còn lại B1 (trước đây của C1) và B2 (trước đây của C2) được gom lại thành cặp (B1, B2) để gán cho C3.
  - *Kết quả cuối cùng:* C1(A0, A2), C2(A1, B0), C3(B1, B2). Dù được tối ưu không làm xáo trộn các Partition cũ, nhưng ở Bước 1, toàn bộ hệ thống vẫn phải trải qua một nhịp "đóng băng" (downtime).
- **Ưu điểm:** Đảm bảo tải phân bổ đồng đều trong khi vẫn giữ lại phần lớn bộ nhớ đệm (State/Cache) của các Consumer, giúp quá trình phục hồi sau Rebalance diễn ra nhanh hơn.
- **Nhược điểm:** Vẫn tồn tại giai đoạn **Stop-The-World** (Bước 2). Dù thực tế số lượng Partition bị thay đổi rất ít, nhưng toàn bộ Group vẫn bị gián đoạn hoạt động hoàn toàn cho đến khi quá trình tái phân bổ hoàn tất.

*▶ Nhóm Cooperative Rebalance (KHÔNG Stop-The-World):*

**4. Cooperative Sticky Assignor (Khuyên dùng từ Kafka 2.4+)**
- **Khái niệm:** Kế thừa thuật toán tối ưu của Sticky Assignor, nhưng nâng cấp toàn diện cơ chế thực thi sang mô hình Cooperative (Hiệp đồng). Nó cho phép các Consumer giữ lại quyền kiểm soát Partition và **tiếp tục xử lý dữ liệu (keep on consuming)** trong khi quá trình Rebalance diễn ra.
- **Các bước hoạt động:**
  - Bước 1: Khi C3 gia nhập, Coordinator kích hoạt Rebalance lần 1.
  - Bước 2: Đột phá lớn nhất: Các Consumer (C1, C2) **KHÔNG DỪNG HOẠT ĐỘNG**. Chúng tiếp tục xử lý các Partition hiện tại. Dựa trên lệnh từ Coordinator, C1 và C2 chỉ chủ động thu hồi (revoke) các Partition bị dư ra (B1, B2) và trả về cho hệ thống.
  - Bước 3: Coordinator nhận được các Partition vừa thu hồi, lập tức kích hoạt Rebalance lần 2 để gán các Partition đang chờ (B1, B2) cho Consumer mới (C3).
- **Ví dụ chạy thực tế (Chi tiết 2 vòng lặp):**
  - *Trạng thái ban đầu:* C1 đang miệt mài xử lý (A0, A2, B1) và C2 đang xử lý (A1, B0, B2).
  - *Sự kiện phát sinh:* C3 xin gia nhập Group. Coordinator nhận tín hiệu và kích hoạt **Rebalance Vòng 1**.
  - *Vòng 1 (Thu hồi hiệp đồng):* Coordinator gửi tín hiệu "Có thành viên mới (C3), yêu cầu san sẻ 2 Partition". Khác biệt cực lớn: C1 và C2 **KHÔNG ĐÓNG BĂNG TOÀN BỘ**. C1 chỉ chủ động ngắt kết nối với B1, C2 ngắt kết nối với B2. Trong khoảnh khắc đó, C1 VẪN liên tục kéo và xử lý dữ liệu từ A0, A2; C2 VẪN miệt mài xử lý A1, B0 bình thường.
  - *Vòng 2 (Gán Partition trống):* Coordinator thu thập B1 và B2 (vừa bị nhả ra thành vô chủ). Kích hoạt **Rebalance Vòng 2** để chính thức gán B1 và B2 cho C3. C3 bắt đầu vào guồng xử lý.
  - *Kết quả cuối cùng:* Hệ thống hoàn tất việc scale từ 2 lên 3 Consumer mà dòng chảy dữ liệu tại A0, A1, A2, B0 chưa hề bị khựng lại một giây nào (Zero Downtime).
- **Ưu điểm:** Loại bỏ triệt để hiện tượng Stop-The-World. Hệ thống duy trì tính sẵn sàng (Zero Downtime) và hiệu suất cao nhất ngay cả khi mở rộng quy mô.
- **Nhược điểm:** Giao thức xử lý ngầm (internal protocol) trải qua nhiều vòng lặp (iterations) phức tạp hơn. Tuy nhiên, rào cản này đã được các thư viện Kafka Client xử lý hoàn toàn trong suốt với lập trình viên.

**5. Cấu hình Static Group Membership (Ngăn chặn Rebalance không cần thiết)**
- **Khái niệm:** Đây không phải là một thuật toán Assignor, mà là một cấu hình cơ sở hạ tầng (`group.instance.id`) giúp Consumer gắn kết tĩnh (Static) với Broker, ngăn chặn việc kích hoạt Rebalance ngay lập tức khi mạng bị gián đoạn tạm thời.
- **Các bước hoạt động:**
  - Bước 1: Cấu hình định danh tĩnh (VD: `group.instance.id = worker-1`) cho C1. C1 đang quản lý A0, A1.
  - Bước 2: C1 bị khởi động lại để triển khai mã nguồn mới (thời gian mất kết nối khoảng 10 giây).
  - Bước 3: Broker phát hiện C1 mất kết nối, nhưng nhận diện được ID tĩnh nên quyết định đưa C1 vào trạng thái chờ (trong khoảng `session.timeout.ms`) thay vì lập tức kích hoạt Rebalance.
  - Bước 4: C1 khởi động thành công và kết nối lại. Broker trả lại đúng A0, A1 cho C1 tiếp tục xử lý.
- **Ví dụ chạy thực tế (Chi tiết đợt Rolling Update):**
  - *Trạng thái ban đầu:* C1 mang ID tĩnh `pod-worker-1` phụ trách (A0, A1). C2 mang ID tĩnh `pod-worker-2` phụ trách (B0, B1).
  - *Sự kiện phát sinh:* Hệ thống triển khai bản Release mới trên Kubernetes (Rolling Update). Pod C1 nhận lệnh `SIGTERM` và khởi động lại. Quá trình tắt và bật lại mất khoảng 15 giây.
  - *Bước 1 (Mất mạng tạm thời):* Broker phát hiện C1 mất tín hiệu heartbeat. Thay vì hoảng loạn và la lên "C1 chết rồi, Rebalance toàn hệ thống ngay!" (hành vi mặc định Eager), Broker dò thấy cấu hình ID tĩnh `pod-worker-1`. Nó ra quyết định: "Tạm thời đóng băng A0, A1 và cho nó cơ hội quay lại trong vòng `session.timeout.ms` (ví dụ 60 giây)".
  - *Bước 2 (Độ ổn định cực cao):* Trong 15 giây C1 vắng mặt đó, C2 VẪN ung dung xử lý B0, B1 bình thường. Hệ thống hoàn toàn không kích hoạt một đợt Rebalance nào.
  - *Bước 3 (Đoàn tụ):* C1 boot xong, mở lại kết nối tới Broker và tự xưng `group.instance.id = pod-worker-1`. Broker nhận ra người quen, lập tức chuyển giao lại A0, A1 cho C1 tiếp tục công việc.
  - *Kết quả cuối cùng:* Một đợt Deploy rủi ro được hoàn thành mượt mà, sự kiện Rebalance vốn cực kỳ đắt đỏ đã bị "vô hiệu hóa" triệt để.
- **Ưu điểm:** Triệt tiêu hoàn toàn sự gián đoạn do Rebalance trong quá trình triển khai hệ thống (Deployment) hoặc khi hạ tầng mạng thiếu ổn định.
- **Nhược điểm:** Đòi hỏi cấu hình vận hành (DevOps) khắt khe. Phải thiết kế hệ thống cấp phát ID động (ví dụ: gán theo Hostname hoặc StatefulSet Pod Name) để đảm bảo không có hai Consumer nào trùng định danh tĩnh.

### 3.4 Consumer Offsets & Delivery Semantics (Lưu điểm nhớ & Độ tin cậy)
Khi Consumer trong một Group đọc và xử lý xong dữ liệu, nó phải định kỳ báo cáo lại cho Kafka biết "Tôi đã xử lý đến Offset số mấy rồi" bằng hành động **Commit Offset**.
- **Topic ẩn `__consumer_offsets`:** Ít ai biết rằng, Kafka lưu trữ các Offset này vào một Topic nội bộ (internal topic) có tên là `__consumer_offsets`. Việc ghi log này diễn ra ở phía Broker chứ không phải lưu ở Client.
- **Phục hồi sau thảm họa (Fault Tolerance):** Nhờ cơ chế này, nếu một Consumer bị chết (die/crash), khi nó sống lại hoặc một Consumer khác lên thay, nó chỉ việc đọc vào Topic `__consumer_offsets` để biết điểm dừng cuối cùng và tiếp tục đọc (read back from where it left off).

**Các chiến lược Commit (Delivery Semantics):**
Mặc định Java Consumer tự động commit (Auto Commit), dẫn đến ngữ nghĩa "At-least-once". Nhưng nếu bạn dùng Manual Commit, bạn sẽ đối mặt với 3 kịch bản:

1. **At-most-once (Tối đa một lần - Rất dễ mất Data):**
   - Offset được commit *ngay lập tức* ngay khi vừa nhận được message từ Kafka (chưa thèm xử lý logic).
   - *Hậu quả:* Nếu trong lúc ứng dụng đang tính toán hoặc lưu DB mà bị crash, message đó bị **mất vĩnh viễn** (vì Kafka đã ghi nhận là đọc rồi và không gửi lại nữa).

2. **At-least-once (Ít nhất một lần - Khuyên dùng/Thường thấy nhất):**
   - Offset chỉ được commit *sau khi* message đã được xử lý xong hoàn toàn (ví dụ: Insert DB thành công).
   - *Hậu quả:* Nếu xử lý DB thành công nhưng chưa kịp Commit mà App bị crash -> Khi khởi động lại, App sẽ đọc và xử lý lại message đó -> Gây ra hiện tượng **Xử lý đúp (Duplicate Processing)**.
   - *Giải pháp:* Ứng dụng của bạn bắt buộc phải có tính **Idempotent** (có chạy lại message đó bao nhiêu lần thì kết quả hệ thống vẫn không thay đổi).

3. **Exactly-once (Chính xác một lần - Chén thánh):**
   - Nếu luồng dữ liệu chỉ nằm gọn trong Kafka (Kafka to Kafka workflows): Sử dụng **Transactional API** (rất dễ làm nếu dùng Kafka Streams API).
   - Nếu luồng dữ liệu đẩy ra hệ thống ngoài (External System workflows như ghi vào Postgres, MySQL): Cực khó. Bắt buộc phải kết hợp *At-least-once* với một **Idempotent Consumer** (thường dùng khóa ngoại hoặc bảng lưu Transaction ID để chặn trùng lặp tuyệt đối).

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
