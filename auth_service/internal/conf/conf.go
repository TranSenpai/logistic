package conf

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Google   GoogleConfig   `yaml:"google"`
	JWT      JWTConfig      `yaml:"jwt"`
}

type ServerConfig struct {
	GrpcPort     string `yaml:"grpc_port" env:"AUTH_GRPC_PORT" env-required:"true"`
	IsProduction bool   `yaml:"is_production" env:"IS_PRODUCTION" env-default:"false"`
}

type DatabaseConfig struct {
	Driver   string `yaml:"driver" env:"DB_DRIVER_NAME" env-required:"true"`
	User     string `env:"POSTGRES_USER" env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
	Host     string `env:"POSTGRES_HOST" env-required:"true"`
	Port     string `env:"POSTGRES_PORT" env-required:"true"`
	DBName   string `env:"DATABASE_NAME" env-required:"true"`
}

type GoogleConfig struct {
	ClientID     string `yaml:"client_id" env:"GOOGLE_CLIENT_ID" env-required:"true"`
	ClientSecret string `yaml:"client_secret" env:"GOOGLE_CLIENT_SECRET" env-required:"true"`
	RedirectURL  string `yaml:"redirect_url" env:"GOOGLE_REDIRECT_URL" env-required:"true"`
}

type JWTConfig struct {
	Secret string `yaml:"secret" env:"JWT_SECRET" env-required:"true"`
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
