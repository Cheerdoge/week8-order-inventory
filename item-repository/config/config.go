package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	KafkaBroker   string
	EtcdEndpoints string
	ServicePort   string
	GRPCPort      string
	PODIP         string
}

var APPConfig *Config

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file", err)
	}
	APPConfig = &Config{
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "3306"),
		DBUser:        getEnv("DB_USER", "root"),
		DBPassword:    getEnv("DB_PASSWORD", "password"),
		DBName:        getEnv("DB_NAME", "ItemRepositoryDB"),
		KafkaBroker:   getEnv("KAFKA_BROKER", "kafka:9092"),
		EtcdEndpoints: getEnv("ETCD_ENDPOINTS", "etcd:2379"),
		ServicePort:   getEnv("SERVICE_PORT", "8081"),
		GRPCPort:      getEnv("GRPC_PORT", "50051"),
		PODIP:         getEnv("POD_IP", "127.0.0.1"),
	}

}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
