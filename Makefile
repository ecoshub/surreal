.PHONY: release
release:
	mkdir -p release
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o ./release/windows-amd64/surreal ./cmd/main.go
	CGO_ENABLED=1 GOOS=windows GOARCH=386 go build -o ./release/windows-386/surreal ./cmd/main.go
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o ./release/linux-amd64/surreal ./cmd/main.go
	CGO_ENABLED=1 GOOS=linux GOARCH=386 go build -o ./release/linux-386/surreal ./cmd/main.go
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o ./release/macos-amd64/surreal ./cmd/main.go
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o ./release/macos-arm/surreal ./cmd/main.go

build:
	go build -o surreal ./cmd/main.go

run:
	go run ./cmd/main.go