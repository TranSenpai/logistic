package di

import (
	"log"
	"matching_service/internal/biz"
	"matching_service/internal/broker/kafka"
	"matching_service/internal/broker/nats_jetstream"
	"matching_service/internal/conf"
	"matching_service/internal/controller"
	"matching_service/internal/mapper/generated"
	"matching_service/internal/repo"

	"google.golang.org/grpc"

	entclient "matching_service/internal/common/ent_client"

	pb "github.com/logistic/api/logistic/matching_service/v1"
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
	appMapper := &generated.AppMapperImpl{}

	repo := repo.NewMatchingRepo(masterClient, salveClient, appMapper)
	engine := biz.NewGeoHashEngine()

	natsConec, err := nats.Connect("nats://" + cfg.NatConfig.Host + ":" + cfg.NatConfig.Port)
	natsCtx, err := natsConec.JetStream()
	natsPub := nats_jetstream.InitPublisher(natsCtx, appMapper)

	brokers := make([]string, 3, 3)
	brokers = append(brokers, cfg.KafkaConfig.Port1, cfg.KafkaConfig.Port2, cfg.KafkaConfig.Port3)
	kafkaPub, err := kafka.NewKafkaPublisher(brokers, appMapper)
	if err != nil {
		return err
	}

	// Tạm thời truyền nil cho pub
	biz := biz.NewMatchingEngine(repo, engine, natsPub, kafkaPub)

	controller := controller.NewMatchingController(biz)
	pb.RegisterMatchingEngineServiceServer(grpcServer, controller)

	return nil
}
