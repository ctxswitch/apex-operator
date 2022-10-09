#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
REPOSITORY=ctx.sh/apex-operator

bash ./vendor/k8s.io/code-generator/generate-groups.sh "all" \
  ${REPOSITORY}/pkg/client ${REPOSITORY}/pkg/apis \
  "apex.ctx.sh:v1" \
  --output-base "${SCRIPT_ROOT}"/../.. \
  --go-header-file "${SCRIPT_ROOT}"/k8s/boilerplate.go.txt
