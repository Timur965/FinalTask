package CensorShip

import (
	storage "FinalTask/Storage"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

type CensorShip struct {
	forbiddenWords map[string]string
}

func NewCensorShip() (*CensorShip, error) {
	var cs CensorShip
	textByte, err := os.ReadFile("forbiddenWords.txt")
	if err != nil {
		return nil, err
	}
	textStr := string(textByte)
	cs.forbiddenWords = make(map[string]string)
	for _, word := range strings.Split(textStr, "\n") {
		word = strings.TrimSpace(word)
		cs.forbiddenWords[word] = strings.Repeat("*", len(word))
	}

	return &cs, nil
}

func (cs *CensorShip) isForbiddenWords(text string) bool {
	for key, _ := range cs.forbiddenWords {
		if strings.Contains(text, key) {
			return true
		}
	}
	return false
}

func (cs *CensorShip) CheckCensor(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	if cs.isForbiddenWords(comment.Content) {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err := json.NewEncoder(w).Encode(comment); err != nil {
		log.Println("Ошибка кодирования ответа:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
	}
}
