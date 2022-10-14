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

KUSTOMIZE_HOME=config
KUSTOMIZE_CRDS=$(KUSTOMIZE_HOME)/crds/

CRD_OPTIONS ?= "crd:maxDescLen=0,generateEmbeddedObjectMeta=true"

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
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=apex-role webhook paths="./pkg/apis/..." output:crd:artifacts:config=$(KUSTOMIZE_CRDS)
	go mod tidy

build: verify
	@CGO_ENABLED=0 GOOS=linux go build -trimpath --ldflags $(LDFLAGS) -o apex

localdev: localdev-deploy install

localdev-deploy:
	@./config/kind/deploy.sh

install: generate manifests
	@cat config/crds/*.yaml | kubectl apply -n apex -f -
	@cat config/rbac/*.yaml | kubectl apply -n apex -f -

# Run inside the localdev kind cluster
run:
	$(eval POD := $(shell kubectl get pods -n apex -l app=dev -o=custom-columns=:metadata.name --no-headers))
	kubectl exec -n apex -it pod/$(POD) -- bash -c "APEX_ENABLE_WEBHOOKS=false go run main.go"

exec:
	$(eval POD := $(shell kubectl get pods -n apex -l app=dev -o=custom-columns=:metadata.name --no-headers))
	kubectl exec -n apex -it pod/$(POD) -- bash

clean:
	@echo "Cleaning up all the generated files"
	@find . -name '*.test' | xargs rm -fv
	@find . -name '*~' | xargs rm -fv
	@find . -name '*.zip' | xargs rm -fv
	@kind delete cluster --name apex
