// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"fmt"

	overcommit "github.com/InditexTech/k8s-overcommit-operator/api/v1alphav1"
	"github.com/InditexTech/k8s-overcommit-operator/internal/metrics"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getTotalClasses(ctx context.Context, client client.Client) error {

	overcommitClasses := &overcommit.OvercommitClassList{}
	err := client.List(ctx, overcommitClasses)
	if err != nil {
		return err
	}

	totalClasses := len(overcommitClasses.Items)
	metrics.K8sOvercommitOperatorTotalClasses.Set(float64(totalClasses))
	metrics.K8sOvercommitOperatorClass.Reset()
	for _, overcommitClass := range overcommitClasses.Items {
		metrics.K8sOvercommitOperatorClass.WithLabelValues(
			overcommitClass.GetName(),
			fmt.Sprintf("%f", overcommitClass.Spec.CpuOvercommit),
			fmt.Sprintf("%f", overcommitClass.Spec.MemoryOvercommit),
			fmt.Sprintf("%t", overcommitClass.Spec.IsDefault),
		).Set(1)
	}
	return nil
}
