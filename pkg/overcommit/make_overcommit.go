// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package overcommit

import (
	"context"
	"os"

	"github.com/InditexTech/k8s-overcommit-operator/internal/metrics"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var podlog = logf.Log.WithName("overcommit")

func mutateContainers(containers []corev1.Container, pod *corev1.Pod, cpuValue float64, memoryValue float64) {
	for i, container := range containers {
		limits := container.Resources.Limits
		requests := container.Resources.Requests
		// If the container doesn't have limits, don't mutate the container
		if limits == nil {
			podlog.Info(
				"Limits is nil, don't mutate the container",
				"containerName", container.Name, "generateName", pod.GenerateName,
			)
		} else if cpuValue == 1 && memoryValue == 1 {
			podlog.Info(
				"Container didn't mutate the cpu and memory",
				"generateName", pod.GenerateName, "cpuValue", cpuValue, "memoryValue", memoryValue,
			)
		} else {
			if cpuLimit, ok := limits[corev1.ResourceCPU]; ok {
				// If the cpu overcommit value is 1, don't mutate the container
				if cpuValue == 1 {
					podlog.Info(
						"Container didn't mutate the cpu",
						"generateName", pod.GenerateName, "cpuValue", cpuValue,
					)
				}
				newCPURequest := float64(cpuLimit.MilliValue()) * cpuValue
				requests[corev1.ResourceCPU] = *resource.NewMilliQuantity(int64(newCPURequest), resource.DecimalSI)
			}
			if memoryLimit, ok := limits[corev1.ResourceMemory]; ok {
				// If the memory overcommit value is 1, don't mutate the container
				if memoryValue == 1 {
					podlog.Info(
						"Container didn't mutate the memory",
						"generateName", pod.GenerateName, "memoryValue", memoryValue,
					)
				}
				newMemoryRequest := float64(memoryLimit.Value()) * memoryValue
				requests[corev1.ResourceMemory] = *resource.NewQuantity(int64(newMemoryRequest), resource.BinarySI)
			}
			containers[i].Resources.Requests = requests
		}
	}
}

func makeOvercommit(pod *corev1.Pod, cpuValue float64, memoryValue float64) {
	mutateContainers(pod.Spec.Containers, pod, cpuValue, memoryValue)
	podlog.Info(
		"Containers mutated", "generateName", pod.GenerateName, "cpuValue", cpuValue, "memoryValue", memoryValue,
	)
}

func makeOvercommitInitContainers(pod *corev1.Pod, cpuValue float64, memoryValue float64) {
	mutateContainers(pod.Spec.InitContainers, pod, cpuValue, memoryValue)
	podlog.Info(
		"InitContainers mutated", "generateName", pod.GenerateName, "cpuValue", cpuValue, "memoryValue", memoryValue,
	)
}

func Overcommit(pod *corev1.Pod, recorder record.EventRecorder, client client.Client) {
	ctx := context.Background()
	podlog.Info("Mutating Pod", "generateGame", pod.GenerateName)
	metrics.K8sOvercommitOperatorPodsRequestedTotal.WithLabelValues(os.Getenv("OVERCOMMIT_CLASS_NAME")).Inc()

	// Get the overcommit values from the labels
	cpuValue, memoryValue := checkOvercommitType(ctx, *pod, client)
	// Check if the values are valid

	// Multiplicate the limits by the overcommit value and set the new value as request
	makeOvercommit(pod, cpuValue, memoryValue)
	podlog.Info(
		"Pod mutated", "generateName", pod.GenerateName, "cpuValue", cpuValue, "memoryValue", memoryValue,
	)

	// If it has initContainers, make the overcommit
	podlog.Info("cheking if pod has initContainer")
	if len(pod.Spec.InitContainers) > 0 {
		podlog.Info("Pod has initContainers, mutating them", "generateName", pod.GenerateName)
		makeOvercommitInitContainers(pod, cpuValue, memoryValue)
	}

	// Increment the metric K8sOvercommitOperatorMutatedPodsTotal
	metrics.K8sOvercommitOperatorMutatedPodsTotal.WithLabelValues(os.Getenv("OVERCOMMIT_CLASS_NAME")).Inc()

	// Add an event to the pod
	recorder.Eventf(
		pod,
		corev1.EventTypeNormal,
		"OvercommitApplied",
		"Applied overcommit to containers of Pod '%s': OvercommitClass = %s, CPU Overcommit = %.2f, Memory Overcommit = %.2f",
		pod.Name,
		os.Getenv("OVERCOMMIT_CLASS_NAME"),
		cpuValue,
		memoryValue,
	)
	if cpuValue == 1 && memoryValue == 1 {
		metrics.K8sOvercommitOperatorPodsNotMutatedTotal.WithLabelValues(
			os.Getenv("OVERCOMMIT_CLASS_NAME"), pod.GenerateName, pod.Namespace, "overcommit values = 1",
		).Inc()

	}
	podlog.Info("Pod mutated", "generateName", pod.GenerateName, "cpuValue", cpuValue, "memoryValue", memoryValue)
}
