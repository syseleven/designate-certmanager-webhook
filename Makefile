IMAGE_NAME := "syseleven/designate-certmanager-webhook"
IMAGE_TAG  ?= $(shell git describe --tags --always --dirty)

build: check
	docker build -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

check:
	@if test -n "$$(find . -not \( \( -wholename "./vendor" \) -prune \) -name "*.go" | xargs gofmt -l)"; then \
		find . -not \( \( -wholename "./vendor" \) -prune \) -name "*.go" | xargs gofmt -d; \
		exit 1; \
	fi

test:
	docker build --file Dockerfile_test . -t $(IMAGE_NAME)-test
	docker run --rm -v $$(pwd):/workspace \
		 -e TEST_ZONE_NAME=$$TEST_ZONE_NAME \
		 -e OS_TENANT_NAME=$$OS_TENANT_NAME \
		 -e OS_TENANT_ID=$$OS_PROJECT_ID \
		 -e OS_DOMAIN_NAME=$$OS_USER_DOMAIN_NAME \
		 -e OS_USERNAME=$$OS_USERNAME \
		 -e OS_PASSWORD=$$OS_PASSWORD \
		 -e OS_AUTH_URL=$$OS_AUTH_URL \
	     $(IMAGE_NAME)-test go test -v .

ci-push:
	echo "$$DOCKER_PASSWORD" | docker login -u "$$DOCKER_USERNAME" --password-stdin
	docker push "$(IMAGE_NAME):$(IMAGE_TAG)"
