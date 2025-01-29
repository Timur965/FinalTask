package main

import (
	"FinalTask/CensorService/CensorShip"
	storage "FinalTask/Storage"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	mux := mux.NewRouter()
	mux.Use(storage.Middleware)
	mux.Use(storage.LoggingMiddleware)

	censor, err := CensorShip.NewCensorShip()

	if err != nil {
		fmt.Errorf("Ошибка создания сервиса цензуры %s", err.Error())
		return
	}

	mux.HandleFunc("/checkCensor", censor.CheckCensor).Methods(http.MethodPost, http.MethodOptions)

	err = (http.ListenAndServe(":8083", mux))
	if err != nil {
		fmt.Errorf("Ошибка запуска сервиса цензуры: %s", err.Error())
	}
}
