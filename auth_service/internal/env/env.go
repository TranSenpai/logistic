package env

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var (
	ErrEnvLoad  = errors.New("failed to load environment variables")
	ErrDBConfig = errors.New("database configuration is missing or invalid")
)

type Env struct {
	isProduction *bool
	dataSource   string
	driverName   string
}

func NewEnv() *Env {
	return &Env{
		isProduction: nil,
		dataSource:   "",
		driverName:   "",
	}
}

func (e *Env) loadProductionEnv() error {
	return nil
}

func (e *Env) loadDevelopEnv() error {
	// Cố gắng đọc từ file, nếu không có thì bỏ qua để Docker tự truyền vào bằng system env
	_ = godotenv.Load("configs/.env")

	driverName := os.Getenv("DB_DRIVER_NAME")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	dbName := os.Getenv("DATABASE_NAME")
	isProductionEnv := os.Getenv("IS_PRODUCTION")

	if driverName == "" || host == "" || dbName == "" || isProductionEnv == "" {
		return fmt.Errorf("%w: missing one or more required fields (driver, host, dbName, isProduction)", ErrDBConfig)
	}
	dataSource := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbName)

	isProduction, err := strconv.ParseBool(isProductionEnv)
	if err != nil {
		return fmt.Errorf("%w: missing one or more required fields (driver, host, dbName, isProduction)", ErrDBConfig)
	}

	e.driverName = driverName
	e.dataSource = dataSource
	e.isProduction = &isProduction

	return nil
}

func (e *Env) LoadEnv() error {
	var err error

	if err = e.loadProductionEnv(); err != nil {
		return err
	}
	if err = e.loadDevelopEnv(); err != nil {
		return err
	}

	return nil
}

func (e *Env) GetDriverName() string {
	return e.driverName
}

func (e *Env) GetDataSource() string {
	return e.dataSource
}
