# Go parameters
GOCMD = go
BUILD = $(GOCMD) build
CLEAN = $(GOCMD) clean
TEST = $(GOCMD) test
GET = $(GOCMD) get
BINARY_NAME = lig
ARCH = amd64
TARGET = target

build:
	@$(BUILD) -o $(TARGET)/$(BINARY_NAME) -v
	@chmod +x $(TARGET)/$(BINARY_NAME)

run: build
	@./$(TARGET)/$(BINARY_NAME) $(args)

test:
	@$(TEST) -v

install: build-linux
	@upx -9 $(TARGET)/$(BINARY_NAME)-linux-$(ARCH)
	@sudo cp $(TARGET)/$(BINARY_NAME)-linux-$(ARCH) /usr/local/bin/$(BINARY_NAME)

clean:
	@$(CLEAN)
	@rm -rf $(TARGET)

deps:
	@$(GET) github.com/markbates/goth
	@$(GET) github.com/markbates/pop

build-linux:
	@CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) $(BUILD) -ldflags '-s -w --extldflags "-static -fpic"' -o $(TARGET)/$(BINARY_NAME)-linux-$(ARCH) -v

build-windows:
	@CGO_ENABLED=0 GOOS=windows GOARCH=$(ARCH) $(BUILD) -ldflags '-s -w --extldflags "-static -fpic"' -o $(TARGET)/$(BINARY_NAME)-windows-$(ARCH) -v

build-darwin:
	@CGO_ENABLED=0 GOOS=darwin GOARCH=$(ARCH) $(BUILD) -ldflags '-s -w --extldflags "-static -fpic"' -o $(TARGET)/$(BINARY_NAME)-darwin-$(ARCH) -v

docker-build:
	@docker run --rm -it -v "$(GOPATH)":/go -w /go/src/bitbucket.org/rsohlich/makepost golang:latest go build -o "$(BINARY_NAME)" -v

all: build-linux build-windows build-darwin
