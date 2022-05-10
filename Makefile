IMG_TAG ?= latest
OS      ?= linux
ARCH    ?= amd64

generate:
	go generate ./pkg/apis/...

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy

unit-test:
	go test -v -coverpkg=./... -coverprofile=/tmp/vela-prism-coverage.txt ./...

helm-charts:
	cp README.md ./charts/

reviewable: generate fmt vet helm-charts

image-apiserver:
	docker build -t oamdev/vela-prism:${IMG_TAG} \
		--build-arg GOPROXY=https://proxy.golang.org \
		--build-arg OS=${OS} \
		--build-arg ARCH=${ARCH} \
		-f cmd/apiserver/Dockerfile \
		.