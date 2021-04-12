package rss

import (
	"context"
	"errors"
	"rss-graber/pkg/models"
	"sync"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/mmcdole/gofeed"
)

const (
	errorURLWasAdded  = "Array contains element"
	errorWrongURL     = "Added wrong URL"
	errorNotFoundElem = "Not found element"
)

type rssObj struct {
	mu       sync.Mutex
	UserUrls map[string]int
}

func NewRss() *rssObj {
	return &rssObj{
		mu:       sync.Mutex{},
		UserUrls: map[string]int{},
	}
}

//AddNewUrl Добавление нового источника данных при его уникальности
func (vl *rssObj) AddNewUrl(url string) (int, error) {
	vl.mu.Lock()
	defer vl.mu.Unlock()

	if validURL := govalidator.IsURL(url); !validURL {
		return 0, errors.New(errorWrongURL)
	}

	_, exists := vl.UserUrls[url]

	if !exists {
		newId := len(vl.UserUrls)
		newId++
		vl.UserUrls[url] = newId
		return vl.UserUrls[url], nil
	}

	return 0, errors.New(errorURLWasAdded)
}

//LoadItems Запуск потока опроса нового ресурса
func (vl *rssObj) LoadItems(url string, id int, CreatedItem chan<- models.Item) error {
	vl.mu.Lock()
	_, exists := vl.UserUrls[url]
	vl.mu.Unlock()

	if exists {
		go readResurses(url, id, 10*time.Second, CreatedItem)
		return nil
	}
	return errors.New(errorNotFoundElem)
}

//Вычитка и парсинг RSS по URL с передачей данных в канал
func readResurses(res_address string, id_resource int, delay time.Duration, CreatedItem chan<- models.Item) {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

		defer cancel()

		fp := gofeed.NewParser()
		feed, err := fp.ParseURLWithContext(res_address, ctx)
		if err != nil {
			break
		}

		var ItemTime time.Time

		for _, vl := range feed.Items {

			if vl.UpdatedParsed != nil {
				ItemTime = *vl.UpdatedParsed
			} else if vl.PublishedParsed != nil {
				ItemTime = *vl.PublishedParsed
			} else {
				continue
			}

			CreatedItem <- models.Item{
				Resourse: id_resource,
				Title:    vl.Title,
				Link:     vl.Link,
				Updated:  ItemTime,
			}
		}

		//Не устраиваем ddos ресурса
		time.Sleep(delay)
	}
}
