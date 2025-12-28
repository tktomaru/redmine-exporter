.PHONY: build test clean install run

BINARY_NAME=redmine-exporter
INSTALL_PATH=/usr/local/bin

build:
	go build -o bin/$(BINARY_NAME) cmd/redmine-exporter/main.go

build-all:
	GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME)-linux-amd64 cmd/redmine-exporter/main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/$(BINARY_NAME)-darwin-amd64 cmd/redmine-exporter/main.go
	GOOS=windows GOARCH=amd64 go build -o bin/$(BINARY_NAME)-windows-amd64.exe cmd/redmine-exporter/main.go

test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

clean:
	rm -rf bin/
	rm -f coverage.out

install: build
	cp bin/$(BINARY_NAME) $(INSTALL_PATH)/

run: build
	./bin/$(BINARY_NAME) -c ../redmine.config -o output.xlsx
