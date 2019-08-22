IMAGE_NAME := "syseleven/designate-certmanager-webhook"
IMAGE_TAG  ?= $(shell git describe --tags --always --dirty)

OUT := $(shell pwd)/_out

$(shell mkdir -p "$(OUT)")

build: check
	docker build -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

check:
	@if test -n "$$(find . -not \( \( -wholename "./vendor" \) -prune \) -name "*.go" | xargs gofmt -l)"; then \
		find . -not \( \( -wholename "./vendor" \) -prune \) -name "*.go" | xargs gofmt -d; \
		exit 1; \
	fi

test:
	go test -v .

ci-push:
	echo "$$DOCKER_PASSWORD" | docker login -u "$$DOCKER_USERNAME" --password-stdin
	docker push "$(IMAGE_NAME):$(IMAGE_TAG)"
