#### BUILDER ####
FROM golang:1.16 AS builder

WORKDIR /go/src/github.com/reshnm/k8s-sample-controller-crd
COPY . .

RUN CGO_ENABLED=0 GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) GO111MODULE=on \
    go install

#### BASE ####
FROM alpine:3.14.0 AS base

RUN apk add --no-cache ca-certificates

#### CONTROLLER ####
FROM base as controller

COPY --from=builder /go/bin/k8s-sample-controller-crd /k8s-sample-controller-crd

WORKDIR /

ENTRYPOINT ["/k8s-sample-controller-crd"]
