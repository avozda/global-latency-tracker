ifneq (,$(wildcard ./.env))
    include .env
    export
endif

.PHONY: migrate-up migrate-down migrate-status migrate-create

migrate-up:
	goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir migrations postgres "$(DATABASE_URL)" down

migrate-status:
	goose -dir migrations postgres "$(DATABASE_URL)" status

migrate-create:
	@read -p "Migration name: " name; \
	goose -dir migrations create $$name sql
