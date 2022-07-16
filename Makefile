PACKAGES=$(shell go list ./...)
OUTDIR=bin
GOBUILD=GOOS=$(1) GOARCH=$(2) go build -o $(OUTDIR)/instx-$(1)-$(2)

.PHONY: build
build:
	go build -o $(OUTDIR)/instx

.PHONY: build-linux-amd64
build-linux-amd64:
	 $(call GOBUILD,linux,amd64)

.PHONY: build-linux-386
build-linux-386:
	$(call GOBUILD,linux,386)

.PHONY: build-windows-amd64
build-windows-amd64:
	$(call GOBUILD,windows,amd64)
	mv $(OUTDIR)/instx-windows-amd64 $(OUTDIR)/instx-windows-amd64.exe

.PHONY: build-windows-386
build-windows-386:
	$(call GOBUILD,windows,386)
	mv $(OUTDIR)/instx-windows-386 $(OUTDIR)/instx-windows-386.exe

.PHONY: build-darwin-amd64
build-darwin-amd64:
	$(call GOBUILD,darwin,amd64)

.PHONY: build-darwin-arm64
build-darwin-arm64:
	$(call GOBUILD,darwin,arm64)

.PHONY: build-macos-amd64
build-macos-amd64: build-darwin-amd64
.PHONY: build-macos-arm64
build-macos-arm64: build-darwin-arm64

.PHONY: test
test:
	go test -v -race $(PACKAGES)

.PHONY: all
all: test \
	 build-linux-amd64 build-linux-386 \
	 build-windows-amd64 build-windows-386 \
	 build-darwin-amd64 build-darwin-arm64

.PHONY: clean
clean:
	go clean

default: build