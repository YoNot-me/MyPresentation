# Presentator

Бэкенд-сервис для **приватного хостинга презентаций**. Каждый бренд получает персональный доступ к своим материалам (works), а администратор централизованно управляет брендами и их презентациями.

Написан на **Go** с чистой feature-ориентированной архитектурой, JWT-аутентификацией поверх Redis, хранением данных в PostgreSQL и полной контейнеризацией через Docker Compose.

---

## Возможности

- **Приватные кабинеты брендов.** Бренд входит по логину/паролю и видит только свои презентации.
- **Двухуровневая авторизация.** Разграничение прав `бренд` и `администратор` на уровне middleware.
- **Админ-панель (API).** Создание/удаление брендов, смена паролей, добавление/редактирование/удаление works.
- **Защищённая раздача файлов.** HTML-презентации отдаются по брендам с изоляцией доступа и защитой от path traversal (`filepath.Clean`).
- **JWT-сессии в Redis.** Токены хранятся в Redis, поддерживается logout (инвалидация сессии).
- **Продакшн-практики.** Graceful shutdown, структурное логирование (zap), таймауты запросов, лимит размера тела запроса.

---

## Технологический стек

| Слой | Технология |
|------|-----------|
| Язык | Go 1.26 |
| HTTP-фреймворк | [Gin](https://github.com/gin-gonic/gin) |
| База данных | PostgreSQL 18 (драйвер [pgx/v5](https://github.com/jackc/pgx), пул соединений) |
| Сессии / кэш | Redis 8 |
| Аутентификация | JWT ([golang-jwt/v5](https://github.com/golang-jwt/jwt)) |
| Логирование | [zap](https://github.com/uber-go/zap) |
| Миграции | [golang-migrate](https://github.com/golang-migrate/migrate) |
| Конфигурация | godotenv (`.env`) |
| Оркестрация | Docker Compose + Makefile |

---

## Архитектура

Проект следует **feature-based** структуре: каждая фича самодостаточна и разбита на слои `transport → service → repository`.

```
.
├── cmd/                        # точка входа: инициализация зависимостей и запуск сервера
│   ├── main.go                 # сборка сервисов, graceful shutdown
│   └── init.go                 # загрузка .env, подключение к Postgres/Redis, логгер
├── internal/
│   ├── core/                   # общее ядро приложения
│   │   ├── server/             # инициализация HTTP-сервера и роутинг
│   │   ├── middleware/         # защита маршрутов (бренд/админ), лимит размера тела
│   │   ├── repository/         # открытие пула соединений с БД
│   │   ├── entity/             # сущности, конфиг, маппинг ошибок → HTTP-статусы
│   │   └── logger/             # инициализация zap
│   └── features/               # бизнес-фичи
│       ├── auth/               # вход брендов, JWT-токены, logout
│       │   └── token/          # выпуск/валидация JWT, хранение в Redis
│       ├── admin/              # управление брендами и works
│       └── fileserving/        # выдача списка и раздача файлов презентаций
├── migrations/                 # SQL-миграции (up/down)
├── public/                     # статические ассеты (страницы auth и presentation)
├── docker-compose.yaml
├── Makefile
└── go.mod
```

Каждая фича слоится единообразно:

- **transport** — HTTP-хендлеры (Gin), парсинг запроса, формирование ответа.
- **service** — бизнес-логика.
- **repository** — доступ к данным (SQL через pgx).

Ошибки описаны как sentinel-значения в `entity/err.go`, а функция `FindStatus` централизованно превращает их в корректные HTTP-коды.

---

## Модель данных

```sql
CREATE SCHEMA presentation;

-- бренды (арендаторы) с хэшем пароля
CREATE TABLE brands (
    id       SERIAL       PRIMARY KEY,
    name     VARCHAR(256) UNIQUE NOT NULL,
    password TEXT         NOT NULL
);

-- презентации, привязанные к бренду
CREATE TABLE works (
    id       SERIAL       PRIMARY KEY,
    brand    VARCHAR(256) NOT NULL,
    workName VARCHAR(256) NOT NULL,
    url      TEXT         NOT NULL DEFAULT '',

    CONSTRAINT presentation_work FOREIGN KEY (brand)
        REFERENCES brands(name) ON DELETE CASCADE,
    CONSTRAINT unique_brand_work UNIQUE (brand, workName)
);

CREATE INDEX work_name_idx ON works(workName);
```

Каскадное удаление гарантирует, что при удалении бренда удаляются все его works.

---

## API

### Публичные маршруты

| Метод | Путь | Описание |
|-------|------|----------|
| `GET`  | `/auth` | Страница/эндпоинт входа бренда |
| `POST` | `/auth/check` | Аутентификация бренда, выдача JWT |
| `POST` | `/logout` | Выход, инвалидация сессии |
| `POST` | `/admin/auth` | Аутентификация администратора |

### Маршруты бренда (требуют JWT-токен бренда)

| Метод | Путь | Описание |
|-------|------|----------|
| `GET` | `/works` | Список презентаций текущего бренда |
| `GET` | `/works/serve` | Отдача HTML-оболочки галереи |
| `GET` | `/presentation/:name/*filepath` | Раздача файлов конкретной презентации |

### Маршруты администратора (требуют JWT админа, префикс `/admin`)

| Метод    | Путь | Описание                            |
|----------|------|-------------------------------------|
| `GET`    | `/admin/brands` | Список всех брендов                 |
| `POST`   | `/admin/brands/add` | Создать бренд                       |
| `PUT`    | `/:brandName/rename` | Переименовать бренд                 |
| `DELETE` | `/admin/:brandName` | Удалить бренд                       |
| `PUT`    | `/admin/:brandName/password` | Сменить пароль бренда               |
| `GET`    | `/admin/:brandName/works` | Список works бренда                 |
| `POST`   | `/admin/:brandName/:workName/add` | Добавить work                       |
| `DELETE` | `/admin/:brandName/remove/:workName` | Удалить work                        |
| `PUT`    | `/admin/:brandName/:workName/change` | Изменить поля work                  |
| `GET`    | `/admin/:brandName/serve/:workName/*filepath` | Раздача файлов work от имени админа |

---

## Конфигурация

Приложение читает настройки из `.env` в корне проекта.

| Переменная | Назначение | Пример |
|------------|-----------|--------|
| `PRES_ADDR` | Адрес прослушивания HTTP-сервера | `:8080` |
| `PRES_DATABASE_URL` | Строка подключения к PostgreSQL | `postgres://user:pass@localhost:5433/presentation?sslmode=disable` |
| `PRES_JWT_SECRET` | Секретный ключ для подписи JWT | `super-secret-key` |
| `PRES_REQ_TIMEOUT` | Таймаут запроса | `180s` |
| `POSTGRES_USER` | Пользователь БД (для Docker/миграций) | `postgres` |
| `POSTGRES_PASSWORD` | Пароль БД | `postgres` |
| `POSTGRES_DB` | Имя БД | `presentation` |
| `REDIS_PASSWORD` | Имя БД | `my_pres_redis123` |

> Redis по умолчанию поднимается на `localhost:6379`.

---

## Запуск

### Предварительные требования

- Docker и Docker Compose
- Go 1.26+ (для локального запуска приложения)

### Через Docker Compose (инфраструктура + миграции)

```bash
# поднять Postgres, Redis, port-forwarder и применить миграции
make run
```

`make run` последовательно:
1. поднимает контейнеры `postgres-database` и `redis` и ждёт готовности БД;
2. запускает `port-forwarder` (socat) на `127.0.0.1:5433 → postgres:5432`;
3. применяет миграции (`migrate-up`).

Затем запустите само приложение:

```bash
make run-local   # go run ./cmd
```

Сервер стартует на `localhost:8080`.

### Полезные команды Makefile

| Команда | Действие |
|---------|----------|
| `make compose-up` | Поднять Postgres и Redis |
| `make compose-down` | Остановить контейнеры |
| `make migrate-up` / `make migrate-down` | Применить / откатить миграции |
| `make migrate-create seq=<name>` | Создать новую миграцию |
| `make run-local` | Запустить приложение локально |
| `make clean-env` | Удалить контейнеры, образы и данные БД |
| `make logs-db` | Логи контейнера PostgreSQL |

---

## Особенности реализации

- **Graceful shutdown.** Сервер слушает `SIGINT`/`SIGTERM`, корректно закрывает HTTP, пул БД и Redis с таймаутом 10 с.
- **Централизованная обработка ошибок.** Sentinel-ошибки → HTTP-статусы через `entity.FindStatus`, единый формат ответа `entity.Response`.
- **Изоляция инфраструктуры.** Доступ к Postgres проброшен через socat-порт-форвардер, что упрощает локальную разработку и деплой.
- **Единая точка сборки зависимостей.** Все сервисы конструируются в `cmd/main.go` через явный DI (без глобального состояния).

---

## Лицензия

См. файл [LICENSE](./LICENSE).
