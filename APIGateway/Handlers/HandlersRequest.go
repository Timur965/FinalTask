package HandlersRequest

import (
	storage "FinalTask/Storage"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
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
	page := r.URL.Query().Get("page")
	response, err := http.Get(ap.NewsServiceURL + "/news?page=" + page)
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
	page := r.URL.Query().Get("page")
	urlAddr := ""

	switch filter {
	case "FullMatchText", "PartialMatchText", "FullMatchHeader", "PartialMatchHeader", "ExcludedPhrases":
		textFilter := url.QueryEscape(r.URL.Query().Get("Text"))
		urlAddr = ap.NewsServiceURL + "/news/filter?filter=" + filter + "&text=" + textFilter + "&page=" + page
	case "SelectionDate":
		date := url.QueryEscape(r.URL.Query().Get("date"))
		urlAddr = ap.NewsServiceURL + "/news/filter?filter=SelectionDate&date=" + date + "&page=" + page
	case "DateRange":
		dateStart := url.QueryEscape(r.URL.Query().Get("dateStart"))
		dateEnd := url.QueryEscape(r.URL.Query().Get("dateEnd"))
		urlAddr = ap.NewsServiceURL + "/news/filter?filter=DateRange&dateStart=" + dateStart + "&dateEnd=" + dateEnd + "&page=" + page
	case "FieldSort":
		field := url.QueryEscape(r.URL.Query().Get("field"))
		urlAddr = ap.NewsServiceURL + "/news/sort?field=" + field + "&page=" + page
	}

	response, err := http.Get(urlAddr)
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
	if idNews == "" {
		http.Error(w, "idNews parameter is required", http.StatusBadRequest)
		return
	}

	type responseResult struct {
		data interface{}
		err  error
	}

	newsURL := ap.NewsServiceURL + "/detailedNews?idNews=" + idNews
	commentsURL := ap.CommentsServiceURL + "/detailedNews?idNews=" + idNews

	results := make(chan responseResult, 2)
	var wg sync.WaitGroup
	wg.Add(2)

	fetchData := func(url string, resultType string) {
		defer wg.Done()
		resp, err := http.Get(url)
		if err != nil {
			results <- responseResult{nil, err}
			return
		}
		defer resp.Body.Close()

		var data interface{}
		if resultType == "news" {
			var oneNews storage.News
			if err := json.NewDecoder(resp.Body).Decode(&oneNews); err != nil {
				results <- responseResult{nil, err}
				return
			}
			data = oneNews
		} else if resultType == "comments" {
			var allComments []storage.Comments
			if err := json.NewDecoder(resp.Body).Decode(&allComments); err != nil {
				results <- responseResult{nil, err}
				return
			}
			data = allComments
		}
		results <- responseResult{data, nil}
	}

	go fetchData(newsURL, "news")
	go fetchData(commentsURL, "comments")

	wg.Wait()
	close(results)

	var detailedResponse storage.DetailedNews
	for result := range results {
		if result.err != nil {
			http.Error(w, result.err.Error(), http.StatusInternalServerError)
			return
		}
		switch data := result.data.(type) {
		case storage.News:
			detailedResponse.OneNews = data
		case []storage.Comments:
			detailedResponse.AllComments = data
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(detailedResponse); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (ap *ApiGateway) HandleAddComments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Ошибка чтения тела запроса:", err)
		http.Error(w, "Ошибка чтения тела запроса", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var comment storage.Comments
	err = json.Unmarshal(body, &comment)
	if err != nil {
		log.Println("Ошибка десериализации JSON:", err)
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}

	if comment.NewsId == 0 {
		log.Println("Ошибка: не указан ID новости")
		http.Error(w, "Некорректный ID новости", http.StatusBadRequest)
		return
	}

	if comment.Content == "" {
		log.Println("Ошибка: не указан текст комментария")
		http.Error(w, "Текст комментария не может быть пустым", http.StatusBadRequest)
		return
	}

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

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		http.Error(w, "Ошибка чтения ответа сервиса", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(response.StatusCode)
	w.Write(responseBody)
}
