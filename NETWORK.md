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

### 3.1. Subnet Mask (Mặt nạ mạng) & Phép toán Bitwise AND
Dải IP luôn đi kèm Subnet Mask (vd `/24` hoặc `255.255.255.0`). Nhiệm vụ duy nhất của mặt nạ mạng là làm cái thớt để "chặt" địa chỉ IP ra làm 2 phần: **Network ID** (Khu phố) và **Host ID** (Số nhà).

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
- **Cơ chế "Ngoại giao" (Peering & Transit):** BGP chính là ngôn ngữ ngoại giao giữa các quốc gia AS này. Các con Router khổng lồ nằm ở biên giới của VNPT sẽ liên tục "cầm loa thông báo" với các Router ở biên giới của Singtel, AWS: *"Ê, dải IP `13.212.x.x` này là đất của tao (ASN Amazon), ai muốn gửi dữ liệu đến IP đó thì cứ quăng qua cáp quang biển APG nối với tao nhé"*.
- **Thuật toán tìm đường:** BGP không chọn đường đi dựa vào đường truyền nhanh nhất, mà chọn đường theo **AS Path** (Đi qua ít quốc gia AS nhất) và theo **Chính sách Kinh doanh** (Ví dụ: Viettel cấu hình Router từ chối đẩy gói tin qua nhà mạng X vì phí chuyển tiếp đắt, thà đi vòng qua nhà mạng Y rẻ hơn).
- *Thảm họa thực tế:* Nhờ BGP mà các Packet luôn tìm được đường tới đích. Nhưng nếu một kỹ sư mạng ở Facebook gõ sai cấu hình, vô tình xóa mất lời "loan báo" BGP của hệ thống Facebook, thì toàn bộ dải IP của Facebook sẽ lập tức "tàng hình" khỏi bản đồ Internet thế giới, khiến mạng lưới sập toàn cầu (Điều này đã từng xảy ra trên thực tế khiến toàn bộ Facebook, Messenger, Instagram sập sạch vào tháng 10/2021).

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

### 5.2. Sự kết hợp hoàn hảo giữa 2 loại Mã hóa
Nếu chỉ dùng 1 loại mã hóa thì mạng Internet sẽ hoặc là "Quá chậm", hoặc là "Quá kém bảo mật". Do đó SSL/TLS đã kết hợp cả 2 loại với nhau:

1. **Mã hóa Bất đối xứng (Asymmetric Cryptography):**
   - **Bản chất:** Gồm 1 cặp chìa: **Public Key** (Chìa công khai, ném cho ai xem cũng được) và **Private Key** (Chìa bí mật, giấu chết ở ổ cứng Server).
   - **Ví dụ thực tế:** Public Key giống như một cái "Ổ khóa hở" (Padlock) mở sẵn. Bạn có thể ném hàng ngàn cái ổ khóa này ra đường cho bất kỳ ai. Người ta bỏ tiền vào rương, rồi lấy ổ khóa đó bấm tạch một phát. Một khi đã bấm khóa lại, thì KHÔNG AI (kể cả thằng vừa bấm) có thể mở ra được nữa. Chỉ duy nhất bạn - người cầm cái chìa **Private Key** - mới đút vào mở được.
   - **Đặc điểm:** Bảo mật tuyệt đối. Dù Hacker có chôm được Public Key cũng vô dụng vì nó chỉ dùng để "khóa" chứ không thể "mở". Nhược điểm chí mạng là thuật toán cực kỳ cồng kềnh, **Cực kỳ chậm**. Nếu dùng nó để mã hóa xem Youtube thì CPU sẽ nổ tung.
   
2. **Mã hóa Đối xứng (Symmetric Cryptography):**
   - **Bản chất:** Chỉ có **1 cái chìa duy nhất** dùng chung cho cả Khóa và Mở (Giống như mật khẩu két sắt nhà bạn, bạn biết, vợ bạn biết thì 2 người cùng mở được).
   - **Đặc điểm:** Xử lý toán học **Cực kỳ nhanh**, phù hợp truyền khối lượng data khổng lồ (Streaming, Tải File). Nhưng nhược điểm là: Làm sao để ném cái chìa khóa này cho bên kia qua môi trường Internet đầy rẫy Hacker mà không bị chôm mất chìa?

### 5.3. Kịch bản Bắt tay TLS (TLS Handshake)
Để lấy "Tốc độ bàn thờ" của Mã hóa đối xứng + "Bảo mật tuyệt đối" của Mã hóa bất đối xứng, máy tính và server dùng kịch bản sau:

1. **Client (Trình duyệt) gọi Server:** *"Ê, tôi muốn kết nối bảo mật!"*.
2. **Server trả lời:** Ném cho Client cái **Chứng chỉ SSL (Certificate)**. Trong cái chứng chỉ này chứa cái Ổ khóa mở sẵn (**Public Key**).
3. **Client kiểm tra Chứng chỉ (Xác thực CA):** Trình duyệt Chrome/Firefox của bạn khi cài đặt đã được nạp sẵn danh sách các Tổ chức cấp phát uy tín toàn cầu (Root CA như Let's Encrypt, DigiCert, GlobalSign). Chrome lấy chữ ký trong Chứng chỉ ra đối chiếu, nếu khớp chữ ký của Root CA -> Server này là thật, không phải lừa đảo (Phishing).
4. **Client tạo chìa khóa tốc độ cao:** Sau khi xác nhận an toàn, Client tự gen ra ngẫu nhiên một mật khẩu cực kỳ nhanh (gọi là **Session Key** - Khóa đối xứng).
5. **Bọc khóa trong khóa:** Client bỏ cái Session Key đó vào một cái rương, dùng cái Ổ khóa mở sẵn (**Public Key**) của Server bấm "TẠCH" một phát. Gửi cái rương đã khóa chặt qua cáp quang biển. (Lúc này, Hacker đứng giữa bắt được cái rương, nhưng đành khóc ròng vì Hacker không có Private Key để mở).
6. **Server nhận hàng:** Server lôi cái **Private Key** (Chìa khóa gốc) giấu kỹ trong ổ cứng ra, chọc vào rương, vặn cái cạch. Lấy được cái **Session Key** tốc độ cao bên trong.
7. **Tận hưởng Tốc độ & Bảo mật:** Kể từ giây phút này, cả Client và Server vứt cặp Public/Private Key qua một bên. Chúng bắt đầu mã hóa toàn bộ hình ảnh, web, video bằng cái **Session Key** tốc độ cao kia. Hacker đứng giữa chỉ nhìn thấy những mảng code mã hóa vô nghĩa bay qua cáp quang với tốc độ ánh sáng! (Cái Session Key này sẽ bị HĐH xóa sổ vĩnh viễn ngay khi bạn tắt tab Trình duyệt).

### 5.4. Luồng Kỹ thuật Thực tế (Packet Flow): Mở đường TCP rồi mới tới TLS
Hãy nhớ rằng, Bắt tay TLS (Layer 7) không thể diễn ra nếu chưa có ống nước kết nối (TCP Socket - Layer 4). Dưới đây là Timeline dòng chảy kỹ thuật thực tế cho một truy cập HTTPS (Port 443):

**GIAI ĐOẠN 1: MỞ KẾT NỐI (TCP 3-Way Handshake - Lớp 4)**
Trước khi bàn chuyện bảo mật, Client phải mở một Socket tới Port 443 của Server để chắc chắn Server đang rảnh.
- `[1] Client -> Server:` **SYN** *(Ê, tôi muốn mở kết nối TCP tới Port 443 của anh)*
- `[2] Server -> Client:` **SYN-ACK** *(OK, tôi đang rảnh, anh kết nối đi)*
- `[3] Client -> Server:` **ACK** *(Xác nhận. Ống TCP đã nối thành công)*
-> *Lúc này 2 bên đã nối dây xong, nhưng CHƯA có một byte dữ liệu Web nào được truyền.*

**GIAI ĐOẠN 2: THỎA THUẬN BẢO MẬT (TLS Handshake - Lớp 7)**
Sử dụng ngay cái ống nước TCP vừa mở, kịch bản lừa Hacker bắt đầu diễn ra:
- `[4] Client -> Server:` **ClientHello** *(Chào, tôi hỗ trợ các chuẩn mã hóa này (Cipher Suites), đây là chuỗi Random 1 của tôi)*
- `[5] Server -> Client:` **ServerHello + Certificate + ServerHelloDone** *(Chào, tôi chốt xài chuẩn mã hóa này. Đây là **Chứng chỉ SSL** chứa Public Key của tôi, và chuỗi Random 2)*
- `[6] Client -> Server:` **ClientKeyExchange + ChangeCipherSpec + Finished** *(Client đã xác thực Chứng chỉ với Root CA thành công. Client gen ra chuỗi bí mật Pre-Master Secret, lấy **Public Key** khóa cái hộp đó lại rồi gửi qua. Sau đó hô lên: "Từ giờ tôi bắt đầu nói chuyện bằng Mã hóa Đối xứng (Session Key) nhé!")*
- `[7] Server -> Client:` **ChangeCipherSpec + Finished** *(Server lấy **Private Key** mở hộp lấy chuỗi bí mật, dùng toán học tạo ra cái **Session Key** y hệt Client. Hô lên: "OK, tôi cũng bắt đầu mã hóa Đối xứng đây!")*

**GIAI ĐOẠN 3: BƠM DỮ LIỆU THẬT (HTTPS Traffic - Đã mã hóa)**
- `[8] Client -> Server:` **HTTP GET /** *(Cái HTTP Payload này đã bị mã hóa nát bét bằng Session Key, gói vào TCP Segment ném qua mạng)*
- `[9] Server -> Client:` **HTTP 200 OK** *(Hình ảnh, HTML trả về cũng bị mã hóa bằng Session Key)*

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
