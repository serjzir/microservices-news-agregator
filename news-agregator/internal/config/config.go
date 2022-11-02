package config

import (
	"encoding/json"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/serjzir/news-agregator/pkg/logging"
	"io/ioutil"
	"log"
	"sync"
)

// ConfigAPI конфигурация всего приложения
type ConfigAPI struct {
	Listen struct {
		BindIP string `yaml:"bind_ip" env-default:"127.0.0.1"`
		Port   string `yaml:"port" env-default:"8080"`
	} `yaml:"listen"`
	Storage StorageConfig  `yaml:"storage"`
	Path    pathConfigJson `yaml:"path"`
}

// ConfigRSS конфигурация для обработчика новостей
type ConfigRSS struct {
	URLS   []string `json:"rss"`
	Period int      `json:"request_period"`
}

type pathConfigJson struct {
	Config string `json:"config"`
}

// ReadRSSConfig читает конфигурацию из json файла для обработчика новостей
func ReadRSSConfig(p string) *ConfigRSS {
	b, err := ioutil.ReadFile(p)
	if err != nil {
		log.Fatal(err)
	}
	var config *ConfigRSS
	err = json.Unmarshal(b, &config)
	if err != nil {
		log.Fatal(err)
	}
	return config
}

// StorageConfig информация для подключения к БД
type StorageConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
}

var instance *ConfigAPI
var configRss *ConfigRSS
var once sync.Once

// GetConfig прочитает единожды конфигурацию и разложить полученую информацию по структурам
func GetConfig() *ConfigAPI {
	once.Do(func() {
		logger := logging.Init()
		logger.Info("Read configuration")
		instance = &ConfigAPI{}
		if err := cleanenv.ReadConfig("config.yml", instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			logger.Info(help)
			logger.Fatal(err)
		}
	})
	return instance
}
