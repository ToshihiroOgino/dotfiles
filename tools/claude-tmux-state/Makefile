LDFLAGS := -s -w

.PHONY: build test clean

build:
	go build -trimpath -ldflags "$(LDFLAGS)" .

test:
	go test ./...

clean:
	go clean -i
