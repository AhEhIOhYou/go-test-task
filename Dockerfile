FROM golang:1.21-alpine as builder

WORKDIR /app
COPY . .
COPY .env ./

RUN go build -o main .

FROM postgres:13

COPY --from=builder /app/main /usr/local/bin/
COPY create_table.sql /docker-entrypoint-initdb.d/

CMD ["main"]
