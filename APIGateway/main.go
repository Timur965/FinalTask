package main

import (
	HandlersRequest "FinalTask/APIGateway/Handlers"
	"fmt"
	"net/http"
)

func main() {
	handler := HandlersRequest.NewApiGateway("http://localhost:8082", "http://localhost:8081")

	http.HandleFunc("/news", handler.HandleGetAllNews)
	http.HandleFunc("/filterNews", handler.HandleFilterNews)
	http.HandleFunc("/detailedNews", handler.HandleDetiledNews)
	http.HandleFunc("/addComment", handler.HandleAddComments)

	err := (http.ListenAndServe(":8080", nil))
	if err != nil {
		fmt.Errorf("Ошибка запуска ApiGateway: %s", err.Error())
	}
}
