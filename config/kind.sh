#!/usr/bin/env bash
set -o errexit

SCRIPT_DIR=$(dirname "${BASH_SOURCE[0]}")
NODE_IMAGE="${KIND_IMAGE:-kindest/node:v1.23.1}"

CLUSTER="$(kind get clusters 2>&1 | grep apex || : )"
if [ "x$CLUSTER" == "x" ] ; then
cat <<EOF | kind create cluster --name=apex --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  image: ${NODE_IMAGE}
  extraMounts:
  - hostPath: ${PWD}
    containerPath: /app
    readOnly: true
- role: worker
  image: ${NODE_IMAGE}
  extraMounts:
  - hostPath: ${PWD}
    containerPath: /app
    readOnly: true
- role: worker
  image: ${NODE_IMAGE}
  extraMounts:
  - hostPath: ${PWD}
    containerPath: /app
    readOnly: true
- role: worker
  image: ${NODE_IMAGE}
  extraMounts:
  - hostPath: ${PWD}
    containerPath: /app
    readOnly: true
EOF
else
echo "Cluster exists, skipping creation"
fi

docker pull golang
kind load docker-image golang --name apex

docker pull datadog/agent:7
kind load docker-image datadog/agent:7 --name apex

docker pull registry.k8s.io/kube-state-metrics/kube-state-metrics:v2.6.0
kind load docker-image registry.k8s.io/kube-state-metrics/kube-state-metrics:v2.6.0 --name apex
