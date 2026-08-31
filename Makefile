IMG_TAG ?= latest
OS      ?= linux
ARCH    ?= amd64

generate:
	go generate ./pkg/apis/...

.PHONY: generate-openapi
generate-openapi:
	go install k8s.io/kube-openapi/cmd/openapi-gen@v0.0.0-20260721132016-d427ff9ee9ad
	$(shell go env GOPATH)/bin/openapi-gen \
	--output-pkg "generated" \
	--output-file zz_generated.openapi.go \
	--output-dir ./pkg/apis/generated \
	--go-header-file ./hack/boilerplate.go.txt \
	./pkg/apis/cluster/v1alpha1 \
	./pkg/apis/applicationresourcetracker/v1alpha1 \
	./pkg/apis/o11y/grafana/v1alpha1 \
	./pkg/apis/o11y/grafanadashboard/v1alpha1 \
	./pkg/apis/o11y/grafanadatasource/v1alpha1 \
	k8s.io/apimachinery/pkg/api/resource \
	k8s.io/apimachinery/pkg/apis/meta/v1 \
	k8s.io/apimachinery/pkg/runtime \
	k8s.io/apimachinery/pkg/version

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

reviewable: generate generate-openapi fmt vet helm-charts

image-apiserver:
	docker build -t oamdev/vela-prism:${IMG_TAG} \
		--build-arg GOPROXY=https://proxy.golang.org \
		--build-arg OS=${OS} \
		--build-arg ARCH=${ARCH} \
		-f cmd/apiserver/Dockerfile \
		.