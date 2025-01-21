# Используем bin в текущей директории для установки плагинов protoc
LOCAL_BIN:=$(CURDIR)/bin

DOCKER_DIR=${CURDIR}/build/dev
DOCKER_YML=${DOCKER_DIR}/docker-compose.yml
ENV_NAME="rd_hub_go"

.PHONY: compose-up
compose-up:
	docker-compose -p ${ENV_NAME} -f ${DOCKER_YML} up -d

.PHONY: compose-down
compose-down: ## terminate local env
	docker-compose -p ${ENV_NAME} -f ${DOCKER_YML} stop

.PHONY: compose-rm
compose-rm: ## remove local env
	docker-compose -p ${ENV_NAME} -f ${DOCKER_YML} rm -fvs

.PHONY: compose-rs
compose-rs: ## remove previously and start new local env
	make compose-rm
	make compose-up

# Скачиваем зависимости
.PHONY: .bin-deps
.bin-deps:
	$(info Installing binary dependencies...)

	GOBIN=$(LOCAL_BIN) go install github.com/pressly/goose/v3/cmd/goose@latest
	GOBIN=$(LOCAL_BIN) go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	GOBIN=$(LOCAL_BIN) go install github.com/gojuno/minimock/v3/cmd/minimock@latest

.PHONY: .sqlc-generate
.sqlc-generate:
	./bin/sqlc -f ./sqlc/sqlc.json generate

.PHONY: .build
.build:
	go build -o bin/rd_hub ./cmd/rd_hub/

.PHONY: .run
.run:
	go run ./cmd/rd_hub/

.PHONY: .run-web
.run-web:
	cd web && npm run dev

.PHONY: .add-migration
.add-migration:
	cd migrations && ../bin/goose create $(MIGRATION_NAME) sql
