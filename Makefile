startdb:
	docker run --name bank_postgres --network bank-network -e POSTGRES_USER=root -e POSTGRES_PASSWORD=12345 -p 5432:5432 -d postgres:latest

createdb:
	docker exec -it bank_postgres createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it bank_postgres dropdb simple_bank

migrateup:
	migrate -path db/migration -database "postgres://root:vihanga123@go-backend-db.cnouamy08kar.ap-southeast-2.rds.amazonaws.com:5432/go-backend-db" -verbose up 

migratedown:
	migrate -path db/migration -database "postgres://root:vihanga123@go-backend-db.cnouamy08kar.ap-southeast-2.rds.amazonaws.com:5432/go-backend-db" -verbose down
	
migratedown1:
	migrate -path db/migration -database "postgres://root:vihanga123@go-backend-db.cnouamy08kar.ap-southeast-2.rds.amazonaws.com:5432/go-backend-db" -verbose down 1

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

.phony: createdb startdb dropdb migrateup migratedown migratedown1 sqlc test psql server mock aws-ecr-login;