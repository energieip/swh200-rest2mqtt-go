COMPONENT=energieip-swh200-rest2mqtt

BINARIES=bin/$(COMPONENT)-armhf bin/$(COMPONENT)-amd64 bin/new-device-amd64 bin/new-device-armhf

.PHONY: $(BINARIES) clean

bin/$(COMPONENT)-armhf:
	env GOOS=linux GOARCH=arm go build -o $@

bin/$(COMPONENT)-amd64:
	go build -o $@


bin/new-device-amd64:
	go build -o $@ -i ./cmd/new-device

bin/new-device-armhf:
	env GOOS=linux GOARCH=arm go build -o $@ ./cmd/new-device

all: $(BINARIES)

prepare:
	glide install

.ONESHELL:
deb-amd64: bin/$(COMPONENT)-amd64 bin/new-device-amd64
	$(eval VERSION := $(shell cat ./version))
	$(eval ARCH := $(shell echo "amd64"))
	$(eval BUILD_NAME := $(shell echo "$(COMPONENT)-$(VERSION)-$(ARCH)"))
	$(eval BUILD_PATH := $(shell echo "build/$(BUILD_NAME)"))
	make deb VERSION=$(VERSION) BUILD_PATH=$(BUILD_PATH) ARCH=$(ARCH) BUILD_NAME=$(BUILD_NAME)

.ONESHELL:
deb-armhf: bin/$(COMPONENT)-armhf bin/new-device-armhf
	$(eval VERSION := $(shell cat ./version))
	$(eval ARCH := $(shell echo "armhf"))
	$(eval BUILD_NAME := $(shell echo "$(COMPONENT)-$(VERSION)-$(ARCH)"))
	$(eval BUILD_PATH := $(shell echo "build/$(BUILD_NAME)"))
	make deb VERSION=$(VERSION) BUILD_PATH=$(BUILD_PATH) ARCH=$(ARCH) BUILD_NAME=$(BUILD_NAME)

deb:
	mkdir -p $(BUILD_PATH)/usr/local/bin $(BUILD_PATH)/etc/$(COMPONENT) $(BUILD_PATH)/etc/systemd/system
	cp -r ./scripts/DEBIAN $(BUILD_PATH)/
	cp ./scripts/config.json $(BUILD_PATH)/etc/$(COMPONENT)/
	cp ./scripts/*.service $(BUILD_PATH)/etc/systemd/system/
	sed -i "s/amd64/$(ARCH)/g" $(BUILD_PATH)/DEBIAN/control
	sed -i "s/VERSION/$(VERSION)/g" $(BUILD_PATH)/DEBIAN/control
	sed -i "s/COMPONENT/$(COMPONENT)/g" $(BUILD_PATH)/DEBIAN/control
	cp ./scripts/Makefile $(BUILD_PATH)/../
	cp bin/new-device-$(ARCH) $(BUILD_PATH)/usr/local/bin/new-device
	cp bin/new_device.sh $(BUILD_PATH)/usr/local/bin/new_device.sh
	chmod +x $(BUILD_PATH)/usr/local/bin/new_device.sh
	cp bin/$(COMPONENT)-$(ARCH) $(BUILD_PATH)/usr/local/bin/$(COMPONENT)
	make -C build DEB_PACKAGE=$(BUILD_NAME) deb

clean:
	rm -rf bin build

run: bin/$(COMPONENT)-amd64
	./bin/$(COMPONENT)-amd64 -c ./scripts/config.json
