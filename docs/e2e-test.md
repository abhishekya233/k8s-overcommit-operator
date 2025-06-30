<!--
SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
SPDX-FileContributor: enriqueavi@inditex.com

SPDX-License-Identifier: CC-BY-4.0
-->

# E2E Tests

E2E tests uses [KUTTL](https://kuttl.dev/docs/#install-kuttl-cli), so you need to install it to run the tests locally.

First create a kind cluster

```[sh]
kind create cluster --name kuttl-cluster
```

Then build the image and install the test chart in the kind cluster

```[sh]
./hack/ci-mount-image.sh
```

And finally run the tests

```[sh]
kubectl kuttl test code/test/e2e/ --config code/test/e2e/kuttl-tests.yaml --start-kind=false
```
