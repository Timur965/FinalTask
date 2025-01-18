package HandlersRequest

import (
	storage "FinalTask/Storage"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type ApiGateway struct {
	NewsServiceURL     string
	CommentsServiceURL string
}

func NewApiGateway(newsServiceURL, commentsServiceURL string) *ApiGateway {
	return &ApiGateway{
		NewsServiceURL:     newsServiceURL,
		CommentsServiceURL: commentsServiceURL,
	}
}

func (ap *ApiGateway) HandleGetAllNews(w http.ResponseWriter, r *http.Request) {
	response, err := http.Get("http://" + r.URL.Query().Get("serviceURL") + "/news")
	if err != nil {
		log.Println("Ошибка при вызове сервиса новостей:", err)
		http.Error(w, "Ошибка сервиса", http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		http.Error(w, "Ошибка чтения ответа сервиса", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(response.StatusCode)
	w.Write(body)
}

func (ap *ApiGateway) HandleFilterNews(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("Filter")
	url := ""

	switch filter {
	case "FullMatchText", "PartialMatchText", "FullMatchHeader", "PartialMatchHeader", "ExcludedPhrases":
		textFilter := r.URL.Query().Get("Text")
		url = ap.NewsServiceURL + "/news/filter?filter=" + filter + "&text=" + textFilter
	case "SelectionDate":
		date := r.URL.Query().Get("date")
		url = ap.NewsServiceURL + "/news/filter?filter=SelectionDate&date=" + date
	case "DateRange":
		dateStart := r.URL.Query().Get("dateStart")
		dateEnd := r.URL.Query().Get("dateEnd")
		url = ap.NewsServiceURL + "/news/filter?filter=DateRange&dateStart=" + dateStart + "&dateEnd=" + dateEnd
	case "FieldSort":
		field := r.URL.Query().Get("field")
		url = ap.NewsServiceURL + "/news/sort?field=" + field
	}

	response, err := http.Get(url)
	if err != nil {
		log.Println("Ошибка при вызове фильтрации новостей:", err)
		http.Error(w, "Ошибка сервиса", http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		http.Error(w, "Ошибка чтения ответа сервиса", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(response.StatusCode)
	w.Write(body)
}

func (ap *ApiGateway) HandleDetiledNews(w http.ResponseWriter, r *http.Request) {
	idNews := r.URL.Query().Get("idNews")
	url := ap.CommentsServiceURL + "/detailedNews?idNews=" + idNews

	response, err := http.Get(url)
	if err != nil {
		log.Println("Ошибка при вызове детальной информации новости:", err)
		http.Error(w, "Ошибка сервиса комментариев", http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		http.Error(w, "Ошибка чтения ответа сервиса", http.StatusInternalServerError)
		return
	}

	url = ap.NewsServiceURL + "/detailedNews?idNews=" + idNews

	response, err = http.Get(url)
	if err != nil {
		log.Println("Ошибка при вызове детальной информации новости:", err)
		http.Error(w, "Ошибка сервиса новостей", http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		http.Error(w, "Ошибка чтения ответа сервиса", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(response.StatusCode)
	w.Write(body)
}

func (ap *ApiGateway) HandleAddComments(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("idNews"))
	if err != nil {
		log.Println("Ошибка получения ID новости:", err)
		http.Error(w, "Некорректный ID новости", http.StatusBadRequest)
		return
	}

	var comment storage.Comments
	comment.NewsId = id
	comment.Content = r.URL.Query().Get("content")
	comment.CreatedAt = time.Now().Unix()

	commentJSON, err := json.Marshal(comment)
	if err != nil {
		log.Println("Ошибка сериализации комментария:", err)
		http.Error(w, "Ошибка обработки комментария", http.StatusInternalServerError)
		return
	}

	url := ap.CommentsServiceURL + "/addComment"
	response, err := http.Post(url, "application/json", bytes.NewBuffer(commentJSON))
	if err != nil {
		log.Println("Ошибка при отправке комментария:", err)
		http.Error(w, "Ошибка сервиса", http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		http.Error(w, "Ошибка чтения ответа сервиса", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(response.StatusCode)
	w.Write(body)
}
