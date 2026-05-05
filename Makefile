IMAGE_NAME := "syseleven/designate-certmanager-webhook"
IMAGE_TAG  ?= $(shell git describe --tags --always --dirty)

build:
	docker build -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

check:
	@if test -n "$$(find . -not \( \( -wholename "./vendor" \) -prune \) -name "*.go" | xargs gofmt -l)"; then \
		find . -not \( \( -wholename "./vendor" \) -prune \) -name "*.go" | xargs gofmt -d; \
		exit 1; \
	fi

test:
	echo "OS_AUTH_URL=$$OS_AUTH_URL"
	echo "OS_APPLICATION_CREDENTIAL_ID=$$OS_APPLICATION_CREDENTIAL_ID"
	docker build --file Dockerfile_test . -t $(IMAGE_NAME)-test
	docker run --rm -v $$(pwd):/workspace \
		 -e TEST_ZONE_NAME=$$TEST_ZONE_NAME \
		 -e OS_APPLICATION_CREDENTIAL_ID=$$OS_APPLICATION_CREDENTIAL_ID \
		 -e OS_APPLICATION_CREDENTIAL_SECRET=$$OS_APPLICATION_CREDENTIAL_SECRET \
		 -e OS_AUTH_TYPE=$$OS_AUTH_TYPE \
		 -e OS_AUTH_URL=$$OS_AUTH_URL \
		 -e OS_ENDPOINT_TYPE=$$OS_ENDPOINT_TYPE \
		 -e OS_IDENTITY_API_VERSION=$$OS_IDENTITY_API_VERSION \
		 -e OS_INTERFACE=$$OS_INTERFACE \
		 -e OS_PROJECT_DOMAIN_NAME=$$OS_PROJECT_DOMAIN_NAME \
		 -e OS_REGION_NAME=$$OS_REGION_NAME \
		 -e OS_USER_DOMAIN_NAME=$$OS_USER_DOMAIN_NAME \
	     $(IMAGE_NAME)-test go test -v .

ci-push:
	echo "$$DOCKER_PASSWORD" | docker login -u "$$DOCKER_USERNAME" --password-stdin
	docker push "$(IMAGE_NAME):$(IMAGE_TAG)"
