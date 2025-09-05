package configs

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type ConfigDatabase struct {
	User     string `envconfig:"L0_WB_DB_USER"`
	Name     string `envconfig:"L0_WB_DB_NAME"`
	Password string `envconfig:"L0_WB_DB_PASSWORD"`
	Host     string `envconfig:"L0_WB_DB_HOST"`
	Port     string `envconfig:"L0_WB_DB_PORT"`
	ModeSSL  string `envconfig:"L0_WB_DB_SSL_MODE"`
}

type ConfigKafka struct {
	Host string `envconfig:"L0_WB_KAFKA_HOST"`
	Port string `envconfig:"L0_WB_KAFKA_PORT"`
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
