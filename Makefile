BUILD_PATH = build
BINARY_BASENAME = domaininator
VERSION = 0.2.0

all: linux windows darwin

linux: | setup
	GOOS=$@ GOARCH=amd64 go build -o $(BUILD_PATH)/$(BINARY_BASENAME)-$(VERSION)-$@_amd64
	GOOS=$@ GOARCH=arm64 go build -o $(BUILD_PATH)/$(BINARY_BASENAME)-$(VERSION)-$@_arm64

windows: | setup
	GOOS=$@ GOARCH=amd64 go build -o $(BUILD_PATH)/$(BINARY_BASENAME)-$(VERSION)-$@_amd64.exe

darwin: | setup
	GOOS=$@ GOARCH=arm64 go build -o $(BUILD_PATH)/$(BINARY_BASENAME)-$(VERSION)-$@_arm64
	GOOS=$@ GOARCH=amd64 go build -o $(BUILD_PATH)/$(BINARY_BASENAME)-$(VERSION)-$@_amd64

setup:
	mkdir -p $(BUILD_PATH)

clean:
	go clean
	rm -rf $(BUILD_PATH)

.PHONY: all clean setup
