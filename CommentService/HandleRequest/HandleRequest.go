package handlerequest

import (
	db "FinalTask/CommentService/DBComment"
	storage "FinalTask/Storage"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type handlerRequests struct {
	db *db.PostgresComments
}

func NewHandler() (*handlerRequests, error) {
	var newHandler handlerRequests
	var err error

	newHandler.db = new(db.PostgresComments)

	newHandler.db, err = db.New(storage.DBCommentStr)
	if err != nil {
		log.Println(err)
	}

	return &newHandler, err
}

func (hr *handlerRequests) HandleGetComments(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("idNews"))

	if err != nil {
		w.Write([]byte(err.Error()))
	}

	news, err := hr.db.GetComments(id)
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	text, err := json.Marshal(news)
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	w.Write(text)
}

func (hr *handlerRequests) AddCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var comment storage.Comments
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		log.Println("Ошибка декодирования запроса:", err)
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}

	if comment.NewsId <= 0 {
		http.Error(w, "Неверный ID новости", http.StatusBadRequest)
		return
	}

	if comment.Content == "" {
		http.Error(w, "Содержимое комментария не может быть пустым", http.StatusBadRequest)
		return
	}

	err := hr.db.AddComments(comment)
	if err != nil {
		http.Error(w, "Не удалось добавить данные в БД", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(comment); err != nil {
		log.Println("Ошибка кодирования ответа:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
	}
}
