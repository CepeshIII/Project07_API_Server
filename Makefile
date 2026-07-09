include .env


.PHONY: migrate-create
migration:
	$(if $(name),,$(error Usage: make migration name=your_name))
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(name)

.PHONY: migrate-up
migrate-up:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) up

.PHONY: migrate-down
migrate-down:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) down $(filter-out $@,$(MAKECMDGOALS))


.PHONY: seed
seed:
	@go run  ./cmd/migrate/seed/main.go