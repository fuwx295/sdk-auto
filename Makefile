.PHONY: build-sdk generate

build-sdk: generate
	go build -o originx-sdk-auto cmd/main.go

generate:
	go mod download
	GO_AUTO_PATH=$$(go list -m -f '{{.Dir}}' "go.opentelemetry.io/auto") && \
	cd $$GO_AUTO_PATH && \
	make generate