# OPERATOR_VERSION defines the project version for the bundle.
# Update this value when you upgrade the version of your project.
# To re-generate a bundle for another specific version without changing the standard setup, you can:
# - use the OPERATOR_VERSION as arg of the bundle target (e.g make bundle OPERATOR_VERSION=0.0.2)
# - use environment variables to overwrite this value (e.g export OPERATOR_VERSION=0.0.2)
OPERATOR_VERSION ?= 0.1.0

export OPERATOR_SDK_VERSION ?= v1.4.0

# CHANNELS define the bundle channels used in the bundle. 
# Add a new line here if you would like to change its default config. (E.g CHANNELS = "preview,fast,stable")
# To re-generate a bundle for other specific channels without changing the standard setup, you can:
# - use the CHANNELS as arg of the bundle target (e.g make bundle CHANNELS=preview,fast,stable)
# - use environment variables to overwrite this value (e.g export CHANNELS="preview,fast,stable")
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif

# DEFAULT_CHANNEL defines the default channel used in the bundle. 
# Add a new line here if you would like to change its default config. (E.g DEFAULT_CHANNEL = "stable")
# To re-generate a bundle for any other default channel without changing the default setup, you can:
# - use the DEFAULT_CHANNEL as arg of the bundle target (e.g make bundle DEFAULT_CHANNEL=stable)
# - use environment variables to overwrite this value (e.g export DEFAULT_CHANNEL="stable")
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

IMAGE_TAG ?= v$(OPERATOR_VERSION)
IMAGE_REGISTRY ?= quay.io/openshift-kni/

# BUNDLE_IMG defines the image:tag used for the bundle. 
# You can use it as an arg. (E.g make bundle-build BUNDLE_IMG=<some-registry>/<project-name-bundle>:<tag>)
BUNDLE_IMG ?= $(IMAGE_REGISTRY)node-label-operator-bundle:$(IMAGE_TAG)

# Image URL to use all building/pushing image targets
IMAGE_NAME ?= node-label-operator
IMG ?= $(IMAGE_REGISTRY)$(IMAGE_NAME):$(IMAGE_TAG)
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

CONTROLLER_GEN = go run $(shell pwd)/vendor/sigs.k8s.io/controller-tools/cmd/controller-gen
KUSTOMIZE = go run $(shell pwd)/vendor/sigs.k8s.io/kustomize/kustomize/v3

all: manager

# Run tests
ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
test: generate fmt vet manifests
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.7.0/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test ./... -tags test -coverprofile cover.out

# Run e2e tests
.PHONY: e2e-test
e2e-test:
	go test ./e2e  -ginkgo.v -test.v

# Build manager binary
manager: generate fmt vet
	hack/build.sh

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

# UnDeploy controller from the configured Kubernetes cluster in ~/.kube/config
undeploy:
	$(KUSTOMIZE) build config/default | kubectl delete -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests:
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	# goimports does more than fmt
	# skip vendor!
	find . -name '*.go' -not -wholename './vendor/*' | while read -r file; do go run golang.org/x/tools/cmd/goimports -w "$$file"; done

# Run go vet against code
vet:
	go vet ./...

# Run lint tests
.PHONY: lint
lint: generate fmt vet manifests bundle
	hack/verify-unchanged.sh

# Generate code
generate:
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build: test
	docker build -t ${IMG} .

# Push the docker image
docker-push:
	docker push ${IMG}

# Download operator sdk if needed
.PHONY: operator-sdk
operator-sdk:
	hack/operator-sdk.sh

# Generate bundle manifests and metadata, then validate generated files.
.PHONY: bundle
bundle: operator-sdk manifests
	./bin/operator-sdk generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests | ./bin/operator-sdk generate bundle -q --overwrite --version $(OPERATOR_VERSION) $(BUNDLE_METADATA_OPTS)
	./bin/operator-sdk bundle validate ./bundle

# Build the bundle image.
.PHONY: bundle-build
bundle-build:
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

# Push the bundle image.
.PHONY: bundle-push
bundle-push:
	docker push $(BUNDLE_IMG)

# Build and push bundle and operator
.PHONY: docker-all
docker-all: bundle bundle-build bundle-push docker-build docker-push

# Run unit tests on CI. Nothing special to do for now, but let's be prepared.
.PHONY: ci-test
ci-test: test

# Run linter tests on CI. Nothing special to do for now, but let's be prepared.
.PHONY: ci-lint
ci-lint: lint

# Run e2e tests on CI. Nothing special to do for now, but let's be prepared.
.PHONY: ci-e2e-test
ci-e2e-test: e2e-test
