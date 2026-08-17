package di

import (
	"log"
	"matching_service/internal/biz"
	"matching_service/internal/broker/kafka"
	"matching_service/internal/broker/nats_jetstream"
	walletclient "matching_service/internal/client/wallet"
	"matching_service/internal/conf"
	"matching_service/internal/controller"
	"matching_service/internal/mapper/generated"
	"matching_service/internal/repo"

	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	entclient "matching_service/internal/common/ent_client"

	pb "github.com/logistic/api/logistic/matching_service/v1"
	pbwallet "github.com/logistic/api/logistic/wallet_service/v1"
	"github.com/nats-io/nats.go"
)

func Injection(grpcServer *grpc.Server, cfg *conf.Config) error {
	if cfg == nil {
		log.Fatalf("[SYSTEM] failed to read config")
	}

	masterClient, salveClient, err := entclient.NewConnection(&cfg.MasterDatabase, &cfg.SlaveDatabase)
	if err != nil {
		return err
	}

	// Khởi tạo Mapper tập trung
	appMapper := &generated.MatchingMapperImpl{}

	repo := repo.NewMatchingRepo(masterClient, salveClient, appMapper)
	engine := biz.NewGeoHashEngine()

	natsConec, err := nats.Connect("nats://" + cfg.NatConfig.Host + ":" + cfg.NatConfig.Port)
	natsCtx, err := natsConec.JetStream()
	natsPub := nats_jetstream.InitPublisher(natsCtx, appMapper)

	brokers := strings.Split(cfg.KafkaConfig.Brokers, ",")
	kafkaPub, err := kafka.NewKafkaPublisher(brokers, appMapper)
	if err != nil {
		return err
	}

	// ─── WALLET SERVICE CLIENT ──────────────────────────────────────────────
	// Dùng để CheckBalance trước khi chốt match. Việc đóng băng tiền cọc thực
	// tế diễn ra bất đồng bộ qua Kafka topic "wallet.hold_deposit" phía trên.
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

	biz := biz.NewMatchingEngine(repo, engine, walletClient, kafkaPub, natsPub)

	controller := controller.NewMatchingController(biz)
	pb.RegisterMatchingEngineServiceServer(grpcServer, controller)

	return nil
}
