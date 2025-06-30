NAMESPACE = 'k8s-overcommit'
VALUES_DEV_FILE = 'hack/tilt/values-dev.yaml'   # original values without modifications
VALUES_FILE = 'hack/tilt/chart/values.yaml'     # values we will use for helm
CHART_PATH = './chart'
RENDERED_MANIFEST = 'hack/tilt/chart/chart.yaml'
CRDS_MANIFEST = 'hack/tilt/chart/crds.yaml'

DEPLOYMENT_NAME = 'k8s-overcommit-controller'

# Initial default image for the first execution
OPERATOR_IMAGE_REF = 'localhost:5000/controller-manager'


local_resource(
    'tilt-build-bin',
    'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o hack/tilt/bin/manager ./cmd/main.go',
    deps=[
     './api',
     './cmd',
     './internal',
     './pkg',
     './go.sum',
     './go.mod'
    ],
)

# 1. Install cert-manager (always runs, but only applies if not present)
local_resource(
    'install-cert-manager',
    '''
    kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.17.0/cert-manager.yaml &&
    kubectl wait --for=condition=available --timeout=60s deployment/cert-manager-webhook -n cert-manager
    '''
)

# 2. Create namespace if it doesn't exist
local_resource(
    'create-namespace',
    'kubectl create namespace ' + NAMESPACE + ' || true',
    resource_deps=['install-cert-manager'],
)



# 3. Build Docker image with Tilt (detects changes and rebuilds)
docker_build(
    OPERATOR_IMAGE_REF,
    '.',
    dockerfile='hack/tilt/Dockerfile',
    only=['hack/tilt/bin'],
)

# 4. Prepare values.yaml (copy or update according to whether it's first execution or not)
local_resource(
    'prepare-values',
    'python3 hack/tilt/scripts/prepare_values.py --namespace ' + NAMESPACE + ' --deployment ' + DEPLOYMENT_NAME + ' --values-dev ' + VALUES_DEV_FILE + ' --values ' + VALUES_FILE + ' --default-image ' + OPERATOR_IMAGE_REF,
    resource_deps=['create-namespace', OPERATOR_IMAGE_REF, 'k8s-overcommit-controller'],
    deps=[
        VALUES_DEV_FILE,
        'hack/tilt/scripts/prepare_values.py',
        'hack/tilt/bin/manager',
    ],
)

# 5. Patch deployments ending in -overcommit-webhook with the current image from core deployment
local_resource(
    'patch-webhooks',
    '''
    set -x
    IMAGE=$(kubectl -n ''' + NAMESPACE + ''' get deployment/''' + DEPLOYMENT_NAME + ''' -o jsonpath='{.spec.template.spec.containers[0].image}')
    echo "Current image to patch webhook deployments: $IMAGE"
    DEPLOYMENTS=$(kubectl -n ''' + NAMESPACE + ''' get deployments -o name | grep -E '-overcommit-webhook$' || true)
    echo "Webhook deployments to patch:"
    echo "$DEPLOYMENTS"
    for dep in $DEPLOYMENTS; do
      echo "Patching $dep with image $IMAGE"
      kubectl -n ''' + NAMESPACE + ''' set image $dep '*=$IMAGE'
    done
    set +x
    ''',
    resource_deps=['prepare-values'],
    deps=[VALUES_FILE],
)

local_resource(
  "render-crds",
  "cat ./chart/crds/*.yaml > hack/tilt/chart/crds.yaml",
  deps=["./chart/crds", 'hack/tilt/chart/values.yaml'],
  resource_deps=["update-overcommit-deployments"]
)


# 6. Render the chart with Helm using updated values.yaml
local_resource(
    'render-chart',
    'helm template my-release ' + CHART_PATH + ' -f ' + VALUES_FILE + ' > ' + RENDERED_MANIFEST,
    resource_deps=['patch-webhooks','render-crds','install-cert-manager'],
    deps=["./chart", 'hack/tilt/chart/values.yaml'],
)

# 7. Apply the rendered chart
k8s_yaml(CRDS_MANIFEST)
k8s_yaml(RENDERED_MANIFEST)
