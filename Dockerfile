FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.20
RUN adduser -D -u 10001 app
COPY --from=build /server /server
COPY web/ /web/
WORKDIR /
ENV APP_PORT=8080
# DATABASE_URL wajib: menunjuk ke container Postgres terpisah (postgres:16)
EXPOSE 8080
USER app
ENTRYPOINT ["/server"]
