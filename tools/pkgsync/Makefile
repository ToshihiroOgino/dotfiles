# Recipes are plain `go` commands so they work under cmd, pwsh, or sh.
LDFLAGS := -s -w

.PHONY: build test clean

# Outputs pkgsync.exe on Windows, pkgsync elsewhere.
build:
	go build -trimpath -ldflags "$(LDFLAGS)" .

test:
	go test ./...

clean:
	go clean -i
