# Stage 1: Build frontend
FROM node:22-alpine AS frontend-build
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Stage 2: Build backend
FROM golang:1.25-alpine AS backend-build
WORKDIR /src
COPY backend/go.* ./
RUN go mod download
COPY backend/ .
# Copy frontend dist into backend cmd/server/dist/ for embed
COPY --from=frontend-build /app/dist ./cmd/server/dist
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

# Stage 3: Runtime
FROM alpine:3.20
RUN adduser -D -u 10001 app
COPY --from=backend-build /server /server
RUN mkdir -p /uploads && chown app:app /uploads
WORKDIR /
EXPOSE 8080
USER app
ENTRYPOINT ["/server"]