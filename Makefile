BUILD_DIR := out
GATEWAY_IMAGE := r2d4/mocker

IMAGE := r2d4/mockerdoby

out/mocker: $(BUILD_DIR)
	CGO_ENABLED=0 go build -o $(BUILD_DIR)/mocker --ldflags '-extldflags "-static"' github.com/r2d4/mockerfile/cmd/mocker

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

.PHONY: build
build:
	sudo buildctl build --frontend=gateway.v0 --frontend-opt=source=$(GATEWAY_IMAGE) --local dockerfile=. --exporter=docker --exporter-opt name=$(IMAGE) | docker load

.PHONY: image
image:
	docker build . -t $(GATEWAY_IMAGE) && docker push $(GATEWAY_IMAGE)

.PHONY: shell
shell:
	docker run -it $(IMAGE) bash

.PHONY: run
run:
	