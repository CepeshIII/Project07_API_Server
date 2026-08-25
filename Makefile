include .env


.PHONY: migrate-create
migration:
	$(if $(name),,$(error Usage: make migration name=your_name))
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(name)

.PHONY: migrate-up
migrate-up:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) up

N ?= 1 

.PHONY: migrate-down
migrate-down:
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_ADDR)" down $(N)


.PHONY: seed
seed:
	@go run  ./cmd/migrate/seed/main.go


.PHONY: gen-docs
gen-docs:
	@swag init -g ./api/main.go -d cmd,internal && swag fmt



.PHONY: fix-dirty
# Extract values using yq paths
DB_NAME := $(shell yq '.services.db.environment.POSTGRES_DB' docker-compose.yml)
DB_USER := $(shell yq '.services.db.environment.POSTGRES_USER' docker-compose.yml)
DB_PASS := $(shell yq '.services.db.environment.POSTGRES_PASSWORD' docker-compose.yml)
HOST_PORT := $(shell yq '.services.db.ports[0]' docker-compose.yml | cut -d':' -f1)

fix-dirty:
	@PGPASSWORD="$(DB_PASS)" psql -h localhost -p $(HOST_PORT) -U $(DB_USER) -d $(DB_NAME) \
		-c "UPDATE schema_migrations SET dirty = false;"