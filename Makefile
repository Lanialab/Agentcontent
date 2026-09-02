.PHONY: test tidy ui build run scan seed

test:
	go test ./...

tidy:
	go mod tidy

ui:
	cd web && npm install && npm run build

build: ui
	mkdir -p bin
	go build -o bin/server ./cmd/server
	go build -o bin/scan ./cmd/scan

run:
	go run ./cmd/server

scan:
	go run ./cmd/scan
