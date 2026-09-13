# Multi-stage: Vite UI → Go server → single process serving both.
# Listens on $PORT (default 8080). SQLite lives at /data/agentcontent.db.

FROM node:20-alpine AS web
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata \
	&& mkdir -p /data /app/web/dist
WORKDIR /app
COPY --from=build /out/server /app/server
COPY --from=web /web/dist /app/web/dist

ENV PORT=8080 \
	DATABASE_PATH=/data/agentcontent.db \
	UI_DIR=/app/web/dist \
	SEED_ON_EMPTY=true

EXPOSE 8080
VOLUME ["/data"]

CMD ["/app/server"]
