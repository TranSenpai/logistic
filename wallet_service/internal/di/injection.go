package di

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"wallet_service/internal/biz"
	"wallet_service/internal/broker/kafka"
	"wallet_service/internal/conf"
	"wallet_service/internal/controller"
	"wallet_service/internal/mapper"
	"wallet_service/internal/mapper/generated"
	"wallet_service/internal/repository"
	"wallet_service/internal/search"

	pb "github.com/logistic/api/logistic/wallet_service/v1"

	"wallet_service/ent"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/logistic/pkg/concurrencyUtils"

	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"
)

type HoldDepositPayload struct {
	DriverID   string  `json:"driver_id"`
	Amount     float64 `json:"amount"`
	ContractID string  `json:"contract_id"`
}

func Injection(ctx context.Context, grpcServer *grpc.Server, cfg *conf.Config) (func(), error) {
	if cfg.Database.DSN == "" {
		return nil, fmt.Errorf("Database DSN is required")
	}

	db, err := entsql.Open(dialect.MySQL, cfg.Database.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	entClient := ent.NewClient(ent.Driver(db))

	esAddresses := strings.Split(cfg.ElasticSearch.Addresses, ",")
	if len(esAddresses) == 0 || esAddresses[0] == "" {
		esAddresses = []string{"http://localhost:9200"}
	}

	esEngine, err := search.NewElasticSearchEngine(esAddresses, cfg.ElasticSearch.Username, cfg.ElasticSearch.Password)
	if err != nil {
		log.Printf("Failed to init elasticsearch: %v. Search will be disabled.", err)
		esEngine = nil
	} else {
		esEngine.EnsureIndices(ctx)
	}

	var walletMapper mapper.WalletMapper = &generated.WalletMapperImpl{}

	uow := repository.NewUnitOfWorkRepository(entClient)
	walletRepo := repository.NewWalletRepository(entClient, walletMapper)
	txRepo := repository.NewTransactionRepository(entClient, walletMapper)

	walletUseCase := biz.NewWalletUseCase(uow, walletRepo, txRepo, esEngine, walletMapper)

	walletController := controller.NewWalletController(walletUseCase, esEngine, walletMapper)
	pb.RegisterWalletServiceServer(grpcServer, walletController)

	brokers := strings.Split(cfg.Kafka.Brokers, ",")
	if len(brokers) == 0 || brokers[0] == "" {
		brokers = []string{"localhost:9092"}
	}

	kafkaConsumer, err := kafka.NewKafkaConsumer(brokers, "wallet-service-group")
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	bufferSize := 100
	msgChan := make(chan []byte, bufferSize)

	holdDepositHandler := func(ctx context.Context, payload []byte) error {
		var event HoldDepositPayload
		if err := json.Unmarshal(payload, &event); err != nil {
			log.Printf("[Wallet] Failed to unmarshal hold_deposit event: %v", err)
			return fmt.Errorf("%w: %v", biz.ErrNonRetryable, err)
		}

		driverID, err := uuid.Parse(event.DriverID)
		if err != nil {
			return fmt.Errorf("%w: invalid driver_id %s", biz.ErrNonRetryable, event.DriverID)
		}

		amount := int64(event.Amount * 100)

		if err := walletUseCase.HoldDeposit(ctx, driverID, amount, event.ContractID); err != nil {
			log.Printf("[Wallet] HoldDeposit failed: %v", err)
			return fmt.Errorf("%w: %v", biz.ErrNonRetryable, err)
		}

		log.Printf("[Wallet] HoldDeposit success: driver=%s amount=%d refID=%s", event.DriverID, amount, event.ContractID)
		return nil
	}

	worker := concurrencyUtils.NewWorker(msgChan, holdDepositHandler)
	worker.Start(ctx)

	go func() {
		err := kafkaConsumer.Consume(ctx, "wallet.hold_deposit", func(c context.Context, bucket []byte) error {
			select {
			case msgChan <- bucket:
			case <-ctx.Done():
			}
			return nil
		})
		if err != nil {
			log.Printf("[Wallet] Kafka consumer stopped: %v", err)
		}
	}()

	log.Println("[Wallet] Kafka consumer started on topic: wallet.hold_deposit")

	cleanup := func() {
		worker.Stop()
		entClient.Close()
		log.Println("[Wallet] Cleanup complete")
	}

	return cleanup, nil
}