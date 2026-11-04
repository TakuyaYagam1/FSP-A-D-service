# Backend

Clean Architecture (Hexagonal Monolith)

## Как проходит запрос

`HTTP router -> v1 handler -> usecase -> outbound port -> adapter`

- **Handler** отвечает за обработку HTTP-запросов и формирование ответов API
- **Usecase** отвечает за бизнес-логику
- **Domain** отвечает за предметную область: модели, правила и ошибки
- **Port** отвечает за интерфейсы хранилищ и внешних сервисов
- **Adapter** отвечает за интеграции с PostgreSQL, Redis, файловой системой и HTTP
