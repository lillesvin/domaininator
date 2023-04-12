BUILD_PATH = build
BINARY_BASENAME = domaininator
VERSION = 0.1.3

all: linux windows darwin

linux: | setup
	GOOS=$@ ARCH=amd64 go build -o $(BUILD_PATH)/$(BINARY_BASENAME)-$(VERSION)-$@_amd64
	GOOS=$@ ARCH=arm64 go build -o $(BUILD_PATH)/$(BINARY_BASENAME)-$(VERSION)-$@_arm64

windows: | setup
	GOOS=$@ ARCH=amd64 go build -o $(BUILD_PATH)/$(BINARY_BASENAME)-$(VERSION)-$@_amd64.exe

darwin: | setup
	GOOS=$@ ARCH=arm64 go build -o $(BUILD_PATH)/$(BINARY_BASENAME)-$(VERSION)-$@_arm64
	GOOS=$@ ARCH=amd64 go build -o $(BUILD_PATH)/$(BINARY_BASENAME)-$(VERSION)-$@_amd64

setup:
	mkdir -p $(BUILD_PATH)

clean:
	go clean
	rm -rf $(BUILD_PATH)

.PHONY: all clean setup
