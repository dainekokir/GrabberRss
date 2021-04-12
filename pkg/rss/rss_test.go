package rss

import (
	"errors"
	"rss-graber/pkg/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRss(t *testing.T) {
	if vl := NewRss(); vl == nil {
		t.Error("Not created rss")
	}
}

func TestAddNewUrl(t *testing.T) {
	var tests = []struct {
		url    string
		id     int
		errVal error
	}{
		{"https://www.reddit.com/.rss", 1, nil},
		{"https://www.reddit.com/.rss", 0, errors.New(errorURLWasAdded)},
		{"https://www.reddit.com/.rss", 0, errors.New(errorURLWasAdded)},
		{"http:feeds.twit.tv/twit.xml", 0, errors.New(errorWrongURL)},
		{"http://feeds.twit.tv/twit.xml", 2, nil},
		{"http://feeds.twit.tv/twit.xml", 0, errors.New(errorURLWasAdded)},
	}

	vl := NewRss()
	if vl == nil {
		t.Error("Not created rss")
	}

	assert := assert.New(t)
	for _, test := range tests {
		id, err := vl.AddNewUrl(test.url)
		assert.Equal(id, test.id)
		assert.Equal(err, test.errVal)
	}
}

func TestLoadItems(t *testing.T) {
	var tests = []struct {
		url    string
		id     int
		errVal error
	}{
		{"https://www.reddit.com/.rss", 1, nil},
		{"http://feeds.twit.tv/twit.xml", 2, nil},
	}

	vl := NewRss()

	assert := assert.New(t)
	for _, test := range tests {
		vl.AddNewUrl(test.url)
	}

	tests = append(tests, struct {
		url    string
		id     int
		errVal error
	}{"https:/ddit.com/.rss", 1, errors.New(errorNotFoundElem)})

	CreatedItem := make(chan models.Item)
	for _, test := range tests {
		err := vl.LoadItems(test.url, test.id, CreatedItem)
		assert.Equal(err, test.errVal)
	}

	time.Sleep(time.Second * 2)
	select {
	case tmp := <-CreatedItem:
		assert.NotNil(tmp)
	default:
		t.Error("Not have data from resource")
	}
}
func TestReadResurses(t *testing.T) {

}
