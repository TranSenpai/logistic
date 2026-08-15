package conf

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	GrpcPort string `env:"VEHICLE_SERVICE_PORT" env-default:"9004"`
}

type DatabaseConfig struct {
	Driver   string `env:"VEHICLE_SERVICE_DB_DRIVER" env-default:"postgres"`
	User     string `env:"VEHICLE_DB_USER" env-required:"true"`
	Password string `env:"VEHICLE_DB_PASSWORD" env-required:"true"`
	Host     string `env:"VEHICLE_DB_HOST" env-required:"true"`
	Port     string `env:"VEHICLE_DB_PORT" env-required:"true"`
	DBName   string `env:"VEHICLE_DB_NAME" env-required:"true"`
}

func (db *DatabaseConfig) GetDataSource() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		db.Host, db.Port, db.User, db.Password, db.DBName)
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	cfg := &Config{}
	err := cleanenv.ReadEnv(cfg)
	return cfg, err
}
