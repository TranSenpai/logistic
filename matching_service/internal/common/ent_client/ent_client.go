package entclient

import (
	"context"
	"errors"
	"fmt"
	"log"
	"matching_service/ent"
	"matching_service/internal/conf"

	_ "matching_service/ent/runtime"

	_ "github.com/lib/pq"
)

var (
	ErrDBMasterConnect = errors.New("failed to open database master connection")
	ErrDBSlaveConnect  = errors.New("failed to open database slave connection")
	ErrDBPing          = errors.New("database is unreachable (ping failed)")
)

func NewConnection(masterCfg *conf.MasterDatabaseConfig, slaveCfg *conf.SlaveDatabaseConfig) (*ent.Client, *ent.Client, error) {
	if masterCfg == nil || slaveCfg == nil {
		log.Fatalf("[SYSTEM] failed to read postgres's config")
	}

	masterClient, err := ent.Open(masterCfg.Driver, masterCfg.GetDataSource())
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrDBMasterConnect, err)
	}

	slaveClient, err := ent.Open(slaveCfg.Driver, slaveCfg.GetDataSource())
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrDBSlaveConnect, err)
	}
	// bật tính tăng sử dụng câu lệnh sql trong code
	// patth file: go-backend/ent/generate.go
	// //go:generate go run -mod=mod entgo.io/ent/cmd/ent generate --feature sql/execquery ./schema// bật tính tăng sử dụng câu lệnh sql trong code
	// patth file: go-backend/ent/generate.go
	// //go:generate go run -mod=mod entgo.io/ent/cmd/ent generate --feature sql/execquery ./schema
	_, err = masterClient.QueryContext(context.Background(), "SELECT 1")
	if err != nil {
		log.Fatalf("[ENT] failed connection to postgres: %v", err)
	}

	// Run the auto migration tool.
	if err := masterClient.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	return masterClient, slaveClient, nil
}
