package api

import (
	"encoding/json"
	"net/http"
	"rss-graber/pkg/models"
	"strconv"
	"strings"

	"github.com/gorilla/schema"
)

type dataSource interface {
	AddNewUrl(string) (int, error)
	LoadItems(string, int, chan<- models.Item) error
}

type dataStore interface {
	AddResourse(string) (int, error)
	GetItems([]string, int, int) ([]models.Item, error)
}

//DefRealization
type DefRealization struct {
	DataHighway chan<- models.Item
	Decoder     *schema.Decoder
	Source      dataSource
	Store       dataStore
}

//Получение диапазона новостей по тегам, данные отсортированы по актуальности даты
func (server *DefRealization) GetItems(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	Tg := r.FormValue("Tags")
	Lim, err := strconv.Atoi(r.FormValue("Limit"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	Offs, err := strconv.Atoi(r.FormValue("Offset"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	response, err := server.Store.GetItems(strings.Split(Tg, ","), Lim, Offs)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	js, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error, " + err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(js))
	return
}

//Добавление источника данных на опрос
func (server *DefRealization) AddUrlRss(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	var newUrl struct{ Url string }
	err = server.Decoder.Decode(&newUrl, r.PostForm)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	//Добавляем источник данных на опрос
	_, err = server.Source.AddNewUrl(newUrl.Url)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	id, err := server.Store.AddResourse(newUrl.Url)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	//fmt.Println(id)
	//Источник добавлен, устанавливаем его на опрос
	err = server.Source.LoadItems(newUrl.Url, id, server.DataHighway)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Was added"))
	return
}
