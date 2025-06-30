#!/usr/bin/env bash

# SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
# SPDX-FileContributor: enriqueavi@inditex.com
#
# SPDX-License-Identifier: Apache-2.0

CLUSTER_NAME="kind-test-overcommit"

mkdir ./hack/tilt/chart
touch ./hack/tilt/chart/chart.yaml
touch ./hack/tilt/chart/crds.yaml

# Create KIND cluster with ctlptl if it doesn't exist
if ! kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
  echo "Creating KIND cluster '${CLUSTER_NAME}' with ctlptl..."
  kind create cluster --name "${CLUSTER_NAME}"
else
  echo "KIND cluster '${CLUSTER_NAME}' already exists."
fi

cleanup() {
  echo "Deleting KIND cluster '${CLUSTER_NAME}' with kind..."
  kind delete cluster --name "${CLUSTER_NAME}"

  echo "Deleting temporary folders..."
  rm -rf ./hack/tilt/bin ./hack/tilt/chart

  echo "Cluster and files deleted."
}
trap cleanup EXIT

# Run tilt and wait for it to finish
echo "Running tilt..."
tilt up --port 10351