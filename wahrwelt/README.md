# Wahrwelt

Wahrwelt - логистическая система контроля экспорта/импорта товаров и услуг.
Действия происходят в научном городе Орле, где строят первый российский узел "Рассвет".
Логистическая служба доставляет для него материалы и оборудование, а Wahrwelt
ведет сведения о поставках и пропусках, лицензии, акты досмотра, договоры,
счета и платежи.

## ВАЖНО

При обновлении зависимостей оставайтесь в пределах текущей мажорной версии.
Используйте только минорные и патч-релизы.
При переходе на минорный релиз убедитесь, что не нарушены контракт API и основные
сценарии сервиса (бизнес логика).

## Учетные записи

Для первого входа подготовлены учетные записи:

- `admin` / `admin`
- `carrier` / `carrier`
- `inspector` / `inspector`

Полноценную RBAC систему внедрять слишком дорого,
поэтому переключайте аккаунты для подтверждения документов.
Иначе сервис бы раздуло еще сильнее :)

## Состав сервиса

- `backend` - Go API, PostgreSQL migrations и Redis sessions
- `frontend` - статические HTML, CSS и JavaScript
- `renderer` - изолированный процесс рендера документов
- `docker-compose.yml` - PostgreSQL, Redis и компоненты Wahrwelt

Все бизнес-ручки находятся под `/api/v1`. Healthcheck доступен по `/healthz`.
Наружу публикуется только frontend на порту `8080`.

## API и генерация

Публичный контракт описан в `backend/api/openapi.yml`. HTTP-код генерируется
через `ogen`, PostgreSQL queries - через `sqlc`, compile-time DI - через Wire.
Goose применяет migrations автоматически при запуске backend.

Следующие файлы и каталоги считаются частью фиксированного контракта сервиса,
изменять их не нужно, считайте, что они уже source of truth:

- `backend/api/openapi.yml`
- `backend/codegen/`
- `backend/db/migrations/`
- `backend/internal/bootstrap/wire.go`
- `backend/go.mod`
- `backend/go.sum`
- `backend/Dockerfile`
- `frontend/Dockerfile`
- `docker-compose.yml`
- `frontend/`

Generated output не редактируется вручную, они генерируются из спеки, *.sql или *.go:

- `backend/internal/adapter/primary/http/api/`
- `backend/internal/adapter/secondary/postgres/sqlc/`
- `backend/internal/bootstrap/wire_gen.go`

## Первый запуск

Билд аймаджей и ап контейнеров:

```sh
make prepare
```

## Обновление во время раунда

Пересобирайте только измененный сервис:

```sh
make backend
make frontend
make renderer
```

Сначала собирается новый аймадж. После успешной сборки compose заменяет только
выбранный контейнер, не перезапуская его зависимости. `make render` является
алиасом для `make renderer`.

Если изменено несколько компонентов:

```sh
make rebuild-all
```

Рестарт контейнеров:

```sh
make restart-all
```

## Диагностика

```sh
make status
make logs
```
