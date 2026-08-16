package conf

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server     ServerConfig
	Cloudinary CloudinaryConfig
}

type ServerConfig struct {
	GrpcPort string `env:"MEDIA_SERVICE_GRPC_PORT" env-default:"9002"`
}

type CloudinaryConfig struct {
	URL string `env:"MEDIA_SERVICE_CLOUDINARY_URL" env-required:"true"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}
	err := cleanenv.ReadEnv(cfg)
	return cfg, err
}
