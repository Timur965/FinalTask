package main

import (
	HandlersRequest "FinalTask/APIGateway/Handlers"
	storage "FinalTask/Storage"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	handler := HandlersRequest.NewApiGateway("http://localhost:8082", "http://localhost:8081")

	mux := mux.NewRouter()
	mux.Use(storage.Middleware)
	mux.Use(storage.LoggingMiddleware)

	mux.HandleFunc("/news", handler.HandleGetAllNews).Methods(http.MethodGet, http.MethodOptions)
	mux.HandleFunc("/filterNews", handler.HandleFilterNews).Methods(http.MethodGet, http.MethodOptions)
	mux.HandleFunc("/detailedNews", handler.HandleDetiledNews).Methods(http.MethodGet, http.MethodOptions)
	mux.HandleFunc("/addComment", handler.HandleAddComments).Methods(http.MethodPost, http.MethodOptions)

	err := (http.ListenAndServe(":8080", mux))
	if err != nil {
		fmt.Errorf("Ошибка запуска ApiGateway: %s", err.Error())
	}
}
