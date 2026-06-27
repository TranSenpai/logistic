# BẢN CHẤT CỐT LÕI CỦA INTERNET: IP, TÊN MIỀN & HÀNH TRÌNH CỦA REQUEST

Tài liệu này giải phẫu chi tiết cách Internet hoạt động ở tầng thấp nhất (Physical, Network & DNS) để hiểu rõ vòng đời của một hệ thống từ khi sinh ra IP đến lúc user truy cập được.

---

## Phần 1: Cấu tạo Thế giới Network & Ngôn ngữ của IP

Trước khi nói về Cloud, chúng ta cần hiểu các thành phần vật lý tạo nên "Mạng" và ngôn ngữ chúng dùng để nói chuyện với nhau.

### 1. Cấu tạo các Component cơ bản trong Network
- **Card mạng (Network Interface Card - NIC / ENI):** Là bộ phận phần cứng cắm vào bo mạch chủ máy tính. Không có Card mạng, máy tính không thể kết nối Internet. Mỗi Card mạng có 1 địa chỉ vật lý duy nhất (MAC Address) và được cấp 1 địa chỉ IP để giao tiếp.
- **Switch (Bộ chuyển mạch):** Trạm trung chuyển **trong cùng một mạng LAN** (mạng nội bộ). Khi các máy tính cắm dây vào Switch, chúng có thể nói chuyện trực tiếp với nhau. Switch không biết Internet là gì, nó chỉ điều phối trong nội bộ tòa nhà.
- **Router (Bộ định tuyến):** Cảnh sát giao thông đứng ở cửa ngõ. Router có nhiệm vụ dẫn đường cho các gói tin đi **từ mạng LAN này sang một mạng LAN khác** (kết nối ra ngoài Internet).

### 2. Cấu tạo IP và Câu chuyện IPv4 vs IPv6
IP (Internet Protocol) là tọa độ của máy tính.
- **IPv4:** Gồm 32-bit, chia thành 4 cụm số (Octet) cách nhau bằng dấu chấm (ví dụ: `192.168.1.5`). Mỗi cụm có giá trị từ 0-255. Do chỉ có 32-bit nên toàn Trái Đất chỉ có khoảng 4.3 tỷ địa chỉ IPv4. Hiện tại đã **CẠN KIỆT** hoàn toàn. Do sự khan hiếm này, IPv4 Public hiện nay rất đắt (bạn đang phải trả tiền thuê IP theo giờ trên AWS).
- **IPv6:** Gồm 128-bit, biểu diễn bằng hệ Hexa dài loằng ngoằng (ví dụ: `2001:0db8:85a3::8a2e:0370:7334`). Số lượng IPv6 nhiều đến mức có thể cấp cho mỗi hạt cát trên Trái Đất 1 cái IP. 

### 3. Subnet Mask, CIDR và Dải IP (Block)
Khi nghe nói mạng này có dải IP `13.212.0.0/16`, `/16` (đọc là xuyệt 16) chính là **Subnet Mask** (Mặt nạ mạng) viết dưới dạng CIDR.
- Một dải IP luôn được chia làm 2 phần: **Network ID** (Tên khu phố) và **Host ID** (Số nhà).
- Con số `/16` có nghĩa là: 16 bit đầu tiên (tương đương 2 cụm số đầu `13.212.x.x`) là **Tên khu phố** (bất di bất dịch). Còn 16 bit phía sau (tương đương 2 cụm cuối) dùng để đánh **Số nhà** cấp cho các máy tính.
- **Toán học Subnet:** Dải `/16` sẽ chứa được $2^{(32-16)} = 2^{16} = 65,536$ địa chỉ IP. Dải `/24` (ví dụ `192.168.1.0/24`) thì 24 bit đầu cố định, chỉ còn 8 bit sau làm số nhà, nên có $2^8 = 256$ địa chỉ IP.

---

## Phần 2: Giao thức BGP - Bản đồ Định tuyến Toàn cầu

Bây giờ bạn đã biết dải `/16` là một "khu phố" gồm 65,536 số nhà. Nhưng làm sao cả Trái Đất biết khu phố `13.212.x.x` đó nằm ở đâu? Đó là nhờ **BGP (Border Gateway Protocol)**.

### 1. Autonomous System (AS) và ASN
Trái Đất không xài chung 1 cái Router. Nó được chia thành hàng vạn "Vương quốc mạng" độc lập, gọi là **AS (Autonomous System)**. 
Ví dụ: VNPT là một AS. FPT là một AS. Tập đoàn AWS là một AS siêu to khổng lồ.
Mỗi AS được tổ chức quản lý Internet cấp cho một mã số định danh duy nhất gọi là **ASN (AS Number)**.

### 2. Luồng hoạt động của BGP (BGP Routing)
BGP là ngôn ngữ ngoại giao để các "Vương quốc AS" nói chuyện với nhau.
1. **Loan báo (Announcement):** Khi AWS (AS của Amazon) xin được dải IP `13.212.0.0/16`, AWS sẽ bật hệ thống Router ở biên giới của họ lên và "phát loa" (BGP Announcement) sang các Vương quốc hàng xóm (như Singtel, StarHub): *"Ê Singtel, nguyên cái dải 13.212.x.x là lãnh thổ của tao, nằm ở Singapore nha!"*.
2. **Truyền miệng:** Singtel ghi chú vào **Bảng định tuyến (Routing Table)** của mình. Sau đó Singtel lại phát loa sang Viettel (qua cáp quang biển APG): *"Ê Viettel, tao biết đường đi tới dải 13.212.x.x của AWS ở Sing, mốt ai ở VN hỏi thì mày cứ đẩy gói tin qua tao"*.
3. **Hình thành Bản đồ:** Cứ thế lan truyền trong vài giây, toàn bộ các Router cốt lõi trên Trái Đất đều cập nhật chung một Bản đồ định tuyến vĩ đại.
4. **Hành trình Gói tin:** Khi bạn ở VN gõ `http://13.212.190.202`:
   - Trình duyệt ném gói tin ra cục Modem. Modem đẩy lên Router của Viettel.
   - Router Viettel tra Bảng định tuyến BGP: *"À, muốn đến `13.212.x.x` thì ném qua tuyến cáp biển APG sang thằng Singtel là đường đi ngắn nhất"*.
   - Singtel nhận được gói tin lại ném thẳng vào cổng Router của AWS. AWS ném vào máy EC2 của bạn.
   Đó là bí mật giúp Internet được kết nối liền mạch toàn cầu!

---

## Phần 3: IP Public/Private và Nghệ thuật NAT của Big Tech

Khi IP (gói tin) đã được chuyển về đến biên giới vương quốc AWS, AWS xử lý nó như thế nào?

### 1. Card mạng và IP Nội bộ (Private IP)
Khi lô IP Public về đến tay AWS, AWS **KHÔNG BAO GIỜ** cắm thẳng IP Public vào máy của bạn! Do sự khan hiếm IPv4, phơi bày trực tiếp máy tính ra Internet cũng rất rủi ro. AWS thiết lập một kiến trúc mạng gọi là **VPC (Virtual Private Cloud)**:
- Bên trong VPC, AWS tạo ra một mạng LAN khổng lồ với các dải **IP Private** (IP Nội bộ, ví dụ `172.31.x.x`). Những IP này là miễn phí, bạn có thể tạo hàng vạn cái, nhưng chúng **hoàn toàn vô hình trên Internet** (nhờ quy ước quốc tế RFC 1918).
- Khi bạn ra lệnh tạo máy EC2, AWS thực chất cắm một sợi cáp ảo vào **Card mạng (ENI)** của máy đó và cấp cho nó 1 cái **IP Private**. 
- Nếu bạn SSH vào Ubuntu và gõ lệnh `ifconfig`, bạn sẽ **CHỈ THẤY** cái IP Private này. Con máy EC2 của bạn hoàn toàn "mù tịt", nó không hề biết bản thân nó đang đứng đại diện cho cái IP Public `13.212.x.x`!

### 2. Bí thuật NAT (Network Address Translation) của Router
Làm sao máy EC2 chỉ mang IP Private lại có thể giao tiếp với Internet qua IP Public? Đó là nhờ thiết bị **Internet Gateway (IGW)** của AWS đóng vai trò làm Router biên giới thực hiện phép thuật **NAT**:

- **NAT Inbound (Từ ngoài Internet đi VÀO):** Khi gói tin từ VN chạy đến cổng IGW của AWS. Tại đây, Gateway lấy danh sách ra dò: *"À, gói tin gửi đến IP Public `13.212.x.x`. Cục IP này đang được cho thuê và trói (map) 1-1 với IP Private `172.31.x.x` của thằng Logistic-Lab"*. Gateway lập tức **bóc cái nhãn IP Public vứt đi**, dán nhãn IP Private vào gói tin rồi quăng vào mạng LAN nội bộ (VPC) để nó chạy trúng đích Card mạng của máy EC2. 
- **NAT Outbound (Từ trong LAN đi RA Internet):** Khi máy EC2 (đang mang IP Private `172.31.x.x`) muốn tải Nginx, gói tin chạy ra đến cổng IGW. Khổ nỗi Internet không chấp nhận IP Private (sẽ bị drop ngay lập tức). Gateway lại làm động tác **lột nhãn Private, dán nhãn IP Public `13.212.x.x` vào** làm "địa chỉ gửi", rồi mới ném ra cáp quang biển. Server chứa file cài đặt Nginx nhìn thấy IP Public nên mới biết đường trả dữ liệu về.
- **Sự lợi hại của Big Tech:** Nhờ trò Ảo thuật NAT này, Big Tech có thể bảo vệ hàng triệu máy ảo an toàn tuyệt đối sau bức tường mạng nội bộ. Khi bạn xóa máy EC2, AWS chỉ việc thu hồi IP Public đó về kho, tháo cái map NAT, và 3 giây sau có thể cho người khác thuê lại.

---

## Phần 4: Name Server (NS) và Quyền lực Phân thân (Subdomain)
IP rất khó nhớ (`13.212.190.202`). Con người thì chỉ nhớ Tên miền (`glolog.dev`). 
Vấn đề là: **Làm sao máy tính biết `glolog.dev` là cái IP nào để kết nối?** 

Chúng ta cần một **Hệ thống Danh bạ Điện thoại**. Hệ thống đó gọi là **DNS** (Domain Name System).
- **Name Server (NS) chính là Những Cuốn Danh Bạ.** Nhiệm vụ của máy chủ NS là chứa các bản ghi (DNS Records) để đối chiếu Tên miền thành IP.
- **Sức mạnh của Cuốn danh bạ:** Bạn mua 1 Domain, không có nghĩa là nó bị dính chặt vào 1 IP. Một cuốn danh bạ có thể lưu vô số số điện thoại của các phòng ban khác nhau (Subdomain). Bạn có thể cấu hình:
  - `api.glolog.dev` -> Trỏ về IP Máy chủ Backend (đặt tại Singapore).
  - `web.glolog.dev` -> Trỏ về IP Máy chủ Frontend (đặt tại Tokyo).
  - `db.glolog.dev` -> Trỏ về IP Máy chủ Database (đặt tại Mỹ).
  Nhờ NS, chỉ với 1 tên miền gốc, bạn có thể điều phối giao thông (Traffic) bay đi khắp thế giới!

- **Trỏ Name Server là gì?** Khi mua tên miền ở GoDaddy, họ cho bạn cuốn danh bạ mặc định. Nhưng bạn muốn xài Cloudflare (để có khiên chống DDoS, WAF lọc SQL Injection). Bạn phải **Trỏ Name Server về Cloudflare**, tức là tuyên bố với thế giới: *"Tôi giao cuốn sổ danh bạ của tôi cho Cloudflare giữ. Mọi luật lệ phân luồng, thêm bớt tính năng bảo mật... Cloudflare sẽ toàn quyền quyết định!"*
- *(Nếu mua domain thẳng trên Cloudflare, họ tự lấy sổ nhà họ ra xài, bạn không cần làm bước trỏ này nữa).*

---

## Phần 5: Hành trình trọn vẹn của 1 Request (Từ User đến Backend)
Bây giờ hãy ghép tất cả lại. Khi một User ở Việt Nam gõ `http://api.glolog.dev` vào trình duyệt, chuyện gì sẽ xảy ra trong vài mili-giây?

### Chặng 1: Truy tìm Danh bạ (DNS Resolution)
1. Trình duyệt không biết IP. Nó hỏi cục Modem Wifi mạng Viettel ở nhà bạn: *"IP của `api.glolog.dev` là gì?"*.
2. Viettel không biết. Viettel chạy lên hỏi Máy chủ gốc (Root Server) của thế giới. Root Server nói: *"Tao không quản lý chi tiết, nhưng đuôi `.dev` thuộc quyền của máy chủ TLD kia, qua đó mà hỏi"*.
3. Viettel chạy qua máy chủ `.dev`. Máy chủ `.dev` mở sổ ra nói: *"À, thằng chủ cái tên miền `glolog.dev` này nó đã **TRỎ NAME SERVER (NS)** về nhà **Cloudflare** rồi. Mày qua Cloudflare mà xin IP!"*.

### Chặng 2: Tổng đài viên Cloudflare trả lời
4. Viettel chạy tới máy chủ NS của Cloudflare hỏi xin IP.
   - **Nếu bạn xài Đám mây Xám (Proxied = false):** Cloudflare thật thà lật sổ ra đọc: *"IP của nó là `13.212.x.x` (IP thật của AWS)"*.
   - **Nếu bạn xài Đám mây Cam (Proxied = true):** Cloudflare "nói dối" để bảo vệ bạn: *"IP của nó là `104.18.25.10` (IP ảo của máy chủ bảo vệ Cloudflare)"*.
5. Cục Modem Viettel ném cái IP vừa nhận được về cho Trình duyệt web của User.

### Chặng 3: Kết nối và Xử lý (TCP/IP)
6. Trình duyệt có IP (của Cloudflare). Nó bắt đầu đóng gói dữ liệu (HTTP Request) và bắn lên cáp quang biển đến máy chủ Cloudflare.
7. Máy chủ bảo vệ của Cloudflare nhận Request. Nó kiểm tra xem thằng User này có phải Hacker hay Bot DDoS không. Nếu là người thật, Cloudflare mới âm thầm bắt một chuyến xe chở Request đó chạy cửa sau về **IP thật của AWS** (`13.212.x.x`).
8. Gói tin đập vào Data Center AWS ở Singapore. Tường lửa (**Security Group** của AWS) ra chặn lại hỏi: *"Mày vào port nào?"*. Gói tin đáp: *"Port 80 (Web)"*. Security Group thấy Port 80 được phép mở, liền cho gói tin chui vào mạng nội bộ qua cổng Gateway. Ở Gateway, phép thuật NAT xảy ra (bóc IP Public, dán IP Private).
9. Bên trong Máy ảo EC2, anh lính gác cổng **Nginx (Reverse Proxy)** đứng đón. Nginx đọc gói tin thấy gọi vào `/api`, liền quay lưng ném gói tin đó sang **Port 8080** cho chương trình **Golang** (chạy bằng Docker/Podman).
10. Golang nhận dữ liệu, gọi xuống Database (Postgres), lấy data, đóng gói thành 1 chuỗi JSON và bắn ngược về con đường cũ... trả lại cho User trên màn hình trình duyệt.

Và tất cả quá trình đồ sộ này diễn ra trong chưa tới 100 mili-giây!

---

## Phần 6: Sự Tiến Hóa Hạ Tầng (Từ Bare Metal đến Cloud Big Tech)

Những khái niệm IP, DNS, Server ở trên là nền tảng cốt lõi không thay đổi, nhưng cách con người "sở hữu" và "vận hành" chúng đã tiến hóa vượt bậc qua 3 thời kỳ:

### 1. Thời kỳ Đồ đá: Bare Metal (Máy chủ vật lý)
- **Cách làm:** Bạn phải ra cửa hàng Phong Vũ mua CPU, RAM, Ổ cứng về lắp thành một cái máy tính cục mịch (Bare Metal). Sau đó bưng máy đó lên trung tâm dữ liệu của nhà mạng (Viettel/FPT) thuê chỗ đặt (Colocation). Viettel sẽ cắm dây mạng, cấp điện, điều hòa và cấp cho bạn 1 cái **IP Public tĩnh**.
- **Nỗi đau:** Máy hư ổ cứng lúc nửa đêm -> Xách xe chạy lên Data Center thay. Cáp mạng lỏng -> Chạy lên cắm lại. Tiền điện, tiền điều hòa, chi phí khấu hao phần cứng cực kỳ đắt đỏ và lãng phí (nếu máy chạy không hết công suất).

### 2. Thời kỳ Đồ Đồng: VPS (Virtual Private Server)
- **Cách làm:** Các nhà cung cấp (Hostinger, DigitalOcean, Vultr...) mua những cái Bare Metal siêu to khổng lồ, cài phần mềm ảo hóa (VMware/KVM) để "băm" cái máy to đó ra thành 100 cái máy nhỏ (VPS). Họ bán cho bạn 1 cái VPS.
- **Giải quyết:** Bạn không cần phải đụng tay vào phần cứng nữa. Mọi thứ thao tác qua màn hình (SSH). Cháy nổ hư hỏng phần cứng nhà mạng tự lo.
- **Nỗi đau:** Khả năng mở rộng (Scale) rất kém. Bạn đang thuê máy 8GB RAM, đợt Sale 11/11 traffic tăng vọt muốn up lên 64GB RAM là cả một cực hình (thường phải tắt máy, backup data, nơm nớp lo sợ mất dữ liệu). Ngoài ra, bạn thuê VPS theo tháng, xài hay không xài vẫn phải trả đủ tiền.

### 3. Thời kỳ Hiện đại: Cloud Computing (Các tập đoàn Big Tech - AWS, GCP, Azure)
- **Cách làm:** Amazon, Google, Microsoft không "bán máy ảo", họ bán **"Tài nguyên tính toán"** y như một loại điện nước. Bạn xài bao nhiêu giây, xài bao nhiêu GB RAM thì tính tiền đúng bấy nhiêu (Pay-as-you-go).
- **Đột phá công nghệ:**
  - **Ảo hóa đỉnh cao (EC2):** Không phải là VPS thông thường. Bạn có thể ra lệnh tạo 1.000 cái máy EC2 trong vòng 3 phút, và xóa sạch chúng trong 1 phút sau khi tính toán xong.
  - **Tách rời Ổ cứng và CPU:** Trên Cloud (như AWS), ổ cứng (EBS Volume) và cục tính toán (EC2 Instance) là 2 thực thể tách rời. Máy bị cháy CPU? Chuyện nhỏ, gỡ cái ổ cứng ra cắm sang một cái máy CPU mới là chạy tiếp như chưa hề có cuộc chia ly.
  - **Infrastructure as Code (IaC - Terraform):** Thay vì lấy chuột click tạo máy mỏi tay, bạn viết code `main.tf`. Code được chạy (apply) thì máy mọc lên, code bị xóa (destroy) thì máy biến mất.
  - **Hệ sinh thái Managed Services:** Không chỉ dừng ở máy ảo, các Big Tech bán luôn các dịch vụ "ăn sẵn". Bạn cần Database? Có Amazon RDS. Bạn cần Message Queue? Có Amazon SQS. Bạn chỉ việc trả tiền, các kĩ sư giỏi nhất thế giới của Google/Amazon sẽ lo phần bảo trì, backup hệ thống thay bạn. Nhiệm vụ của bạn chỉ còn tập trung vào code Logic App.
