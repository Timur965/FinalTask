// API приложения GoNews.
package api

import (
	db "FinalTask/NewsService/DB"
	storage "FinalTask/Storage"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

const limit int = 15

type API struct {
	db *db.PostgresNews
	r  *mux.Router
}

func New(db *db.PostgresNews) *API {
	a := API{db: db, r: mux.NewRouter()}
	a.endpoints()
	return &a
}
func (api *API) Router() *mux.Router {
	return api.r
}
func (api *API) endpoints() {
	api.r.HandleFunc("/news", api.getAllNews).Methods(http.MethodGet, http.MethodOptions)
	api.r.HandleFunc("/detailedNews", api.getDetailedNews).Methods(http.MethodGet, http.MethodOptions)
	api.r.HandleFunc("/news/filter", api.getFilterNews).Methods(http.MethodGet, http.MethodOptions)
}

func (api *API) getAllNews(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodOptions {
		return
	}
	news, err := api.db.GetAllNews()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if page < 1 {
		page = 1
	}

	var pageNews []storage.News
	if page*limit > len(news) {
		pageNews = nil
	} else {
		pageNews = news[(page-1)*limit : page*limit]
	}

	numPages := len(news) / limit
	resp := storage.ResponseWithPagiantion{pageNews, storage.Pagination{numPages, page, limit}}

	json.NewEncoder(w).Encode(resp)
}

func (api *API) getDetailedNews(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		return
	}

	idNews, err := strconv.ParseInt(r.URL.Query().Get("idNews"), 10, 32)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	news, err := api.db.GetCurrentNew(int(idNews))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(news)
}

func (api *API) getFilterNews(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}

	filter := r.URL.Query().Get("filter")
	var news []storage.News
	var err error
	var strToint map[string]int

	strToint = make(map[string]int)
	strToint["FullMatchText"] = storage.FullMatchText
	strToint["PartialMatchText"] = storage.PartialMatchText
	strToint["FullMatchHeader"] = storage.FullMatchHeader
	strToint["PartialMatchHeader"] = storage.PartialMatchHeader
	strToint["ExcludedPhrases"] = storage.ExcludedPhrases
	strToint["SelectionDate"] = storage.SelectionDate
	strToint["DateRange"] = storage.DateRange
	strToint["FieldSort"] = storage.FieldSort

	switch filter {
	case "FullMatchText", "PartialMatchText", "FullMatchHeader", "PartialMatchHeader", "ExcludedPhrases":
		if textFilter, ok := strToint[filter]; ok {
			news, err = api.db.GetFilterNews(textFilter, r.URL.Query().Get("text"))
		}
	case "SelectionDate":
		if dateFilter, ok := strToint[filter]; ok {
			news, err = api.db.GetFilterNews(dateFilter, r.URL.Query().Get("date"))
		}
	case "DateRange":
		if dateFilter, ok := strToint[filter]; ok {
			news, err = api.db.GetFilterNews(dateFilter, r.URL.Query().Get("dateStart"), r.URL.Query().Get("dateEnd"))
		}
	case "FieldSort":
		if sortFilter, ok := strToint[filter]; ok {
			news, err = api.db.GetFilterNews(sortFilter, r.URL.Query().Get("field"))
		}
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(news) == 0 {
		http.Error(w, "array empty", http.StatusInternalServerError)
		return
	}

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if page < 1 {
		page = 1
	}

	var pageNews []storage.News
	if page*limit > len(news) {
		pageNews = nil
	} else {
		pageNews = news[(page-1)*limit : page*limit]
	}

	numPages := len(news) / limit
	resp := storage.ResponseWithPagiantion{pageNews, storage.Pagination{numPages, page, limit}}

	json.NewEncoder(w).Encode(resp)
}
