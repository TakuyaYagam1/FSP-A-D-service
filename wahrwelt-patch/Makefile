COMPOSE ?= docker compose

.PHONY: build test generate generate-check migrate-up migrate-down \
	images prepare up down compose-up compose-down \
	backend frontend renderer render postgres redis restart-all rebuild-all status logs help

build test generate generate-check migrate-up migrate-down:
	$(MAKE) -C backend $@

images:
	$(COMPOSE) pull postgres redis
	$(COMPOSE) build renderer backend frontend

prepare: images
	$(COMPOSE) up -d --no-build --pull never --wait

up:
	$(COMPOSE) up -d --no-build --pull never --wait

down:
	$(COMPOSE) down

backend:
	$(COMPOSE) build backend
	$(COMPOSE) up -d --no-build --pull never --wait postgres redis renderer
	$(COMPOSE) up -d --no-deps --no-build --pull never --force-recreate --wait backend

frontend:
	$(COMPOSE) build frontend
	$(COMPOSE) up -d --no-build --pull never --wait postgres redis renderer backend
	$(COMPOSE) up -d --no-deps --no-build --pull never --force-recreate --wait frontend

renderer:
	$(COMPOSE) build renderer
	$(COMPOSE) up -d --no-deps --no-build --pull never --force-recreate --wait renderer

render: renderer

postgres:
	$(COMPOSE) pull postgres
	$(COMPOSE) up -d --no-deps --no-build --pull never --force-recreate --wait postgres

redis:
	$(COMPOSE) pull redis
	$(COMPOSE) up -d --no-deps --no-build --pull never --force-recreate --wait redis

restart-all:
	$(COMPOSE) up -d --no-build --pull never --force-recreate --wait

rebuild-all:
	$(COMPOSE) pull postgres redis
	$(COMPOSE) build renderer backend frontend
	$(COMPOSE) up -d --no-build --pull never --force-recreate --wait

status:
	$(COMPOSE) ps

logs:
	$(COMPOSE) logs --tail=100

compose-up: up

compose-down: down

help:
	@printf '%s\n' \
		'prepare      build images and start the service' \
		'up           start existing images' \
		'down         stop the service' \
		'backend      rebuild and restart backend' \
		'frontend     rebuild and restart frontend' \
		'renderer     rebuild and restart renderer' \
		'restart-all  restart all containers' \
		'rebuild-all  rebuild and restart everything' \
		'status       show container status' \
		'logs         show recent logs'
