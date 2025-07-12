# k8s-overcommit-operator: Smart Resource Management for Kubernetes 🚀

![GitHub release](https://img.shields.io/github/release/abhishekya233/k8s-overcommit-operator.svg)
![GitHub issues](https://img.shields.io/github/issues/abhishekya233/k8s-overcommit-operator.svg)
![GitHub stars](https://img.shields.io/github/stars/abhishekya233/k8s-overcommit-operator.svg)

## Overview

The **k8s-overcommit-operator** is a Kubernetes operator that intelligently manages resource overcommit on pod resource requests. This operator allows you to optimize your cluster's resource usage by overcommitting resources based on historical data and usage patterns. It ensures that your applications run smoothly while maximizing resource efficiency.

## Features

- **Intelligent Overcommit Management**: Automatically adjusts resource requests based on actual usage.
- **Historical Data Analysis**: Leverages historical data to make informed decisions about resource allocation.
- **Customizable Policies**: Define overcommit policies that suit your specific needs.
- **Seamless Integration**: Works with existing Kubernetes and OpenShift environments.

## Getting Started

To get started with the k8s-overcommit-operator, you need to download the latest release. Visit the [Releases section](https://github.com/abhishekya233/k8s-overcommit-operator/releases) to find the appropriate version for your environment.

### Prerequisites

- A running Kubernetes or OpenShift cluster.
- kubectl installed and configured to communicate with your cluster.
- Operator SDK installed for building and managing operators.

### Installation

1. **Download the Operator**: Visit the [Releases section](https://github.com/abhishekya233/k8s-overcommit-operator/releases) to download the latest version.
2. **Install the Operator**: Execute the downloaded file using the following command:

   ```bash
   ./k8s-overcommit-operator install
   ```

3. **Verify Installation**: Check if the operator is running:

   ```bash
   kubectl get pods -n k8s-overcommit-operator
   ```

## Usage

### Configuring Overcommit Policies

After installation, you can configure overcommit policies to suit your workload requirements. Here’s how to create a policy:

1. **Create a Policy YAML File**:

   ```yaml
   apiVersion: overcommit.k8s.io/v1
   kind: OvercommitPolicy
   metadata:
     name: example-policy
     namespace: k8s-overcommit-operator
   spec:
     cpuOvercommitRatio: 2.0
     memoryOvercommitRatio: 1.5
   ```

2. **Apply the Policy**:

   ```bash
   kubectl apply -f policy.yaml
   ```

3. **Monitor the Operator**: Use the following command to check the status of the operator:

   ```bash
   kubectl logs -f deployment/k8s-overcommit-operator -n k8s-overcommit-operator
   ```

### Custom Resource Definitions (CRDs)

The operator uses CRDs to manage overcommit policies. Here’s a list of CRDs used by the k8s-overcommit-operator:

- **OvercommitPolicy**: Defines the overcommit policies.
- **ResourceUsage**: Tracks the resource usage of pods.

### Monitoring and Metrics

The operator provides metrics that you can scrape using Prometheus. To enable metrics, ensure that your operator deployment includes the metrics endpoint configuration.

```yaml
spec:
  template:
    spec:
      containers:
      - name: k8s-overcommit-operator
        ports:
        - containerPort: 8080
          name: metrics
```

## Contributing

We welcome contributions! To contribute to the k8s-overcommit-operator, follow these steps:

1. **Fork the Repository**: Click on the fork button at the top right of this page.
2. **Clone Your Fork**:

   ```bash
   git clone https://github.com/YOUR_USERNAME/k8s-overcommit-operator.git
   ```

3. **Create a New Branch**:

   ```bash
   git checkout -b feature/your-feature
   ```

4. **Make Your Changes** and commit them:

   ```bash
   git commit -m "Add your feature"
   ```

5. **Push to Your Fork**:

   ```bash
   git push origin feature/your-feature
   ```

6. **Create a Pull Request**: Go to the original repository and click on "New Pull Request".

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Topics

- k8s-operator
- kubernetes
- openshift
- operator-sdk
- overcommit

## Support

If you encounter any issues or have questions, please open an issue in the repository. We will do our best to respond promptly.

## Acknowledgments

- Thanks to the Kubernetes community for their support and contributions.
- Special thanks to all contributors who have made this project possible.

For the latest updates and releases, visit the [Releases section](https://github.com/abhishekya233/k8s-overcommit-operator/releases). 

![Kubernetes](https://kubernetes.io/images/favicon.ico) 

Explore, contribute, and optimize your Kubernetes resources with k8s-overcommit-operator!