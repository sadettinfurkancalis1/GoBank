postgres:
	docker run --name go_postgres --restart always -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -p 5432:5432 -d postgres:12-alpine

createdb:
	docker exec -it go_postgres createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it go_postgres dropdb simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down

test:
	C:\Users\sfcalis\ProgramFiles\go\bin\go.exe test -v -cover ./...

go:
	C:\Users\sfcalis\ProgramFiles\go\bin\go $(param)
	
gov2:
	C:\Users\sfcalis\go\bin\ $(param)

sqlc:
	C:\Users\sfcalis\go\bin\sqlc $(param)




# bunlar folder degil, command oldugu icin phony olarak tanimliyoruz sadece komut çalıştırmasını sağlar.
.PHONY: postgres createdb dropdb migrateup migratedown go sqlc test