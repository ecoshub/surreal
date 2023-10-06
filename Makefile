.PHONY: release
release:
	mkdir -p release
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o ./release/windows-amd64/surreal-windows-amd64 ./cmd/main.go
	CGO_ENABLED=1 GOOS=windows GOARCH=386 go build -ldflags "-s -w" -o ./release/windows-386/surreal-windows-386 ./cmd/main.go
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o ./release/linux-amd64/surreal-linux-amd64 ./cmd/main.go
	CGO_ENABLED=1 GOOS=linux GOARCH=386 go build -ldflags "-s -w" -o ./release/linux-386/surreal-linux-386 ./cmd/main.go
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o ./release/macos-amd64/surreal-macos-amd64 ./cmd/main.go
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o ./release/macos-arm/surreal-macos-arm ./cmd/main.go

build:
	go build -o surreal ./cmd/main.go

run:
	go run ./cmd/main.go
