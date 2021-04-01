#!/bin/bash
set -ex

GIT_VERSION=$(git describe --always --tags || true)
VERSION=${CI_UPSTREAM_VERSION:-${GIT_VERSION}}
GIT_COMMIT=$(git rev-list -1 HEAD || true)
COMMIT=${CI_UPSTREAM_COMMIT:-${GIT_COMMIT}}
BUILD_DATE=$(date --utc -Iseconds)

mkdir -p bin

LDFLAGS="-s -w "
LDFLAGS+="-X github.com/openshift-kni/node-label-operator/pkg.Version=${VERSION} "
LDFLAGS+="-X github.com/openshift-kni/node-label-operator/pkg.GitCommit=${COMMIT} "
LDFLAGS+="-X github.com/openshift-kni/node-label-operator/pkg.BuildDate=${BUILD_DATE} "
GOFLAGS=-mod=vendor CGO_ENABLED=0 GOOS=linux go build -ldflags="${LDFLAGS}" -o bin/manager github.com/openshift-kni/node-label-operator
