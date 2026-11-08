COMPOSE ?= docker compose

.PHONY: build install-deps generate generate-check migrate-up migrate-down \
	images prepare up down \
	backend frontend renderer render postgres redis restart-all rebuild-all status logs help

build: ## Собрать Go backend на хосте
	$(MAKE) -C backend $@

install-deps: ## Установить Go-зависимости и инструменты кодогенерации для локальной правки
	$(MAKE) -C backend $@

generate: ## Перегенерировать API, SQL и код сборки зависимостей
	$(MAKE) -C backend $@

generate-check: ## Проверить актуальность сгенерированного кода
	$(MAKE) -C backend $@

migrate-up: ## Применить миграции PostgreSQL
	$(MAKE) -C backend $@

migrate-down: ## Откатить миграции PostgreSQL
	$(MAKE) -C backend $@

images: ## Скачать PostgreSQL и Redis, собрать образы сервиса
	$(COMPOSE) pull postgres redis
	$(COMPOSE) build renderer backend frontend

prepare: images ## Собрать образы и запустить весь сервис
	$(COMPOSE) up -d --no-build --pull never --wait

up: ## Запустить сервис из готовых образов
	$(COMPOSE) up -d --no-build --pull never --wait

down: ## Остановить сервис
	$(COMPOSE) down

backend: ## Пересобрать и перезапустить backend
	$(COMPOSE) build backend
	$(COMPOSE) up -d --no-build --pull never --wait postgres redis renderer
	$(COMPOSE) up -d --no-deps --no-build --pull never --force-recreate --wait backend

frontend: ## Пересобрать и перезапустить frontend
	$(COMPOSE) build frontend
	$(COMPOSE) up -d --no-build --pull never --wait postgres redis renderer backend
	$(COMPOSE) up -d --no-deps --no-build --pull never --force-recreate --wait frontend

renderer: ## Пересобрать и перезапустить renderer
	$(COMPOSE) build renderer
	$(COMPOSE) up -d --no-deps --no-build --pull never --force-recreate --wait renderer

render: renderer ## Короткое имя для renderer

postgres: ## Обновить и перезапустить PostgreSQL
	$(COMPOSE) pull postgres
	$(COMPOSE) up -d --no-deps --no-build --pull never --force-recreate --wait postgres

redis: ## Обновить и перезапустить Redis
	$(COMPOSE) pull redis
	$(COMPOSE) up -d --no-deps --no-build --pull never --force-recreate --wait redis

restart-all: ## Перезапустить все контейнеры без сборки
	$(COMPOSE) up -d --no-build --pull never --force-recreate --wait

rebuild-all: ## Пересобрать и перезапустить все контейнеры
	$(COMPOSE) pull postgres redis
	$(COMPOSE) build renderer backend frontend
	$(COMPOSE) up -d --no-build --pull never --force-recreate --wait

status: ## Показать состояние контейнеров
	$(COMPOSE) ps

logs: ## Показать последние строки логов
	$(COMPOSE) logs --tail=100

help: ## Показать описание команд
	@awk '/^[a-z][a-z0-9-]*:.*##[[:space:]]*/ { target = $$0; sub(/:.*/, "", target); description = $$0; sub(/^.*##[[:space:]]*/, "", description); printf "%-16s %s\n", target, description }' $(MAKEFILE_LIST)

