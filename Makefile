PWD := $(shell pwd)
LDFLAGS ?= "-s -w -X main.Version=$(VERSION)"
TMPFILE := $(shell mktemp)
GOPATH := $(shell go env GOPATH)
GOARCH := $(shell go env GOARCH)
GOOS := $(shell go env GOOS)

ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

CRD_OPTIONS ?= "crd:maxDescLen=0,generateEmbeddedObjectMeta=true"
RBAC_OPTIONS ?= "rbac:roleName=apex-role"
WEBHOOK_OPTIONS ?= "webhook"
OUTPUT_OPTIONS ?= "output:crd:artifacts:config=config/base/crd"

certs:
	@./config/gen-certs.sh

controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.10.0 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(shell go env GOPATH)/bin/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

getdeps:
	@echo "Checking dependencies"
	@which golangci-lint 1>/dev/null || (echo "Installing golangci-lint" && curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v1.36.0)
	@which docker 1>/dev/null || (echo "It's 2021.  Why no docker?" && exit 1)
	@GO111MODULE=off go get sigs.k8s.io/controller-tools/cmd/controller-gen

verify: govet gotest lint

lint:
	@echo "Running $@ check"
	@GO111MODULE=on ${GOPATH}/bin/golangci-lint cache clean
	@GO111MODULE=on ${GOPATH}/bin/golangci-lint run --timeout=5m --config ./.golangci.yml

govet:
	@go vet ./...

gotest:
	@go test -race ./...

generate: controller-gen
	@./k8s/update-codegen.sh

manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) $(RBAC_OPTIONS) $(WEBHOOK_OPTIONS) paths="./pkg/apis/..." $(OUTPUT_OPTIONS)
	go mod tidy

build: verify
	@CGO_ENABLED=0 GOOS=linux go build -trimpath --ldflags $(LDFLAGS) -o apex

local: kind install example

kind:
	@./config/kind.sh

install:
	kubectl apply -k config/overlays/dev

example:
	@cat config/examples/ns.yaml | kubectl apply -n example -f -
	@cat config/examples/app.yaml | kubectl apply -n example -f -
	@cat config/examples/ddagent.yaml | kubectl apply -n example -f -
	@cat config/examples/kube-state-metrics.yaml | kubectl apply -n kube-system -f -

example-scraper:
	@cat config/examples/scraper.yaml | kubectl apply -n example -f -

run:
	$(eval POD := $(shell kubectl get pods -n apex-system -l name=apex-operator -o=custom-columns=:metadata.name --no-headers))
	kubectl exec -n apex-system -it pod/$(POD) -- bash -c "go run main.go -enable-leader-election"

exec:
	$(eval POD := $(shell kubectl get pods -n apex-system -l name=apex-operator -o=custom-columns=:metadata.name --no-headers))
	kubectl exec -n apex-system -it pod/$(POD) -- bash

clean:
	@echo "Cleaning up all the generated files"
	@find . -name '*.test' | xargs rm -fv
	@find . -name '*~' | xargs rm -fv
	@find . -name '*.zip' | xargs rm -fv
	@kind delete cluster --name apex
