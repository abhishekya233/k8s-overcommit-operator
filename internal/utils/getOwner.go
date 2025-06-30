// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetPodOwner retrieves the owner of a Pod. If the owner is a ReplicaSet, it fetches the Deployment.
func GetPodOwner(ctx context.Context, k8sClient client.Client, pod *corev1.Pod) (string, string, error) {
	// Check if the Pod has an owner reference
	if len(pod.OwnerReferences) == 0 {
		return pod.Name, "pod", nil
	}

	ownerRef := pod.OwnerReferences[0] // Assume the first owner reference is the relevant one

	// If the owner is a ReplicaSet, fetch its Deployment
	if ownerRef.Kind == "ReplicaSet" {
		replicaSet := &appsv1.ReplicaSet{}
		err := k8sClient.Get(ctx, types.NamespacedName{Name: ownerRef.Name, Namespace: pod.Namespace}, replicaSet)
		if err != nil {
			return "", "", fmt.Errorf("failed to get ReplicaSet %s: %v", ownerRef.Name, err)
		}

		// Check if the ReplicaSet has an owner reference
		if len(replicaSet.OwnerReferences) == 0 {
			return "", "", fmt.Errorf("replicaSet %s has no owner", replicaSet.Name)
		}

		rsOwnerRef := replicaSet.OwnerReferences[0]
		if rsOwnerRef.Kind == "Deployment" {
			return rsOwnerRef.Name, rsOwnerRef.Kind, nil
		}

		return "", "", fmt.Errorf("replicaSet %s owner is not a Deployment", replicaSet.Name)
	}

	// If the owner is not a ReplicaSet, find the root owner

	ownerObj := &unstructured.Unstructured{}
	ownerObj.SetKind(ownerRef.Kind)
	ownerObj.SetAPIVersion(ownerRef.APIVersion)
	err := k8sClient.Get(ctx, types.NamespacedName{Name: ownerRef.Name, Namespace: pod.Namespace}, ownerObj)
	if err != nil {
		return "", "", fmt.Errorf("failed to get owner object %s: %v", ownerRef.Name, err)
	}

	rootOwner, err := findRootOwner(ctx, k8sClient, ownerObj)
	if err != nil {
		return "", "", fmt.Errorf("failed to find root owner: %v", err)
	}

	u, ok := rootOwner.(*unstructured.Unstructured)
	if !ok {
		return rootOwner.GetName(), "", fmt.Errorf("rootOwner is not *unstructured.Unstructured")
	}
	return u.GetName(), u.GetKind() + "/" + u.GetAPIVersion(), nil
}

func findRootOwner(ctx context.Context, c client.Client, obj metav1.Object) (metav1.Object, error) {
	owners := obj.GetOwnerReferences()
	if len(owners) == 0 {
		return obj, nil
	}

	ownerRef := owners[0]
	ownerObj := &unstructured.Unstructured{}
	ownerObj.SetKind(ownerRef.Kind)
	ownerObj.SetAPIVersion(ownerRef.APIVersion)

	// Search for the owner by name and namespace
	err := c.Get(ctx, types.NamespacedName{
		Name:      ownerRef.Name,
		Namespace: obj.GetNamespace(),
	}, ownerObj)
	if err != nil {
		return nil, err
	}

	return findRootOwner(ctx, c, ownerObj)
}
