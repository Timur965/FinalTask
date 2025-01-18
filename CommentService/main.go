package main

import (
	HandlersRequest "FinalTask/CommentService/HandleRequest"
	"fmt"
	"net/http"
)

func main() {
	handler, err := HandlersRequest.NewHandler()
	if err != nil {
		fmt.Errorf("Ошибка создания обработчика комментариев")
	}

	http.HandleFunc("/detailedNews", handler.HandleGetComments)
	http.HandleFunc("/addComment", handler.AddCommentHandler)

	err = (http.ListenAndServe(":8081", nil))
	if err != nil {
		fmt.Errorf("Ошибка запуска сервиса комментариев: %s", err.Error())
	}
}
