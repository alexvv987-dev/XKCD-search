# XKCD Search — Web-интерфейс

## Demo

<video src="demo.mp4" controls width="900"></video>

## Описание

Веб-интерфейс для поиска комиксов XKCD.
Реализован как отдельный микросервис на Go с html/template.

### Страницы
- `/` — поиск комиксов по фразе (fulltext и index режимы)
- `/login` — авторизация администратора
- `/admin` — управление базой (update/drop/status)