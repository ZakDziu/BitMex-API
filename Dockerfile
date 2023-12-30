FROM golang:1.21.1

RUN apt-get update && apt-get -y install postgresql-client

COPY ./ /app

WORKDIR /app

RUN chmod +x wait-for-postgres.sh

RUN go mod download
RUN go build -o bitmex-api ./cmd/main.go

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest


CMD ["./wait-for-postgres.sh", "postgres", "/go/bin/migrate -path=db/migrations -database 'postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@$POSTGRES_HOST:$POSTGRES_PORT/$POSTGRES_DB?sslmode=disable' up && ./bitmex-api"]
