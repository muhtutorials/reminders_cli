.PHONY: client
.PHONY: server

client:
	@echo "Building the client binary"
	go build -o bin/client.exe cmd/client/main.go

server:
	@echo "Building the server binary"
	go build -o bin/server.exe cmd/server/main.go