ARG BASE_IMAGE
# Build the manager binary
FROM golang:1.19-alpine as builder
ARG OS
ARG ARCH

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/apiserver/main.go cmd/apiserver/main.go
COPY pkg/ pkg/

RUN CGO_ENABLED=0 \
    GOOS=${OS} \
    GOARCH=${ARCH} \
    go build \
      -a -ldflags "-s -w" \
      -o vela-prism \
      cmd/apiserver/main.go

FROM alpine:3.18.3
# This is required by daemon connnecting with cri
RUN apk add --no-cache ca-certificates bash expat
RUN apk add curl

WORKDIR /
COPY --from=builder /workspace/vela-prism /usr/local/bin/vela-prism

CMD ["vela-prism"]
