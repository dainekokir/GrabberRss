package store

import (
	"database/sql"
	"fmt"
	"rss-graber/pkg/models"
	"strconv"
	"strings"
)

type DBStore struct {
	DB *sql.DB
}

//GetItems Получение записей из БД
func (store *DBStore) GetItems(tags []string, limit int, offset int) ([]models.Item, error) {

	LimitStr := ""
	Offset := ""
	TagsStr := ""
	Query := `SELECT
					title,
					link,
					updated
				FROM
					public."item"
				#LIKE_TG#
				ORDER BY Updated desc 
				#LIMIT#
				#OFFSET#`

	if limit > 0 {
		LimitStr = " LIMIT " + strconv.Itoa(limit)
	}
	if offset > 0 {
		Offset = " OFFSET " + strconv.Itoa(offset)
	}
	if len(tags) > 0 {
		for i, vl := range tags {
			tags[i] = "'%" + vl + "%'"
		}
		TagsStr = " WHERE title ILIKE ANY (ARRAY[" + strings.Join(tags, ",") + "])"
	}

	Query = strings.ReplaceAll(Query, "#LIKE_TG#", TagsStr)
	Query = strings.ReplaceAll(Query, "#LIMIT#", LimitStr)
	Query = strings.ReplaceAll(Query, "#OFFSET#", Offset)

	rows, err := store.DB.Query(Query)

	defer rows.Close()

	if err != nil {
		return nil, err
	}

	var response []models.Item
	for rows.Next() {
		temp := models.Item{}
		if err := rows.Scan(&temp.Title, &temp.Link, &temp.Updated); err != nil {
			return nil, err
		}
		response = append(response, temp)
	}

	return response, nil
}

func (store *DBStore) Close() {
	store.DB.Close()
}

//Источник перед постановкой на опрос, нужно записать в БД (при его отсутствии) и поучить его ID
func (store *DBStore) AddResourse(url string) (int, error) {
	rows, err := store.DB.Query(`INSERT INTO public."InformationResource" (feedlink,link,title)
								VALUES ($1,'','')
								ON CONFLICT("feedlink") DO UPDATE SET feedlink=EXCLUDED.feedlink
								returning id;`, url)
	defer rows.Close()

	if err != nil {
		return 0, err
	}
	var id int
	for rows.Next() {
		if err = rows.Scan(&id); err != nil {
			return 0, err
		}
	}

	return id, nil
}

// Чтение из канала поступивших новостей и запись в БД
// канал sem с буфером, одномоментно может выполняться до (глубина буфера) горутин
// тем самым при долгом timeout от БД не будет создана лавинная нагрузка
// защита на случай, если в БД нет ограничение пула запросов
func (store *DBStore) AddItems(item <-chan models.Item) {
	var sem = make(chan int, 10)
	for oneFeed := range item {
		sem <- 1

		go func(F models.Item) {
			rows, err := store.DB.Query(`INSERT INTO public."item" (resourse, title, link, categories, updated)
										VALUES ($1, $2, $3, $4, $5)
										ON CONFLICT DO NOTHING;`, F.Resourse, F.Title, F.Link, strings.Join(F.Categories, ","), F.Updated)

			defer rows.Close()
			if err != nil {
				//fmt.Println(err)
			}
			<-sem
		}(oneFeed)
	}
}

func (store *DBStore) CreateDB(BaseName string) error {
	_, err := store.DB.Query(fmt.Sprintf("DROP DATABASE IF EXISTS %s;", BaseName))
	if err != nil {
		return err
	}

	_, err = store.DB.Query(fmt.Sprintf("CREATE DATABASE %s ENCODING = 'UTF8';", BaseName))
	if err != nil {
		return err
	}

	return nil
}

func (store *DBStore) CreateTables() error {
	err := store.DB.Ping()
	if err != nil {
		return err
	}

	_, err = store.DB.Query(`CREATE TABLE IF NOT EXISTS public."InformationResource"
	(
		id serial NOT NULL,
		link text NOT NULL,
		title varchar(500) NOT NULL,
		feedlink text NOT NULL,
		PRIMARY KEY (id),
		CONSTRAINT feedlink UNIQUE (feedlink)
	);
	
	ALTER TABLE public."InformationResource"
		OWNER to postgres;`)

	if err != nil {
		return err
	}

	_, err = store.DB.Query(`CREATE TABLE IF NOT EXISTS public.Item
	(
		id serial NOT NULL,
		resourse integer NOT NULL,
		title varchar(500) NOT NULL,
		link text NOT NULL,
		categories text,
		updated timestamp without time zone,
		PRIMARY KEY (id),
		UNIQUE (link),
		FOREIGN KEY (resourse)
			REFERENCES public."InformationResource" (id) MATCH SIMPLE
			ON UPDATE CASCADE
			ON DELETE CASCADE
	);
	
	ALTER TABLE public.item
		OWNER to postgres;`)

	if err != nil {
		return err
	}

	return nil
}
