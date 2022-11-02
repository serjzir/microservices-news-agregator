package main

import (
	"context"
	"fmt"
	"github.com/serjzir/news-agregator/internal/config"
	"github.com/serjzir/news-agregator/pkg/api"
	"github.com/serjzir/news-agregator/pkg/client/postgresql"
	"github.com/serjzir/news-agregator/pkg/clientrss"
	"github.com/serjzir/news-agregator/pkg/logging"
	"github.com/serjzir/news-agregator/pkg/storage"
	"time"
)

func main() {
	// инициализация логгера
	logger := logging.Init()
	logger.Info("Run learn application")
	logger.Info("Create new database client")

	// получение конфигурации для API
	configAPI := config.GetConfig()

	// создание клиента к БД
	clientDB, err := postgresql.NewClient(context.Background(), 3, *configAPI)
	if err != nil {
		logger.Fatalf("%v", err)
	}

	// создание репозитория
	db := storage.NewRepository(clientDB, logger)

	// инициализация API
	api := api.New(db)

	// запуск парсинга новостей в отдельном потоке
	chPosts := make(chan []storage.Post)
	chErrs := make(chan error)
	// конфигурация для обработчика RSS
	configRSS := config.ReadRSSConfig(configAPI.Path.Config)
	for _, url := range configRSS.URLS {
		go ParserUrl(url, chPosts, chErrs, configRSS.Period)
	}

	// запись новостей в БД
	go func() {
		logger.Info("Запуск обработчика новостей записывающего новости в БД")
		for posts := range chPosts {
			db.Create(context.Background(), posts)
		}
	}()

	// обработка потока ошибок
	go func() {
		fmt.Println("Запуск обработчика ошибок")
		for err := range chErrs {
			logger.Error(err)
		}
	}()

	// запуск веб-сервера
	api.GetRouter(configAPI.Listen.BindIP, configAPI.Listen.Port)
}

// ParserUrl асинхронное чтение потока RSS, обработанные новости и ошибки пишутся в каналы
func ParserUrl(url string, posts chan<- []storage.Post, errs chan<- error, period int) {
	for {
		news, err := clientrss.Parse(url)
		if err != nil {
			errs <- err
			continue
		}
		posts <- news
		time.Sleep(time.Minute * time.Duration(period))
	}
}
