# Recipes are plain `go` commands so they work under cmd, pwsh, or sh.
LDFLAGS := -s -w

.PHONY: build install clean

# Outputs ime.exe on Windows, ime elsewhere.
build:
	go build -trimpath -ldflags "$(LDFLAGS)" .

# Installs into $(go env GOBIN) or ~/go/bin, which must be on PATH.
install:
	go install -ldflags "$(LDFLAGS)" .

clean:
	go clean -i
