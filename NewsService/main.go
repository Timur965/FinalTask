package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	db "FinalTask/NewsService/DB"
	"FinalTask/NewsService/api"
	"FinalTask/NewsService/rss"
	storage "FinalTask/Storage"
)

// конфигурация приложения
type config struct {
	URLS   []string `json:"rss"`
	Period int      `json:"request_period"`
}

func main() {
	storage.InitGetEnv()
	dbNews, err := db.NewNews(storage.DBNewsStr)
	if err != nil {
		log.Fatal(err)
	}
	api := api.New(dbNews)
	api.Router().Use(storage.Middleware)
	api.Router().Use(storage.LoggingMiddleware)

	b, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Fatal(err)
	}
	var config config
	err = json.Unmarshal(b, &config)
	if err != nil {
		log.Fatal(err)
	}

	channelPosts := make(chan []storage.News)
	channelErrs := make(chan error)
	for _, url := range config.URLS {
		go parseURL(url, channelPosts, channelErrs, config.Period)
	}

	go func() {
		for posts := range channelPosts {
			dbNews.AddNews(posts)
		}
	}()

	go func() {
		for err := range channelErrs {
			log.Println("ошибка:", err)
		}
	}()

	err = http.ListenAndServe(":8082", api.Router())
	if err != nil {
		log.Fatal(err)
	}
}
func parseURL(url string, posts chan<- []storage.News, errs chan<- error, period int) {
	for {
		news, err := rss.Parse(url)
		if err != nil {
			errs <- err
			continue
		}
		posts <- news
		time.Sleep(time.Minute * time.Duration(period))
	}
}
