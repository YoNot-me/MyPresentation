include .env
export

export PROJECT_ROOT=$(shell pwd)

compose-up:
	@docker compose up -d postgres-database redis
	@echo "Waiting for PostgreSQL to become ready..."
	@docker compose exec postgres-database sh -c 'until pg_isready -U $$POSTGRES_USER -d $$POSTGRES_DB; do sleep 1; done'

compose-down:
	@docker compose down

clean-env:

	@read -p "Clean all volumes files env? [Y/N]: " answ; \
	if [ "$$answ" = "y" ] || [ "$$answ" = "Y" ]; then \
		docker compose down --rmi local -v --remove-orphans && \
		rm -rf out/pgdata && \
		echo "Success"; \
	else \
		echo "Canceled"; \
	fi

migrate-create:
	@if [ -z "$(seq)" ]; then \
  		echo "WRONG seq = nil, write migration name seq=..."; \
  		exit 1; \
	fi; \
	docker compose run --rm postgres-migrate \
		create \
		-ext sql \
		-dir /migrations \
		-seq "$(seq)"

migrate-up:
	@$(MAKE) migrate-action action=up

migrate-down:
	@$(MAKE) migrate-action action=down

migrate-action:
	@if [ -z "$(action)" ]; then \
			echo "WRONG action = nil, write action=up... or action=down..."; \
			exit 1; \
	fi; \
	docker compose run --rm postgres-migrate \
    		-path /migrations \
    		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@presentation-postgres:5432/${POSTGRES_DB}?sslmode=disable \
    		"$(action)"

run:
	@$(MAKE) compose-up
	@$(MAKE) migrate-up
	@echo "server started localhost:8080"

clean:
	@$(MAKE) compose-clean-env

logs-db:
	docker logs presentation-postgres

run-local:
	@go run ./cmd