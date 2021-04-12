package models

import "time"

//Feed Новостной ресурс
type Feed struct {
	Id       int
	Link     string
	Title    string
	Feedlink string
	Items    []Item
}

//Item Новостной пост
type Item struct {
	Id         int
	Resourse   int
	Title      string
	Link       string
	Categories []string
	Updated    time.Time
}

//Входные данные на API, для получения инфы с БД
type Neededtems struct {
	Tags   []string
	Limit  int
	Offset int
}

//Канал для транспорта новых событий от источника к БД
type rWItem struct {
	GetedItem chan Item
}

func NewRWItem() *rWItem {
	return &rWItem{GetedItem: make(chan Item)}
}

func (rwI *rWItem) CloseChan() {
	close(rwI.GetedItem)
}
