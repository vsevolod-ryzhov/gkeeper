# Password storage app

## TODO

### server
[+] user registration, authentication, and authorization;
[+-] storage of private data;
[?] data synchronization between multiple authorized clients of a single owner;
[+] transfer of private data to the owner upon request.

### client
[+] user authentication and authorization on the remote server;
[+] access to private data upon request.

### general
[+-] creation, editing, and deletion of data on the server or client side;
[+] new user registration form;
[+] selection of storage location and data storage format;
[+] ensuring security of data transfer and storage;
[+] client-server interaction protocol;
[+] user authentication and information access authorization mechanisms.

### other
[+] TUI
[-] Tests >= 70%


## Useful commands
- docker-compose up -d
- docker-compose down
- cd api/proto && protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --go_opt=default_api_level=API_OPAQUE gkeeper.proto && cd ../..
- migrate create -ext sql -dir ./migrations -seq <create_tableName_table>
- migrate -database "postgres://postgres_user:postgres_password@localhost:5432/postgres_db?sslmode=disable" -path ./migrations up
- go run cmd/server/main.go -d="host=localhost user=postgres_user password=postgres_password dbname=postgres_db sslmode=disable" -a="localhost:8000"

## testing
- go test ./... -coverprofile cover.out
- go tool cover -html=cover.out
- go test $(go list ./... | grep -v -E '/api/proto|/model$|/cmd/|/tui') -coverprofile=coverage.out && go tool cover -func=coverage.out | grep total
- go test ./internal/storage/ -tags=integration -v