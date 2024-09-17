MIGRATIONS_DIR = ./migrations

run:
	@go run ./...

generate:
	@templ generate
	
create-migration:
	@goose -dir ${MIGRATIONS_DIR} create ${NAME} sql

migrate:
	@GOOSE_DRIVER=sqlite3 GOOSE_DBSTRING=${DB_URL} goose -dir ${MIGRATIONS_DIR} up

rollback:
	@GOOSE_DRIVER=sqlite3 GOOSE_DBSTRING=${DB_URL} goose -dir ${MIGRATIONS_DIR} down

sqlc-generate:
	@sqlc generate

.PHONY: run create-migration migrate rollback sqlc-generate