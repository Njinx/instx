VERSION=1.0.1
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
build-windows-amd64: install-go-msi
	$(call GOBUILD,windows,amd64)
	mv $(OUTDIR)/instx-windows-amd64 $(OUTDIR)/instx-windows-amd64.exe
	$(call generate_msi,instx-windows-amd64,64bit,$(VERSION))

.PHONY: build-windows-386
build-windows-386: install-go-msi
	$(call GOBUILD,windows,386)
	mv $(OUTDIR)/instx-windows-386 $(OUTDIR)/instx-windows-386.exe
	$(call generate_msi,instx-windows-386,32bit,$(VERSION))

# generate_msi name 32bit|64bit version
define generate_msi
	cp "wix_configs/windows-$(2)-wix.json" "wix.json"
	go-msi set-guid -f
	go-msi make --msi "bin/$(1).msi" --version $(3)
	rm -f "wix.json"
endef

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

.PHONY: install-go-msi
install-go-msi:
	go get "github.com/mh-cbon/go-msi"
	go install "github.com/mh-cbon/go-msi"

.PHONY: all
all: test \
	 build-linux-amd64 build-linux-386 \
	 build-windows-amd64 build-windows-386 \
	 build-darwin-amd64 build-darwin-arm64

.PHONY: clean
clean:
	go clean

default: build