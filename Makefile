.PHONY: vendor

SHELL=/bin/bash # Use bash syntax


vendor:
	go mod vendor

lint:
	golangci-lint run --modules-download-mode vendor -v --max-same-issues 10

dev-image:
	docker build --tag "leslie-wang/clusterd:1.0" -f dockerfiles/dev-image .

dev:
	docker build --tag "leslie-wang/clusterd" -f dockerfiles/dev-run .
	docker rm -f clusterd
	docker run \
		--name clusterd --hostname clusterd \
		--privileged --cap-add=ALL -v /dev:/dev -v /lib/modules:/lib/modules \
		-v "${PWD}:/go/src/github.com/leslie-wang/clusterd" \
		--net host --dns-search local \
		-it "leslie-wang/clusterd" -d bash
