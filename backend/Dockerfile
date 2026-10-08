FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.20
RUN adduser -D -u 10001 app
COPY --from=build /server /server
RUN mkdir -p /web/uploads && chown app:app /web/uploads
WORKDIR /
EXPOSE 8080
USER app
ENTRYPOINT ["/server"]
