# CẨM NANG TOÀN TẬP VỀ INTERNET & KIẾN TRÚC MẠNG (MASTER NETWORK GUIDE)

Tài liệu này là sự kết hợp toàn vẹn mọi kiến thức nền tảng thiết yếu nhất về mạng máy tính (Networking) và bảo mật hạ tầng mà một DevOps, Backend Engineer, và Data Engineer cần phải nắm vững. Cẩm nang được biên soạn theo chuẩn học thuật (Network Engineering) nhưng đi kèm các ví dụ kỹ thuật thực tiễn.

---

## 1. Mô hình Tham chiếu & Quá trình Đóng gói Dữ liệu (Encapsulation)

### 1.1. Mô hình OSI (7 Lớp) và TCP/IP (4 Lớp)
- **Layer 7 - Application (Ứng dụng):** Nơi phần mềm (Chrome, Nginx) hoạt động. Giao thức: HTTP, HTTPS, SSH, DNS.
- **Layer 6 - Presentation (Trình diễn):** Mã hóa, nén dữ liệu (SSL/TLS mã hóa Payload ở đây).
- **Layer 5 - Session (Phiên):** Quản lý phiên làm việc giữa các ứng dụng.
- **Layer 4 - Transport (Giao vận):** Đảm bảo truyền tải. Quản lý các Cổng (Port). Giao thức: TCP, UDP.
- **Layer 3 - Network (Mạng):** Định tuyến đường đi toàn cầu. Giao thức: IP, ICMP (Ping).
- **Layer 2 - Data Link (Liên kết dữ liệu):** Truyền dữ liệu trong mạng LAN vật lý bằng MAC Address. Giao thức: Ethernet, Wi-Fi.
- **Layer 1 - Physical (Vật lý):** Truyền tải tín hiệu điện/quang/sóng vô tuyến qua dây dẫn.

### 1.2. Cấu tạo Gói tin & Quá trình Đóng gói (Encapsulation)
1. **Payload (Lớp 7):** Khối dữ liệu gốc thực sự (Ví dụ: chuỗi JSON chứa thông tin user).
2. **Segment / Datagram (Lớp 4):** Payload được bọc thêm **TCP Header** (chứa Source/Dest Port) thành **Segment**.
3. **Packet (Lớp 3):** Segment được bọc thêm **IP Header** (chứa Source/Dest IP) thành **Packet**.
4. **Frame (Lớp 2):** Packet đi xuống Card mạng, được bọc thêm **MAC Header** và đuôi kiểm lỗi FCS thành **Frame**.
5. **Bitstream (Lớp 1):** Băm nhỏ thành các xung điện (0 và 1) phóng ra cáp mạng.

### 1.3. Triết lý Thiết kế: Tại sao lại sinh ra MAC, IP và Port?
Nhiều người thắc mắc: *"Tại sao phải đẻ ra tận 3 loại địa chỉ lằng nhằng này? Sao không dùng chung 1 cái?"*. Đây là câu trả lời triệt để:

- **MAC Address (Định danh Phần cứng):** Nó giống như **Căn cước Công dân (CCCD)** hoặc Số khung xe máy. Nó là định danh bất di bất dịch, gán chết vào con chip phần cứng của Card mạng từ nhà máy. Nó giúp nhận diện đích danh thiết bị trong phạm vi hẹp (LAN). Nhưng CCCD không cho biết bạn đang sống ở quốc gia nào! Nếu chỉ dùng MAC, Internet sẽ không thể nào tìm đường (Route) để đẩy gói tin.
- **IP Address (Định danh Vị trí / Địa lý):** Nó giống như **Địa chỉ Thường trú (Số nhà, Tên đường, Thành phố)**. Khi bạn xách laptop từ Việt Nam sang Mỹ, số MAC (CCCD) của bạn không đổi, nhưng số IP (Địa chỉ nhà) của bạn bắt buộc phải đổi để thế giới biết bạn đang ở Mỹ mà gửi thư đến.
  - *Ai là người cấp phát IP?* IP không tự sinh ra. Khi máy tính cắm dây vào mạng (hoặc bắt Wi-Fi), cục Modem/Router của nhà mạng sẽ chạy một giao thức gọi là **DHCP (Dynamic Host Configuration Protocol)** để động cấp phát (cho thuê) một IP Private và gắn nó với cái MAC Address của bạn. Khi bạn ngắt mạng rời đi, Router sẽ thu hồi IP đó lại và cấp cho người khác.
  - *Tại sao 1 máy (1 MAC) lại có nhiều IP?* Một máy tính có thể cắm cáp LAN (nhận IP A), đồng thời bật Wi-Fi (nhận IP B), đồng thời chạy phần mềm VPN (nhận IP ảo C). IP là thứ vô hình, đại diện cho một kết nối logic (Logical Interface) chứ không bị trói buộc 1-1 với phần cứng.
- **Port (Định danh Tiến trình Phần mềm):** Nó giống như **Số Phòng trong 1 Tòa nhà**. Khi bưu tá (Hệ điều hành) dựa vào IP để mang bưu phẩm đến đúng tòa nhà (Máy tính), họ mở cửa bước vào và thấy trong nhà có... 100 phần mềm đang chạy cùng lúc (Chrome, Zalo, Nginx, Database...). Bưu phẩm này dành cho phần mềm nào? Port chính là Số phòng. Gói tin gắn Port 80 sẽ được HĐH ném cho Nginx, gắn Port 5432 sẽ được ném cho Postgres. Một IP có 65535 Port, giúp máy tính có thể làm hàng vạn việc song song!

---

## 2. Cấu trúc Hạ tầng Vật lý & Logic Truyền tải

### 2.1. Card mạng (NIC) & Chuyển đổi Tín hiệu (Layer 1 & 2)
- **Modulation (Điều chế tín hiệu):** Máy tính chỉ hiểu 0 và 1. NIC dịch các bit này thành vật lý.
  - *Cáp đồng:* Dịch bit `1` thành điện áp +5V, bit `0` thành 0V.
  - *Cáp quang:* Kích hoạt đèn tia laser chớp tắt cực nhanh.
  - *Wifi:* Biến thiên tần số hoặc biên độ sóng vô tuyến (Radio Waves).

### 2.2. Switch & Cơ chế Broadcast / ARP (Layer 2)
Switch là "Trạm thu phí" nội bộ của mạng LAN.
- **MAC Address Table (Lưu trên Switch):** Bảng nhớ lưu ánh xạ giữa Cổng (Port vật lý cắm dây) và MAC Address. *(Lưu ý cực kỳ quan trọng: Bảng này KHÔNG chứa IP. Switch hoàn toàn "mù" về IP. Nó chỉ biết "Lỗ cắm cáp số 1 đang nối với thiết bị có MAC là A")*.
- **ARP Table (Lưu trên Máy tính / Router):** Đây mới là bảng lưu ánh xạ (Mapping) giữa `IP Address` và `MAC Address` (Ví dụ: `192.168.1.1` đi liền với MAC `00:1A:2B...`).
- **Cơ chế Broadcast & Flooding (ARP):** Khi Máy A biết IP của Máy B, nhưng chưa biết MAC, Máy A gửi một Frame **Broadcast** tới MAC đích là `FF:FF:FF:FF:FF:FF` (nghĩa là "Gửi tất cả"). Switch nhận được liền "hét lên" (Flooding) đẩy dòng điện ra tất cả các cổng. Máy B nghe thấy, trả lời lại số MAC thật. Từ đó Switch cập nhật bảng nhớ và chỉ truyền trực tiếp (Unicast).

### 2.3. Router & Bản đồ Định tuyến (Layer 3)
Router kết nối các mạng LAN khác nhau. Router "xé" Lớp 2 (Frame), đọc Lớp 3 (Packet) lấy **Destination IP**. Router dò **Bảng định tuyến (Routing Table)** để tìm đường (Cổng mạng) gần nhất, đóng bọc Lớp 2 mới và đẩy tín hiệu đi (Hop-by-Hop).

---

## 3. Bản chất IP, Subnet và Giao thức BGP Toàn cầu

### 3.1. Phân lớp Địa chỉ IPv4 (Classful Addressing)
Vào thời kỳ sơ khai của Internet, tổ chức IANA chưa có khái niệm chia nhỏ Subnet (CIDR) như bây giờ. Họ chặt không gian 4.3 tỷ địa chỉ IPv4 thành 5 Lớp (Classes) cố định dựa vào những bit đầu tiên của địa chỉ. Việc hiểu rõ 5 lớp này là nền tảng để phân biệt đâu là mạng khổng lồ, đâu là mạng nhỏ, và đâu là mạng chuyên dụng.

- **Lớp A (Class A):** Dành cho các Mạng lưới Khổng lồ (Chính phủ, Nhà mạng)
  - **Dải IP:** Từ `1.0.0.0` đến `126.255.255.255`.
  - **Đặc điểm:** Mặc định sử dụng Subnet Mask `/8`. Nghĩa là chỉ dùng 1 Octet đầu để định danh Mạng, dành tận 3 Octet sau cho Host. Một mạng Lớp A có thể chứa tới **16.7 triệu máy tính**.
  - *(Lưu ý: Dải `127.x.x.x` bị bỏ qua vì nó được dùng làm Loopback Address (Localhost) để máy tính tự gọi chính mình).*

- **Lớp B (Class B):** Dành cho các Tổ chức/Doanh nghiệp quy mô Vừa và Lớn
  - **Dải IP:** Từ `128.0.0.0` đến `191.255.255.255`.
  - **Đặc điểm:** Mặc định sử dụng Subnet Mask `/16`. Dùng 2 Octet cho Mạng, 2 Octet cho Host. Mỗi mạng Lớp B chứa được khoảng **65,536 máy tính**.

- **Lớp C (Class C):** Dành cho các Mạng Nhỏ (Quán Cafe, Gia đình, Văn phòng chi nhánh)
  - **Dải IP:** Từ `192.0.0.0` đến `223.255.255.255`.
  - **Đặc điểm:** Mặc định sử dụng Subnet Mask `/24`. Dùng 3 Octet cho Mạng, chỉ chừa 1 Octet cuối cho Host. Một mạng Lớp C chỉ chứa tối đa **254 máy tính**.

- **Lớp D (Class D):** Chuyên dụng cho Multicast (Truyền phát đa hướng)
  - **Dải IP:** Từ `224.0.0.0` đến `239.255.255.255`.
  - **Đặc điểm:** Lớp này KHÔNG dùng để cấp cho máy tính cá nhân (Không có Network ID hay Host ID). Nó được dùng để truyền tin **Multicast** (Gửi 1 gói tin nhưng nhiều máy nhận cùng lúc).
  - *Ví dụ thực tiễn:* Truyền hình cáp IPTV (Khi VTV phát 1 trận bóng đá trực tiếp, họ không gửi 1 triệu gói tin cho 1 triệu tivi, họ chỉ gửi 1 gói tin Multicast vào dải Lớp D, tivi nào muốn xem thì "Subscribe" vào nhóm Multicast đó để nhận). Hoặc dùng cho các con Router nói chuyện với nhau (Giao thức OSPF).

- **Lớp E (Class E):** Dành cho Nghiên cứu & Quân sự (Reserved)
  - **Dải IP:** Từ `240.0.0.0` đến `255.255.255.255`.
  - **Đặc điểm:** Lớp này bị niêm phong, không bao giờ được cấp phát ra Internet công cộng. Cục IETF giữ lại để nghiên cứu và dự phòng. Riêng địa chỉ cuối cùng `255.255.255.255` được quy ước làm địa chỉ **Broadcast** toàn cầu.

> 💡 **Bí ẩn Toán học: Tại sao lại sinh ra các con số 126, 128, 192?**
> Vào những năm 1980, bộ định tuyến (Router) cực kỳ yếu. Để giúp Router nhận diện một IP thuộc Lớp nào nhanh như chớp, các nhà khoa học máy tính đã dùng cơ chế **"Bit dẫn đầu" (High-Order Bits)** của Octet đầu tiên làm lá cờ hiệu (Flag):
> - **Lớp A:** Bit dẫn đầu luôn khóa cứng là `0`. (Dải nhị phân từ `00000000` đến `01111111` ➔ Đổi ra thập phân chính là **0 đến 127**).
> - **Lớp B:** Bit dẫn đầu luôn khóa cứng là `10`. (Dải nhị phân từ `10000000` đến `10111111` ➔ Đổi ra thập phân là **128 đến 191**).
> - **Lớp C:** Bit dẫn đầu luôn khóa cứng là `110`. (Dải nhị phân từ `11000000` đến `11011111` ➔ Đổi ra thập phân là **192 đến 223**).
> - **Lớp D:** Khóa cứng `1110` ➔ Thập phân **224 đến 239**.
> - **Lớp E:** Khóa cứng `1111` ➔ Thập phân **240 đến 255**.
> Chính luật khóa bit thiên tài này đã tự động sinh ra những ranh giới toán học bất di bất dịch của Internet sơ khai mà không cần đến Subnet Mask!

> ⚠️ **Sự nhầm lẫn kinh điển (Public vs Private):**
> Đừng nhầm lẫn giữa Lớp IP (Classful) và IP Nội bộ (Private IP). Ví dụ: Lớp A kéo dài từ `1` đến `126`, nhưng IANA chỉ trích ra ĐÚNG MỘT DẢI DUY NHẤT trong đó là `10.0.0.0/8` để làm IP Private. Toàn bộ các IP Lớp A còn lại (như `8.8.8.8` của Google, hay `1.1.1.1` của Cloudflare) đều là IP Public được mua bằng tiền tỷ! Tương tự với Lớp B (trích dải `172.16`) và Lớp C (trích dải `192.168`).

### 3.2. Subnet Mask (Mặt nạ mạng), CIDR & Phép toán Bitwise AND
Dải IP luôn đi kèm Subnet Mask (vd `/24` hoặc `255.255.255.0`). Nhiệm vụ duy nhất của mặt nạ mạng là làm cái thớt để "chặt" địa chỉ IP ra làm 2 phần: **Network ID** (Khu phố) và **Host ID** (Số nhà). Việc tự do gán Subnet Mask không màng tới Lớp A,B,C được gọi là chuẩn CIDR (Classless Inter-Domain Routing).

**Cách xác định Khúc nào là Network, Khúc nào là Host:**
- Bản chất IP IPv4 gồm 4 cụm số (4 Octet), tổng cộng 32-bit.
- Con số `/24` mang ý nghĩa: **24 bit đầu tiên** (tương đương 3 cụm số đầu) bị "khóa cứng" làm **Network ID**. 8 bit còn lại (cụm số cuối) được tự do cấp phát làm **Host ID**.
- Nhìn dưới góc độ số thập phân, `/24` tương đương với `255.255.255.0`. Số `255` mang ý nghĩa "Khóa chặt phần này làm Network", số `0` mang ý nghĩa "Phần này là Host".
- *Ví dụ với IP `192.168.1.5` đi kèm mask `255.255.255.0`:*
  - **Network ID:** `192.168.1.0` (3 cụm đầu bị khóa cứng. Đây chính là "Tên khu phố").
  - **Host ID:** `.5` (Cụm cuối cùng tự do. Đây là "Số nhà" của bạn trong khu phố đó).

**Sự đánh đổi (Trade-off) giữa Network và Host:**
- Tổng số bit luôn là 32. Nếu bạn lấy nhiều bit làm Network, bạn sẽ có ít bit làm Host (số IP cấp phát được ít đi), và ngược lại.
- Với Subnet `/24` (24 bit Network, 8 bit Host), Router/DHCP của bạn chỉ có thể cấp tối đa $2^8 - 2 = 254$ số nhà (IP) cho các thiết bị truy cập. (Lý do trừ 2: IP `.0` dành riêng cho tên Network, IP `.255` dành riêng cho lệnh Broadcast hét lên toàn mạng).
- Nếu hệ thống của bạn có 1000 user truy cập, dải `/24` sẽ không đủ IP để cấp! Lúc này bạn phải mượn thêm bit, ví dụ xài dải `/22` (22 bit Network, 10 bit Host). Lúc này Router sẽ có $2^{10} - 2 = 1022$ cái IP để tha hồ cấp phát.
- *Câu hỏi đặt ra:* **Tại sao không xài luôn Subnet `/8` để có tận 16 triệu IP ($2^{24}$) cấp cho sướng, khỏi lo hết IP?**
  1. **Thảm họa Bão Broadcast (Broadcast Storm):** Nhớ lại giao thức ARP ở Lớp 2, khi 1 máy muốn tìm MAC, nó phải hét lên (Broadcast) cho TẤT CẢ các máy trong cùng mạng LAN. Nếu bạn nhét 16 triệu thiết bị vào chung 1 mạng LAN `/8`, chỉ cần 1% số máy đó hét lên tìm MAC thôi là phần cứng Switch sẽ bốc cháy, toàn bộ 16 triệu máy tính phải dừng công việc lại để lắng nghe rác. Mạng lưới sẽ sập ngay lập tức!
  2. **Vấn đề Bảo mật (Security):** Trong cùng 1 mạng LAN, các máy tính nói chuyện trực tiếp với nhau qua Switch mà không qua tường lửa của Router. Nếu nhét cả Server DB, máy Kế Toán, và Guest Wi-Fi vào chung 1 dải `/8`, một con virus từ máy khách có thể scan và lây lan sang máy Server ngay lập tức. Bằng cách băm nhỏ ra thành nhiều Subnet `/24` khác nhau, muốn đi từ Subnet Guest sang Subnet DB bắt buộc phải đi vòng qua Router, lúc này Tường lửa trên Router sẽ chặn đứng con virus lại!
  3. **Giới hạn phần cứng Switch:** Switch lưu bảng MAC Address Table trên bộ nhớ RAM đắt tiền (CAM Table). Không có con Switch vật lý nào đủ RAM để chứa được 16 triệu cái MAC Address cả. Khi đầy RAM, nó sẽ biến thành một cục Hub ngu ngốc, xả rác ra mọi cổng.

**Cơ chế định tuyến nội bộ thông minh của HĐH:**
Làm sao máy tính của bạn biết lúc nào thì ném gói tin cho Switch (trong nhà), lúc nào thì ném cho Router (ra ngoài Internet)? Nó dùng **Phép toán Bitwise AND**.
*Ví dụ: Máy tính của bạn có IP `192.168.1.5`, Subnet Mask là `255.255.255.0`.*
- *Lưu ý toán học:* Tại sao `192.168.1.5 [AND] 255.255.255.0` lại ra `192.168.1.0`? Vì phép `AND` nhị phân quy định `1 AND 1 = 1`, còn lại bằng `0`. Số `255` là `11111111`, khi AND với số nào thì giữ nguyên số đó. Còn số `0` là `00000000`, khi AND với số `5` (tức `00000101`) sẽ triệt tiêu về `0`. Do đó kết quả bóc tách ra được phần Network ID là `192.168.1.0`!
- **Trường hợp 1 (Gửi cho máy tính kế bên `192.168.1.10`):** Máy bạn lấy IP đích `192.168.1.10 [AND] 255.255.255.0 = 192.168.1.0`. Máy bạn so sánh thấy Network ID này **TRÙNG** với Network ID của nó. Thế là nó tra bảng ARP lấy MAC rồi quăng thẳng Frame vào **Switch**.
- **Trường hợp 2 (Gửi cho Google `8.8.8.8`):** Máy bạn lấy IP đích `8.8.8.8 [AND] 255.255.255.0 = 8.8.8.0`. Máy bạn so sánh thấy **KHÁC** với Network ID của mình (`192.168.1.0`). Nó bọc gói tin lại và ném thẳng cho **Router (Default Gateway)** để Router tự lo liệu tìm đường.

### 3.2. BGP (Border Gateway Protocol) - Bản Đồ Định Tuyến Của Internet
Làm sao máy tính ở Việt Nam biết đường bắn gói tin sang tận máy chủ AWS ở Mỹ mà không bị lạc giữa hàng triệu rễ cây cáp quang dưới đáy biển? Câu trả lời là nhờ giao thức **BGP**.

- **Hệ thống tự trị (Autonomous System - AS):** Internet không có một máy chủ trung tâm hay "Giám đốc" nào điều phối cả. Nó là một mạng lưới chắp vá khổng lồ từ hàng vạn hệ thống mạng độc lập của các tập đoàn (Ví dụ: Mạng của Viettel là một AS, mạng của FPT là một AS, mạng của Google/AWS là một AS). Mỗi tổ chức này được tổ chức quản lý mạng thế giới cấp một mã số định danh duy nhất gọi là **ASN**.
- **Cơ chế "Ngoại giao" (Peering & Transit):** BGP là ngôn ngữ ngoại giao giữa các quốc gia AS. Các Router khổng lồ ở biên giới sẽ liên tục phát các bản tin **BGP Announcements (Loan báo định tuyến)**. Trong thế giới thực, hành động "Ngoại giao" này được vận hành dựa trên 2 hình thức thương mại (Tiền bạc) cốt lõi:
  - **Peering (Kết nối ngang hàng - Miễn phí):** Hai AS (Ví dụ: VNPT và Google) kéo cáp nối trực tiếp Router với nhau (thường tại các trạm trung chuyển Internet - IXP). Họ thỏa thuận trao đổi dữ liệu miễn phí (Settlement-free). Nhờ đó, user VNPT xem Youtube cực nhanh vì data chạy thẳng từ máy chủ Google sang thẳng VNPT mà không qua trung gian. BGP lúc này chỉ loan báo dải IP của riêng 2 bên cho nhau.
  - **Transit (Mua đường truyền - Trả phí):** VNPT không thể tự kéo cáp nối trực tiếp với 100,000 cái mạng AS trên toàn thế giới được. VNPT bắt buộc phải vác tiền đi thuê một "Ông trùm" mạng viễn thông cấp 1 (Tier-1 ISP như AT&T, Tata). BGP Router của Tier-1 sẽ loan báo cho VNPT: *"Tao có bản đồ đường đi đến TOÀN BỘ thế giới, cứ quăng hết data cho tao, tao tính tiền cước theo dung lượng (GB)"*.
- **Thuật toán tìm đường:** BGP không hề chọn đường đi dựa vào cáp quang nào nhanh nhất, mà nó chọn đường theo **AS Path** (Chặng đường đi qua ít quốc gia AS nhất) và đặc biệt là theo **Chính sách Kinh doanh**.
  - *Ví dụ thực tiễn:* Kỹ sư mạng VNPT sẽ viết code cấu hình BGP ưu tiên đẩy Traffic qua đường **Peering** (vì nó miễn phí). Chỉ khi cáp Peering đứt, Router mới ngậm ngùi bẻ lái (Failover) đẩy traffic qua đường **Transit** đắt tiền. Sự điều hướng dòng chảy Data trên toàn cầu thực chất được quyết định bởi... Bài toán Kinh tế!
- *Thảm họa thực tế:* Nhờ BGP mà các Packet luôn tìm được đường tới đích. Nhưng nếu một kỹ sư mạng ở Facebook gõ sai cấu hình, vô tình xóa mất lời "loan báo" BGP của hệ thống Facebook, thì toàn bộ dải IP của Facebook sẽ lập tức "tàng hình" khỏi bản đồ Internet thế giới, khiến mạng lưới sập toàn cầu (Điều này đã từng xảy ra trên thực tế khiến toàn bộ Facebook, Messenger, Instagram sập sạch vào tháng 10/2021).

### 3.3. Ứng dụng Thực tiễn: Quy hoạch VPC trên Cloud (AWS)
Từ mớ lý thuyết khô khan về Subnet Mask và Private IP ở trên, khi áp dụng vào thực tế triển khai Cloud (AWS/GCP), các Kỹ sư Hệ thống sẽ tư duy quy hoạch hạ tầng (Infrastructure as Code) như sau:

**1. Khai báo Mạng ảo (Virtual Private Cloud - VPC):**
AWS yêu cầu bạn phải tạo một "Mảnh đất tổng" (VPC) để chứa toàn bộ Server. Tiêu chuẩn vàng (Golden Standard) luôn là chọn dải IP **Lớp A (`10.0.0.0/16`)**. Tại sao?
- **Tránh đụng độ Lớp C:** Tuyệt đối không dùng `192.168.0.0/16` vì 99% cục Router Wi-Fi gia đình đang xài dải này. Nếu bạn bật VPN từ nhà kết nối lên Cloud, HĐH máy tính của bạn sẽ bị "lú" vì Routing Table bị trùng lặp, không phân biệt được IP Server trên Cloud và IP cái Tivi thông minh ở nhà.
- **Tránh đụng độ Lớp B:** Không dùng dải `172.x.x.x` vì mạng mặc định của AWS (Default VPC) luôn lấy dải `172.31.0.0/16`, và mạng ảo của Docker Containers cũng tự động xí phần dải `172.17.0.0/16`.
- Dải **Lớp A (`10.0.0.0/16`)** cực kỳ rộng lớn (65,536 IP) và là dải sạch sẽ nhất, an toàn tuyệt đối khi cần đấu nối mạng nối mạng ghép (VPC Peering/Transit Gateway) với công ty đối tác.

**2. Băm nhỏ Mảnh đất (Subnetting):**
Từ mảnh đất khổng lồ `/16`, hệ thống bắt buộc phải cắt nhỏ thành các khu phố (Subnet) hẹp hơn, thường dùng Subnet `/24` (256 IP mỗi khu):
- `10.0.1.0/24`: **Public Subnet** (Khu mặt tiền). Cắm Internet Gateway, cấp Public IP. Dành cho Web Server, Load Balancer.
- `10.0.2.0/24`: **Private Subnet** (Khu hầm ngầm). Không có đường ra Internet. Dành cho Database.

Logic phân lô này bằng Toán học Nhị phân chính là công cụ mạnh nhất để cô lập rủi ro bảo mật (Blast Radius). Khi áp dụng thực tế vào file code `network.tf` của Terraform, chỉ cần nhìn dòng chữ `cidr_block = "10.0.1.0/24"`, một kỹ sư mạng sẽ tự động hiểu rõ quy mô của toàn bộ hệ thống!

---

## 4. NAT (Network Address Translation)

Vào những năm 1990, thế giới nhận ra thảm họa: IPv4 chỉ có 4.3 tỷ địa chỉ Public, sớm muộn gì cũng cạn kiệt. Nếu mỗi cái điện thoại, máy tính, tủ lạnh đều chiếm 1 IP Public thì Internet đã "chết" từ năm 2000. Giải pháp vĩ đại được đẻ ra chính là NAT.

Tổ chức IANA quy hoạch ra các dải **IP Private** (`10.x.x.x`, `172.16.x.x`, `192.168.x.x`). IP Private hoàn toàn "vô hình" trên Internet (Các Router công cộng giữa đại dương sẽ DROP ngay lập tức nếu thấy IP Private). NAT cho phép hàng nghìn máy tính trong mạng LAN dùng IP Private, nhưng khi ra Internet thì "núp bóng" chung **Duy nhất 1 IP Public** của Router.

**Cách NAT hoạt động (PAT - Port Address Translation):**
Làm sao 1 con Router với 1 IP Public có thể gánh cho 10 người cùng lướt web mà không bị gửi nhầm dữ liệu của người này sang màn hình người kia? Bí mật nằm ở việc Router không chỉ tráo IP, mà nó tráo luôn cả Cổng (Port)!

- **1. NAT Outbound (Chiến dịch tráo nhãn đi ra):**
  - Bạn mở Chrome, máy bạn (`192.168.1.5`) dùng Port ngẫu nhiên `51000` đẩy Packet ra Router.
  - Router chặn Packet lại, xé bọc Lớp 3 (Network) và Lớp 4 (Transport).
  - Nó xóa `IP Nguồn: 192.168.1.5` và xóa luôn `Port Nguồn: 51000`.
  - Nó đè **IP Public** của nó vào (vd `14.22.33.44`), và chọn ngẫu nhiên một cái Port trống của nó (vd `33333`) gắn vào Packet. Rồi ném ra Internet.
  - Đồng thời, Router lấy sổ tay ra ghi chú vào **NAT Session Table**: *"Mình vừa lấy Port `33333` của mình để đi chợ dùm thằng cu `192.168.1.5:51000` trong nhà"*.

- **2. NAT Inbound (Chiến dịch tráo nhãn trả về):**
  - Web Server (như Facebook) xử lý xong, gửi Response trả về đúng IP Public và Port của Router (`14.22.33.44:33333`).
  - Router nhận được gói hàng. Nó mở sổ tay **NAT Session Table** ra dò: *"À, gói hàng gửi vào Port 33333 là đồ của thằng cu 192.168.1.5:51000"*.
  - Router bóc lớp vỏ `14.22.33.44:33333` vứt đi, dán đè nhãn `192.168.1.5:51000` vào và bơm ngược gói tin vào mạng LAN thẳng tới máy tính của bạn. Gói tin đi xuyên suốt mà Trình duyệt không hề hay biết đã bị tráo nhãn!

- **3. NAT Gateway trên Cloud (AWS/GCP):**
  - Trên AWS, các máy chủ Backend/Database luôn được đặt ở **Private Subnet** (chỉ có IP Private để bảo mật tuyệt đối, chống Hacker từ Internet rà quét).
  - Nhưng nếu không có IP Public, làm sao Server ra ngoài tải được bản cập nhật Linux (apt update)?
  - Lúc này DevOps phải tạo một con **NAT Gateway** (hoặc NAT Instance) đặt ở Public Subnet. Con NAT Gateway này hoạt động y chang cục Router ở nhà bạn: Nó gom tất cả request mồ côi từ các Server nội bộ, lấy danh nghĩa IP Public của nó đi tải bản cập nhật về, rồi bóc nhãn chuyển trả vào lại cho từng Server nội bộ tương ứng một cách an toàn.

---

## 5. Mật mã học Cơ bản & Nguyên lý SSL/TLS

### 5.1. Bài toán: Tại sao cần SSL/TLS?
Internet nguyên thủy truyền dữ liệu dạng văn bản rõ (Clear Text / HTTP). Một gói tin đi từ VN sang Mỹ phải nhảy qua hàng tá Router (Viettel, Singtel...). Bất kỳ kỹ sư mạng nào ở các trạm trung gian (hoặc Hacker nghe lén Wi-Fi quán cafe) đều có thể bắt gói tin (Sniffing) và đọc được Mật khẩu, Thẻ tín dụng của bạn. 
Giải pháp: Giao thức **SSL/TLS (HTTPS)** sinh ra để "băm nát" và mã hóa cái Payload trước khi ném xuống tầng TCP.

### 5.2. Sự kết hợp hoàn hảo giữa 2 loại Mã hóa (Cơ chế Toán học)
Nếu chỉ dùng 1 loại mã hóa thì mạng Internet sẽ hoặc là "Quá chậm", hoặc là "Quá kém bảo mật". Do đó SSL/TLS đã kết hợp cả 2 loại với nhau:

1. **Mã hóa Bất đối xứng (Asymmetric Cryptography - Tiêu biểu: RSA, ECC):**
   - **Bản chất Toán học:** Dựa trên các bài toán hàm một chiều (Trapdoor function) mà máy tính dễ dàng tính xuôi nhưng gần như không thể giải ngược (nếu không có tham số bí mật). Ví dụ, thuật toán RSA dựa trên bài toán **Phân tích thừa số nguyên tố**: Rất dễ để lấy 2 số nguyên tố siêu lớn nhân với nhau ra một tích khổng lồ, nhưng các siêu máy tính cũng cần hàng triệu năm để từ tích khổng lồ đó mò ngược ra 2 số nguyên tố gốc.
   - **Cơ chế hoạt động (Dòng chảy Kỹ thuật Thực tế):** Toán học sinh ra một cặp khóa có liên hệ mật thiết: Khóa công khai (**Public Key**) và Khóa bí mật (**Private Key**). 
     - **Bước 1 (Server phân phối khóa):** Server tạo ra cặp khóa, lưu trữ Private Key vào phân vùng tuyệt mật của hệ điều hành. Sau đó, Server phát Public Key công khai ra Internet cho bất kỳ Client (User) nào muốn kết nối.
     - **Bước 2 (Client mã hóa):** Trình duyệt của User sử dụng Public Key vừa nhận được để mã hóa dữ liệu nhạy cảm (ví dụ: mã hóa cái Khóa đối xứng). Theo quy luật hàm một chiều, dữ liệu sau khi mã hóa bởi Public Key sẽ biến thành một khối byte hỗn độn mà **chỉ có Private Key tương ứng mới có khả năng tính toán dịch ngược lại được**.
     - **Bước 3 (Server giải mã):** Server nhận được khối byte mã hóa, sử dụng Private Key của mình làm tham số bí mật để giải phương trình toán học ngược, từ đó thu lại được dữ liệu gốc của User.
     - **Góc nhìn của Hacker:** Hacker đứng lén giữa mạng lưới bắt được trọn vẹn cả "Public Key" (ở Bước 1) và "Dữ liệu đã mã hóa" (ở Bước 2). Tuy nhiên, vì không có Private Key, Hacker phải sử dụng siêu máy tính để cố gắng Brute-force (mò) ngược từ Public Key ra Private Key. Dựa trên rào cản của bài toán Phân tích thừa số nguyên tố (phải tìm 2 số nguyên tố cấu thành một tích dài hàng nghìn bit), mạng lưới siêu máy tính của Hacker sẽ phải chạy liên tục trong **vài triệu năm** mới giải ra được. Điều này khiến nỗ lực bẻ khóa trở nên hoàn toàn vô vọng!
   - **Đặc điểm:** Giải quyết triệt để bài toán phân phối khóa trên môi trường Public không an toàn. Nhược điểm chí mạng là phép toán lũy thừa số lớn (Modular Exponentiation) cực kỳ cồng kềnh, **Tốc độ cực kỳ chậm**, ngốn rất nhiều tài nguyên CPU.
   
2. **Mã hóa Đối xứng (Symmetric Cryptography - Tiêu biểu: AES, ChaCha20):**
   - **Bản chất Toán học:** Hoạt động dựa trên các phép toán biến đổi ma trận (Substitution-Permutation Network) và các toán hạng Bitwise tuyến tính (như phép XOR). Dữ liệu gốc (Plaintext) được băm thành các khối (Block Cipher) hoặc dòng (Stream Cipher), sau đó trải qua nhiều vòng (rounds) trộn lẫn, dịch bit và hoán vị với Khóa.
   - **Cơ chế hoạt động:** Chỉ tồn tại **một Khóa duy nhất (Symmetric Key)** cho cả hai quá trình Mã hóa và Giải mã. Nếu Plaintext ⊕ Key = Ciphertext, thì Ciphertext ⊕ Key = Plaintext (Bản chất của phép XOR).
   - **Đặc điểm:** Các phép toán dịch bit và XOR được các dòng CPU hiện đại hỗ trợ thẳng ở cấp độ phần cứng (như tập lệnh AES-NI). Do đó, **Tốc độ xử lý cực kỳ nhanh**, đáp ứng được băng thông khổng lồ của Streaming Video hay truyền File lớn. Nhược điểm chí mạng là: Làm sao để Client gửi được chiếc Khóa duy nhất này sang cho Server an toàn?

-> **Giải pháp thiên tài của TLS (Hybrid Encryption - Mã hóa Lai):**
Thay vì đắn đo chọn 1 trong 2, TLS đã kết hợp cả hai để tạo ra một kịch bản hoàn hảo:
- **Bước 1 (Dùng Bất đối xứng để bọc khóa):** Trong vài mili-giây đầu tiên, máy tính của bạn (Client) tự ngẫu nhiên sinh ra một cái "Khóa đối xứng" (Session Key). Sau đó, Client lấy Public Key của Server để mã hóa chính cái Session Key này rồi gửi qua cáp quang biển. 
- **Bước 2 (Giải mã an toàn):** Nhờ cơ chế hàm một chiều RSA, Hacker đứng giữa mạng bắt được gói tin này cũng phải khóc ròng. Chỉ duy nhất Server cầm Private Key mới giải mã ra và thu thập được cái Session Key nguyên vẹn. Vậy là bài toán "Phân phối Khóa đối xứng an toàn qua Internet" đã được giải quyết triệt để!
- **Bước 3 (Chuyển sang Đối xứng tốc độ cao):** Lúc này, cả Client và Server đã cùng nắm giữ một Session Key bí mật chung. Chúng lập tức vứt bỏ bộ khóa Bất đối xứng nặng nề. Kể từ giây phút này, toàn bộ hình ảnh, video, dữ liệu Web sẽ được mã hóa bằng Session Key (AES) với tốc độ ánh sáng!

> 💡 **Giải phẫu Toán học: Sự thật đằng sau Public Key và Dữ liệu là gì?**
> Sự thật: Trong máy tính không hề có khái niệm "Ổ khóa" hay "Chìa khóa" vật lý, tất cả chỉ là những **con số nguyên khổng lồ**.
> 
> **1. Dữ liệu (Payload):** 
> Giả sử bạn muốn gửi con số **M = 4** (Đại diện cho một mẩu dữ liệu nhỏ).
> 
> **2. Bức tranh Toàn cảnh về Định lý Euler (The Big Picture):**
> Trước khi đi vào tính toán, hãy cùng hiểu gốc rễ: **Tại sao thuật toán RSA lại chọn định lý Euler làm nền tảng?**
> 
> - **Mục đích (Dùng để làm gì?):** 
>   Mục tiêu tối thượng của mã hóa Bất đối xứng là tạo ra một "Hàm một chiều" (Trapdoor function). Tức là một hàm toán học mà: Đi xuôi (mã hóa) thì cực kỳ dễ, nhưng Đi ngược (giải mã) thì vô phương cứu chữa... trừ khi bạn có một "cửa sập bí mật" (Private Key). Định lý Euler chính là công cụ toán học hoàn hảo để xây dựng cái "cửa sập" này.
> 
> - **Nguyên lý cốt lõi (Nó hoạt động dựa trên nguyên lý nào?):**
>   Mọi thứ bắt nguồn từ sự bất đối xứng trong bài toán thừa số nguyên tố:
>   - Rất DỄ để lấy 2 số nguyên tố khổng lồ $p$ và $q$ nhân lại với nhau để ra một tích số $n$ (Máy tính mất $0.001$ giây).
>   - Nhưng cực kỳ KHÓ để làm ngược lại: Đưa cho siêu máy tính tích số $n$, yêu cầu nó phân tích ngược ra $p$ và $q$. 
>   *Góc nhìn Toán học (Tại sao lại khó?)*: Với một con số hợp số bình thường (như 24), nó có vô số cách phân tích ($2 \times 12, 3 \times 8, 4 \times 6$). Máy tính chỉ cần thử vài phép chia là có thể "khoanh vùng" tìm ra quy luật rất nhanh. Tuy nhiên, tích $n$ của 2 số nguyên tố là một "vùng đất chết" (Semi-prime). Nó CHỈ chia hết cho $p$ và $q$. Mà các số nguyên tố thì xuất hiện hoàn toàn ngẫu nhiên trên trục số, không tuân theo bất kỳ phương trình dự đoán nào. Do đó, máy tính không có "đường tắt" (shortcut), buộc phải mò mẫm (brute-force) chia thử cho từng số lẻ một từ 3 cho tới tận căn bậc hai của $n$. Khi $n$ là một số dài 600 chữ số, tập hợp các phép thử này lớn hơn tổng số nguyên tử trong vũ trụ. Quá trình chia mò mẫm vô vọng này ngốn của siêu máy tính... hàng triệu năm.
>   Định lý Phi Euler, ký hiệu là $\phi(n)$, chính là cầu nối. Xin lưu ý: **$\phi(n)$ hoàn toàn khác với $n$**. 
>   Số $n$ đơn thuần chỉ là tích ($n = p \times q$). Còn $\phi(n)$ là "tổng số lượng các con số nguyên tố cùng nhau với $n$". 
>   *(Ghi chú: Hai số được gọi là "nguyên tố cùng nhau" - Coprime - nếu chúng KHÔNG có bất kỳ ước số chung nào ngoài số 1. Ví dụ đếm thử $\phi(10)$: Đếm các số từ 1 đến 9 xem số nào "trong sạch" với 10.*
>   *- Số 1, 3, 7, 9: Nhận (Vì không có ước chung với 10).*
>   *- Số 2, 4, 6, 8: Loại (Vì cùng chia hết cho 2).*
>   *- Số 5: Loại (Vì cùng chia hết cho 5).*
>   *-> Tổng cộng đếm được 4 số. Vậy $\phi(10) = 4$).*
>   
>   Điểm kỳ diệu của toán học nằm ở đây:
>   - **Mục đích của $\phi(n)$ không phải là lấy các con số coprime đó ra phân phát cho User**. $\phi(n)$ chỉ là một "hằng số đếm" trung gian. Khi có được số đếm này, Toán học Euler chứng minh rằng $M^{\phi(n)} \equiv 1 \pmod{n}$. Từ cái mỏ neo bằng 1 này, Server mới chế tạo ra được cặp khóa $(e, d)$ sao cho $e \times d \equiv 1 \pmod{\phi(n)}$. Tạo xong cặp khóa thì giá trị $\phi(n)$ hết nhiệm vụ. Server chỉ cần tạo 1 cặp khóa duy nhất này để dùng chung cho hàng triệu User.
>   - Theo định nghĩa, nếu $p$ là số nguyên tố, mọi số nhỏ hơn $p$ đều không chia hết cho nó. Suy ra $\phi(p) = p - 1$.
>   - Vì hàm Phi Euler có tính chất nhân (multiplicative) đối với 2 số nguyên tố cùng nhau, nên $\phi(n) = \phi(p \times q) = \phi(p) \times \phi(q) = (p - 1) \times (q - 1)$.
>   - Nhờ vậy, nếu bạn nắm giữ $p$ và $q$ (như Server), bạn có thể tính ra lượng số $\phi(n)$ cực nhanh bằng công thức tắt: $(p-1) \times (q-1)$.
>   - Còn nếu bạn CHỈ bắt được số $n$ trôi nổi trên mạng (như Hacker), bạn không có bất kỳ công thức nào để tính ra $\phi(n)$, trừ khi bạn giải được bài toán đập vỡ $n$ ra thành $p$ và $q$ (mất hàng triệu năm).
> 
> - **Áp dụng vào bài toán RSA (Tại sao nó giải quyết được bài toán?):**
>   Toán học gia Euler đã chứng minh một định lý vĩ đại: Mọi con số $M$ (nếu nguyên tố cùng nhau với $n$) khi đem lũy thừa cho $\phi(n)$ rồi chia lấy dư cho $n$, thì phần dư LUÔN LUÔN BẰNG 1. Công thức: $M^{\phi(n)} \equiv 1 \pmod{n}$.
>   *Ví dụ chứng minh:* Ở trên ta tính được $\phi(10) = 4$. Hãy chọn một số $M = 3$ (nguyên tố cùng nhau với 10). Ta thử lấy $3^{\phi(10)} = 3^4 = 81$. Đem $81$ chia lấy dư cho $10$, ta được phần dư chính xác bằng $1$. (Dù bạn chọn $M=7$, thì $7^4 = 2401$, đem chia 10 cũng vẫn dư 1!).
> 
>   Từ "cái mỏ neo bằng 1" tuyệt đẹp này, thuật toán RSA chế tạo ra cặp khóa sinh đôi $e$ (Public) và $d$ (Private) sao cho tích $e \times d$ lớn hơn một bội số của $\phi(n)$ đúng 1 đơn vị. Nghĩa là: $e \times d = k \times \phi(n) + 1$.
>   Hệ quả sinh ra một phép màu Toán học:
>   $M^{(e \times d)} = M^{k \times \phi(n) + 1} = (M^{\phi(n)})^k \times M$
>   Mà theo định lý Euler, cụm $M^{\phi(n)}$ luôn bằng $1$. Nên phương trình lập tức rút gọn thành: $1^k \times M = M$.
>   **Kết luận tối thượng:** **$M^{(e \times d)} \equiv M \pmod{n}$**.
>   Nghĩa là: Nếu bạn lấy dữ liệu gốc $M$, đem mũ $e$ (Mã hóa bằng Public Key), rồi lại đem mũ tiếp cho $d$ (Giải mã bằng Private Key), số mũ sẽ tự động triệt tiêu và trả về đúng dữ liệu $M$ ban đầu một cách hoàn hảo!
> 
> - **Sự phân vai Đích đáng (Ai cầm cái gì?):**
>   Server (Người tạo khóa) biết $p$ và $q$ -> Dễ dàng tính được $\phi(n)$ -> Tính ra được chìa khóa bí mật $d$.
>   Hacker (Kẻ cắp) chỉ nhìn thấy số $n$ trên mạng -> Không thể tính ra $\phi(n)$ -> Vĩnh viễn không thể tìm ra $d$. 
>   Đây chính là bức tường thành Toán học bảo vệ toàn bộ mạng lưới Internet.
> 
> **3. Khóa là gì? (Ứng dụng thực tế Định lý Euler với 3 con số `e`, `d`, `n`):** 
> Server chọn 2 số nguyên tố ngẫu nhiên (Lấy ví dụ với số cực nhỏ: $p=11$, $q=3$).
> - Tính số **n** (Modulus - Vũ trụ số học): $n = p \times q = 33$. (Mọi phép tính mã hóa sẽ bị giam lỏng trong vòng tròn 33 này).
> - Tính **Hàm Phi Euler** $\phi(n)$ (Số lượng các số nguyên tố cùng nhau với n): Công thức $\phi(n) = (p-1) \times (q-1) = 10 \times 2 = 20$.
> - Tìm số **e** (Encrypt Exponent - Khóa mã hóa): Chọn một số `e` sao cho `e` nhỏ hơn $\phi(n)$ và không có ước chung nào với $\phi(n)$. Ta chọn được **e = 3** (vì 3 và 20 là nguyên tố cùng nhau).
> - Tìm số **d** (Decrypt Exponent - Khóa giải mã): Tìm số `d` sao cho $(d \times e)$ chia lấy dư cho $\phi(n)$ bằng 1. Tức là $(d \times 3) \pmod{20} = 1$. Nhẩm tính nhanh: $7 \times 3 = 21$; $21$ chia $20$ dư $1$. Vậy ta tính ra được **d = 7**.
> -> Server ném Public Key **(e=3, n=33)** ra Internet công cộng. Server giấu chặt Private Key **(d=7, n=33)** trong két sắt.
> 
> **4. Client Mã hóa (Công thức: `C = M^e mod n`):** 
> Trình duyệt nhận được `(3, 33)`. Nó mã hóa số `M = 4` bằng phương trình duy nhất: Lũy thừa và chia lấy phần dư.
> Lấy $4^3 = 64$. Đem $64$ chia lấy phần dư cho $33$ ta được phần dư là $31$.
> -> Kết quả sinh ra con số rác rưởi là **C = 31** (Ciphertext). Khối byte `31` này được gửi qua cáp đại dương.
> 
> **5. Hàm một chiều chặn đứng Hacker:** 
> Phép chia lấy dư (Modulo) là **Hàm một chiều**. Hacker bắt lén trên mạng được số dư là `31`, vĩnh viễn không thể suy ngược ra số gốc là `4` (vì có hàng ngàn số đem chia 33 cũng dư 31). Do đó, Public Key `e` chỉ có thể tiến tới chứ không thể lùi lại.
> 
> **6. Server giải mã (Công thức: `M = C^d mod n`):** 
> Server nhận số `C = 31`. Nó lôi chìa khóa vạn năng `d = 7` ra để giải mã:
> Lấy $31^7 = 27,512,614,111$.
> Đem $27,512,614,111$ chia lấy phần dư cho $33$. Kết quả trả về phần dư chính xác bằng... **4**. 
> BÙM! Dữ liệu gốc `M = 4` đã được phục hồi hoàn hảo mà không cần gửi chìa khóa $d$ đi qua mạng.
> 
> *Góc nhìn Hacker (Hệ phương trình Vô vọng):* 
> Để tìm được khóa bí mật `d`, Hacker bắt buộc phải tính được giá trị $\phi(n)$. 
> Hãy thử tư duy ngược: Nếu Hacker BẰNG MỘT PHÉP MÀU NÀO ĐÓ biết trước được con số đếm $\phi(n) = 20$ (cùng với $n = 33$ công khai trên mạng). Hắn sẽ lập tức giải ra $p$ và $q$ thông qua hệ phương trình Viète lớp 9:
> 1. Tích hai số: $p \times q = 33$
> 2. Dựa vào $\phi(n)$: $(p-1) \times (q-1) = 20 \implies (p \times q) - p - q + 1 = 20 \implies 33 - (p+q) + 1 = 20 \implies p + q = 14$
> -> Hacker có Tổng = 14, Tích = 33. Lập tức nhẩm ra ngay 2 nghiệm là $11$ và $3$. Hacker lấy được $p=11, q=3$ và bẻ khóa toàn bộ hệ thống!
> 
> **NHƯNG... LÀM SAO ĐỂ BIẾT ĐƯỢC $\phi(n)$?** 
> Hacker trên mạng chỉ bắt được duy nhất con số $n$. Hắn KHÔNG THỂ lập hệ phương trình vì thiếu mảnh ghép $\phi(n)$. Để đếm được $\phi(n)$ của một số $n$ dài 600 chữ số, siêu máy tính của Hacker buộc phải đếm thủ công nghiệm Coprime nghiệm từ 1 lên tới $n$. Quá trình đếm mò mẫm này ngốn thời gian... lớn hơn hàng tỷ lần tuổi thọ của Vũ trụ. Đó chính là sự bất lực tuyệt đối của Hacker trước bức tường Toán học!

### 5.3. Kịch bản Truyền tải Thực tế (TCP & TLS Handshake)
Để đạt được sự tối ưu giữa "Tốc độ của mã hóa đối xứng" và "Tính bảo mật của mã hóa bất đối xứng", máy tính (Client) và máy chủ (Server) bắt buộc phải thực hiện một chuỗi giao thức bắt tay nghiêm ngặt theo đúng trình tự thời gian. Hãy nhớ rằng: **Bắt tay TLS (Layer 7) không thể diễn ra nếu chưa có ống nước kết nối (TCP Socket - Layer 4)**. 

Dưới đây là Dòng chảy Kỹ thuật thuần túy (Packet Flow) cho một truy cập HTTPS (Port 443):

**GIAI ĐOẠN 1: MỞ KẾT NỐI (TCP 3-Way Handshake - Lớp 4)**
Trước khi bàn chuyện bảo mật, Client phải thiết lập một Socket tới Port 443 của Server.
1. `[Client -> Server]:` **SYN** *(Đề nghị mở kết nối TCP).*
2. `[Server -> Client]:` **SYN-ACK** *(Đồng ý mở kết nối, anh xác nhận đi).*
3. `[Client -> Server]:` **ACK** *(Xác nhận. Ống TCP đã nối thành công).*
-> *Kết quả: Kênh truyền tải đã sẵn sàng, nhưng CHƯA có byte dữ liệu Web nào được gửi.*

**GIAI ĐOẠN 2: THỎA THUẬN BẢO MẬT (TLS Handshake - Lớp 7)**
Sử dụng ống TCP vừa mở, quá trình trao đổi khóa mã hóa bắt đầu diễn ra:

4. **ClientHello** `[Client -> Server]:` Client khởi tạo luồng kết nối bảo mật, gửi bản tin bao gồm: Phiên bản giao thức (ví dụ: TLS 1.2/1.3), danh sách các bộ thuật toán mã hóa (Cipher Suites) được hỗ trợ, và một chuỗi ngẫu nhiên `Client Random`.
5. **ServerHello** `[Server -> Client]:` Server phản hồi, chốt chọn một bộ Cipher Suite chung mạnh nhất, kèm theo chuỗi ngẫu nhiên `Server Random` của riêng mình.
6. **Certificate Exchange (Chứng thư số)** `[Server -> Client]:` Server gửi tiếp Chứng chỉ số X.509 (SSL Certificate). Điểm cốt lõi là bên trong chứng chỉ này có đính kèm **Public Key** của Server.
7. **CA Validation (Xác thực nội bộ tại Client):** Trình duyệt Client dùng danh sách Root CA (như Let's Encrypt, DigiCert) cài sẵn trong HĐH để xác minh chữ ký số trên Chứng chỉ. Nếu chữ ký hợp lệ, Client tin tưởng Server này là thật (chống tấn công giả mạo Man-in-the-Middle).
8. **Key Exchange (Trao đổi Khóa bí mật):** 
   - Client tự tạo ra một chuỗi bí mật tên là `Pre-Master Secret`.
   - Client dùng **Public Key** của Server để bọc (mã hóa) chuỗi `Pre-Master Secret` này và truyền qua mạng. Do cơ chế hàm một chiều, gói tin này bất khả xâm phạm đối với Hacker chặn giữa.
9. **Session Key Generation (Sinh Khóa phiên):** 
   - Server nhận gói tin, lôi **Private Key** giấu tuyệt mật trong ổ cứng ra để giải mã, thu thập thành công `Pre-Master Secret`. 
   - Lúc này, cả hai bên độc lập sử dụng tập hợp (`Client Random` + `Server Random` + `Pre-Master Secret`) đưa vào hàm băm toán học để tính ra chung một Khóa phiên duy nhất (**Master Secret / Session Key** - Khóa đối xứng).
   - *(Giải ngố: Tại sao lại cần ghép thêm `Client Random` và `Server Random`? Mục đích là để chống lại **Replay Attack** (Tấn công phát lại). Giả sử không có chuỗi Random, Hacker dù không giải mã được gói tin nhưng hắn copy y nguyên gói tin đó và ngày mai gửi lại cho Server, Server vẫn tưởng đó là bạn. Việc bơm thêm chuỗi Random ở hai phía đảm bảo rằng mỗi một phiên làm việc (Session) sẽ luôn sinh ra một chìa khóa độc nhất vô nhị dựa trên thời gian thực. Sang phiên khác, chuỗi Random đổi, khóa Session Key đổi, gói tin cũ thu âm từ hôm qua mang ra gửi lại sẽ tự động trở thành rác rưởi không thể giải mã).*
10. **Secure Payload (Thiết lập kênh truyền):** Cả hai bên gửi bản tin `ChangeCipherSpec` và `Finished`, cam kết từ giây phút này mọi dữ liệu sẽ được mã hóa bằng Khóa phiên đối xứng tốc độ cao (như thuật toán AES-GCM). Bộ khóa Bất đối xứng (RSA) chính thức bị vứt bỏ vì đã hoàn thành nhiệm vụ trao đổi.

**GIAI ĐOẠN 3: BƠM DỮ LIỆU THỰC TẾ (HTTPS Traffic)**
11. `[Client -> Server]:` **HTTP GET /** *(Toàn bộ Payload HTTP lúc này đã bị băm nát và mã hóa bằng Khóa phiên Đối xứng, đóng vào TCP Segment ném qua mạng).*
12. `[Server -> Client]:` **HTTP 200 OK** *(Server giải mã bằng Khóa phiên, xử lý logic Web, và trả về HTML/JSON cũng được mã hóa bằng Khóa phiên).*
-> *Kết quả: Khóa phiên đối xứng chỉ tồn tại tạm thời trong RAM và sẽ bị HĐH xóa sổ vĩnh viễn khi tắt tab trình duyệt (Cơ chế Perfect Forward Secrecy).*

### 5.5. Góc đính chính: Hacker chôm Public Key và Lỗi CORS
Có hai thắc mắc kinh điển mà các Lập trình viên hay nhầm lẫn khi cấu hình Web Server (như Nginx kết nối FE với BE):

**Câu hỏi 1: Tại sao Server ném Public Key công khai qua mạng mà không sợ Hacker bắt được và giả mạo (Man-in-the-Middle Attack)?**
- *Trả lời:* Hacker CÓ THỂ bắt được Public Key trên đường bay, nhưng **Vô dụng**. Hacker không thể dùng Public Key đó để bẻ khóa (vì luật Toán học quy định chỉ có Private Key giấu ở Server mới mở được).
- *Vậy nếu Hacker tráo Public Key của Hacker vào thì sao?* Hacker chặn Chứng chỉ thật, tự tạo một Chứng chỉ giả chứa Public Key của Hacker rồi gửi cho Client. Client tưởng thật, dùng khóa của Hacker bọc Session Key, Hacker sẽ mở được! ĐỂ CHỐNG LẠI TRÒ NÀY, ta có **Chữ ký số (Digital Signature) của Tổ chức CA**. Chứng chỉ gốc đã được ký niêm phong (hashed) bằng khóa bí mật của Tổ chức Root CA (như Let's Encrypt). Bất kỳ ai sửa đổi dù chỉ 1 ký tự trong Chứng chỉ (tráo Public Key), chữ ký đó sẽ bị vỡ. Trình duyệt Chrome check thấy chữ ký vỡ lập tức văng màn hình đỏ chót *"Your connection is not private"* và chặt đứt kết nối.

**Câu hỏi 2: Có phải chứng chỉ SSL/CA bị sai là nguyên nhân gây ra lỗi CORS khi FE gọi API sang BE không?**
- *Trả lời:* Về mặt lý thuyết mạng lưới, **SSL/TLS (Mã hóa đường truyền) KHÔNG HỀ LIÊN QUAN GÌ ĐẾN CORS (Chính sách Trình duyệt)**.
  - **SSL/TLS:** Đảm bảo dữ liệu không bị nghe lén.
  - **CORS (Cross-Origin Resource Sharing):** Là chính sách bảo vệ do Trình duyệt (Chrome/Safari) áp đặt. Nó ngăn cấm web `hacker.com` lén dùng Javascript gọi API sang `api.bank.com` để ăn cắp tiền. Lỗi này là do Backend của Bank quên chưa cấu hình trả về HTTP Header `Access-Control-Allow-Origin`. Nó xảy ra kể cả khi cả 2 bên xài HTTP không mã hóa.
- *Tại sao Dev cực kỳ hay nhầm lẫn 2 cái này làm một?* Khi Dev làm việc ở Local (FE chạy `http://localhost:3000`, BE chạy `https://api.dev.com`). Nếu BE xài chứng chỉ SSL **tự ký (Self-signed - không được Root CA bảo chứng)**, Trình duyệt Chrome sẽ ngầm đánh giá kết nối SSL này là nguy hiểm và chặn toàn bộ các lệnh `fetch/axios` gửi đi. Thay vì báo lỗi "SSL Invalid", cái cửa sổ Console của Chrome đôi lúc lại văng ra dòng chữ ngu ngốc: `CORS error` hoặc `Network Error`. Thế là Dev cuống cuồng đi sửa file config CORS trong Backend suốt 3 ngày mà không fix được, trong khi thực tế chỉ cần click chuột vào nút "Bỏ qua cảnh báo an toàn SSL" hoặc cài chứng chỉ thật là xong!

---

## 6. Tường lửa (Firewall):

Một khi mạng đã thông, dữ liệu đã được mã hóa TLS, bài toán sống còn tiếp theo là: *"Làm sao để cấm một gã ất ơ nào đó ở Nga không được quyền chọc vào Database của tôi?"*. Đó là lúc Firewall ra đời. Firewall giống như ông bảo vệ tòa nhà, đứng xét hỏi từng gói tin (Packet) đi qua.

### 6.1. Tường lửa Lớp 3/4 (Network Firewall)
Đây là loại tường lửa cổ điển và phổ biến nhất, hoạt động ở tầng **Network (IP)** và **Transport (Port)**. Trên Cloud AWS, nó chính là **Security Group** hoặc **Network ACL**.

- **Cách hoạt động:** Ông bảo vệ Lớp 4 khá "nguyên tắc nhưng mù quáng". Ông ta chỉ nhìn vào cái phong bì (Header), kiểm tra `IP Nguồn`, `IP Đích` và `Port`. Ông không thèm bóc phong bì ra xem bên trong (Payload) chứa mã độc hay không.
- **Ví dụ thực tiễn:** Cấu hình Security Group cho EC2 chạy Web Server:
  - Cho phép `Inbound Port 443` từ `0.0.0.0/0` (Mở cửa cho toàn thế giới truy cập Web HTTPS).
  - Cấm `Inbound Port 22` từ mọi nơi, Cụ thể chỉ cho phép IP tĩnh của Văn phòng Cty `14.55.x.x` được phép truy cập để SSH vào bảo trì hệ thống.
- **Stateful vs Stateless:** 
  - **Stateful (Lưu trữ trạng thái):** Điển hình là AWS Security Group. Nếu bạn cho phép một IP đi **VÀO** cổng 443, Firewall sẽ lưu vào trí nhớ tạm thời và tự động mở cửa cho phép luồng dữ liệu trả lời đi **RA** mà không cần bạn phải cấu hình thêm luật chiều ra.
  - **Stateless (Không lưu trữ trạng thái):** Điển hình là AWS Network ACL. Trí nhớ bằng 0. Cấm hay cho phép chiều VÀO thì DevOps cũng phải tự tay viết thêm một luật tương ứng cho phép chiều RA, nếu không gói tin đi vào được nhưng lúc gửi trả Response lại bị kẹt mất hút.

### 6.2. Tường lửa Lớp 7 (WAF - Web Application Firewall)
Bảo vệ vòng ngoài bằng Lớp 4 là chưa đủ. Giả sử Hacker gọi đúng vào Port 443 được phép, nhưng bên trong gói hàng hắn nhét một dòng code SQL Injection (`' OR 1=1 --`) hòng xóa sạch Database thì sao? Ông bảo vệ Lớp 4 mù quáng sẽ cho đi qua tuốt vì vỏ thư ghi đúng Port 443! Lúc này ta cần WAF (Như Cloudflare WAF, AWS WAF).

- **Cách hoạt động:** Ông bảo vệ Lớp 7 (WAF) là một chuyên gia soi mói. Ông ta bắt khách hàng mở khóa phong bì ra. Ông ta soi xét từng chữ trong cái Payload HTTP (Body JSON, URL query, Headers, Cookies). 
- **Quyền năng của WAF:**
  - Nếu thấy URL có chứa mã độc SQL/XSS: *Chặn đứng tắp lự (DROP).*
  - Nếu thấy 1 IP từ một quốc gia lạ gọi Port 443 đến 500 lần/giây (Tấn công DDOS lụt mạng): *Ban IP.*
  - Nếu Header `User-Agent` rỗng hoặc giống một con Bot cào dữ liệu (Scraping): *Bật ngay màn hình CAPTCHA bắt giải đố.*

### 6.3. Kiến trúc Defense-in-Depth (Phòng thủ nhiều lớp)
Kỹ sư hạ tầng (Platform Engineer) không bao giờ giao sinh mạng của hệ thống cho 1 loại Firewall. Họ quy hoạch mạng lưới theo nguyên tắc Phòng thủ nhiều lớp (Defense-in-Depth):
1. **Ngoài cùng (Cloudflare / Edge WAF):** Hứng bão, lọc mọi cuộc tấn công DDOS Lớp 7, soi kỹ Payload để chặn SQL Injection/XSS trước cả khi gói tin chạm tới Cloud của bạn.
2. **Lớp giữa (Public Subnet / DMZ):** Nơi cắm Load Balancer (ELB) hoặc Nginx. Dùng Security Group (Lớp 4) bóp chặt chỉ mở đúng Port 80 và 443. Mọi truy cập vào port khác lập tức bị từ chối.
3. **Lõi (Private Subnet):** Trái tim của hệ thống chứa Database và Backend. Lớp Firewall ở đây thiết lập gắt gao nhất: Cắt hoàn toàn kết nối từ Internet, chỉ cho phép Inbound Port 3306/5432 **ĐỘC NHẤT** từ địa chỉ IP Private của chính con Backend/Nginx nội bộ. Bất kỳ Server nào khác trong mạng dù biết IP cũng không thể đụng vào DB.

---

## 7. CASE STUDY TOÀN DIỆN: Hành trình End-to-End của 1 Request

Hãy theo dõi một kịch bản siêu chi tiết: User (IP LAN `192.168.1.5`) gõ URL mở App HTTPS được host trên Cloud AWS (IP Private nội bộ AWS `10.0.0.5`, IP Public `13.212.0.5`).

### CHIỀU ĐI (REQUEST JOURNEY)
**1. Khởi tạo (Layer 7):** HĐH chạy bắt tay SSL/TLS (tạo Session Key). Trình duyệt gen ra khối JSON HTTP Request, rồi dùng Session Key mã hóa nó thành đống ký tự vô nghĩa. Khối này gọi là **Payload**.
**2. Đóng gói TCP (Layer 4):** HĐH chia nhỏ Payload, bọc **TCP Header** (Ghi rõ: `Source Port: 51234`, `Dest Port: 443`). Tạo thành **TCP Segment**.
**3. Đóng gói IP (Layer 3):** Bọc **IP Header** (`Source IP: 192.168.1.5`, `Dest IP: 13.212.0.5`). Tạo thành **IP Packet**.
**4. Đóng gói Frame (Layer 2):** Máy tính tính Subnet thấy IP Đích khác LAN. Tra bảng ARP lấy MAC của Router Wifi. Bọc **MAC Header**. Tạo thành **Ethernet Frame**.
**5. Truyền tải (Layer 1):** Card mạng dịch Frame thành xung điện, đẩy qua cáp đến Switch gia đình.
**6. Switch Nội bộ xử lý:** Đọc MAC Header, Flooding/Unicast đẩy xung điện trúng cổng Router Wifi.
**7. NAT Outbound tại Router:** Router xé Lớp 2. Đọc IP Packet. Thực hiện **NAT**, sửa `Source IP` từ `192.168.1.5` thành IP Public (`14.22.33.44`). Tạo Frame mới ném ra cáp quang Internet.
**8. Định tuyến Internet (Hop-by-Hop):** Qua hàng tá Router của Viettel, APG, Singtel... Tại mỗi trạm, Router chỉ xé Lớp 2 (Frame), tra Routing Table dựa vào Dest IP (Layer 3), rồi đóng một Frame mới ném đi. **Lưu ý: Suốt chặng giữa này không có NAT, IP Packet bên trong được giữ nguyên vẹn.**
**9. NAT Inbound tại Đích (AWS IGW):** Packet tới Internet Gateway (IGW) của AWS. IGW có luật Load Balancer, bóc IP Public đích (`13.212.0.5`), đổi thành IP Private máy ảo EC2 (`10.0.0.5`).
**10. Switch VPC AWS:** VPC đóng gói Frame bọc MAC Address của EC2 và truyền tới EC2.
**11. Giải nén (Decapsulation) tại Server:** 
- Card mạng EC2 nhận xung điện -> Frame. Đúng MAC mình -> Xé bọc Lớp 2.
- Linux đọc Packet. Đúng IP mình `10.0.0.5` -> Xé bọc Lớp 3.
- Linux đọc TCP Segment. Thấy `Dest Port 443` -> Ném Payload cho Nginx.
- Nginx dùng Session Key (đã thỏa thuận từ bước bắt tay TLS) giải mã cái Payload vô nghĩa kia ra thành chuỗi HTTP JSON gốc.
- Nginx xử lý điều hướng `proxy_pass` ném sang Process Backend (Golang ở port 8080).

### CHIỀU VỀ (RESPONSE JOURNEY)
Mọi thứ diễn ra ngược lại hoàn toàn. Backend trả HTTP Response -> Bọc TCP -> Bọc IP (Source: 10.0.0.5, Dest: 14.22.33.44) -> AWS IGW làm NAT đổi Source IP thành Public -> Bay qua đại dương về Router Wifi nhà bạn -> Router dò NAT Table bóc Dest IP Public trả lại thành 192.168.1.5 -> Switch ném về máy tính -> Trình duyệt Giải mã SSL -> Render UI!

---

## 7. Tường lửa (Security Group Layer 4)

- **Stateful Firewall:** Tường lửa hoạt động tại ranh giới mạng ảo (Layer 4). Nó can thiệp bằng cách đọc TCP/UDP Port. Chữ "Stateful" nghĩa là nó Ghi nhớ trạng thái phiên TCP. Nếu gói tin Ingress (Request) vượt qua tường lửa thành công, gói Egress (Response) tương ứng sẽ tự động được xả trạm mà không cần tạo Rule chiều ra.
- **Zero Trust:** Đóng mọi Ingress từ Internet (`0.0.0.0/0`), chỉ mở tối thiểu (80/443). Tuyệt đối không mở các Port Database/Cache ra ngoài Public. Khóa ICMP (Ping) để chống dò quét.

---

## 8. Phân giải Tên miền (DNS) & Đảo ngược Proxy tại Biên

- **Name Server (NS) là Quyền Lực Tối Cao:** Khi đổi NS về Cloudflare, bạn giao quyền sinh sát phân luồng traffic cho Edge Network của họ.
- **Proxying (TLS Termination):** Thay vì trỏ DNS A-Record vào thẳng AWS (dễ bị DDoS). Request đập vào Đám mây Cam Cloudflare. Tại đây Cloudflare đứng ra **Bắt tay SSL (TLS Termination)** với User, giải mã Lớp 7 ra quét WAF (chống SQL Injection), sau đó Cloudflare tự dùng Session Key nội bộ mã hóa lại và mở phiên TCP/IP gửi ngầm về Origin IP của AWS. IP thật của Server hoàn toàn tàng hình.

---

## 9. Sự Tiến Hóa Cơ Sở Hạ Tầng (Infrastructure Evolution)

Tất cả kiến trúc mạng đồ sộ trên đã tiến hóa qua 3 thời kỳ:
- **Bare Metal (Máy chủ vật lý):** Mua thiết bị, tự cắm cáp đồng, setup Switch vật lý tại Data Center. Tiền đầu tư khổng lồ, Scale kém.
- **VPS (Virtual Private Server):** Ảo hóa chia nhỏ máy tính. Giảm chi phí cứng nhưng xài hay không xài vẫn trả đủ tiền.
- **Cloud Computing (AWS/GCP):** Mọi thứ (Kể cả Card mạng, Switch, Router) đều được ảo hóa (Software-Defined Networking - SDN). Hạ tầng quản lý bằng Code (Terraform), tính tiền theo giây (Pay-as-you-go). Kỹ sư thoát khỏi phần cứng để tập trung 100% vào Kiến trúc & Bảo mật.

---

## 10. Xương sống của Internet (The Internet Backbone & Bức màn Bí ẩn)

Rất nhiều kỹ sư nhầm tưởng rằng Internet thuộc về Google, AWS hay Quân đội Mỹ. Thực chất, Internet hiện đại không có Tổng giám đốc, không có máy chủ trung tâm. Nó được thống trị bởi một liên minh các "Ông trùm" thầm lặng nằm dưới đáy đại dương.

### 10.1. Kim tự tháp Quyền lực Internet
- **Đỉnh Kim tự tháp (Tier-1 Networks):** Đây là các "Ông trùm" đích thực (Ví dụ: AT&T, Lumen, Arelion, NTT). Họ sở hữu hệ thống cáp quang biển liên lục địa vĩ đại nhất. **Đặc điểm sống còn:** Họ kết nối ngang hàng (Peering) với nhau và KHÔNG BAO GIỜ phải trả tiền cước cho bất kỳ ai trên thế giới. Nếu 3 trong số họ rút điện, một nửa thế giới sẽ quay về thời đồ đá.
- **Tầng giữa (Tier-2 ISPs):** Điển hình là các nhà mạng quốc gia (VNPT, Viettel, FPT). Họ phủ sóng cáp đến từng hộ gia đình, nhưng họ không có cáp vươn tới Mỹ. Do đó, VNPT bắt buộc phải đem hàng triệu USD đi **Mua đường truyền (Transit)** của các Tier-1 để người dùng quốc gia có thể kết nối ra quốc tế.

### 10.2. Giải mã hiểu lầm: "Mua IP" vs "Mua Đường truyền"
Có một sự nhầm lẫn kinh điển: *"Google mua IP của Tier-1 để làm dịch vụ"*. Đây là nhận định hoàn toàn sai lầm. Bạn cần phân biệt rõ 2 loại tài nguyên:
1. **Tài nguyên Logic (IP Address / ASN):** Được cấp phát bởi Tổ chức quản lý phi lợi nhuận toàn cầu (IANA). Cả Tier-1, Viettel hay Google đều phải làm đơn xin IANA cấp IP. Xin IP giống như được cấp cái "Số nhà".
2. **Tài nguyên Vật lý (Băng thông cáp quang):** Đây là thứ mà Tier-1 sở hữu. Có số nhà (IP) nhưng không có đường nhựa (Cáp quang) thì dữ liệu không thể đi đâu được. 

**Tại sao dạo này Google/Meta lại tự rải cáp quang biển?**
Google và Meta có IP riêng, nhưng ngày xưa họ vẫn phải làm "khách hàng" trả tiền cước (Transit) cho Tier-1 để mượn đường cáp vận chuyển data. Tuy nhiên, lưu lượng Youtube và Cloud hiện nay chiếm tới hơn 20% toàn bộ Internet thế giới. Trả tiền cho Tier-1 quá đắt đỏ và bị kẹt xe. Thế là Google tự vác tàu đi xây luôn hệ thống cáp riêng xuyên đại dương (Ví dụ cáp quang Curie, Dunant) để Server của họ nói chuyện trực tiếp với nhau với độ trễ cực thấp mà không cần dựa dẫm vào Tier-1 nữa.

### 10.3. Bí mật Vật lý dưới đáy đại dương (Submarine Cables)
Nhiều người nghĩ cáp biển truyền dữ liệu bằng xung điện (Volt). Thực tế, xung điện (Volt) truyền trên cáp đồng gặp phải kẻ thù chí mạng là **Điện trở**. Đi vài chục cây số là tín hiệu điện bị hao mòn (Attenuation) suy yếu về 0.

- **Dữ liệu truyền bằng Ánh sáng:** Để đi qua Thái Bình Dương xa hàng vạn kilomet, cáp biển sử dụng các sợi quang học làm từ **Thủy tinh siêu tinh khiết (Ultra-pure silica glass)** mỏng hơn sợi tóc người. Dữ liệu (0 và 1) được truyền đi bằng các nhịp chớp tắt của đèn Laser. Ánh sáng này phản xạ toàn phần bên trong sợi thủy tinh để lao đi với vận tốc cực khủng.
- **Nghịch lý của dòng điện 10.000 Volt:** Dù thủy tinh siêu trong suốt, ánh sáng đi được khoảng 50-100km dưới đáy biển vẫn bị mờ đi. Người ta bắt buộc phải đặt các Cục kích sáng (Optical Amplifiers / Repeaters) dọc đường để hút ánh sáng yếu vào và bắn ánh sáng mạnh ra. **Và để nuôi sống các cục Repeater này ở độ sâu 4.000 mét dưới đáy biển, lớp vỏ của sợi cáp quang chứa một ống đồng khổng lồ truyền dòng điện một chiều (DC) lên tới 10.000 Volt!** 

> **Kết luận:** Dưới đáy biển, **Dữ liệu chạy bằng Ánh sáng Laser, còn Dòng điện 10.000 Volt chạy song song dọc theo cáp chỉ để làm duy nhất một việc: Cấp nguồn nuôi hệ thống kích sáng (Repeater) dọc đường!**
