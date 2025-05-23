
postgres:
	docker run --name bank_postgres -e POSTGRES_USER=root -e POSTGRES_PASSWORD=12345 -p 5432:5432 -d postgres:latest

createdb:
	docker exec -it bank_postgres createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it bank_postgres dropdb simple_bank

migrateup:
	migrate -path db/migration -database "postgres://root:12345@localhost:5432/simple_bank?sslmode=disable" -verbose up 

migratedown:
	migrate -path db/migration -database "postgres://root:12345@localhost:5432/simple_bank?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test:
	go test -count=1 -timeout 30s -v -cover ./...

psql:
	doker exec -it simple_bank_db psql -U root -d simple_bank

server:
	go run main.go

.phony: createdb postgres dropdb migrateup migratedown sqlc test psql server