ifneq (,$(wildcard ./.env))
    include .env
    export
endif

.PHONY: test test-integration migrate-up migrate-down migrate-status migrate-create

test:
	go test ./...

test-integration:
	TEST_DATABASE_URL="$(if $(TEST_DATABASE_URL),$(TEST_DATABASE_URL),$(DATABASE_URL))" go test -tags=integration ./...

migrate-up:
	goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir migrations postgres "$(DATABASE_URL)" down

migrate-status:
	goose -dir migrations postgres "$(DATABASE_URL)" status

migrate-create:
	@read -p "Migration name: " name; \
	goose -dir migrations create $$name sql
