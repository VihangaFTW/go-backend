DB_URL = postgres://root:12345@localhost:5432/simple_bank?sslmode=disable

startdb:
	docker run --name postgres  -e POSTGRES_USER=root -e POSTGRES_PASSWORD=12345 -p 5432:5432 -d postgres:latest

createdb:
	docker exec -it postgres createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres dropdb simple_bank

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up 

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down
	
migratedown1:
	migrate -path db/migration -database "$(DB_URL)" -verbose down 1

sqlc:
	sqlc generate

test:
	go test -count=1 -timeout 30s -v -cover ./...

psql:
	doker exec -it simple_bank_db psql -U root -d simple_bank

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/VihangaFTW/Go-Backend/db/sqlc Store

aws-ecr-login:
	aws ecr get-login-password --region $(AWS_REGION) | docker login --username AWS --password-stdin $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com

db_docs:
	dbdocs build doc/db.dbml 

db_schema:
	dbml2sql --postgres -o doc/schema.sql doc/db.dbml

proto:
	rm -f pb/*.go
	rm -f doc/swagger/*.swagger.json
	protoc --grpc-gateway_out=pb --grpc-gateway_opt paths=source_relative --proto_path=proto --go_out=pb --go_opt=paths=source_relative --go-grpc_out=pb --openapiv2_out ./doc/swagger/ --openapiv2_opt=allow_merge=true,merge_file_name=simple_bank,json_names_for_fields=false --go-grpc_opt=paths=source_relative proto/*.proto

evans:
	evans --host localhost --port 9090 -r repl

.PHONY: createdb startdb dropdb migrateup migratedown migratedown1 sqlc test psql server mock aws-ecr-login db_docs db_schema proto evans