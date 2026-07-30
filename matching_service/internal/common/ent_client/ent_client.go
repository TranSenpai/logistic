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
	ErrDBConnect = errors.New("failed to open database connection")
	ErrDBPing    = errors.New("database is unreachable (ping failed)")
)

func NewConnection(dbCfg conf.DatabaseConfig) (*ent.Client, error) {
	client, err := ent.Open(dbCfg.Driver, dbCfg.GetDataSource())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDBConnect, err)
	}
	// bật tính tăng sử dụng câu lệnh sql trong code
	// patth file: go-backend/ent/generate.go
	// //go:generate go run -mod=mod entgo.io/ent/cmd/ent generate --feature sql/execquery ./schema// bật tính tăng sử dụng câu lệnh sql trong code
	// patth file: go-backend/ent/generate.go
	// //go:generate go run -mod=mod entgo.io/ent/cmd/ent generate --feature sql/execquery ./schema
	_, err = client.QueryContext(context.Background(), "SELECT 1")
	if err != nil {
		log.Fatalf("[ENT] failed connection to postgres: %v", err)
	}

	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	return client, nil
}
