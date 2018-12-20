package:
	@ mkdir -p ./bin
	@ mkdir -p ./downloads
	@ tar -c ./bin > myapps.tar

build-server:
	@ go build -o ./bin/server cmd/server/main.go

build-client:
	@ go build -o ./bin/client cmd/server/main.go
