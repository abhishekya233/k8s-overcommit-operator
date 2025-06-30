<!--
SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
SPDX-FileContributor: enriqueavi@inditex.com

SPDX-License-Identifier: CC-BY-4.0
-->

# 📊 Available Metrics

This document describes the metrics exposed by the **k8s-overcommit-operator**. These metrics can be used to monitor the operator's behavior, performance, and the resources it manages.

## 📋 Table of Contents

- [Counter Metrics](#-counter-metrics)
- [Gauge Metrics](#-gauge-metrics)
- [Metric Usage](#-metric-usage)
- [Monitoring Setup](#-monitoring-setup)
- [Example Queries](#-example-queries)

---

## 📈 Counter Metrics

### k8s_overcommit_operator_pods_requested_total

**Type:** Counter
**Description:** Total number of pods requested to be mutated by the webhook.

**Labels:**
- `class`: Overcommit class applied to the pod

**Example:**
```
k8s_overcommit_operator_pods_requested_total{class="high-density"} 150
k8s_overcommit_operator_pods_requested_total{class="default"} 89
```

---

### k8s_overcommit_operator_mutated_pods_total

**Type:** Counter
**Description:** Total number of pods successfully mutated by the k8s-overcommit-operator webhook.

**Labels:**
- `class`: Overcommit class that was applied

**Example:**
```
k8s_overcommit_operator_mutated_pods_total{class="high-density"} 145
k8s_overcommit_operator_mutated_pods_total{class="default"} 82
```

---

### k8s_overcommit_operator_pods_not_mutated_total

**Type:** Counter
**Description:** Total number of pods that were not mutated by the operator's webhook.

**Labels:**
- `class`: Overcommit class (if any)
- `generate_name`: Generated name of the pod
- `namespace`: Namespace where the pod was created
- `reason`: Reason why the pod was not mutated

**Common reasons:**
- `no_limits`: Pod has no resource limits defined
- `excluded_namespace`: Namespace is in the exclusion list
- `no_class_found`: No matching overcommit class found
- `validation_error`: Pod spec validation failed

**Example:**
```
k8s_overcommit_operator_pods_not_mutated_total{class="",generate_name="app-",namespace="kube-system",reason="excluded_namespace"} 25
k8s_overcommit_operator_pods_not_mutated_total{class="",generate_name="worker-",namespace="default",reason="no_limits"} 12
```

---

### k8s_overcommit_operator_pod_mutated

**Type:** Counter
**Description:** Detailed counter for individual pod mutations with full context.

**Labels:**
- `class`: Overcommit class applied
- `kind`: Kubernetes resource kind (usually "Pod")
- `name`: Name of the pod
- `namespace`: Namespace of the pod

**Example:**
```
k8s_overcommit_operator_pod_mutated{class="high-density",kind="Pod",name="web-app-7f8b9c",namespace="production"} 1
k8s_overcommit_operator_pod_mutated{class="default",kind="Pod",name="worker-abc123",namespace="default"} 1
```

---

## 📊 Gauge Metrics

### k8s_overcommit_operator_total_classes

**Type:** Gauge
**Description:** Total number of OvercommitClass resources currently defined in the cluster.

**Labels:** None

**Example:**
```
k8s_overcommit_operator_total_classes 5
```

---

### k8s_overcommit_operator_version

**Type:** Gauge
**Description:** Version information of the k8s-overcommit-operator.

**Labels:**
- `version`: Semantic version of the operator

**Example:**
```
k8s_overcommit_operator_version{version="1.0.0"} 1
```

---

### k8s_overcommit_operator_class

**Type:** Gauge
**Description:** Information about each OvercommitClass resource.

**Labels:**
- `name`: Name of the OvercommitClass
- `cpu`: CPU overcommit ratio (0.0-1.0)
- `memory`: Memory overcommit ratio (0.0-1.0)
- `isDefault`: Whether this is the default class ("true"/"false")

**Example:**
```
k8s_overcommit_operator_class{name="high-density",cpu="0.2",memory="0.8",isDefault="true"} 1
k8s_overcommit_operator_class{name="moderate",cpu="0.5",memory="0.9",isDefault="false"} 1
k8s_overcommit_operator_class{name="conservative",cpu="0.8",memory="0.95",isDefault="false"} 1
```

---

## 🔧 Metric Usage

### Accessing Metrics

The operator exposes metrics on the `/metrics` endpoint, typically on port `8080`:

```bash
# Direct access to metrics
curl http://<operator-pod-ip>:8080/metrics

# Through port-forward
kubectl port-forward -n k8s-overcommit-operator-system deployment/k8s-overcommit-operator-controller-manager 8080:8080
curl http://localhost:8080/metrics
```

### Filtering Metrics

To see only k8s-overcommit-operator metrics:

```bash
curl -s http://localhost:8080/metrics | grep k8s_overcommit_operator
```

---

## 📊 Monitoring Setup

### Prometheus Configuration

Add the following ServiceMonitor to scrape metrics:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: k8s-overcommit-operator-metrics
  namespace: k8s-overcommit-operator-system
spec:
  selector:
    matchLabels:
      control-plane: controller-manager
  endpoints:
  - port: https
    scheme: https
    tlsConfig:
      insecureSkipVerify: true
    bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
    path: /metrics
```

### Grafana Dashboard

Create dashboards to visualize:

1. **Pod Mutation Rate**: `rate(k8s_overcommit_operator_mutated_pods_total[5m])`
2. **Mutation Success Rate**: `rate(k8s_overcommit_operator_mutated_pods_total[5m]) / rate(k8s_overcommit_operator_pods_requested_total[5m])`
3. **Active Classes**: `k8s_overcommit_operator_total_classes`

You have a example dashboard in grafana_dashboard.json

---

## 📈 Example Queries

### PromQL Examples

#### Mutation Success Rate by Class
```promql
rate(k8s_overcommit_operator_mutated_pods_total[5m]) by (class)
```

#### Pods Not Mutated by Reason
```promql
k8s_overcommit_operator_pods_not_mutated_total by (reason)
```

#### Overall Mutation Success Rate
```promql
sum(rate(k8s_overcommit_operator_mutated_pods_total[5m])) /
sum(rate(k8s_overcommit_operator_pods_requested_total[5m])) * 100
```

#### Top Namespaces with Excluded Pods
```promql
topk(10,
  sum(k8s_overcommit_operator_pods_not_mutated_total{reason="excluded_namespace"}) by (namespace)
)
```

#### Active OvercommitClasses
```promql
count(k8s_overcommit_operator_class) by (isDefault)
```

### Sample Grafana Queries

#### Mutation Rate Panel
```promql
sum(rate(k8s_overcommit_operator_mutated_pods_total[5m])) by (class)
```

#### Error Rate Panel
```promql
sum(rate(k8s_overcommit_operator_pods_not_mutated_total[5m])) by (reason)
```

---

## 🔍 Troubleshooting

### Common Metric Issues

1. **Metrics not appearing**: Check if the operator pod is running and metrics port is accessible
2. **Stale metrics**: Verify the operator is processing pod admission requests
3. **Missing labels**: Ensure OvercommitClass resources have proper labels defined

### Debug Commands

```bash
# Check operator pod status
kubectl get pods -n k8s-overcommit-operator-system

# View operator logs
kubectl logs -n k8s-overcommit-operator-system deployment/k8s-overcommit-operator-controller-manager

# Test metrics endpoint
kubectl port-forward -n k8s-overcommit-operator-system svc/k8s-overcommit-operator-controller-manager-metrics-service 8080:8443
curl -k https://localhost:8080/metrics
```

---

<div align="center">

**[⬆️ Back to Top](#-available-metrics)**

</div>
