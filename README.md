# RSS Grabber

1. Запуск PostgreSql из Docker:

**docker run --name some-postgres -e POSTGRES_PASSWORD=mysecretpassword -d -p 5432:5432 postgres**

2. Запуск проекта:

**ВАЖНО:** При каждом старте база пересоздается, создано для примера разворачивания, на проде использовались бы миграции

3. Адреса для проверки:
   - http://feeds.twit.tv/twit.xml
   - https://www.reddit.com/.rss

4. Команды через API:
- Добавить источник данных

http://127.0.0.1:1300/api/v1/AddUrl

[POST] x-www-uri

| Key| Values|
| :------------ | :------------ |
|Url| http://feeds.twit.tv/twit.xml |

- Получить данные из БД

http://127.0.0.1:1300/api/v1/GetItems

[GET]

| Key| Values|
| :------------ | :------------ |
|Limit| 9 |
|Offset|0|
|Tags|Zoo,8|

>Пример: http://127.0.0.1:1300/api/v1/GetItems?Limit=0&Offset=0&Tags=Zoo,8
>

