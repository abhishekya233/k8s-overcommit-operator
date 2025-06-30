// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package overcommit

import (
	"context"
	"fmt"

	"github.com/InditexTech/k8s-overcommit-operator/internal/metrics"
	"github.com/InditexTech/k8s-overcommit-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

// GetLabelValue gets the value of a label in the pod or in the namespace of the pod
func getNamespaceOvercommit(ctx context.Context, pod *corev1.Pod, client client.Client, label, ownerName, ownerKind string) (float64, float64) {
	// Get the namespace of the pod
	namespaceName := pod.ObjectMeta.Namespace
	var ns corev1.Namespace
	namespace, err := getNamespaceYAML(ctx, namespaceName, client)
	if err != nil {
		podlog.Error(err, "Error getting the namespace yaml", "namespace", namespaceName)
	}
	err = yaml.Unmarshal([]byte(namespace), &ns)
	if err != nil {
		podlog.Error(err, "Error unmarshaling the namespace", "namespace", namespaceName)
	}

	// Check if the overcommit class label is in the namespace
	if val, ok := ns.Labels[label]; ok {
		podlog.Info("Namespace class found", "class", val)
		overcommitClass, err := utils.GetOvercommitClassSpec(ctx, val, client)
		if err != nil {
			podlog.Error(err, "Error getting the overcommit class", "overcommitClassLabel", val)
		}
		metrics.K8sOvercommitPodMutated.WithLabelValues(val, ownerKind, ownerName, pod.Namespace).Inc()
		return overcommitClass.CpuOvercommit, overcommitClass.MemoryOvercommit
	} else {
		podlog.Info("Overcommit class not found in the namespace, using the default", "namespace", ns.Name)
		defaultClass, error := utils.GetDefaultSpec(client)
		if error != nil {
			podlog.Error(error, "Error getting the default overcommit class")
		}
		metrics.K8sOvercommitPodMutated.WithLabelValues("default", ownerKind, ownerName, pod.Namespace).Inc()
		return defaultClass.CpuOvercommit, defaultClass.MemoryOvercommit
	}
}

// getNamespaceYAML gets the YAML of a namespace using the ServiceAccount token
func getNamespaceYAML(ctx context.Context, namespaceName string, k8sClient client.Client) (string, error) {
	// Create a variable to store the Namespace object
	var namespace corev1.Namespace

	// Get the Namespace object from the API server
	err := k8sClient.Get(ctx, client.ObjectKey{
		Name: namespaceName,
	}, &namespace)
	if err != nil {
		return "", fmt.Errorf("error getting the namespace: %v", err)
	}

	// Convert the Namespace object to YAML
	nsYAML, err := yaml.Marshal(namespace)
	if err != nil {
		return "", fmt.Errorf("error converting the namespace to YAML: %v", err)
	}

	return string(nsYAML), nil
}

func checkOvercommitType(ctx context.Context, pod corev1.Pod, client client.Client) (float64, float64) {
	var cpuValue float64
	var memoryValue float64
	ownerName, ownerKind, err := utils.GetPodOwner(ctx, client, &pod)
	if err != nil {
		podlog.Error(err, "Error getting the pod owner")
	}

	label, err := utils.GetOvercommitLabel(ctx, client)
	if err != nil {
		podlog.Error(err, "Error getting the overcommit label")
	}
	//  Check if the pod has the overcommit class label
	value, exists := pod.Labels[label]
	podlog.Info(
		"Checking if pod has overcommit class label",
		"overcommitClassLabel", value,
		"exists", exists,
	)
	if exists {
		// Overcommit class found in pod
		overcommitClass, err := utils.GetOvercommitClassSpec(ctx, value, client)
		if err != nil {
			podlog.Error(err, "Error getting the overcommit class", "overcommitClassLabel", value)
			// Overcommit class not found or some error
			cpuValue, memoryValue = getNamespaceOvercommit(ctx, &pod, client, label, ownerName, ownerKind)
		} else {
			cpuValue, memoryValue = overcommitClass.CpuOvercommit, overcommitClass.MemoryOvercommit
		}
		metrics.K8sOvercommitPodMutated.WithLabelValues(value, ownerKind, ownerName, pod.Namespace).Inc()
	} else {
		// Overcommit class not found, checking the overcommit labels
		podlog.Info("Overcommit class label not found in pod, checking the namespace")
		cpuValue, memoryValue = getNamespaceOvercommit(ctx, &pod, client, label, ownerName, ownerKind)

	}
	return cpuValue, memoryValue
}
