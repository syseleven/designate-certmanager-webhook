IMAGE_NAME := "syseleven/designate-certmanager-webhook"
IMAGE_TAG  ?= $(shell git describe --tags --always --dirty)

OUT := $(shell pwd)/_out

$(shell mkdir -p "$(OUT)")

verify:
	go test -v .

build:
	docker build -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

.PHONY: rendered-manifest.yaml
rendered-manifest.yaml:
	helm template \
	    --name webhook \
        --set image.repository=$(IMAGE_NAME) \
        --set image.tag="$(IMAGE_TAG)" \
        deploy/designate-certmanager-webhook > "$(OUT)/rendered-manifest.yaml"
