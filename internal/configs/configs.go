package configs

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type ConfigDatabase struct {
	User     string `envconfig:"L0_WB_DB_USER" default:"postgres"`
	Name     string `envconfig:"L0_WB_DB_NAME" default:"l0_wb"`
	Password string `envconfig:"L0_WB_DB_PASSWORD" default:"postgres"`
	Host     string `envconfig:"L0_WB_DB_HOST" default:"localhost"`
	Port     string `envconfig:"L0_WB_DB_PORT" default:"5432"`
	ModeSSL  string `envconfig:"L0_WB_DB_SSL_MODE" default:"disable"`
}

type ConfigKafka struct {
	Host string `envconfig:"L0_WB_KAFKA_HOST" default:"localhost"`
	Port string `envconfig:"L0_WB_KAFKA_PORT" default:"9092"`
}

type ConfigCache struct {
	Size         int `envconfig:"L0_WB_CACHE_SIZE" default:"10000"`
	PreloadLimit int `envconfig:"L0_WB_CACHE_PRELOAD_LIMIT" default:"0"`
}

func NewConfigDB() (*ConfigDatabase, error) {
	var dbConfig ConfigDatabase
	err := envconfig.Process("l0_wb_db", &dbConfig)
	if err != nil {
		return nil, fmt.Errorf("could not process database env: %s", err.Error())
	}
	return &dbConfig, nil
}

func NewConfigKafka() (*ConfigKafka, error) {
	var kafkaConfig ConfigKafka
	err := envconfig.Process("l0_wb_kafka", &kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("could not process kafka env: %s", err.Error())
	}
	return &kafkaConfig, nil
}

func NewConfigCache() (*ConfigCache, error) {
	var cacheConfig ConfigCache
	if err := envconfig.Process("l0_wb_cache", &cacheConfig); err != nil {
		return nil, fmt.Errorf("could not process cache env: %s", err.Error())
	}
	return &cacheConfig, nil
}
