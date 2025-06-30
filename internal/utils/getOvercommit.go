// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"errors"
	"fmt"

	overcommit "github.com/InditexTech/k8s-overcommit-operator/api/v1alphav1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetOvercommitLabel(ctx context.Context, k8sClient client.Client) (string, error) {
	if k8sClient == nil {
		return "", errors.New("client parameter cannot be nil")
	}

	// Create a new OvercommitClass object
	var overcommitObject overcommit.Overcommit

	// Search for the OvercommitClass with the name "cluster"
	err := k8sClient.Get(ctx, client.ObjectKey{
		Name: "cluster",
	}, &overcommitObject)

	if err != nil {
		podlog.Error(err, "Error getting the CR cluster of Overcommit", "name", "cluster")
		return "", fmt.Errorf("error getting OvercommitClass with name '%s': %w", "cluster", err)
	}

	// Return the overcommit label
	return overcommitObject.Spec.OvercommitLabel, nil
}

func GetOvercommit(ctx context.Context, k8sClient client.Client) (overcommit.Overcommit, error) {
	// Create a new OvercommitClass object
	var overcommitObject overcommit.Overcommit

	if k8sClient == nil {
		return overcommitObject, errors.New("client parameter cannot be nil")
	}

	// Search for the OvercommitClass with the name "cluster"
	err := k8sClient.Get(ctx, client.ObjectKey{
		Name: "cluster",
	}, &overcommitObject)

	if err != nil {
		podlog.Error(err, "Error getting the CR cluster of Overcommit", "name", "cluster")
		return overcommitObject, fmt.Errorf("error getting OvercommitClass with name '%s': %w", "cluster", err)
	}

	// Return the overcommit label
	return overcommitObject, nil
}
