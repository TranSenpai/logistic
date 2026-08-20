package di

import (
	"log"
	"strings"

	"matching_service/internal/biz"
	"matching_service/internal/broker/kafka"
	"matching_service/internal/broker/nats_jetstream"
	brokerrabbit "matching_service/internal/broker/rabbitmq"
	walletclient "matching_service/internal/client/wallet"
	"matching_service/internal/conf"
	"matching_service/internal/controller"
	"matching_service/internal/mapper/generated"
	"matching_service/internal/repo"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	entclient "matching_service/internal/common/ent_client"

	pb "github.com/logistic/api/logistic/matching_service/v1"
	pbwallet "github.com/logistic/api/logistic/wallet_service/v1"
	"github.com/logistic/pkg/mq"
	"github.com/nats-io/nats.go"
)

// Container giữ các tài nguyên nền cần đóng lúc tắt service.
type Container struct {
	MQConn      *mq.Connection
	MQPublisher *mq.Publisher
}

func (c *Container) Close() {
	if c == nil {
		return
	}
	if c.MQPublisher != nil {
		if err := c.MQPublisher.Close(); err != nil {
			log.Printf("[matching_service] closing rabbitmq publisher failed: %v", err)
		}
	}
	if c.MQConn != nil {
		if err := c.MQConn.Close(); err != nil {
			log.Printf("[matching_service] closing rabbitmq failed: %v", err)
		}
	}
}

func Injection(grpcServer *grpc.Server, cfg *conf.Config) (*Container, error) {
	if cfg == nil {
		log.Fatalf("[SYSTEM] failed to read config")
	}

	masterClient, salveClient, err := entclient.NewConnection(&cfg.MasterDatabase, &cfg.SlaveDatabase)
	if err != nil {
		return nil, err
	}

	// Mapper tập trung, dùng chung cho repo / controller / broker.
	appMapper := &generated.MatchingMapperImpl{}

	matchingRepo := repo.NewMatchingRepo(masterClient, salveClient, appMapper)
	engine := biz.NewGeoHashEngine()

	// ─── NATS JETSTREAM (đẩy realtime tới app đang mở) ──────────────────────
	natsConn, err := nats.Connect("nats://" + cfg.NatConfig.Host + ":" + cfg.NatConfig.Port)
	if err != nil {
		return nil, err
	}
	natsCtx, err := natsConn.JetStream()
	if err != nil {
		return nil, err
	}
	natsPub := nats_jetstream.InitPublisher(natsCtx, appMapper)

	// ─── KAFKA (nhật ký sự kiện lâu dài) ────────────────────────────────────
	brokers := strings.Split(cfg.KafkaConfig.Brokers, ",")
	kafkaPub, err := kafka.NewKafkaPublisher(brokers, appMapper)
	if err != nil {
		return nil, err
	}

	container := &Container{}

	// ─── RABBITMQ (thông báo bền tới người dùng) ────────────────────────────
	// Đây là kênh phục vụ hai nghiệp vụ chính:
	//   1. Chủ hàng đăng đơn -> báo cho các tài xế tiềm năng.
	//   2. Ghép được xe      -> báo cho cả chủ hàng lẫn tài xế.
	// Dựng không được thì dùng NoopNotifier: việc ghép đơn vẫn chạy bình thường,
	// chỉ mất phần thông báo.
	var notifier biz.Notifier = biz.NoopNotifier{}
	if cfg.RabbitMQ.Enabled {
		mqConn, mqErr := mq.Connect(mq.Config{
			Host:     cfg.RabbitMQ.Host,
			Port:     cfg.RabbitMQ.Port,
			User:     cfg.RabbitMQ.User,
			Password: cfg.RabbitMQ.Password,
			VHost:    cfg.RabbitMQ.VHost,
		})
		if mqErr != nil {
			log.Printf("[matching_service] CẢNH BÁO: không kết nối được RabbitMQ (%v) — "+
				"người dùng sẽ KHÔNG nhận được thông báo ghép đơn", mqErr)
		} else {
			publisher, pErr := mq.NewPublisher(mqConn, cfg.RabbitMQ.Exchange, "matching_service")
			if pErr != nil {
				log.Printf("[matching_service] CẢNH BÁO: không tạo được publisher RabbitMQ (%v)", pErr)
				_ = mqConn.Close()
			} else {
				container.MQConn = mqConn
				container.MQPublisher = publisher
				notifier = brokerrabbit.NewNotifier(publisher)
				log.Printf("[matching_service] RabbitMQ sẵn sàng — exchange=%s", cfg.RabbitMQ.Exchange)
			}
		}
	} else {
		log.Printf("[matching_service] Publisher RabbitMQ bị tắt bằng cấu hình")
	}

	// ─── WALLET SERVICE CLIENT ──────────────────────────────────────────────
	// Dùng để CheckBalance trước khi chốt match. Việc đóng băng tiền cọc thực
	// tế diễn ra bất đồng bộ qua Kafka topic "wallet.hold_deposit".
	var walletClient biz.WalletClient
	walletConn, err := grpc.NewClient(
		cfg.WalletService.GrpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("Failed to connect to wallet_service via gRPC: %v. Falling back to mock wallet client.", err)
		walletClient = biz.NewMockWalletClient()
	} else {
		walletClient = walletclient.NewGrpcClient(pbwallet.NewWalletServiceClient(walletConn))
	}

	matchingEngine := biz.NewMatchingEngine(matchingRepo, engine, walletClient, kafkaPub, natsPub, notifier)

	matchingController := controller.NewMatchingController(matchingEngine)
	pb.RegisterMatchingEngineServiceServer(grpcServer, matchingController)

	return container, nil
}
