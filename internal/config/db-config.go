package config

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	customerrors "AvitoTest/internal/custom-errors"

	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

type DBPSQLConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	MaxConnLifetime int
}

// LoadDBConfig загружает конфиг БД
func LoadDBConfig(envPath string) (*DBPSQLConfig, error) {
	if envPath != "" {
		if err := loadEnvFile(envPath); err != nil {
			return nil, fmt.Errorf("ошибка загрузки .env файла: %w", err)
		}
	}

	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")
	sslmode := os.Getenv("POSTGRES_SSLMODE")

	if host == "" || port == "" || user == "" || pass == "" || dbname == "" || sslmode == "" {
		return nil, fmt.Errorf("ошибка конфигурации БД: %w", customerrors.ErrParamNotFound)
	}

	maxOpen := getIntOrDefault("POSTGRES_MAX_OPEN_CONNS", 25)
	maxIdle := getIntOrDefault("POSTGRES_MAX_IDLE_CONNS", 10)
	maxLife := getIntOrDefault("POSTGRES_CONN_MAX_LIFETIME", 15)

	return &DBPSQLConfig{
		Host:            host,
		Port:            port,
		User:            user,
		Password:        pass,
		DBName:          dbname,
		SSLMode:         sslmode,
		MaxOpenConns:    maxOpen,
		MaxIdleConns:    maxIdle,
		MaxConnLifetime: maxLife,
	}, nil
}

func getIntOrDefault(key string, def int) int {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	v, err := strconv.Atoi(val)
	if err != nil {
		return def
	}
	return v
}

// InitDB создает и настраивает подключение к БД
func InitDB(cfg *DBPSQLConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBCreation, err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.MaxConnLifetime) * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBCreation, err)
	}

	return db, nil
}

// loadEnvFile загружает переменные окружения из .env файла
func loadEnvFile(path string) error {
	if err := godotenv.Load(path); err != nil {
		return nil
	}
	return nil
}
