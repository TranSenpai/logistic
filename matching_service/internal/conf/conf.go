package conf

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server         ServerConfig
	MasterDatabase MasterDatabaseConfig
	SlaveDatabase  SlaveDatabaseConfig
	NatConfig      NatConfig
	KafkaConfig    KafkaConfig
}

type ServerConfig struct {
	GrpcPort     string `env:"MATCHING_SERVICE_GRPC_PORT" env-required:"true"`
	IsProduction bool   `env:"GLOBAL_IS_PRODUCTION" env-default:"false"`
}

type MasterDatabaseConfig struct {
	Driver   string `env:"MATCHING_SERVICE_DB_DRIVER" env-required:"true"`
	User     string `env:"MATCHING_DB_MASTER_USER" env-required:"true"`
	Password string `env:"MATCHING_DB_MASTER_PASSWORD" env-required:"true"`
	Host     string `env:"MATCHING_DB_MASTER_HOST" env-required:"true"`
	Port     string `env:"MATCHING_DB_MASTER_PORT" env-required:"true"`
	DBName   string `env:"MATCHING_DB_MASTER_DB" env-required:"true"`
}

type SlaveDatabaseConfig struct {
	Driver   string `env:"MATCHING_SERVICE_DB_DRIVER" env-required:"true"`
	User     string `env:"MATCHING_DB_SLAVE_USER" env-required:"true"`
	Password string `env:"MATCHING_DB_SLAVE_PASSWORD" env-required:"true"`
	Host     string `env:"MATCHING_DB_SLAVE_HOST" env-required:"true"`
	Port     string `env:"MATCHING_DB_SLAVE_PORT" env-required:"true"`
	DBName   string `env:"MATCHING_DB_SLAVE_DB" env-required:"true"`
}

type NatConfig struct {
	Host string `env:"NATS_HOST" env-required:"true"`
	Port string `env:"NATS_PORT" env-required:"true"`
}

type KafkaConfig struct {
	ClusterId string `env:"KAFKA_KRAFT_CLUSTER_ID" env-required:"true"`
	Brokers   string `env:"KAFKA_BROKERS" env-required:"true"`
}

func (db *MasterDatabaseConfig) GetDataSource() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		db.Host, db.Port, db.User, db.Password, db.DBName)
}

func (db *SlaveDatabaseConfig) GetDataSource() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		db.Host, db.Port, db.User, db.Password, db.DBName)
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	cfg := &Config{}
	err := cleanenv.ReadEnv(cfg)
	return cfg, err
}
