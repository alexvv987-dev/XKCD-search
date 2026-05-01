# XKCD Search Service

Микросервисная система для поиска комиксов XKCD, объединяющая REST API, gRPC-сервисы и событийную архитектуру. Проект демонстрирует навыки бэкенд-разработки на Go, работы с БД, брокерами сообщений и Docker-контейнеризацией.

## Архитектура
Система построена по принципам микросервисной архитектуры:

- **api:** HTTP REST Gateway (управление поиском, авторизация JWT).
- **words:** gRPC-сервис нормализации слов (стемминг).
- **update:** gRPC-сервис синхронизации данных с xkcd.com.
- **postgres:** Хранилище комиксов.
- **nats:** Брокер сообщений для событий (pub/sub).

## Технологический стек
- **Backend:** Go
- **Communication:** gRPC, REST API
- **Database:** PostgreSQL (pgx/sqlx)
- **Message Broker:** NATS
- **Infrastructure:** Docker, Docker Compose
- **Auth:** JWT (golang-jwt/jwt/v5)
- **Testing:** Unit & Integration tests

## Быстрый старт

### Требования
- Docker & Docker Compose

### Запуск
```bash
# Сборка и запуск системы
make up

# Запуск полного цикла тестирования
make test
```

## API Краткий обзор

| Метод | Эндпоинт | Описание |
| :--- | :--- | :--- |
| `POST` | `/api/login` | Авторизация и получение JWT |
| `GET` | `/api/search` | Поиск комиксов по фразе |
| `POST` | `/api/db/update` | Запуск обновления БД |
| `GET` | `/api/db/stats` | Статистика по комиксам |

## Реализованные навыки
- Разработка микросервисов на Go.
- Проектирование REST API и gRPC контрактов.
- Работа с транзакциями в PostgreSQL и миграциями.
- Реализация Middleware (Auth, Rate Limiting, Concurrency Control).
- Событийная архитектура (NATS).
- CI/CD и контейнеризация инфраструктуры.

## Авторы
[Твое Имя]
