package conf

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Google   GoogleConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	GrpcPort     string `env:"AUTH_SERVICE_GRPC_PORT" env-required:"true"`
	IsProduction bool   `env:"GLOBAL_IS_PRODUCTION" env-default:"false"`
}

type DatabaseConfig struct {
	Driver   string `env:"AUTH_SERVICE_DB_DRIVER" env-required:"true"`
	User     string `env:"AUTH_SERVICE_DB_USER" env-required:"true"`
	Password string `env:"AUTH_SERVICE_DB_PASSWORD" env-required:"true"`
	Host     string `env:"AUTH_SERVICE_DB_HOST" env-required:"true"`
	Port     string `env:"AUTH_SERVICE_DB_PORT" env-required:"true"`
	DBName   string `env:"AUTH_SERVICE_DB_NAME" env-required:"true"`
}

type GoogleConfig struct {
	ClientID     string `env:"AUTH_SERVICE_GOOGLE_CLIENT_ID" env-required:"true"`
	ClientSecret string `env:"AUTH_SERVICE_GOOGLE_CLIENT_SECRET" env-required:"true"`
	RedirectURL  string `env:"AUTH_SERVICE_GOOGLE_REDIRECT_URL" env-required:"true"`
}

type JWTConfig struct {
	Secret string `env:"AUTH_SERVICE_JWT_SECRET" env-required:"true"`
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
