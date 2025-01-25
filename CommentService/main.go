package main

import (
	HandlersRequest "FinalTask/CommentService/HandleRequest"
	storage "FinalTask/Storage"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	handler, err := HandlersRequest.NewHandler()
	if err != nil {
		fmt.Errorf("Ошибка создания обработчика комментариев")
	}

	mux := mux.NewRouter()
	mux.Use(storage.Middleware)
	mux.Use(storage.LoggingMiddleware)

	mux.HandleFunc("/detailedNews", handler.HandleGetComments).Methods(http.MethodGet, http.MethodOptions)
	mux.HandleFunc("/addComment", handler.AddCommentHandler).Methods(http.MethodPost, http.MethodOptions)

	err = (http.ListenAndServe(":8081", mux))
	if err != nil {
		fmt.Errorf("Ошибка запуска сервиса комментариев: %s", err.Error())
	}
}
