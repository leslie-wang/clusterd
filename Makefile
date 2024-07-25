.PHONY: vendor

SHELL=/bin/bash # Use bash syntax
VERSION_COMPILE := -ldflags "-X github.com/leslie-wang/clusterd/common/release.Version=$$(cat LATEST_RELEASE)"

vendor:
	go mod vendor

lint:
	golangci-lint run --modules-download-mode vendor -v --max-same-issues 10

dev-image:
	docker build --tag "qiwang/clusterd:1.0" -f dockerfiles/dev-image .

dev:
	docker build --tag "qiwang/clusterd" -f dockerfiles/dev-run .
	docker rm -f clusterd
	docker run \
		--name clusterd --hostname clusterd \
		--privileged --cap-add=ALL -v /dev:/dev -v /lib/modules:/lib/modules \
		-v "${PWD}:/go/src/github.com/leslie-wang/clusterd" \
		--net host --dns-search local \
		-it "leslie-wang/clusterd" -d bash

release:
	echo "release-$$(date +%m-%d-%y.%H.%M.%S)-$$(git rev-parse --short=8 HEAD)" >LATEST_RELEASE
	go install -v $(VERSION_COMPILE) ./cmd/...

install:
	go install -v ./cmd/...

integration-test-sqlite: install
	go test  -v ./tests/integration-sqlite/
